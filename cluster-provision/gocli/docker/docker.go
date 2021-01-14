package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"golang.org/x/crypto/ssh/terminal"
)

func GetPrefixedContainers(cli *client.Client, prefix string) ([]types.Container, error) {
	args, err := filters.ParseFlag("name="+prefix, filters.NewArgs())
	if err != nil {
		return nil, err
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: args,
		All:     true,
	})
	return containers, err
}

func GetPrefixedVolumes(cli *client.Client, prefix string) ([]*types.Volume, error) {
	args, err := filters.ParseFlag("name="+prefix, filters.NewArgs())
	if err != nil {
		return nil, err
	}
	volumes, err := cli.VolumeList(context.Background(), args)
	if err != nil {
		return nil, err
	}
	return volumes.Volumes, nil
}

func ImagePull(cli *client.Client, ctx context.Context, ref string, options types.ImagePullOptions) error {

	if !strings.ContainsAny(ref, ":@") {
		ref = ref + ":latest"
	}
	ref = strings.TrimPrefix(ref, "docker.io/")

	images, err := cli.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		return err
	}

	for _, img := range images {
		for _, tag := range append(img.RepoTags, img.RepoDigests...) {
			if tag == ref {
				logrus.Infof("Using local image %s", ref)
				return nil
			}
		}
	}
	logrus.Infof("Using remote image %s", ref)

	for _, i := range []int{0, 1, 2, 6} {
		time.Sleep(time.Duration(i) * time.Second)
		reader, err := cli.ImagePull(ctx, ref, options)
		if err != nil {
			log.Printf("failed to download %s: %v\n", ref, err)
			continue
		}
		err = PrintProgress(reader, os.Stdout)
		if err != nil {
			log.Printf("failed to download %s: %v\n", ref, err)
			continue
		}
		return nil
	}
	return fmt.Errorf("failed to download %s four times, giving up.", ref)
}

func Exec(cli *client.Client, container string, args []string, out io.Writer) (bool, error) {
	ctx := context.Background()
	id, err := cli.ContainerExecCreate(ctx, container, types.ExecConfig{
		Privileged:   true,
		Tty:          true,
		Detach:       false,
		Cmd:          args,
		AttachStdout: true,
		AttachStderr: true,
	})

	if err != nil {
		return false, err
	}

	attached, err := cli.ContainerExecAttach(ctx, id.ID, types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		Tty:          true,
	})
	if err != nil {
		return false, err
	}
	defer attached.Close()

	io.Copy(out, attached.Reader)

	resp, err := cli.ContainerExecInspect(ctx, id.ID)
	if err != nil {
		return false, err
	}
	return resp.ExitCode == 0, nil
}

func Terminal(cli *client.Client, container string, args []string, file *os.File) (int, error) {

	if !terminal.IsTerminal(int(file.Fd())) {
		return 1, fmt.Errorf("failure calling terminal out of TTY")
	}

	ctx := context.Background()
	id, err := cli.ContainerExecCreate(ctx, container, types.ExecConfig{
		Privileged:   true,
		Tty:          terminal.IsTerminal(int(file.Fd())),
		Detach:       false,
		Cmd:          args,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
	})

	if err != nil {
		return -1, err
	}

	attached, err := cli.ContainerExecAttach(ctx, id.ID, types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		AttachStdin:  true,
		Tty:          terminal.IsTerminal(int(file.Fd())),
	})
	if err != nil {
		return -1, err
	}
	defer attached.Close()

	state, err := terminal.MakeRaw(int(file.Fd()))
	if err != nil {
		return -1, err
	}

	errChan := make(chan error)

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		<-interrupt
		close(errChan)
	}()

	go func() {
		_, err := io.Copy(file, attached.Conn)
		errChan <- err
	}()

	go func() {
		_, err := io.Copy(attached.Conn, file)
		errChan <- err
	}()

	defer func() {
		terminal.Restore(int(file.Fd()), state)
	}()

	err = <-errChan

	if err != nil {
		return -1, err
	}

	resp, err := cli.ContainerExecInspect(ctx, id.ID)
	if err != nil {
		return -1, err
	}
	return resp.ExitCode, nil
}

func NewCleanupHandler(cli *client.Client, cleanupChan chan error, errWriter io.Writer, forceClean bool) (containers chan string, volumes chan string, done chan error) {

	ctx := context.Background()

	containers = make(chan string)
	volumes = make(chan string)
	done = make(chan error)

	go func() {
		createdContainers := []string{}
		createdVolumes := []string{}
		defer close(done)

		for {
			select {
			case container := <-containers:
				createdContainers = append(createdContainers, container)
			case volume := <-volumes:
				createdVolumes = append(createdVolumes, volume)
			case err := <-cleanupChan:
				log := false
				if err != nil {
					log = true
				}
				if err != nil || forceClean {
					for _, c := range createdContainers {
						if log {
							reader, err := cli.ContainerLogs(ctx, c, types.ContainerLogsOptions{ShowStderr: true, ShowStdout: true, Details: true})
							if err == nil {
								fmt.Fprintf(os.Stderr, "\n===== %s ====\n", c)
								io.Copy(os.Stderr, reader)
							}
						}
						err := cli.ContainerRemove(ctx, c, types.ContainerRemoveOptions{Force: true})
						if err != nil {
							fmt.Fprintf(errWriter, "%v\n", err)
						}
					}

					for _, v := range createdVolumes {
						err := cli.VolumeRemove(ctx, v, true)
						fmt.Printf("volume: %v\n", v)
						if err != nil {
							fmt.Fprintf(errWriter, "%v\n", err)
						}
					}
				}
				return
			}
		}
	}()

	return
}

func PrintProgress(progressReader io.ReadCloser, writer *os.File) error {
	isTerminal := terminal.IsTerminal(int(writer.Fd()))
	w, _, err := terminal.GetSize(int(writer.Fd()))

	if isTerminal && err == nil {
		scanner := bufio.NewScanner(progressReader)
		for scanner.Scan() {
			line := scanner.Text()
			if err := checkForError(line); err != nil {
				return err
			}
			clearLength := w - len(line)
			if clearLength < 0 {
				clearLength = 0
			}
			fmt.Print("\r" + line + strings.Repeat(" ", clearLength))
		}
	} else {
		fmt.Fprint(writer, "Downloading ...")
		scanner := bufio.NewScanner(progressReader)
		for scanner.Scan() {
			line := scanner.Text()
			if err := checkForError(line); err != nil {
				return err
			}
			fmt.Print(".")
		}
		fmt.Print("\n")
	}
	return nil
}

func checkForError(line string) error {
	pullStatus := &PullStatus{}
	if line != "" {
		err := json.Unmarshal([]byte(line), pullStatus)
		if err != nil {
			return err
		}
	}
	if pullStatus.Error != "" {
		return fmt.Errorf("%s", pullStatus.Error)
	}
	return nil
}

type PullStatus struct {
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

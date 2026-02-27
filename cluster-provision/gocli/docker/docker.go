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
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

func GetPrefixedContainers(cli *client.Client, prefix string) ([]types.Container, error) {
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}
	return filterByPrefix(containers, prefix), nil
}

func filterByPrefix(containers []types.Container, prefix string) []types.Container {
	prefixedConatiners := []types.Container{}
	for i, c := range containers {
		for _, name := range c.Names {
			if strings.HasPrefix(name, prefix) || strings.HasPrefix(name, "/"+prefix) {
				prefixedConatiners = append(prefixedConatiners, containers[i])
			}
		}
	}
	return prefixedConatiners
}

func GetPrefixedVolumes(cli *client.Client, prefix string) ([]*volume.Volume, error) {
	args := filters.NewArgs(filters.KeyValuePair{
		Key:   "name",
		Value: prefix,
	})
	options := volume.ListOptions{
		Filters: args,
	}
	volumes, err := cli.VolumeList(context.Background(), options)
	if err != nil {
		return nil, err
	}
	return volumes.Volumes, nil
}

func ImagePull(cli *client.Client, ctx context.Context, ref string, options image.PullOptions) error {

	if !strings.ContainsAny(ref, ":@") {
		ref = ref + ":latest"
	}
	ref = strings.TrimPrefix(ref, "docker.io/")

	images, err := cli.ImageList(ctx, image.ListOptions{All: true})
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

func Exec(cli *client.Client, containerID string, args []string, out io.Writer) (bool, error) {
	ctx := context.Background()
	id, err := cli.ContainerExecCreate(ctx, containerID, container.ExecOptions{
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

	attached, err := cli.ContainerExecAttach(ctx, id.ID, container.ExecStartOptions{
		Detach: false,
		Tty:    true,
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

func Terminal(cli *client.Client, containerID string, args []string, file *os.File) (int, error) {

	if !terminal.IsTerminal(int(file.Fd())) {
		return 1, fmt.Errorf("failure calling terminal out of TTY")
	}

	ctx := context.Background()
	id, err := cli.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Privileged:   true,
		Tty:          true,
		Detach:       false,
		Cmd:          args,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
	})

	if err != nil {
		return -1, err
	}

	attached, err := cli.ContainerExecAttach(ctx, id.ID, container.ExecStartOptions{
		Detach: false,
		Tty:    true,
	})
	if err != nil {
		return -1, err
	}
	defer attached.Close()

	state, err := terminal.MakeRaw(int(file.Fd()))
	if err != nil {
		return -1, err
	}

	resizeCh := make(chan os.Signal, 1)
	signal.Notify(resizeCh, syscall.SIGWINCH)
	resizeTerminal(ctx, cli, id.ID, file)
	go func() {
		for range resizeCh {
			resizeTerminal(ctx, cli, id.ID, file)
		}
	}()
	defer signal.Stop(resizeCh)

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
		containersFailedToRemove := []string{}

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
							reader, err := cli.ContainerLogs(ctx, c, container.LogsOptions{ShowStderr: true, ShowStdout: true, Details: true})
							if err == nil {
								fmt.Fprintf(os.Stderr, "\n===== %s ====\n", c)
								io.Copy(os.Stderr, reader)
							}
						}
						err := cli.ContainerRemove(ctx, c, container.RemoveOptions{Force: true})
						if err != nil {
							fmt.Fprintf(errWriter, "%v\n", err)
							containersFailedToRemove = append(containersFailedToRemove, c)
						}
					}
					for _, c := range containersFailedToRemove {
						err := cli.ContainerRemove(ctx, c, container.RemoveOptions{Force: true})
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
			if err := parseAndCheckForError(line, &PullStatus{}); err != nil {
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
		// Map to store which state was last printed for each ctr image layer
		lastReportedState := make(map[string]PullStatus)
		for scanner.Scan() {
			line := scanner.Text()
			// Discard empty lines
			if line == "" {
				continue
			}
			// Parse the progress json message into PullStatus struct
			pullStatus := &PullStatus{}
			if err := parseAndCheckForError(line, pullStatus); err != nil {
				return err
			}

			lastStatus, ok := lastReportedState[pullStatus.Id]
			toReport := false
			if ok {
				// This later was seen before
				if pullStatus.Status != lastStatus.Status {
					toReport = true
				} else {
					// Current status is same as last seen status
					// This is true only for Downloading and Extracting states
					lastProgress := float64(lastStatus.ProgressDetail.Current) / float64(lastStatus.ProgressDetail.Total)
					currProgress := float64(pullStatus.ProgressDetail.Current) / float64(pullStatus.ProgressDetail.Total)
					if currProgress-lastProgress >= 0.1 { // 10% progress
						toReport = true
					}
				}
			} else {
				// If this layer hasn't been seen before, print its status
				toReport = true
			}

			if toReport {
				fmt.Fprintf(writer, "%s\t%s\t%s\n", pullStatus.Id, pullStatus.Status, pullStatus.Progress)
				// Update the last reported status for this layer
				lastReportedState[pullStatus.Id] = *pullStatus
			}

		}
	}
	return nil
}

func parseAndCheckForError(line string, pullStatus *PullStatus) error {
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

type PullProgressDetail struct {
	Current int64 `json:"current"`
	Total   int64 `json:"total"`
}

type PullStatus struct {
	Id             string             `json:"id,omitempty"`
	Status         string             `json:"status,omitempty"`
	ProgressDetail PullProgressDetail `json:"progressDetail,omitempty"`
	Progress       string             `json:"progress,omitempty"`
	Error          string             `json:"error,omitempty"`
}

func resizeTerminal(ctx context.Context, cli *client.Client, execID string, file *os.File) {
	if w, h, err := terminal.GetSize(int(file.Fd())); err == nil {
		cli.ContainerExecResize(ctx, execID, container.ResizeOptions{
			Height: uint(h),
			Width:  uint(w),
		})
	}
}

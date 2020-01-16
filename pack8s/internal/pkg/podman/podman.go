package podman

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	logger "github.com/apsdehal/go-logger"
	"github.com/varlink/go/varlink"

	"github.com/fromanirh/pack8s/internal/pkg/images"
	"github.com/fromanirh/pack8s/pkg/varlinkapi/virtwriter"

	"github.com/fromanirh/pack8s/iopodman"
)

const (
	DefaultSocket   string = "unix:/run/podman/io.podman"
	LabelGeneration string = "io.kubevirt/pack8s.generation"
)

func SprintError(methodname string, err error) string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "Error calling %s: ", methodname)
	switch e := err.(type) {
	case *iopodman.ImageNotFound:
		//error ImageNotFound (name: string)
		fmt.Fprintf(buf, "'%v' name='%s'\n", e, e.Id)

	case *iopodman.ContainerNotFound:
		//error ContainerNotFound (name: string)
		fmt.Fprintf(buf, "'%v' name='%s'\n", e, e.Id)

	case *iopodman.NoContainerRunning:
		//error NoContainerRunning ()
		fmt.Fprintf(buf, "'%v'\n", e)

	case *iopodman.PodNotFound:
		//error PodNotFound (name: string)
		fmt.Fprintf(buf, "'%v' name='%s'\n", e, e.Name)

	case *iopodman.PodContainerError:
		//error PodContainerError (podname: string, errors: []PodContainerErrorData)
		fmt.Fprintf(buf, "'%v' podname='%s' errors='%v'\n", e, e.Podname, e.Errors)

	case *iopodman.NoContainersInPod:
		//error NoContainersInPod (name: string)
		fmt.Fprintf(buf, "'%v' name='%s'\n", e, e.Name)

	case *iopodman.ErrorOccurred:
		//error ErrorOccurred (reason: string)
		fmt.Fprintf(buf, "'%v' reason='%s'\n", e, e.Reason)

	case *iopodman.RuntimeError:
		//error RuntimeError (reason: string)
		fmt.Fprintf(buf, "'%v' reason='%s'\n", e, e.Reason)

	case *varlink.InvalidParameter:
		fmt.Fprintf(buf, "'%v' parameter='%s'\n", e, e.Parameter)

	case *varlink.MethodNotFound:
		fmt.Fprintf(buf, "'%v' method='%s'\n", e, e.Method)

	case *varlink.MethodNotImplemented:
		fmt.Fprintf(buf, "'%v' method='%s'\n", e, e.Method)

	case *varlink.InterfaceNotFound:
		fmt.Fprintf(buf, "'%v' interface='%s'\n", e, e.Interface)

	case *varlink.Error:
		fmt.Fprintf(buf, "'%v' parameters='%v'\n", e, e.Parameters)

	default:
		if err == io.EOF {
			fmt.Fprintf(buf, "Connection closed\n")
		} else if err == io.ErrUnexpectedEOF {
			fmt.Fprintf(buf, "Connection aborted\n")
		} else {
			fmt.Fprintf(buf, "%T - '%v'\n", err, err)
		}
	}
	return buf.String()
}

type PullProgressReporter interface {
	GetInterval() time.Duration
	Report(ref string, elapsed, completed uint64, err error) error
}

type Handle struct {
	PullReporter PullProgressReporter
	socket       string
	ctx          context.Context
	conn         *varlink.Connection
	log          *logger.Logger
}

func NewHandle(ctx context.Context, socket string, log *logger.Logger) (*Handle, error) {
	if socket == "" {
		socket = DefaultSocket
	}
	conn, err := varlink.NewConnection(ctx, socket)
	log.Infof("connected to %s", socket)
	hnd := Handle{
		PullReporter: pullProgressReporter{Log: log},
		socket:       socket,
		ctx:          ctx,
		conn:         conn,
		log:          log,
	}
	return &hnd, err
}

func (hnd *Handle) reconnect() (bool, error) {
	if hnd.conn == nil {
		var err error
		hnd.conn, err = varlink.NewConnection(hnd.ctx, hnd.socket)
		hnd.log.Noticef("reconnected to %s (err=%v)", hnd.socket, err)
		return true, err
	}
	hnd.log.Debugf("already connected to %s", hnd.socket)
	return false, nil
}

func (hnd *Handle) disconnect() error {
	var err error
	if hnd.conn != nil {
		err = hnd.conn.Close()
		hnd.conn = nil
		hnd.log.Infof("disconnected from %s (err=%v)", hnd.socket, err)
	}
	return err
}

type ReaderContext interface {
	Read(context.Context, []byte) (int, error)
}

type readerCtx struct {
	ReaderContext
	ctx context.Context
}

func (r readerCtx) Read(in []byte) (int, error) {
	return r.ReaderContext.Read(r.ctx, in)
}

func (hnd *Handle) Exec(container string, args []string, out io.Writer) error {
	_, err := hnd.reconnect()
	if err != nil {
		return err
	}
	defer hnd.disconnect()

	rwc, err := ExecContainer().Call(hnd.ctx, hnd.conn, iopodman.ExecOpts{
		Name:       container,
		Tty:        true,
		Privileged: true,
		Cmd:        args,
	})

	if err != nil {
		return err
	}

	rd := readerCtx{
		ReaderContext: rwc,
		ctx:           context.TODO(),
	}

	ecChan := make(chan int, 1)
	errChan := make(chan error, 1)
	go func() {
		// Read from the wire and direct to stdout or stderr
		err := virtwriter.Reader(rd, out, os.Stderr, nil, ecChan)
		errChan <- err
	}()

	err = <-errChan
	if err != nil {
		return err
	}
	rc := <-ecChan
	if rc != 0 {
		return fmt.Errorf("exec failed: rc=%d", rc)
	}
	return nil
}

func (hnd *Handle) GetPrefixedContainers(prefix string) ([]iopodman.Container, error) {
	_, err := hnd.reconnect()
	if err != nil {
		return nil, err
	}

	ret := []iopodman.Container{}
	containers, err := iopodman.ListContainers().Call(hnd.ctx, hnd.conn)
	if err != nil {
		return ret, err
	}

	hnd.log.Infof("found %d containers in the system - prefix=[%s]", len(containers), prefix)
	for _, cont := range containers {
		// TODO: why is it Name*s*? there is a bug lurking here? docs are unclear.
		if strings.HasPrefix(cont.Names, prefix) {
			hnd.log.Debugf("matching container: %s (%s)\n", cont.Names, cont.Id)
			ret = append(ret, cont)
		}
	}
	return ret, nil
}

func (hnd *Handle) GetPrefixedVolumes(prefix string) ([]iopodman.Volume, error) {
	_, err := hnd.reconnect()
	if err != nil {
		return nil, err
	}

	ret := []iopodman.Volume{}
	args := []string{}
	all := true
	volumes, err := iopodman.GetVolumes().Call(hnd.ctx, hnd.conn, args, all)
	if err != nil {
		return ret, err
	}

	hnd.log.Infof("found %d volumes in the system", len(volumes))
	for _, vol := range volumes {
		if strings.HasPrefix(vol.Name, prefix) {
			hnd.log.Debugf("matching volume: %s @(%s)\n", vol.Name, vol.MountPoint)
			ret = append(ret, vol)
		}
	}
	return ret, err
}

func (hnd *Handle) FindPrefixedContainer(prefixedName string) (iopodman.Container, error) {
	_, err := hnd.reconnect()
	if err != nil {
		return iopodman.Container{}, err
	}

	containers := []iopodman.Container{}

	containers, err = hnd.GetPrefixedContainers(prefixedName)
	if err != nil {
		return iopodman.Container{}, err
	}

	if len(containers) != 1 {
		return iopodman.Container{}, fmt.Errorf("failed to found the container with name %s", prefixedName)
	}
	return containers[0], nil
}

//PruneVolumes removes all unused volumes on the host.
func (hnd *Handle) PruneVolumes() error {
	_, err := hnd.reconnect()
	if err != nil {
		return err
	}

	_, _, err = iopodman.VolumesPrune().Call(hnd.ctx, hnd.conn)
	return err
}

//GetAllVolumes returns all volumes
func (hnd *Handle) GetAllVolumes() ([]iopodman.Volume, error) {
	_, err := hnd.reconnect()
	if err != nil {
		return nil, err
	}

	return iopodman.GetVolumes().Call(hnd.ctx, hnd.conn, []string{}, true)
}

func (hnd *Handle) RemoveVolumes(volumes []iopodman.Volume) error {
	_, err := hnd.reconnect()
	if err != nil {
		return err
	}

	volumeNames := []string{}
	for _, vol := range volumes {
		hnd.log.Infof("removing volume %s @%s", vol.Name, vol.MountPoint)
		volumeNames = append(volumeNames, vol.Name)
	}
	_, _, err = iopodman.VolumeRemove().Call(hnd.ctx, hnd.conn, iopodman.VolumeRemoveOpts{
		Volumes: volumeNames,
		Force:   true,
	})
	return err
}

func (hnd *Handle) RemoveContainer(cont iopodman.Container, force, removeVolumes bool) (string, error) {
	_, err := hnd.reconnect()
	if err != nil {
		return "", err
	}

	hnd.log.Infof("trying to remove: %s (%s) force=%v removeVolumes=%v\n", cont.Names, cont.Id, force, removeVolumes)
	return iopodman.RemoveContainer().Call(hnd.ctx, hnd.conn, cont.Id, force, removeVolumes)
}

func (hnd *Handle) CreateNamedVolume(name string) (string, error) {
	_, err := hnd.reconnect()
	if err != nil {
		return "", err
	}

	return iopodman.VolumeCreate().Call(hnd.ctx, hnd.conn, iopodman.VolumeCreateOpts{
		VolumeName: name,
	})
}

func (hnd *Handle) CreateContainer(conf iopodman.Create) (string, error) {
	_, err := hnd.reconnect()
	if err != nil {
		return "", err
	}

	return iopodman.CreateContainer().Call(hnd.ctx, hnd.conn, conf)
}

func (hnd *Handle) StopContainer(name string, timeout int64) (string, error) {
	_, err := hnd.reconnect()
	if err != nil {
		return "", err
	}

	return iopodman.StopContainer().Call(hnd.ctx, hnd.conn, name, timeout)
}

func (hnd *Handle) StartContainer(contID string) (string, error) {
	_, err := hnd.reconnect()
	if err != nil {
		return "", err
	}

	return iopodman.StartContainer().Call(hnd.ctx, hnd.conn, contID)
}

func (hnd *Handle) WaitContainer(name string, interval int64) (int64, error) {
	_, err := hnd.reconnect()
	if err != nil {
		return 0, err
	}

	return iopodman.WaitContainer().Call(hnd.ctx, hnd.conn, name, interval)
}

//ListImages returns all images on host
func (hnd *Handle) ListImages() ([]iopodman.Image, error) {
	_, err := hnd.reconnect()
	if err != nil {
		return nil, err
	}

	return iopodman.ListImages().Call(hnd.ctx, hnd.conn)
}

func (hnd *Handle) PullImageFromRegistry(registry, image string) error {
	imageRef := registry + "/" + image // TODO: is that a path? an URL? something else?
	return hnd.PullImage(imageRef)
}

func (hnd *Handle) PullImage(ref string) error {
	_, err := hnd.reconnect()
	if err != nil {
		return err
	}

	hnd.log.Noticef("pulling image: %s", ref)

	tries := []int{0, 1, 2, 6}
	interval := hnd.PullReporter.GetInterval() * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for idx, i := range tries {
		time.Sleep(time.Duration(i) * time.Second)

		prefix := fmt.Sprintf("attempt #%d", idx)
		hnd.log.Infof("%s to download '%s' - progress every %v\n", prefix, ref, interval)
		err := hnd.pullImage(ticker, prefix, ref)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("failed to download %s %d times, giving up.", ref, len(tries))
}

func (hnd *Handle) PullClusterImages(reqs images.Requests, clusterRegistry, clusterImage string) error {
	var err error
	err = hnd.PullImageFromRegistry(clusterRegistry, clusterImage)
	if err != nil {
		return err
	}
	err = hnd.PullImage(images.DockerRegistryImage)
	if err != nil {
		return err
	}
	if reqs.WantsNFS() {
		err = hnd.PullImage(images.NFSGaneshaImage)
		if err != nil {
			return err
		}
	}
	if reqs.WantsCeph() {
		err = hnd.PullImage(images.CephImage)
		if err != nil {
			return err
		}
	}
	if reqs.WantsFluentd() {
		err = hnd.PullImage(images.FluentdImage)
		if err != nil {
			return err
		}
	}
	return nil
}

func (hnd *Handle) pullImage(ticker *time.Ticker, prefix, ref string) error {
	var err error
	errChan := make(chan error)
	go func() {
		_, err := iopodman.PullImage().Call(hnd.ctx, hnd.conn, ref)
		errChan <- err
	}()

	hnd.PullReporter.Report(ref, 0, 0, nil)
	for {
		select {
		case <-ticker.C:
			hnd.PullReporter.Report(ref, 0, 1, nil)
		case err = <-errChan:
			return hnd.PullReporter.Report(ref, 1, 1, err)
		}
	}

	return fmt.Errorf("pull failed - internal error") // can't be reached
}

type pullProgressReporter struct {
	Log *logger.Logger
}

func (ppr pullProgressReporter) GetInterval() time.Duration {
	return 10 // assuming NOT-interactive report
}

func (ppr pullProgressReporter) Report(ref string, elapsed, completed uint64, err error) error {
	if err != nil {
		ppr.Log.Warningf("download failed for %s: %v\n", ref, err)
	} else if completed != 0 && elapsed == completed {
		ppr.Log.Noticef("downloaded completed for %s", ref)
	} else {
		ppr.Log.Infof("downloading %s...", ref)
	}
	return err
}

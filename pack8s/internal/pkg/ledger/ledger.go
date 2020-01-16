package ledger

import (
	"fmt"
	"io"

	logger "github.com/apsdehal/go-logger"

	"github.com/fromanirh/pack8s/internal/pkg/podman"
	"github.com/fromanirh/pack8s/iopodman"
)

type Ledger struct {
	hnd        *podman.Handle
	containers chan string
	volumes    chan string
	Done       chan error
	log        *logger.Logger
}

func NewLedger(hnd *podman.Handle, errWriter io.Writer, log *logger.Logger) Ledger {
	containers := make(chan string)
	volumes := make(chan string)
	done := make(chan error)

	go func() {
		createdContainers := []string{}
		createdVolumes := []string{}

		for {
			select {
			case container := <-containers:
				createdContainers = append(createdContainers, container)
			case volume := <-volumes:
				createdVolumes = append(createdVolumes, volume)
			case err := <-done:
				if err != nil {
					/*
						for _, c := range createdContainers {
							name, err := hnd.RemoveContainer(c, true, false)
							if err == nil {
								fmt.Printf("removed container: %v (%v)\n", name, c)
							} else {
								fmt.Fprintf(errWriter, "error removing container %v: %v\n", c, err)
							}
						}
					*/
					for _, v := range createdVolumes {
						//err := conn.VolumeRemove(ctx, v, true)
						fmt.Printf("volume: %v - can't do it yet", v)
						if err != nil {
							fmt.Fprintf(errWriter, "%v\n", err)
						}
					}
				}
			}
		}
	}()

	defer log.Infof("ledger ready")
	return Ledger{
		hnd:        hnd,
		log:        log,
		containers: containers,
		volumes:    volumes,
		Done:       done,
	}
}

func (ld Ledger) MakeVolume(name string) (string, error) {
	volName, err := ld.hnd.CreateNamedVolume(name)
	if err != nil {
		return volName, err
	}

	ld.volumes <- volName
	ld.log.Infof("tracked volume %s", volName)
	return volName, err
}

func (ld Ledger) RunContainer(conf iopodman.Create) (string, error) {
	ld.log.Debugf("running container...")
	contID, err := ld.hnd.CreateContainer(conf)
	if err != nil {
		return contID, err
	}

	ld.containers <- contID
	ld.log.Infof("tracked container %s", contID)
	if _, err := ld.hnd.StartContainer(contID); err != nil {
		return contID, err
	}
	ld.log.Infof("started container %s", contID)
	return contID, nil
}

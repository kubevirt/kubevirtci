package cri

// maybe just create wrappers around bash after all
type ContainerClient interface {
	ImagePull(image string) error
	Create(image string, co *CreateOpts) (string, error)
	Start(containerID string) error
	Remove(containerID string) error
	Inspect(containerID, format string) ([]byte, error)
	Build(tag, containerFile string, buildArgs map[string]string) error
	Run(runArgs []string) error
}

type CreateOpts struct {
	Privileged    bool
	Mounts        map[string]string
	Name          string
	Ports         map[string]string
	RestartPolicy string
	Network       string
	Command       []string
	Remove        bool
	Capabilities  []string
}

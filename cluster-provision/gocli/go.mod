module kubevirt.io/kubevirtci/cluster-provision/gocli

go 1.13

require (
	github.com/Microsoft/go-winio v0.4.7 // indirect
	github.com/Sirupsen/logrus v1.4.1
	github.com/docker/distribution v2.6.2+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.3.0
	github.com/docker/go-units v0.3.3 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/opencontainers/runc v0.1.1 // indirect
	github.com/pkg/errors v0.8.0 // indirect
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/spf13/cobra v0.0.2
	github.com/spf13/pflag v1.0.5
	github.com/stevvooe/resumable v0.0.0-20180830230917-22b14a53ba50 // indirect
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	k8s.io/apimachinery v0.18.4
)

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.1
	k8s.io/apimachinery/resource => k8s.io/apimachinery/resource v0.18.4
)

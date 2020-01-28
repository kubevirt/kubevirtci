package cmdutil

import (
	"fmt"
	"path/filepath"

	"github.com/fromanirh/pack8s/iopodman"

	"github.com/fromanirh/pack8s/internal/pkg/images"
	"github.com/fromanirh/pack8s/internal/pkg/ledger"
	"github.com/fromanirh/pack8s/internal/pkg/mounts"
	"github.com/fromanirh/pack8s/internal/pkg/podman"
)

func SetupRegistry(ldgr ledger.Ledger, prefix, network, registryVolume string, privileged bool) error {
	var err error
	// TODO: how to use the user-supplied name?
	var registryMounts mounts.MountMapping
	if registryVolume != "" {
		registryMounts, err = mounts.NewVolumeMappings(ldgr, []mounts.MountInfo{
			mounts.MountInfo{
				Name: fmt.Sprintf("%s-registry", prefix),
				Path: "/var/lib/registry",
				Type: "volume",
			},
		})
		if err != nil {
			return err
		}
	}

	registryName := fmt.Sprintf("%s-registry", prefix)
	registryMountsStrings := registryMounts.ToStrings()
	registryLabels := []string{fmt.Sprintf("%s=0001", podman.LabelGeneration)}
	_, err = ldgr.RunContainer(iopodman.Create{
		Args:       []string{images.DockerRegistryImage},
		Name:       &registryName,
		Label:      &registryLabels,
		Mount:      &registryMountsStrings,
		Network:    &network,
		Privileged: &privileged,
	})
	return err
}

func SetupNFS(ldgr ledger.Ledger, prefix, network, nfsData string, privileged bool) error {
	var err error
	nfsData, err = filepath.Abs(nfsData)
	if err != nil {
		return err
	}

	nfsName := fmt.Sprintf("%s-nfs", prefix)
	nfsMounts := []string{fmt.Sprintf("type=bind,source=%s,destination=/data/nfs", nfsData)}
	nfsLabels := []string{fmt.Sprintf("%s=010", podman.LabelGeneration)}
	_, err = ldgr.RunContainer(iopodman.Create{
		Args:       []string{images.NFSGaneshaImage},
		Name:       &nfsName,
		Label:      &nfsLabels,
		Mount:      &nfsMounts,
		Network:    &network,
		Privileged: &privileged,
	})
	return err
}

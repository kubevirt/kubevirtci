package mounts

import (
	"fmt"

	"github.com/fromanirh/pack8s/internal/pkg/ledger"
	"github.com/fromanirh/pack8s/iopodman"
)

type MountInfo struct {
	Name string
	Path string
	Type string
}

type MountMapping struct {
	data []iopodman.ContainerMount
}

func NewVolumeMappings(ldgr ledger.Ledger, mountInfos []MountInfo) (MountMapping, error) {
	mm := MountMapping{}
	var err error
	for _, mountItem := range mountInfos {
		volName, err := ldgr.MakeVolume(mountItem.Name)
		if err != nil {
			return mm, err
		}
		mm.data = append(mm.data, iopodman.ContainerMount{
			Type:        mountItem.Type,
			Source:      volName,
			Destination: mountItem.Path,
		})
	}
	return mm, err
}

func (mm MountMapping) ToStrings() []string {
	res := []string{}
	for _, mmItem := range mm.data {
		res = append(res, fmt.Sprintf("type=%s,source=%s,destination=%s", mmItem.Type, mmItem.Source, mmItem.Destination))
	}
	return res
}

package aaq

import (
	"embed"
	"fmt"
	"regexp"

	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
)

//go:embed manifests/*
var f embed.FS

type AaqOpt struct {
	client        k8s.K8sDynamicClient
	customVersion string
}

func NewAaqOpt(c k8s.K8sDynamicClient, cv string) *AaqOpt {
	return &AaqOpt{
		client:        c,
		customVersion: cv,
	}
}

func (o *AaqOpt) Exec() error {
	operator, err := f.ReadFile("manifests/operator.yaml")
	if err != nil {
		return err
	}
	cr, err := f.ReadFile("manifests/cr.yaml")
	if err != nil {
		return err
	}
	if o.customVersion != "" {
		pattern := `v[0-9]+\.[0-9]+\.[0-9]+(.*)?$`
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}
		operatorNewVersion := regex.ReplaceAllString(string(operator), o.customVersion)
		operator = []byte(operatorNewVersion)
	}

	for i, manifest := range [][]byte{operator, cr} {
		if err := o.client.Apply(manifest); err != nil {
			return fmt.Errorf("error applying manifest at index %d, %s", i, err)
		}
	}
	return nil
}

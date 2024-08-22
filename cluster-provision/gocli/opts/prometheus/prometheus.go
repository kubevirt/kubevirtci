package prometheus

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/sirupsen/logrus"

	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

//go:embed manifests/*
var f embed.FS

type prometheusOpt struct {
	grafanaEnabled      bool
	alertmanagerEnabled bool

	client k8s.K8sDynamicClient
}

func NewPrometheusOpt(c k8s.K8sDynamicClient, grafanaEnabled, alertmanagerEnabled bool) *prometheusOpt {
	return &prometheusOpt{
		grafanaEnabled:      grafanaEnabled,
		alertmanagerEnabled: alertmanagerEnabled,
		client:              c,
	}
}

func (o *prometheusOpt) Exec() error {
	for _, dir := range []string{"prometheus-operator", "prometheus", "monitors", "kube-state-metrics", "node-exporter"} {
		err := fs.WalkDir(f, "manifests/"+dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && filepath.Ext(path) == ".yaml" {
				yamlData, err := f.ReadFile(path)
				if err != nil {
					return err
				}
				yamlDocs := bytes.Split(yamlData, []byte("\n---\n"))
				for _, yamlDoc := range yamlDocs {
					if len(yamlDoc) == 0 {
						continue
					}

					obj, err := k8s.SerializeIntoObject(yamlDoc)
					if err != nil {
						logrus.Info(err.Error())
						continue
					}
					if err := o.client.Apply(obj); err != nil {
						return fmt.Errorf("error applying manifest %s", err)
					}
				}
			}
			return nil
		})

		if err != nil {
			return err
		}
	}

	if o.alertmanagerEnabled {
		for _, dir := range []string{"alertmanager", "alertmanager-rules"} {
			err := fs.WalkDir(f, "manifests/"+dir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && filepath.Ext(path) == ".yaml" {
					yamlData, err := f.ReadFile(path)
					if err != nil {
						return err
					}
					yamlDocs := bytes.Split(yamlData, []byte("---\n"))
					for _, yamlDoc := range yamlDocs {
						if len(yamlDoc) == 0 {
							continue
						}

						obj, err := k8s.SerializeIntoObject(yamlDoc)
						if err != nil {
							logrus.Info(err.Error())
							continue
						}
						if err := o.client.Apply(obj); err != nil {
							return fmt.Errorf("error applying manifest %s", err)
						}
					}
				}
				return nil
			})

			if err != nil {
				return err
			}
		}
	}

	if o.grafanaEnabled {
		err := fs.WalkDir(f, "manifests/grafana", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && filepath.Ext(path) == ".yaml" {
				yamlData, err := f.ReadFile(path)
				if err != nil {
					return err
				}
				yamlDocs := bytes.Split(yamlData, []byte("---\n"))
				for _, yamlDoc := range yamlDocs {
					if len(yamlDoc) == 0 {
						continue
					}

					obj, err := k8s.SerializeIntoObject(yamlDoc)
					if err != nil {
						logrus.Info(err.Error())
						continue
					}
					if err := o.client.Apply(obj); err != nil {
						return fmt.Errorf("error applying manifest %s", err)
					}
				}
			}
			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

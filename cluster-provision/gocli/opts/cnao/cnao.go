package cnao

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"

	"github.com/sirupsen/logrus"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/*
var f embed.FS

type cnaoOpt struct {
	client        k8s.K8sDynamicClient
	sshClient     libssh.Client
	multusEnabled bool
	dncEnabled    bool
	skipCR        bool
}

func NewCnaoOpt(c k8s.K8sDynamicClient, sshClient libssh.Client, multusEnabled, dncEnabled, skipCR bool) *cnaoOpt {
	return &cnaoOpt{
		client:        c,
		sshClient:     sshClient,
		multusEnabled: multusEnabled,
		skipCR:        skipCR,
		dncEnabled:    dncEnabled,
	}
}

func (o *cnaoOpt) Exec() error {
	err := fs.WalkDir(f, "manifests", func(path string, d fs.DirEntry, err error) error {
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

				if path == "manifests/network-addons-config-example.cr.yaml" {
					if o.skipCR {
						continue
					}

					if o.multusEnabled {
						re := regexp.MustCompile("(?m)[\r\n]+^.*multus:.*$")
						res := re.ReplaceAllString(string(yamlDoc), "")
						yamlDoc = []byte(res)
					}

					if !o.dncEnabled {
						re := regexp.MustCompile("(?m)[\r\n]+^.*multusDynamicNetworks:.*$")
						res := re.ReplaceAllString(string(yamlDoc), "")
						yamlDoc = []byte(res)
					}
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

	if err := o.sshClient.Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf wait deployment -n cluster-network-addons cluster-network-addons-operator --for condition=Available --timeout=200s"); err != nil {
		return err
	}
	return nil
}

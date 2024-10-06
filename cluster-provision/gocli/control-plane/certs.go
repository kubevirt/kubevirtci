package controlplane

import (
	"path"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/control-plane/certs"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/control-plane/crypto"
)

type CertsPhase struct {
	pkiPath string
}

func NewCertsPhase(pkiPath string) *CertsPhase {
	return &CertsPhase{
		pkiPath: pkiPath,
	}
}

func (p *CertsPhase) Run() error {
	componentCerts := certs.GetAllCertificates()
	ca, caKey, err := crypto.GenerateKubernetesCAKeyPair()
	if err != nil {
		return err
	}

	err = crypto.WriteKeyAndCertToFile(ca, caKey, path.Join(p.pkiPath, "ca.crt"), path.Join(p.pkiPath, "key.pem"))
	if err != nil {
		return err
	}

	for _, componentCert := range componentCerts {
		cert, key, err := crypto.GenerateCertKeyPairWithCA(componentCert.Config, ca, caKey)
		if err != nil {
			return err
		}
		err = crypto.WriteKeyAndCertToFile(cert, key, path.Join(p.pkiPath, componentCert.Name+".crt"), path.Join(p.pkiPath, componentCert.Name+".pem"))
		if err != nil {
			return err
		}
	}
	return nil
}

package tests

import (
	"flag"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

var (
	kubeconfig string
	k8sClient  k8s.K8sDynamicClient
)

func TestTests(t *testing.T) {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")

	RegisterFailHandler(Fail)
	RunSpecs(t, "Tests Suite")
}

var _ = BeforeEach(func() {
	config, err := k8s.NewConfig(kubeconfig, 36443)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = k8s.NewDynamicClient(config)
	Expect(err).NotTo(HaveOccurred())
})

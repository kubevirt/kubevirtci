package podman_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPodman(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Podman Suite")
}

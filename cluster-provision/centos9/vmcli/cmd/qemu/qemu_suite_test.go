package qemu_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestQemu(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Qemu Suite")
}

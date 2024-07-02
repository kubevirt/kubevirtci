package cmd

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

func TestCmd(t *testing.T) {
	suite.Run(t, new(TestSuite))
	RegisterFailHandler(Fail)
	RunSpecs(t, "Provision Manager Suite")
}

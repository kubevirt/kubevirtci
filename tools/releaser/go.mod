module kubevirt.io/kubevirtci/tools/releaser

go 1.13

require (
	github.com/go-git/go-git/v5 v5.1.1-0.20200721083337-cded5b685b8a
	github.com/google/go-containerregistry v0.0.0-20200115214256-379933c9c22b
	github.com/google/go-github/v32 v32.1.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/otiai10/copy v1.0.2
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/sosedoff/gitkit v0.2.1-0.20191202022816-7182d43c6254
	github.com/udhos/equalfile v0.3.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v9.0.0+incompatible
	k8s.io/test-infra v0.0.0-20200728085909-4407d8aec1ee
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	k8s.io/client-go => k8s.io/client-go v0.17.3
)

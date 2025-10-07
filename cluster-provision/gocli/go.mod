module kubevirt.io/kubevirtci/cluster-provision/gocli

go 1.24

toolchain go1.24.7

require (
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/alessio/shellescape v1.4.2
	github.com/bramvdbogaerde/go-scp v1.2.0
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/docker/docker v26.1.5+incompatible
	github.com/docker/go-connections v0.5.0
	github.com/onsi/ginkgo/v2 v2.18.0
	github.com/onsi/gomega v1.33.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.74.0
	github.com/rook/rook/pkg/apis v0.0.0-20240801164516-9c98467f3c32
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/afero v1.11.0
	github.com/spf13/cobra v1.8.0
	github.com/spf13/pflag v1.0.5
	go.uber.org/mock v0.4.0
	golang.org/x/crypto v0.35.0
	golang.org/x/net v0.33.0
	gopkg.in/yaml.v3 v3.0.1
	istio.io/operator v0.0.0-20200714085832-f408beefc360
	k8s.io/api v0.30.3
	k8s.io/apiextensions-apiserver v0.30.1
	k8s.io/apimachinery v0.30.3
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/pod-security-admission v0.22.0-beta.0.0.20240531013614-68e02f3c6735
	kubevirt.io/application-aware-quota-api v1.2.3
	kubevirt.io/containerized-data-importer-api v1.59.1-0.20240610172909-253d5a6e7f95
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/Microsoft/hcsshim v0.12.3 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containernetworking/cni v1.2.0-rc1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/emicklei/go-restful/v3 v3.12.1 // indirect
	github.com/evanphx/json-patch v5.9.0+incompatible // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-jose/go-jose/v4 v4.0.5 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-test/deep v1.1.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.8 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.6 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-5 // indirect
	github.com/hashicorp/vault/api v1.14.0 // indirect
	github.com/hashicorp/vault/api/auth/approle v0.6.0 // indirect
	github.com/hashicorp/vault/api/auth/kubernetes v0.6.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/k8snetworkplumbingwg/network-attachment-definition-client v1.7.0 // indirect
	github.com/kube-object-storage/lib-bucket-provisioner v0.0.0-20221122204822-d1a8c34382f1 // indirect
	github.com/libopenstorage/secrets v0.0.0-20240416031220-a17cf7f72c6c // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/sys/user v0.1.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/natefinch/lumberjack v2.0.0+incompatible // indirect
	github.com/nxadm/tail v1.4.11 // indirect
	github.com/openshift/api v0.0.0 // indirect
	github.com/openshift/custom-resource-status v1.1.2 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.45.0 // indirect
	go.opentelemetry.io/otel v1.20.0 // indirect
	go.opentelemetry.io/otel/metric v1.20.0 // indirect
	go.opentelemetry.io/otel/trace v1.20.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/oauth2 v0.21.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/grpc v1.62.1 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gotest.tools/v3 v3.5.1 // indirect
	istio.io/api v1.22.0 // indirect
	istio.io/pkg v0.0.0-20191029184635-5c2f5ef63692 // indirect
	k8s.io/component-base v0.30.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20240620174524-b456828f718b // indirect
	k8s.io/utils v0.0.0-20240502163921-fe8a2dddb1d0 // indirect
	kubevirt.io/application-aware-quota v0.0.0-20240609100014-e72e988ec98a // indirect
	kubevirt.io/controller-lifecycle-operator-sdk/api v0.0.0-20220329064328-f3cc58c6ed90 // indirect
	sigs.k8s.io/controller-runtime v0.17.2 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/containerd/containerd v1.7.17 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/pprof v0.0.0-20240424215950-a892ee059fd6 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/term v0.29.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect; gindirect
)

replace (
	github.com/kubernetes-incubator/external-storage => github.com/libopenstorage/external-storage v0.20.4-openstorage-rc3
	github.com/libopenstorage/secrets => github.com/rook/secrets v0.0.0-20240315053144-3195f6906937
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.1.1-0.20240131200429-02120488a4c0
	github.com/openshift/api => github.com/openshift/api v0.0.0-20230406152840-ce21e3fe5da2
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20230324103026-3f1513df25e0
	github.com/operator-framework/operator-lifecycle-manager => github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190128024246-5eb7ae5bdb7a
	github.com/portworx/sched-ops => github.com/portworx/sched-ops v0.20.4-openstorage-rc3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.28.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.28.4
	k8s.io/client-go => k8s.io/client-go v0.30.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.28.4
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.28.4
	k8s.io/component-base => k8s.io/component-base v0.28.4
	k8s.io/cri-api => k8s.io/cri-api v0.28.4
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.28.4
	k8s.io/endpointslice => k8s.io/staging/src/k8s.io/endpointslice v0.28.4
	k8s.io/klog => k8s.io/klog v0.4.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.28.4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.28.4
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.28.4
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.28.4
	k8s.io/kubectl => k8s.io/kubectl v0.28.4
	k8s.io/kubelet => k8s.io/kubelet v0.28.4
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.28.4
	k8s.io/metrics => k8s.io/metrics v0.28.4
	k8s.io/node-api => k8s.io/node-api v0.28.4
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.28.4
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.28.4
	k8s.io/sample-controller => k8s.io/sample-controller v0.28.4
	kubevirt.io/qe-tools => kubevirt.io/qe-tools v0.1.8
)

exclude github.com/kubernetes-incubator/external-storage v0.20.4-openstorage-rc2

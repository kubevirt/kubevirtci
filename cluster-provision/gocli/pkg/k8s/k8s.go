package utils

import (
	"context"
	"fmt"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	"github.com/sirupsen/logrus"
	istiov1alpha1 "istio.io/operator/pkg/apis"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	admissionv1 "k8s.io/pod-security-admission/admission/api/v1"
	aaqv1alpha1 "kubevirt.io/application-aware-quota-api/pkg/apis/core/v1alpha1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	"github.com/cenkalti/backoff/v4"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/testing"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

type K8sDynamicClient interface {
	Get(gvk schema.GroupVersionKind, name, ns string) (*unstructured.Unstructured, error)
	Apply(obj *unstructured.Unstructured) error
	List(gvk schema.GroupVersionKind, ns string) (*unstructured.UnstructuredList, error)
	Delete(gvk schema.GroupVersionKind, name, ns string) error
}

type k8sDynamicClientImpl struct {
	scheme *runtime.Scheme
	client dynamic.Interface
}
type ReactorConfig struct {
	verb      string
	resource  string
	reactfunc func(action testing.Action) (bool, runtime.Object, error)
}

func NewConfig(manifestPath string, apiServerPort uint16) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", manifestPath)
	if err != nil {
		return nil, fmt.Errorf("Error building kubeconfig: %v", err)
	}
	config.Host = "https://127.0.0.1:" + fmt.Sprintf("%d", apiServerPort)
	config.Insecure = true
	config.CAData = []byte{}
	return config, nil
}

func NewDynamicClient(config *rest.Config) (*k8sDynamicClientImpl, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Error creating dynamic client: %v", err)
	}
	s := initSchema()
	return &k8sDynamicClientImpl{
		client: dynamicClient,
		scheme: s,
	}, nil
}

func NewTestClient(reactors ...ReactorConfig) *k8sDynamicClientImpl {
	s := initSchema()
	dynamicClient := fake.NewSimpleDynamicClient(s)
	for _, r := range reactors {
		dynamicClient.PrependReactor(r.verb, r.resource, r.reactfunc)
	}

	return &k8sDynamicClientImpl{
		client: dynamicClient,
		scheme: s,
	}
}

func NewReactorConfig(v string, r string, reactfunc func(action testing.Action) (bool, runtime.Object, error)) ReactorConfig {
	return ReactorConfig{
		verb:      v,
		resource:  r,
		reactfunc: reactfunc,
	}
}

func (c *k8sDynamicClientImpl) Get(gvk schema.GroupVersionKind, name, ns string) (*unstructured.Unstructured, error) {
	resourceClient, err := c.initResourceClientForGVKAndNamespace(gvk, ns)
	if err != nil {
		return nil, err
	}

	obj, err := resourceClient.Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (c *k8sDynamicClientImpl) List(gvk schema.GroupVersionKind, ns string) (*unstructured.UnstructuredList, error) {
	resourceClient, err := c.initResourceClientForGVKAndNamespace(gvk, ns)
	if err != nil {
		return nil, err
	}

	objs, err := resourceClient.List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return objs, nil
}

func (c *k8sDynamicClientImpl) Apply(obj *unstructured.Unstructured) error {
	gvk := obj.GroupVersionKind()
	resourceClient, err := c.initResourceClientForGVKAndNamespace(gvk, obj.GetNamespace())
	if err != nil {
		return err
	}

	err = c.createWithExponentialBackoff(resourceClient, obj)
	if err != nil {
		return err
	}
	logrus.Infof("Object %v applied successfully", obj.GetName())
	return nil
}

func SerializeIntoObject(scheme *runtime.Scheme, manifest []byte) (*unstructured.Unstructured, error) {
	jsonData, err := yaml.YAMLToJSON(manifest)
	if err != nil {
		return nil, fmt.Errorf("Error converting YAML to JSON: %v\n", err)
	}

	obj := &unstructured.Unstructured{}
	dec := serializer.NewCodecFactory(scheme).UniversalDeserializer()
	_, _, err = dec.Decode(jsonData, nil, obj)
	if err != nil {
		return nil, fmt.Errorf("Error decoding JSON to Unstructured object: %v\n", err)
	}
	return obj, nil
}

func (c *k8sDynamicClientImpl) createWithExponentialBackoff(resourceClient dynamic.ResourceInterface, obj *unstructured.Unstructured) error {
	operation := func() error {
		_, err := resourceClient.Create(context.TODO(), obj, v1.CreateOptions{})
		return err
	}

	backoffStrategy := backoff.NewExponentialBackOff()
	backoffStrategy.InitialInterval = 3 * time.Second
	backoffStrategy.MaxElapsedTime = 1 * time.Minute

	err := backoff.Retry(operation, backoffStrategy)
	if err != nil {
		return err
	}
	return nil
}

func (c *k8sDynamicClientImpl) Delete(gvk schema.GroupVersionKind, name, ns string) error {
	resourceClient, err := c.initResourceClientForGVKAndNamespace(gvk, ns)
	if err != nil {
		return err
	}

	err = resourceClient.Delete(context.TODO(), name, v1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func initSchema() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)

	_ = apiextensionsv1.AddToScheme(s)
	_ = cephv1.AddToScheme(s)
	_ = monitoringv1alpha1.AddToScheme(s)
	_ = monitoringv1.AddToScheme(s)
	_ = istiov1alpha1.AddToScheme(s)
	_ = admissionv1.AddToScheme(s)
	_ = cdiv1beta1.AddToScheme(s)
	_ = aaqv1alpha1.AddToScheme(s)
	_ = corev1.AddToScheme(s)
	return s
}

func (c *k8sDynamicClientImpl) initResourceClientForGVKAndNamespace(gvk schema.GroupVersionKind, ns string) (dynamic.ResourceInterface, error) {
	restMapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{gvk.GroupVersion()})
	restMapper.Add(gvk, meta.RESTScopeNamespace)
	mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	resourceClient := c.client.Resource(mapping.Resource).Namespace(ns)
	if ns == "" {
		resourceClient = c.client.Resource(mapping.Resource)
	}
	return resourceClient, nil
}

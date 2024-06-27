package utils

import (
	"bytes"
	"context"
	"fmt"
	"time"

	cephv1 "github.com/aerosouund/rook/pkg/apis/ceph.rook.io/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/sirupsen/logrus"
	istiov1alpha1 "istio.io/operator/pkg/apis"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	admissionv1 "k8s.io/pod-security-admission/admission/api/v1"
	aaqv1alpha1 "kubevirt.io/application-aware-quota-api/pkg/apis/core/v1alpha1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

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
	Apply(manifest []byte) error
	List(gvk schema.GroupVersionKind, ns string) (*unstructured.UnstructuredList, error)
	Delete(gvk schema.GroupVersionKind, name, ns string) error
}

type K8sDynamicClientImpl struct {
	scheme *runtime.Scheme
	client dynamic.Interface
}
type ReactorConfig struct {
	verb      string
	resource  string
	reactfunc func(action testing.Action) (bool, runtime.Object, error)
}

func InitConfig(manifestPath string, apiServerPort uint16) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", manifestPath)
	if err != nil {
		return nil, fmt.Errorf("Error building kubeconfig: %v", err)
	}
	config.Host = "https://127.0.0.1:" + fmt.Sprintf("%d", apiServerPort)
	config.Insecure = true
	config.CAData = []byte{}
	return config, nil
}

func NewDynamicClient(config *rest.Config) (*K8sDynamicClientImpl, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Error creating dynamic client: %v", err)
	}
	s := runtime.NewScheme()
	scheme.AddToScheme(s)

	apiextensionsv1.AddToScheme(s)
	cephv1.AddToScheme(s)
	monitoringv1alpha1.AddToScheme(s)
	monitoringv1.AddToScheme(s)
	istiov1alpha1.AddToScheme(s)
	admissionv1.AddToScheme(s)
	cdiv1beta1.AddToScheme(s)
	aaqv1alpha1.AddToScheme(s)

	return &K8sDynamicClientImpl{
		client: dynamicClient,
		scheme: s,
	}, nil
}

func NewTestClient(reactors ...ReactorConfig) K8sDynamicClient {
	s := runtime.NewScheme()
	scheme.AddToScheme(s)

	apiextensionsv1.AddToScheme(s)
	cephv1.AddToScheme(s)
	monitoringv1alpha1.AddToScheme(s)
	monitoringv1.AddToScheme(s)
	istiov1alpha1.AddToScheme(s)
	admissionv1.AddToScheme(s)
	cdiv1beta1.AddToScheme(s)
	aaqv1alpha1.AddToScheme(s)

	dynamicClient := fake.NewSimpleDynamicClient(s)
	for _, r := range reactors {
		dynamicClient.PrependReactor(r.verb, r.resource, r.reactfunc)
	}

	return &K8sDynamicClientImpl{
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

func (c *K8sDynamicClientImpl) Get(gvk schema.GroupVersionKind, name, ns string) (*unstructured.Unstructured, error) {
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

func (c *K8sDynamicClientImpl) List(gvk schema.GroupVersionKind, ns string) (*unstructured.UnstructuredList, error) {
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

func (c *K8sDynamicClientImpl) Apply(manifest []byte) error {
	yamlDocs := bytes.Split(manifest, []byte("---\n"))
	for _, yamlDoc := range yamlDocs {
		if len(yamlDoc) == 0 {
			continue
		}

		jsonData, err := yaml.YAMLToJSON(yamlDoc)
		if err != nil {
			fmt.Printf("Error converting YAML to JSON: %v\n", err)
			continue
		}

		obj := &unstructured.Unstructured{}
		dec := serializer.NewCodecFactory(c.scheme).UniversalDeserializer()
		_, _, err = dec.Decode(jsonData, nil, obj)
		if err != nil {
			fmt.Printf("Error decoding JSON to Unstructured object: %v\n", err)
			continue
		}

		gvk := obj.GroupVersionKind()
		resourceClient, err := c.initResourceClientForGVKAndNamespace(gvk, obj.GetNamespace())
		if err != nil {
			return err
		}

		err = c.createWithRetries(resourceClient, obj)
		if err != nil {
			return err
		}
		logrus.Infof("Object %v applied successfully\n", obj.GetName())
	}

	return nil
}

func (c *K8sDynamicClientImpl) createWithRetries(resourceClient dynamic.ResourceInterface, obj *unstructured.Unstructured) error {
	const maxRetries = 3
	var err error

	for i := 0; i < maxRetries; i++ {
		_, err = resourceClient.Create(context.TODO(), obj, v1.CreateOptions{})
		if err == nil {
			return nil
		}
		logrus.Infof("Attempt %d: Error applying manifest: %v for object %s\n", i+1, err, obj.GetName())
		time.Sleep(3 * time.Second)
	}

	return fmt.Errorf("Error applying manifest after %d attempts: %v for object %s", maxRetries, err, obj.GetName())
}

func (c *K8sDynamicClientImpl) Delete(gvk schema.GroupVersionKind, name, ns string) error {
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

func (c *K8sDynamicClientImpl) initResourceClientForGVKAndNamespace(gvk schema.GroupVersionKind, ns string) (dynamic.ResourceInterface, error) {
	restMapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{gvk.GroupVersion()})
	restMapper.Add(gvk, meta.RESTScopeNamespace)
	mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	var resourceClient dynamic.ResourceInterface
	resourceClient = c.client.Resource(mapping.Resource).Namespace(ns)
	if ns == "" {
		resourceClient = c.client.Resource(mapping.Resource)
	}
	return resourceClient, nil
}

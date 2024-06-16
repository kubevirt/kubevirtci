package utils

import (
	"bytes"
	"context"
	"embed"
	"fmt"

	cephv1 "github.com/aerosouund/rook/pkg/apis/ceph.rook.io/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	istiov1alpha1 "istio.io/operator/pkg/apis"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	admissionv1 "k8s.io/pod-security-admission/admission/api/v1"

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
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

type K8sDynamicClient interface {
	Get(gvk schema.GroupVersionKind, name, ns string) (*unstructured.Unstructured, error)
	Apply(fs embed.FS, manifestPath string) error
	List(gvk schema.GroupVersionKind, ns string) (*unstructured.UnstructuredList, error)
	Delete(gvk schema.GroupVersionKind, name, ns string) error
}

type K8sDynamicClientImpl struct {
	scheme *runtime.Scheme
	client dynamic.Interface
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

func NewDynamicClient(config *rest.Config) (K8sDynamicClient, error) {
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

	return &K8sDynamicClientImpl{
		client: dynamicClient,
		scheme: s,
	}, nil
}

func NewTestClient() K8sDynamicClient {
	s := runtime.NewScheme()
	scheme.AddToScheme(s)
	apiextensionsv1.AddToScheme(s)
	cephv1.AddToScheme(s)
	monitoringv1alpha1.AddToScheme(s)
	monitoringv1.AddToScheme(s)
	istiov1alpha1.AddToScheme(s)

	dynamicClient := fake.NewSimpleDynamicClient(s)

	return &K8sDynamicClientImpl{
		client: dynamicClient,
		scheme: s,
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

func (c *K8sDynamicClientImpl) Apply(fs embed.FS, manifestPath string) error {
	yamlData, err := fs.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("Error reading YAML file: %v", err)

	}
	yamlDocs := bytes.Split(yamlData, []byte("---\n"))
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

		_, err = resourceClient.Create(context.TODO(), obj, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("Error applying manifest: %v", err)
		}

		fmt.Printf("Object %v applied successfully!\n", obj.GetName())
	}

	return nil
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

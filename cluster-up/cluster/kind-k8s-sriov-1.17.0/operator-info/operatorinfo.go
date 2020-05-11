package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang/glog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Extracts SR-IOV related information from a kubernetes cluster.
// The exported information is persisted in the `artifactsDir` directory.
type SRIOVReporter struct {
	client       *kubernetes.Clientset
	namespace    string
	artifactsDir string
}

const (
	sriovEntityURITemplate = "/apis/sriovnetwork.openshift.io/v1/namespaces/%s/%s/"
	sriovNodeNetworkPolicyEntity = "sriovnetworknodepolicies"
	sriovNodeStateEntity = "sriovnetworknodestates"
	sriovNetworks = "sriovnetworks"
	sriovOperatorConfigs = "sriovoperatorconfigs"
)

// Gather all SR-IOV info into an 'sriov' folder in the specified artifact
// directory.
func (s *SRIOVReporter) DumpInfo() {
	sriovDir := filepath.Join(s.artifactsDir, "sriov")
	if err := os.MkdirAll(sriovDir, 0777); err != nil {
		glog.Errorf("failed to create directory: %v\n", err)
		return
	}

	s.logSRIOVNodeState(filepath.Join(sriovDir, "nodestate.log"))
	s.logSRIOVNodeNetworkPolicies(filepath.Join(sriovDir, "nodenetworkpolicies.log"))
	s.logNetworks(filepath.Join(sriovDir, "networks.log"))
	s.logOperatorConfigs(filepath.Join(sriovDir, "operatorconfigs.log"))
}

func (s *SRIOVReporter) logSRIOVNodeState(nodeStateLogPath string) {
	s.dumpK8sEntityToFile(sriovNodeStateEntity, nodeStateLogPath)
}

func (s *SRIOVReporter) logSRIOVNodeNetworkPolicies(nodeNetworkPolicyLogPath string) {
	s.dumpK8sEntityToFile(sriovNodeNetworkPolicyEntity, nodeNetworkPolicyLogPath)
}

func (s *SRIOVReporter) logNetworks(networksPath string) {
	s.dumpK8sEntityToFile(sriovNetworks, networksPath)
}

func (s *SRIOVReporter) logOperatorConfigs(operatorConfigPath string) {
	s.dumpK8sEntityToFile(sriovOperatorConfigs, operatorConfigPath)
}

func (s *SRIOVReporter) dumpK8sEntityToFile(entityName string, outputFilePath string) {
	requestURI := fmt.Sprintf(sriovEntityURITemplate, s.namespace, entityName)
	glog.V(4).Infof("Reading entity: %s from URI: %s", entityName, requestURI)
	f, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		glog.Errorf("failed to open the file: %v\n", err)
		return
	}
	defer f.Close()

	response, err := s.client.RESTClient().Get().RequestURI(requestURI).Do().Get()
	if err != nil {
		glog.Errorf("failed to dump SR-IOV [%s] entities: %v", entityName, err)
		return
	}

	j, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		glog.Errorf("Failed to marshall SR-IOV [%s] state objects", entityName)
		return
	}
	fmt.Fprintln(f, string(j))
}

func main() {
	artifacts := flag.String("artifacts", "", "the artifact dir")
	kubeconfig := flag.String("kubeconfig", "", "the path of kubeconfig")
	namespace := flag.String("namespace", corev1.NamespaceAll, "the namespace where the sr-iov operator does its black magic")
	flag.Parse()

	if *artifacts == "" {
		*artifacts = os.Getenv("ARTIFACTS")
		if *artifacts != "" {
			glog.V(4).Infof("Using implicit env artifacts dir %s", *artifacts)
		} else {
			glog.Warning("Could not derive the artifact output dir from the current environment")
			flag.Usage()
			return
		}
	}

	if *kubeconfig == "" {
		*kubeconfig = os.Getenv("KUBECONFIG")
		if *kubeconfig != "" {
			glog.V(4).Infof("Using implicit env kubeconfig %s", *kubeconfig)
		} else {
			glog.Warning("Could not derive kubeconfig from the current environment")
			flag.Usage()
			return
		}
	}

	client, err := getK8sClient(kubeconfig)
	if err != nil {
		glog.Fatalf("Error creating KubeVirt API: %v", err)
		return
	}

	sriovReporter := SRIOVReporter{
		client:       client,
		namespace:    *namespace,
		artifactsDir: *artifacts,
	}
	sriovReporter.DumpInfo()
}

func getK8sClient(kubeconfig *string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

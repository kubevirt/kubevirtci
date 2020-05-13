package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang/glog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	sriovEntityURITemplate       = "/apis/sriovnetwork.openshift.io/v1/namespaces/%s/%s/"
	sriovNodeNetworkPolicyEntity = "sriovnetworknodepolicies"
	sriovNodeStateEntity         = "sriovnetworknodestates"
	sriovNetworks                = "sriovnetworks"
	sriovOperatorConfigs         = "sriovoperatorconfigs"
)

// Extracts SR-IOV related information from a kubernetes cluster.
// The exported information is persisted in the `outputDir` directory.
type SRIOVReporter struct {
	client    *kubernetes.Clientset
	namespace string
	outputDir string
}

// Gather all SR-IOV info into an 'sriov' folder in the specified artifact
// directory.
func (s *SRIOVReporter) DumpInfo() {
	if err := os.MkdirAll(s.outputDir, 0777); err != nil {
		glog.Errorf("failed to create directory: %v", err)
		return
	}

	s.logSRIOVNodeState(filepath.Join(s.outputDir, "nodestate.log"))
	s.logSRIOVNodeNetworkPolicies(filepath.Join(s.outputDir, "nodenetworkpolicies.log"))
	s.logNetworks(filepath.Join(s.outputDir, "networks.log"))
	s.logOperatorConfigs(filepath.Join(s.outputDir, "operatorconfigs.log"))

	pods := s.client.CoreV1().Pods(s.namespace)
	sriovPods, err := pods.List(metav1.ListOptions{})
	if err != nil {
		glog.Errorf("Could not list pods in sriov namespace. Error: %v", err)
		return
	}

	events := s.client.CoreV1().Events(s.namespace)
	for _, sriovPod := range sriovPods.Items {
		glog.V(4).Infof("Iterating sriov pod %s", sriovPod.Name)
		s.serializePodSpec(sriovPod)
		s.fetchPodLogs(pods, sriovPod.Name)
		s.fetchPodEvents(events, sriovPod)
	}
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
		glog.Errorf("failed to open file: %v", err)
		return
	}
	defer f.Close()

	response, err := s.client.RESTClient().Get().RequestURI(requestURI).Do().Raw()
	if err != nil {
		glog.Errorf("failed to dump SR-IOV [%s] entities: %v", entityName, err)
		return
	}
	glog.V(4).Infof("Entities for %s: %v", entityName, string(response))

	var prettyJson bytes.Buffer
	err = json.Indent(&prettyJson, response, "", "    ")
	if err != nil {
		glog.Errorf("Failed to marshall SR-IOV [%s] state objects", entityName)
		return
	}
	fmt.Fprintln(f, string(prettyJson.Bytes()))
}

func (s *SRIOVReporter) serializePodSpec(pod corev1.Pod) {
	outputFilePath := fmt.Sprintf("%s/%s-spec.log", s.outputDir, pod.Name)
	f, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		glog.Errorf("failed to open file: %v", err)
		return
	}
	defer f.Close()

	jsonStruct, err := json.MarshalIndent(pod, "", "    ")
	if err != nil {
		glog.Errorf("Failed to marshall pod %s", pod.Name)
		return
	}
	glog.V(4).Infof("Stored pod spec for pod %s", pod.Name)
	fmt.Fprintln(f, string(jsonStruct))
}

func (s *SRIOVReporter) fetchPodLogs(pods v1.PodInterface, podName string) {
	s.persistRawLogs(pods, podName, &corev1.PodLogOptions{}, fmt.Sprintf("%s/%s.log", s.outputDir, podName))
	s.persistRawLogs(pods, podName, &corev1.PodLogOptions{Previous: true}, fmt.Sprintf("%s/%s_previous.log", s.outputDir, podName))
}

func (s *SRIOVReporter) persistRawLogs(pods v1.PodInterface, podName string, logOptions *corev1.PodLogOptions, outputFilePath string) {
	f, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		glog.Errorf("failed to open file: %v", err)
		return
	}
	defer f.Close()

	rawLogs, err := pods.GetLogs(podName, logOptions).DoRaw()
	if err != nil {
		glog.Errorf("Could not get pod logs for pod %s. Error: %v", podName, err)
		return
	}
	glog.V(4).Infof("Stored pod log for pod %s in file %s", podName, outputFilePath)
	fmt.Fprintln(f, string(rawLogs))
}

func (s *SRIOVReporter) fetchPodEvents(events v1.EventInterface, pod corev1.Pod) {
	outputFilePath := fmt.Sprintf("%s/%s_events.log", s.outputDir, pod.Name)
	f, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		glog.Errorf("failed to open file: %v", err)
		return
	}
	defer f.Close()

	podUID := string(pod.UID)
	eventFilterType := "Pod"
	fieldSelector := events.GetFieldSelector(nil, nil, &eventFilterType, &podUID).String()
	glog.V(4).Infof("Event field selector: %s", fieldSelector)

	eventList, err := events.List(metav1.ListOptions{FieldSelector: fieldSelector})
	if err != nil {
		glog.Errorf("Could not list events. Error: %v", err)
	}

	jsonStruct, err := json.MarshalIndent(eventList, "", "    ")
	if err != nil {
		glog.Errorf("Failed to marshall event list for pod %s", pod.Name)
		return
	}
	glog.V(4).Infof("Stored pod %s events in file %s", pod.Name, outputFilePath)
	fmt.Fprintln(f, string(jsonStruct))
}

func main() {
	defer glog.Flush()

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
		client:    client,
		namespace: *namespace,
		outputDir: filepath.Join(*artifacts, "sriov"),
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

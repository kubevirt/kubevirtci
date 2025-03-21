package k8scomponents

import (
	"fmt"
	"strings"
	"time"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type componentName string

const (
	componentKubeAPIServer     componentName = "kube-apiserver"
	componentKubeControllerMgr componentName = "kube-controller-manager"
	componentKubeScheduler     componentName = "kube-scheduler"

	searchFeatureGatesInFileFormat                    = "awk '/feature-gates/' %s"
	getComponentCommandFormat                         = "kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system -l component=%s -o jsonpath='{.items[0].spec.containers[*].command}'"
	getComponentReadyContainersFormat                 = "kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system -l component=%s -o=jsonpath='{.items[0].status.containerStatuses[*].ready}'"
	getNodeReadyStatusCommand                         = "kubectl --kubeconfig=/etc/kubernetes/admin.confget nodes -o=jsonpath='{.items[*].status.conditions[?(@.type==\"Ready\")].status}'"
	addFlagsToComponentCommandFormat                  = `sudo sed -i '/- %s/a\    - %s' /etc/kubernetes/manifests/%s.yaml`
	searchComponentsFilesCommand                      = "ls /etc/kubernetes/manifests"
	addFeatureGatesFieldToKubeletConfigCommand        = "sudo echo -e 'featureGates:' >> /var/lib/kubelet/config.yaml"
	addFeatureGatesToKubeletConfigCommandFormatFormat = `sudo sed -i 's/featureGates:/featureGates:\n  %s/g' /var/lib/kubelet/config.yaml`
	kubeletRestartCommand                             = "sudo systemctl restart kubelet"

	featureGateExistInKubeletError                = "feature gates should not exist in kubelet by default"
	featureGateExistInComponentCommandErrorFormat = "feature gates should not exist in %v command by default"
)

type component interface {
	verifyComponent() error
	prepareCommandsForConfiguration() error
	runCommandsToConfigure() error
	waitForComponentReady() error
	requiredOnlyForMaster() bool
}

func componentReady(component componentName, sshClient libssh.Client, waitingForFeatureGate bool, waitingForRuntimeConfig bool) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	timeoutChannel := time.After(5 * time.Minute)

	select {
	case <-timeoutChannel:
		return fmt.Errorf("timed out after 5 minutes waiting for %v to be ready", component)
	case <-ticker.C:
		ready, reason := componentReadyHelper(component, sshClient, waitingForFeatureGate, waitingForRuntimeConfig)
		if ready {
			return nil
		}
		fmt.Printf(reason)
	}

	return nil
}

func componentReadyHelper(component componentName, sshClient libssh.Client, waitingForFeatureGate bool, waitingForRuntimeConfig bool) (bool, string) {
	if waitingForFeatureGate || waitingForRuntimeConfig {
		output, err := sshClient.CommandWithNoStdOut(getComponentCommand(component))
		if err != nil {
			return false, fmt.Sprintf("modifying flags, waiting for apiserver to respord after %v restart err:%v\n", component, err)
		}

		if waitingForFeatureGate && !strings.Contains(output, "feature-gate") {
			return false, fmt.Sprintf("modifying flags, waiting for %v pods to restart\n", component)
		}

		if waitingForRuntimeConfig && !strings.Contains(output, "runtime-config") {
			return false, fmt.Sprintf("modifying flags, waiting for %v pods to restart\n", component)
		}
	}

	output, err := sshClient.CommandWithNoStdOut(getComponentReadyContainers(component))
	if err != nil {
		return false, fmt.Sprintf("modifying flags, waiting for apiserver to respord after %v restart err:%v\n", component, err)
	}
	if strings.Contains(output, "false") {
		return false, fmt.Sprintf("modifying flags, waiting for %v pods to be ready\n", component)
	}

	return true, ""
}

func runCommands(commands []string, sshClient libssh.Client) error {
	for _, cmd := range commands {
		err := sshClient.Command(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func verifyComponentCommandHelper(sshClient libssh.Client, c componentName) error {
	output, err := sshClient.CommandWithNoStdOut(searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", c)))
	if err != nil {
		return err
	}
	fgsExist := len(output) > 0
	if fgsExist {
		return fmt.Errorf(fmt.Sprintf(featureGateExistInComponentCommandErrorFormat, c))
	}
	return nil
}

func searchFeatureGatesInFile(file string) string {
	return fmt.Sprintf(searchFeatureGatesInFileFormat, file)
}

func getComponentCommand(component componentName) string {
	return fmt.Sprintf(getComponentCommandFormat, component)
}

func getComponentReadyContainers(component componentName) string {
	return fmt.Sprintf(getComponentReadyContainersFormat, component)
}

func addFlagsToComponentCommand(component componentName, flag string) string {
	return fmt.Sprintf(addFlagsToComponentCommandFormat, component, flag, component)
}

func addFeatureGatesToKubeletConfigCommand(feature string) string {
	return fmt.Sprintf(addFeatureGatesToKubeletConfigCommandFormatFormat, feature)
}

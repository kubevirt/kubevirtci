package k8scomponents

import (
	"fmt"
	"strings"
	"time"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type componentName string

const (
	getNodeReadyStatusCommand    = "kubectl --kubeconfig=/etc/kubernetes/admin.conf get nodes -o=jsonpath='{.items[*].status.conditions[?(@.type==\"Ready\")].status}'"
	searchComponentsFilesCommand = "ls /etc/kubernetes/manifests"

	featureGateExistInComponentCommandErrorFormat = "feature gates should not exist in %v command by default"
)

type component interface {
	validateComponent() error
	configureComponent() error
	waitForComponentReady() error
	requiredOnlyForMaster() bool
}

func waitUntilReady(timeout time.Duration, interval time.Duration, checkFunc func() (bool, string)) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	timeoutChannel := time.After(timeout)

	for {
		select {
		case <-timeoutChannel:
			return fmt.Errorf("timed out after %v", timeout)
		case <-ticker.C:
			ready, reason := checkFunc()
			if ready {
				return nil
			}
			if reason != "" {
				fmt.Println(reason)
			}
		}
	}
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

func validateComponentCommandHelper(sshClient libssh.Client, c componentName) error {
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
	searchFeatureGatesInFileFormat := "awk '/feature-gates/' %s"
	return fmt.Sprintf(searchFeatureGatesInFileFormat, file)
}

func getComponentCommand(component componentName) string {
	getComponentCommandFormat := "kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system -l component=%s -o jsonpath='{.items[0].spec.containers[*].command}'"
	return fmt.Sprintf(getComponentCommandFormat, component)
}

func getComponentReadyContainers(component componentName) string {
	getComponentReadyContainersFormat := "kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system -l component=%s -o=jsonpath='{.items[0].status.containerStatuses[*].ready}'"
	return fmt.Sprintf(getComponentReadyContainersFormat, component)
}

func addFlagsToComponentCommand(component componentName, flag string) string {
	addFlagsToComponentCommandFormat := `sudo sed -i '/- %s/a\    - %s' /etc/kubernetes/manifests/%s.yaml`
	return fmt.Sprintf(addFlagsToComponentCommandFormat, component, flag, component)
}

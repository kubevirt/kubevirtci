package featuregate

import (
	"fmt"
	"strings"
	"time"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type featureGateOpt struct {
	sshClient    libssh.Client
	featureGates []string
}

type component string

const (
	componentKubeAPIServer     component = "kube-apiserver"
	componentKubeControllerMgr component = "kube-controller-manager"
	componentKubeScheduler     component = "kube-scheduler"

	searchFeatureGatesInFileFormat                    = "awk '/feature-gates/' %s"
	getComponentCommandFormat                         = "kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system -l component=%s -o jsonpath='{.items[0].spec.containers[*].command}'"
	getComponentReadyContainersFormat                 = "kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system -l component=%s -o=jsonpath='{.items[0].status.containerStatuses[*].ready}'"
	addFeatureGatesToComponentCommandFormat           = `sudo sed -i 's/- %s/- %s\n    - %s/g' /etc/kubernetes/manifests/%s.yaml`
	searchComponentsFilesCommand                      = "ls /etc/kubernetes/manifests"
	addFeatureGatesFieldToKubeletConfigCommand        = "sudo /bin/su -c \"echo -e 'featureGates:' >> /var/lib/kubelet/config.yaml\""
	addFeatureGatesToKubeletConfigCommandFormatFormat = `sudo sed -i 's/featureGates:/featureGates:\n  %s: true/g' /var/lib/kubelet/config.yaml`
	kubeletRestartCommand                             = "sudo systemctl restart kubelet"

	featureGateExistInKubeletError                = "feature gates should not exist in kubelet by default"
	featureGateExistInComponentCommandErrorFormat = "feature gates should not exist in %v command by default"
)

func NewFeatureGatesOpt(sc libssh.Client, featureGates []string) *featureGateOpt {
	return &featureGateOpt{
		sshClient:    sc,
		featureGates: featureGates,
	}
}

func (o *featureGateOpt) Exec() error {
	output, err := o.sshClient.CommandWithNoStdOut(searchFeatureGatesInFile("/var/lib/kubelet/config.yaml"))
	if err != nil {
		return err
	}
	fgsExist := len(output) > 0
	if fgsExist {
		return fmt.Errorf(featureGateExistInKubeletError)
	}

	err = o.sshClient.Command(addFeatureGatesFieldToKubeletConfigCommand)
	if err != nil {
		return err
	}

	for _, fg := range o.featureGates {
		err = o.sshClient.Command(addFeatureGatesToKubeletConfigCommand(fg))
		if err != nil {
			return err
		}
	}

	err = o.sshClient.Command(kubeletRestartCommand)
	if err != nil {
		return err
	}

	output, err = o.sshClient.CommandWithNoStdOut(searchComponentsFilesCommand)
	if err != nil {
		return err
	}
	onMasterNode := len(output) > 0
	if onMasterNode { //should modify components files only on master node
		for _, component := range []component{componentKubeAPIServer, componentKubeControllerMgr, componentKubeScheduler} {
			output, err = o.sshClient.CommandWithNoStdOut(searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", component)))
			if err != nil {
				return err
			}
			fgsExist = len(output) > 0
			if fgsExist {
				return fmt.Errorf(fmt.Sprintf(featureGateExistInComponentCommandErrorFormat, component))
			}
		}

		for _, component := range []component{componentKubeAPIServer, componentKubeControllerMgr, componentKubeScheduler} {
			err = o.sshClient.Command(addFeatureGatesToComponentCommand(component, o.featureGates))
			if err != nil {
				return err
			}
			for {
				ready, reason := o.componentReady(component)
				if !ready {
					fmt.Printf(reason)
					time.Sleep(5 * time.Second)
					continue
				}
				break
			}
		}
	}
	return nil
}

func (o *featureGateOpt) componentReady(component component) (bool, string) {
	output, err := o.sshClient.CommandWithNoStdOut(getComponentCommand(component))
	if err != nil {
		return false, fmt.Sprintf("enabling feature gates, waiting for apiserver to respord after %v restart err:%v\n", component, err)
	}
	if !strings.Contains(output, "feature-gate") {
		return false, fmt.Sprintf("enabling feature gates, waiting for %v pods to restart\n", component)
	}

	output, err = o.sshClient.CommandWithNoStdOut(getComponentReadyContainers(component))
	if err != nil {
		return false, fmt.Sprintf("enabling feature gates, waiting for apiserver to respord after %v restart err:%v\n", component, err)
	}
	if strings.Contains(output, "false") {
		return false, fmt.Sprintf("enabling feature gates, waiting for %v pods to be ready\n", component)
	}

	return true, ""
}

func searchFeatureGatesInFile(file string) string {
	return fmt.Sprintf(searchFeatureGatesInFileFormat, file)
}

func getComponentCommand(component component) string {
	return fmt.Sprintf(getComponentCommandFormat, component)
}

func getComponentReadyContainers(component component) string {
	return fmt.Sprintf(getComponentReadyContainersFormat, component)
}

func addFeatureGatesToComponentCommand(component component, featureGates []string) string {
	var formattedFeatures []string
	for _, feature := range featureGates {
		formattedFeatures = append(formattedFeatures, fmt.Sprintf("%s=true", feature))
	}
	result := strings.Join(formattedFeatures, ",")
	finalResult := fmt.Sprintf("--feature-gates=%s", result)
	return fmt.Sprintf(addFeatureGatesToComponentCommandFormat, component, component, finalResult, component)
}

func addFeatureGatesToKubeletConfigCommand(feature string) string {
	return fmt.Sprintf(addFeatureGatesToKubeletConfigCommandFormatFormat, feature)
}

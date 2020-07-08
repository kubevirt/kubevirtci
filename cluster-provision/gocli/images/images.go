package images

var (
	K8S118SHA     = ""
	K8S117SHA     = ""
	K8S116SHA     = ""
	K8S115SHA     = ""
	K8S114SHA     = ""
	SHAByProvider map[string]string
)

func init() {
	SHAByProvider = map[string]string{
		"k8s-1.18": K8S118SHA,
		"k8s-1.17": K8S117SHA,
		"k8s-1.16": K8S116SHA,
		"k8s-1.15": K8S115SHA,
		"k8s-1.14": K8S114SHA,
	}
}

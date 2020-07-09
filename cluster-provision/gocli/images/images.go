package images

var (
	K8S118SUFFIX     = ""
	K8S117SUFFIX     = ""
	K8S116SUFFIX     = ""
	K8S115SUFFIX     = ""
	K8S114SUFFIX     = ""
	SuffixByProvider map[string]string
)

func init() {
	SuffixByProvider = map[string]string{
		"k8s-1.18": K8S118SUFFIX,
		"k8s-1.17": K8S117SUFFIX,
		"k8s-1.16": K8S116SUFFIX,
		"k8s-1.15": K8S115SUFFIX,
		"k8s-1.14": K8S114SUFFIX,
	}
}

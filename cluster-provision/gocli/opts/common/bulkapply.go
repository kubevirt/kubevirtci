package common

import (
	"bytes"
	"unicode"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

const documentSeparator = "---\n"

func ApplyYAML(multiDocYAML []byte, client k8s.K8sDynamicClient) error {
	yamlDocs := bytes.Split(multiDocYAML, []byte(documentSeparator))
	for _, yamlDoc := range yamlDocs {
		if len(yamlDoc) == 0 {
			continue
		}

		var obj *unstructured.Unstructured
		var err error
		if obj, err = k8s.SerializeIntoObject(yamlDoc); err != nil {
			if beginsInComment(yamlDoc) {
				continue
			}
			return err
		}

		if err = client.Apply(obj); err != nil {
			return err
		}
	}
	return nil
}

func beginsInComment(doc []byte) bool {
	const commentRune = '#'
	
	for i := 0; i < len(doc); i++ {
		if !unicode.IsSpace(rune(doc[i])) {
			return doc[i] == commentRune
		}
	}
	return false
}

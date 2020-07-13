package main

import (
	"bufio"
	"bytes"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"log"
	"os"
	"path/filepath"
)

var deserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <manifest-file|manifest-dir>", os.Args[0])
	}
	manifestFileOrDirectory := os.Args[1]
	fileInfo, err := os.Stat(manifestFileOrDirectory)
	if os.IsNotExist(err) {
		log.Fatalf("Failed to open %s: %v", manifestFileOrDirectory, err)
	}
	filesWithOffendingPullPolicies := map[string]map[string]corev1.PullPolicy{}
	if fileInfo.IsDir() {
		err = filepath.Walk(manifestFileOrDirectory, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			checkFileForOffendingPullPolicies(path, filesWithOffendingPullPolicies)
			return nil
		})
		if err != nil {
			log.Fatalf("Error on walking path %s: %v", manifestFileOrDirectory, err)
		}
	} else {
		checkFileForOffendingPullPolicies(manifestFileOrDirectory, filesWithOffendingPullPolicies)
	}
	fmt.Printf("%d manifest files with offending pull policies detected\n", len(filesWithOffendingPullPolicies))
	if len(filesWithOffendingPullPolicies) > 0 {
		for filePath, offendingPullPolicies := range filesWithOffendingPullPolicies {
			fmt.Printf("File: %s\n", filePath)
			for image, pullPolicy := range offendingPullPolicies {
				fmt.Printf("\tImage: %s\n", image)
				if pullPolicy == "" {
					pullPolicy = corev1.PullAlways+" (implicit)"
				}
				fmt.Printf("\t\tPullPolicy: %s\n", pullPolicy)
			}
		}
		os.Exit(1)
	}
}

func checkFileForOffendingPullPolicies(manifestFile string, filesWithOffendingPullPolicies map[string]map[string]corev1.PullPolicy) {
	file, err := os.Open(manifestFile)
	if err != nil {
		log.Fatalf("Error on opening file %s: %v", manifestFile, err)
	}
	//noinspection GoUnhandledErrorResult
	defer file.Close()

	offendingPullPolicies := map[string]corev1.PullPolicy{}
	scanner := bufio.NewScanner(file)
	var bufferString *bytes.Buffer
	for scanner.Scan() {
		if scanner.Text() == "---" {
			if checkHasOffendingPullPolicies(bufferString, offendingPullPolicies) {
				filesWithOffendingPullPolicies[manifestFile] = offendingPullPolicies
			}
		} else {
			if bufferString == nil {
				bufferString = bytes.NewBufferString(scanner.Text())
			} else {
				bufferString.WriteString("\n" + scanner.Text())
			}
		}
	}
	if checkHasOffendingPullPolicies(bufferString, offendingPullPolicies) {
		filesWithOffendingPullPolicies[manifestFile] = offendingPullPolicies
	}
	return
}

func checkHasOffendingPullPolicies(bufferString *bytes.Buffer, offendingPullPolicies map[string]corev1.PullPolicy) (hasOffendingPullPolicies bool) {
	if bufferString == nil {
		return false
	}
	object, err := deserializeManifestAsObject(bufferString)
	if err != nil {
		log.Fatalf("Failed to deserialize buffer %v: %v", bufferString, err)
	}
	kind := object.GetObjectKind().GroupVersionKind().Kind
	switch kind {
	case "Deployment":
		object, err := deserializeManifestAs(bufferString, &appsv1.Deployment{})
		if err != nil {
			log.Fatalf("Failed to deserialize buffer %v: %v", bufferString, err)
		}
		return appendOffendingPullPoliciesForDeployment(object.(*appsv1.Deployment), offendingPullPolicies)
	case "StatefulSet":
		object, err := deserializeManifestAs(bufferString, &appsv1.StatefulSet{})
		if err != nil {
			log.Fatalf("Failed to deserialize buffer %v: %v", bufferString, err)
		}
		return appendOffendingPullPoliciesForStatefulSet(object.(*appsv1.StatefulSet), offendingPullPolicies)
	case "ReplicaSet":
		object, err := deserializeManifestAs(bufferString, &appsv1.ReplicaSet{})
		if err != nil {
			log.Fatalf("Failed to deserialize buffer %v: %v", bufferString, err)
		}
		return appendOffendingPullPoliciesForReplicaSet(object.(*appsv1.ReplicaSet), offendingPullPolicies)
	case "DaemonSet":
		object, err := deserializeManifestAs(bufferString, &appsv1.DaemonSet{})
		if err != nil {
			log.Fatalf("Failed to deserialize buffer %v: %v", bufferString, err)
		}
		return appendOffendingPullPoliciesForDaemonSet(object.(*appsv1.DaemonSet), offendingPullPolicies)
	default:
		return false
	}
}

func deserializeManifestAsObject(bufferString *bytes.Buffer) (object runtime.Object, err error) {
	object, _, err = deserializer.Decode(bufferString.Bytes(), nil, &appsv1.Deployment{})
	return
}

func deserializeManifestAs(bufferString *bytes.Buffer, what runtime.Object) (object runtime.Object, err error) {
	object, _, err = deserializer.Decode(bufferString.Bytes(), nil, what)
	return
}

func appendOffendingPullPoliciesForDeployment(deployment *appsv1.Deployment, offendingPullPolicies map[string]corev1.PullPolicy) (hasOffendingPullPolicies bool) {
	if appendOffendingPullPolicies(deployment.Spec.Template.Spec.Containers, offendingPullPolicies) {
		hasOffendingPullPolicies = true
	}
	if appendOffendingPullPolicies(deployment.Spec.Template.Spec.InitContainers, offendingPullPolicies) {
		hasOffendingPullPolicies = true
	}
	return
}

func appendOffendingPullPoliciesForStatefulSet(statefulSet *appsv1.StatefulSet, offendingPullPolicies map[string]corev1.PullPolicy) (hasOffendingPullPolicies bool) {
	if appendOffendingPullPolicies(statefulSet.Spec.Template.Spec.Containers, offendingPullPolicies) {
		hasOffendingPullPolicies = true
	}
	if appendOffendingPullPolicies(statefulSet.Spec.Template.Spec.InitContainers, offendingPullPolicies) {
		hasOffendingPullPolicies = true
	}
	return
}

func appendOffendingPullPoliciesForDaemonSet(daemonSet *appsv1.DaemonSet, offendingPullPolicies map[string]corev1.PullPolicy) (hasOffendingPullPolicies bool) {
	if appendOffendingPullPolicies(daemonSet.Spec.Template.Spec.Containers, offendingPullPolicies) {
		hasOffendingPullPolicies = true
	}
	if appendOffendingPullPolicies(daemonSet.Spec.Template.Spec.InitContainers, offendingPullPolicies) {
		hasOffendingPullPolicies = true
	}
	return
}

func appendOffendingPullPoliciesForReplicaSet(replicaSet *appsv1.ReplicaSet, offendingPullPolicies map[string]corev1.PullPolicy) (hasOffendingPullPolicies bool) {
	if appendOffendingPullPolicies(replicaSet.Spec.Template.Spec.Containers, offendingPullPolicies) {
		hasOffendingPullPolicies = true
	}
	if appendOffendingPullPolicies(replicaSet.Spec.Template.Spec.InitContainers, offendingPullPolicies) {
		hasOffendingPullPolicies = true
	}
	return
}

func appendOffendingPullPolicies(containers []corev1.Container, offendingPullPolicies map[string]corev1.PullPolicy) (hasOffendingPullPolicies bool) {
	for _, container := range containers {
		if container.ImagePullPolicy == corev1.PullAlways || container.ImagePullPolicy == "" {
			offendingPullPolicies[container.Image] = container.ImagePullPolicy
			hasOffendingPullPolicies = true
		}
	}
	return
}

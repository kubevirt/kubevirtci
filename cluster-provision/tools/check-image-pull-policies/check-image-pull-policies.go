package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type options struct {
	manifestSource string
	verbose        bool
	dryRun         bool
}

func flagOptions() options {
	o := options{}
	flag.StringVar(&o.manifestSource, "manifest-source", "", "The directory with manifest files or the manifest file to check. If a directory is given, all files are tried to parse as manifest")
	flag.BoolVar(&o.dryRun, "dry-run", true, "Whether to exit with a non-zero exit code if the check fails")
	flag.BoolVar(&o.verbose, "verbose", false, "Whether to output all parsed information or only the information on where the checks failed")
	flag.Parse()
	return o
}

var deserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()

func main() {
	options := flagOptions()

	if options.manifestSource == "" {
		log.Fatal("No manifest-source given!")
	}

	fileInfo, err := os.Stat(options.manifestSource)
	if os.IsNotExist(err) {
		log.Fatalf("Failed to open %s: %v", options.manifestSource, err)
	}

	filesWithPullPolicies := map[string]map[string]corev1.PullPolicy{}
	if fileInfo.IsDir() {
		err = filepath.Walk(options.manifestSource, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			checkFileForPullPolicies(path, filesWithPullPolicies)
			return nil
		})
		if err != nil {
			log.Fatalf("Error on walking path %s: %v", options.manifestSource, err)
		}
	} else {
		checkFileForPullPolicies(options.manifestSource, filesWithPullPolicies)
	}

	fmt.Printf("%d files with pull policies detected\n", len(filesWithPullPolicies))
	hasOffendingPolicies := false
	for filePath, pullPolicies := range filesWithPullPolicies {
		for image, pullPolicy := range pullPolicies {
			imageParts := strings.Split(image, ":")
			var imageTag string
			if len(imageParts) > 0 {
				imageTag = imageParts[1]
			}
			offending := pullPolicy == corev1.PullAlways ||
				(pullPolicy == "" && (imageTag == "" || imageTag == "latest"))
			if offending {
				hasOffendingPolicies = true
			}
			if !offending && !options.verbose {
				continue
			}
			fmt.Printf("File: %s\n", filePath)
			fmt.Printf("\tImage: %s\n", image)
			// https://kubernetes.io/docs/concepts/containers/images/#updating-images
			if offending {
				fmt.Printf("\t\t-> PullPolicy: %s\n", pullPolicy)
			} else {
				fmt.Printf("\t\t   PullPolicy: %s\n", pullPolicy)
			}
		}
	}
	if hasOffendingPolicies {
		if options.dryRun {
			fmt.Println("WARNING: detected pull policies that will always pull images!")
		} else {
			log.Fatal("ERROR: detected pull policies that will always pull images!")
		}
	}
}

func checkFileForPullPolicies(manifestFile string, filesWithPullPolicies map[string]map[string]corev1.PullPolicy) {
	file, err := os.Open(manifestFile)
	if err != nil {
		log.Fatalf("Error on opening file %s: %v", manifestFile, err)
	}
	//noinspection GoUnhandledErrorResult
	defer file.Close()

	pullPolicies := map[string]corev1.PullPolicy{}
	scanner := bufio.NewScanner(file)
	var bufferString *bytes.Buffer
	for scanner.Scan() {
		if scanner.Text() == "---" {
			appendPullPoliciesFromManifest(bufferString, pullPolicies)
		} else {
			if bufferString == nil {
				bufferString = bytes.NewBufferString(scanner.Text())
			} else {
				bufferString.WriteString("\n" + scanner.Text())
			}
		}
	}
	appendPullPoliciesFromManifest(bufferString, pullPolicies)
	if len(pullPolicies) > 0 {
		filesWithPullPolicies[manifestFile] = pullPolicies
	}
}

func appendPullPoliciesFromManifest(bufferString *bytes.Buffer, pullPolicies map[string]corev1.PullPolicy) {
	if bufferString == nil {
		return
	}
	object, err := deserializeManifestAsObject(bufferString)
	if err != nil {
		log.Fatalf("Failed to deserialize buffer %v: %v", bufferString, err)
	}
	kind := object.GetObjectKind().GroupVersionKind().Kind
	switch kind {
	case "Deployment":
		deployment := deserializeManifestAs(bufferString, &appsv1.Deployment{}).(*appsv1.Deployment)
		appendPullPoliciesFromContainers(deployment.Spec.Template.Spec.Containers, deployment.Spec.Template.Spec.InitContainers, pullPolicies)
	case "StatefulSet":
		statefulSet := deserializeManifestAs(bufferString, &appsv1.StatefulSet{}).(*appsv1.StatefulSet)
		appendPullPoliciesFromContainers(statefulSet.Spec.Template.Spec.Containers, statefulSet.Spec.Template.Spec.InitContainers, pullPolicies)
	case "ReplicaSet":
		replicaSet := deserializeManifestAs(bufferString, &appsv1.ReplicaSet{}).(*appsv1.ReplicaSet)
		appendPullPoliciesFromContainers(replicaSet.Spec.Template.Spec.Containers, replicaSet.Spec.Template.Spec.InitContainers, pullPolicies)
	case "DaemonSet":
		daemonSet := deserializeManifestAs(bufferString, &appsv1.DaemonSet{}).(*appsv1.DaemonSet)
		appendPullPoliciesFromContainers(daemonSet.Spec.Template.Spec.Containers, daemonSet.Spec.Template.Spec.InitContainers, pullPolicies)
	case "Pod":
		pod := deserializeManifestAs(bufferString, &corev1.Pod{}).(*corev1.Pod)
		appendPullPoliciesFromContainers(pod.Spec.Containers, pod.Spec.InitContainers, pullPolicies)
	default:
		break
	}
}

func deserializeManifestAsObject(bufferString *bytes.Buffer) (object runtime.Object, err error) {
	object, _, err = deserializer.Decode(bufferString.Bytes(), nil, &appsv1.Deployment{})
	return
}

func deserializeManifestAs(bufferString *bytes.Buffer, what runtime.Object) (object runtime.Object) {
	object, _, err := deserializer.Decode(bufferString.Bytes(), nil, what)
	if err != nil {
		log.Fatalf("Failed to deserialize buffer %v: %v", bufferString, err)
	}
	return
}

func appendPullPoliciesFromContainers(containers []corev1.Container, initContainers []corev1.Container, pullPolicies map[string]corev1.PullPolicy) {
	allContainers := []corev1.Container{}
	allContainers = append(allContainers, containers...)
	allContainers = append(allContainers, initContainers...)
	for _, container := range allContainers {
		pullPolicies[container.Image] = container.ImagePullPolicy
	}
}

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

	filesWithPullPolicies := map[string]map[string]corev1.PullPolicy{}
	err := walkFiles(options, filesWithPullPolicies)
	if err != nil {
		log.Fatalf("Failed to open %s: %v", options.manifestSource, err)
	}

	bufferString := bytes.NewBufferString("")
	hasOffendingPolicies := writeCheckResultToBuffer(options, filesWithPullPolicies, bufferString)
	fmt.Print(bufferString.String())
	if hasOffendingPolicies && !options.dryRun {
		os.Exit(1)
	}
}

func walkFiles(options options, filesWithPullPolicies map[string]map[string]corev1.PullPolicy) (err error) {
	fileInfo, err := os.Stat(options.manifestSource)
	if os.IsNotExist(err) {
		return err
	}

	if fileInfo.IsDir() {
		err = filepath.Walk(options.manifestSource, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			checkFileForPullPolicies(path, filesWithPullPolicies)
			return nil
		})
		return err
	} else {
		checkFileForPullPolicies(options.manifestSource, filesWithPullPolicies)
		return nil
	}
}

func writeCheckResultToBuffer(options options, filesWithPullPolicies map[string]map[string]corev1.PullPolicy, bufferString *bytes.Buffer) (hasOffendingPolicies bool) {
	bufferString.WriteString(fmt.Sprintf("%d files with pull policies detected\n", len(filesWithPullPolicies)))
	hasOffendingPolicies = false
	for filePath, pullPolicies := range filesWithPullPolicies {
		for image, pullPolicy := range pullPolicies {
			imageParts := strings.Split(image, ":")
			var imageTag string
			if len(imageParts) > 0 {
				imageTag = imageParts[1]
			}
			offending := isEffectivelyPullAlways(pullPolicy, imageTag)
			if offending {
				hasOffendingPolicies = true
			}
			if !offending && !options.verbose {
				continue
			}
			bufferString.WriteString(fmt.Sprintf("File: %s\n", filePath))
			bufferString.WriteString(fmt.Sprintf("\tImage: %s\n", image))
			if offending {
				bufferString.WriteString(fmt.Sprintf("\t\t-> PullPolicy: %s\n", pullPolicy))
			} else {
				bufferString.WriteString(fmt.Sprintf("\t\t   PullPolicy: %s\n", pullPolicy))
			}
		}
	}
	if hasOffendingPolicies {
		if options.dryRun {
			bufferString.WriteString(fmt.Sprintf("WARNING: detected pull policies that will always pull images!\n"))
		} else {
			bufferString.WriteString(fmt.Sprintf("ERROR: detected pull policies that will always pull images!\n"))
		}
	}
	return hasOffendingPolicies
}

// isEffectivelyPullAlways checks by looking at the combination of pullPolicy and imageTag
// whether this will lead to effectively always pulling the image, as described in
// https://kubernetes.io/docs/concepts/containers/images/#updating-images
func isEffectivelyPullAlways(pullPolicy corev1.PullPolicy, imageTag string) bool {
	offending := pullPolicy == corev1.PullAlways ||
		(pullPolicy == "" && (imageTag == "" || imageTag == "latest"))
	return offending
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

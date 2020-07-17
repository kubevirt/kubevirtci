package main

import (
	"bytes"
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"testing"
)

func TestCheckFileForPullPolicies(t *testing.T) {

	testCases := map[string]map[string]corev1.PullPolicy{
		"testdata/cni.yaml": map[string]corev1.PullPolicy{
			"calico/cni:v3.12.0":                "",
			"calico/kube-controllers:v3.12.0":   "",
			"calico/node:v3.12.0":               "",
			"calico/pod2daemon-flexvol:v3.12.0": "",
		},
		"testdata/cdi/cdi-operator.yaml": map[string]corev1.PullPolicy{
			"kubevirt/cdi-operator:v1.18.0": corev1.PullIfNotPresent,
		},
		"testdata/cnao/operator.yaml": map[string]corev1.PullPolicy{
			"quay.io/kubevirt/cluster-network-addons-operator:0.35.0": corev1.PullAlways,
		},
	}

	for testCaseFile, expectedData := range testCases {
		filesWithPullPolicies := map[string]map[string]corev1.PullPolicy{}
		checkFileForPullPolicies(testCaseFile, filesWithPullPolicies)

		if filesWithPullPolicies[testCaseFile] == nil {
			t.Errorf("Expected pull policies for %s", testCaseFile)
		}
		if !reflect.DeepEqual(expectedData, filesWithPullPolicies[testCaseFile]) {
			t.Errorf("Expected differs from actual for file %s: %v != %v", testCaseFile, expectedData, filesWithPullPolicies[testCaseFile])
		}
	}
}

type WriteCheckResultToBufferTestData struct {
	options               options
	filesWithPullPolicies map[string]map[string]corev1.PullPolicy
	expectedBufferContent string
}

func TestWriteCheckResultToBuffer(t *testing.T) {

	testCases := []WriteCheckResultToBufferTestData{
		{
			options: options{verbose: true, dryRun: true},
			filesWithPullPolicies: map[string]map[string]corev1.PullPolicy{
				"testdata/cdi/cdi-operator.yaml": {
					"kubevirt/cdi-operator:v1.18.0": corev1.PullIfNotPresent,
				},
			},
			expectedBufferContent:
			`1 files with pull policies detected
File: testdata/cdi/cdi-operator.yaml
	Image: kubevirt/cdi-operator:v1.18.0
		   PullPolicy: IfNotPresent
`,
		},
		{
			options: options{verbose: false, dryRun: true},
			filesWithPullPolicies: map[string]map[string]corev1.PullPolicy{
				"testdata/cdi/cdi-operator.yaml": {
					"kubevirt/cdi-operator:v1.18.0": corev1.PullIfNotPresent,
				},
			},
			expectedBufferContent:
			`1 files with pull policies detected
`,
		},
		{
			options: options{verbose: true, dryRun: true},
			filesWithPullPolicies: map[string]map[string]corev1.PullPolicy{"testdata/cnao/operator.yaml": {
				"quay.io/kubevirt/cluster-network-addons-operator:0.35.0": corev1.PullAlways,
			},
			},
			expectedBufferContent:
			`1 files with pull policies detected
File: testdata/cnao/operator.yaml
	Image: quay.io/kubevirt/cluster-network-addons-operator:0.35.0
		-> PullPolicy: Always
WARNING: detected pull policies that will always pull images!
`,
		},
		{
			options: options{verbose: true, dryRun: false},
			filesWithPullPolicies: map[string]map[string]corev1.PullPolicy{"testdata/cnao/operator.yaml": {
				"quay.io/kubevirt/cluster-network-addons-operator:0.35.0": corev1.PullAlways,
			},
			},
			expectedBufferContent:
			`1 files with pull policies detected
File: testdata/cnao/operator.yaml
	Image: quay.io/kubevirt/cluster-network-addons-operator:0.35.0
		-> PullPolicy: Always
ERROR: detected pull policies that will always pull images!
`,
		},
	}

	for _, testData := range testCases {
		bufferString := bytes.NewBufferString("")
		writeCheckResultToBuffer(testData.options, testData.filesWithPullPolicies, bufferString)
		if testData.expectedBufferContent != bufferString.String() {
			t.Errorf("Expected differs from actual: \n'%v'\n != \n'%v'\n", testData.expectedBufferContent, bufferString.String())
		}
	}
}

type IsEffectivelyPullAlwaysTestData struct {
	pullPolicy corev1.PullPolicy
	imageTag   string
	expected   bool
}

func TestIsEffectivelyPullAlways(t *testing.T) {
	testCases := []IsEffectivelyPullAlwaysTestData{
		{
			pullPolicy: corev1.PullAlways,
			imageTag:   "whatever",
			expected:   true,
		},
		{
			pullPolicy: "",
			imageTag:   "latest",
			expected:   true,
		},
		{
			pullPolicy: "",
			imageTag:   "",
			expected:   true,
		},
		{
			pullPolicy: corev1.PullIfNotPresent,
			imageTag:   "",
			expected:   false,
		},
		{
			pullPolicy: corev1.PullIfNotPresent,
			imageTag:   "latest",
			expected:   false,
		},
		{
			pullPolicy: corev1.PullNever,
			imageTag:   "v1.0",
			expected:   false,
		},
	}

	for _, testData := range testCases {
		actual := isEffectivelyPullAlways(testData.pullPolicy, testData.imageTag)
		if testData.expected != actual {
			t.Errorf("%v, %s: Expected differs from actual: %v != %v", testData.pullPolicy, testData.imageTag, testData.expected, actual)
		}
	}
}

type WalkFilesTestData struct {
	options options
	expectedEntries int
	expectedErr bool
}

func TestWalkFiles(t *testing.T) {
	testCases := []WalkFilesTestData{
		{
			options: options{manifestSource: "testdata/", verbose: true, dryRun: true},
			expectedEntries: 5,
		},
		{
			options: options{manifestSource: "testdata/cnao/operator.yaml", verbose: true, dryRun: true},
			expectedEntries: 1,
		},
		{
			options: options{manifestSource: "testdata/whatever", verbose: true, dryRun: true},
			expectedEntries: 0,
			expectedErr: true,
		},
	}

	for _, testData := range testCases {
		filesWithPullPolicies := map[string]map[string]corev1.PullPolicy{}
		err := walkFiles(testData.options, filesWithPullPolicies)
		if testData.expectedErr && err == nil {
			t.Errorf("Expected err!")
		}
		if testData.expectedEntries != len(filesWithPullPolicies) {
			t.Errorf("Expected result to have %d entries, but was %d!", testData.expectedEntries, len(filesWithPullPolicies))
		}
	}
}

func TestFlagOptions(t *testing.T) {
	expected := options{manifestSource: "", dryRun: true, verbose: false}
	actual := flagOptions()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected options to be %v, but was %v", expected, actual)
	}
}

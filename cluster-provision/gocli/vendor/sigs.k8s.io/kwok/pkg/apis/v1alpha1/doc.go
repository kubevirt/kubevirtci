/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// +k8s:deepcopy-gen=package
// +k8s:defaulter-gen=TypeMeta
// +groupName=kwok.x-k8s.io

// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=nodes/status,verbs=patch;update
// +kubebuilder:rbac:groups="",resources=pods,verbs=delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups="",resources=pods/status,verbs=patch;update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=create;get;list;patch;update;watch

// Package v1alpha1 implements the v1alpha1 apiVersion of kwok's configuration
package v1alpha1

/*
Copyright 2022.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NamespaceSelectorsSpec struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

type GenerateToSpec struct {
	NamespaceSelectors []NamespaceSelectorsSpec `json:"namespaceSelectors,omitempty"`
}

// ClusterConfigMapSpec defines the desired state of ClusterConfigMap
type ClusterConfigMapSpec struct {
	Data map[string]string `json:"data,omitempty"`

	GenerateTo GenerateToSpec `json:"generateTo,omitempty"`
}

// ClusterConfigMapStatus defines the observed state of ClusterConfigMap
type ClusterConfigMapStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ClusterConfigMap is the Schema for the clusterconfigmaps API
type ClusterConfigMap struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterConfigMapSpec   `json:"spec,omitempty"`
	Status ClusterConfigMapStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterConfigMapList contains a list of ClusterConfigMap
type ClusterConfigMapList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterConfigMap `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterConfigMap{}, &ClusterConfigMapList{})
}

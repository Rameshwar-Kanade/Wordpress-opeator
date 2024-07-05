package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WordpressSpec struct {
	Image                          string            `json:"image,omitempty"`
	Replicas                       int32             `json:"replicas,omitempty"`
	ConfigData                     map[string]string `json:"configData,omitempty"`
	DBUserName                     string            `json:"dbUsername,omitempty"`
	DBPassword                     string            `json:"dbPassword,omitempty"`
	MinReplicas                    int32             `json:"minReplicas,omitempty"`
	MaxReplicas                    int32             `json:"maxReplicas,omitempty"`
	TargetCPUUtilizationPercentage int32             `json:"targetCPUUtilizationPercentage,omitempty"`
}

// WordpressStatus defines the observed state of Wordpress
type WordpressStatus struct {
	// Insert additional status field - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Wordpress is the Schema for the wordpresses API
type Wordpress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WordpressSpec   `json:"spec,omitempty"`
	Status WordpressStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WordpressList contains a list of Wordpress
type WordpressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Wordpress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Wordpress{}, &WordpressList{})
}

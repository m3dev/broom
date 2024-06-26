/*
Copyright 2024.

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

package v1alpha1

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type BroomTarget struct {
	Name      string            `json:"name,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Namespace string            `json:"namespace,omitempty"`
}

type BroomAdjustmentType string

const (
	AddAdjustment BroomAdjustmentType = "Add"
	MulAdjustment BroomAdjustmentType = "Mul"
)

type BroomAdjustment struct {
	// Adjustment type. `Add` or `Mul`.
	Type BroomAdjustmentType `json:"type"`
	// Adjustment value. For `Add` type, it is the value to be added to the current memory. For `Mul` type, it is the value to be multiplied with the current memory.
	Value string `json:"value"`
	// Maximum limit for the memory. If the memory after adjustment is greater than this value, the memory is set to this value.
	MaxLimit resource.Quantity `json:"maxLimit,omitempty"`
}

// AdjustMemory adjusts the memory based on the adjustment type, value and maxLimit.
func (adj BroomAdjustment) AdjustMemory(m *resource.Quantity) error {
	switch adj.Type {
	case AddAdjustment:
		y, err := resource.ParseQuantity(adj.Value)
		if err != nil {
			return fmt.Errorf("unable to parse value to resource.Quantity: %w", err)
		}
		m.Add(y)
	case MulAdjustment:
		y, err := strconv.Atoi(adj.Value)
		if err != nil {
			return fmt.Errorf("unable to parse value to int: %w", err)
		}
		m.Mul(int64(y))
	}

	if adj.MaxLimit.IsZero() { // maxLimit is not set
		return nil
	}
	if m.Cmp(adj.MaxLimit) == 1 { // m is greater than maxLimit
		*m = adj.MaxLimit
	}
	return nil
}

type BroomRestartPolicy string

const (
	RestartOnOOMPolicy         BroomRestartPolicy = "OnOOM"
	RestartOnSpecChangedPolicy BroomRestartPolicy = "OnSpecChanged"
	RestartNeverPolicy         BroomRestartPolicy = "Never"
)

type BroomSlackWebhookSecret struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Key       string `json:"key"`
}

type BroomSlackWebhook struct {
	Secret  BroomSlackWebhookSecret `json:"secret"`
	Channel string                  `json:"channel,omitempty"`
}

// BroomSpec defines the desired state of Broom
type BroomSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Target        BroomTarget        `json:"target,omitempty"`
	Adjustment    BroomAdjustment    `json:"adjustment"`
	RestartPolicy BroomRestartPolicy `json:"restartPolicy"`
	SlackWebhook  BroomSlackWebhook  `json:"slackWebhook"`
}

// BroomStatus defines the observed state of Broom
type BroomStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Broom is the Schema for the brooms API
type Broom struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BroomSpec   `json:"spec,omitempty"`
	Status BroomStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BroomList contains a list of Broom
type BroomList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Broom `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Broom{}, &BroomList{})
}

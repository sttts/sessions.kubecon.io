package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//
// Session is a KubeCon session.
type Session struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SessionSpec   `json:"spec"`
	Status SessionStatus `json:"status"`
}

type SessionType string

const (
	SessionTypeDeepDive SessionType = "deepdive"
	SessionTypeTalk     SessionType = "talk"
)

type SessionSpec struct {
	Type     SessionType `json:"type"`
	Capacity int         `json:"capacity"`
	Title    string      `json:"title"`
}

type SessionStatus struct {
	Attendees  int                `json:"attendees,omitempty"`
	Conditions []SessionCondition `json:"conditions,omitempty"`
}

type SessionConditionType string

const (
	SessionConditionTypeStarted SessionConditionType = "Started"
)

type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

type SessionCondition struct {
	Type               SessionConditionType `json:"type"`
	Status             ConditionStatus      `json:"status"`
	LastTransitionTime metav1.Time          `json:"lastTransitionTime,omitempty"`
	Reason             string               `json:"reason,omitempty"`
	Message            string               `json:"message,omitempty"`
}

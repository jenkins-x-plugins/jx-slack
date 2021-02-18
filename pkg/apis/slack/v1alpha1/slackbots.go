package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SlackBotList is a list of Slack Bots available
type SlackBotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []SlackBot `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// SlackBot represents a Slack Bot
type SlackBot struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec SlackBotSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// SlackBotSpec provides details of a Slack Bot
type SlackBotSpec struct {
	Namespace      string            `json:"namespace,omitempty" protobuf:"bytes,1,name=namespace"`
	TokenReference ResourceReference `json:"tokenReference,omitempty" protobuf:"bytes,5,name=tokenReference"`
	PullRequests   []SlackBotMode    `json:"pullRequests,omitempty" protobuf:"bytes,6,name=pullRequests"`
	Pipelines      []SlackBotMode    `json:"pipelines,omitempty" protobuf:"bytes,7,name=pipelines"`
	Statuses       Statuses          `json:"statuses,omitempty" protobuf:"bytes,2,name=statuses"`
}

type ResourceReference struct {
	// API version of the referent.
	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,5,opt,name=apiVersion"`
	// Kind of the referent.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	Kind string `json:"kind" protobuf:"bytes,1,opt,name=kind"`
	// Name of the referent.
	// More info: http://kubernetes.io/docs/user-guide/identifiers#names
	Name string `json:"name" protobuf:"bytes,3,opt,name=name"`
	// UID of the referent.
	// More info: http://kubernetes.io/docs/user-guide/identifiers#uids
	UID types.UID `json:"uid,omitempty" protobuf:"bytes,4,opt,name=uid,casttype=k8s.io/apimachinery/pkg/types.UID"`
}

type SlackBotMode struct {
	DirectMessage   bool     `json:"directMessage" protobuf:"bytes,1,name=directMessage"`
	NotifyReviewers bool     `json:"notifyReviewers" protobuf:"bytes,2,name=notifyReviewers"`
	Channel         string   `json:"channel" protobuf:"bytes,3,name=channel"`
	Orgs            []Org    `json:"orgs" protobuf:"bytes,4,name=orgs"`
	IgnoreLabels    []string `json:"ignoreLabels" protobuf:"bytes,5,name=ignoreLabels"`
}

type Org struct {
	Name  string   `json:"name,omitempty" protobuf:"bytes,1,name=name"`
	Repos []string `json:"repos" protobuf:"bytes,2,name=repos"`
}

type Statuses struct {
	Succeeded     *Status `json:"succeeded,omitempty" protobuf:"bytes,1,name=succeeded"`
	Failed        *Status `json:"failed,omitempty" protobuf:"bytes,2,name=failed"`
	NotApproved   *Status `json:"notApproved,omitempty" protobuf:"bytes,3,name=notApproved"`
	Approved      *Status `json:"approved,omitempty" protobuf:"bytes,4,name=approved"`
	Running       *Status `json:"running,omitempty" protobuf:"bytes,5,name=running"`
	Hold          *Status `json:"hold,omitempty" protobuf:"bytes,6,name=hold"`
	NeedsOkToTest *Status `json:"needsOkToTest,omitempty" protobuf:"bytes,7,name=needsOkToTest"`
	Merged        *Status `json:"merged,omitempty" protobuf:"bytes,8,name=merged"`
	Pending       *Status `json:"pending,omitempty" protobuf:"bytes,9,name=pending"`
	Errored       *Status `json:"errored,omitempty" protobuf:"bytes,10,name=errored"`
	Aborted       *Status `json:"aborted,omitempty" protobuf:"bytes,11,name=aborted"`
	LGTM          *Status `json:"lgtm,omitempty" protobuf:"bytes,12,name=lgtm"`
	Unknown       *Status `json:"unknown,omitempty" protobuf:"bytes,13,name=unknown"`
	Closed        *Status `json:"closed,omitempty" protobuf:"bytes,14,name=closed"` // Closed means the PR is closed but not merged
}

type Status struct {
	Emoji string `json:"emoji,omitempty" protobuf:"bytes,1,name=emoji"`
	Text  string `json:"text,omitempty" protobuf:"bytes,2,name=text"`
}

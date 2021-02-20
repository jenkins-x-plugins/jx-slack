package slackbot

import (
	"github.com/jenkins-x-plugins/slack/pkg/slacker"
	"github.com/jenkins-x/go-scm/scm"
	jenkinsv1client "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx-gitops/pkg/apis/gitops/v1alpha1"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cmdrunner"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient"
	"k8s.io/client-go/kubernetes"
)

type SlackOptions struct {
	Dir           string `env:"GIT_DIR"`
	SlackToken    string `env:"SLACK_TOKEN"`
	SlackURL      string `env:"SLACK_URL"`
	GitURL        string `env:"GIT_URL"`
	Name          string
	Namespace     string
	FakeTimestamp string
}

// SlackBotOptions contains options for the SlackBot
type SlackBotOptions struct {
	SlackOptions
	KubeClient        kubernetes.Interface
	JXClient          jenkinsv1client.Interface
	SlackClient       slacker.Interface
	ScmClient         *scm.Client
	SourceConfigs     *v1alpha1.SourceConfig
	Statuses          Statuses
	Timestamps        map[string]map[string]*MessageReference
	SlackUserResolver SlackUserResolver
	GitClient         gitclient.Interface
	CommandRunner     cmdrunner.CommandRunner
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

type MessageReference struct {
	ChannelID string
	Timestamp string
}

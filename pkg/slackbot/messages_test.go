package slackbot

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"testing"

	"github.com/jenkins-x-plugins/jx-gitops/pkg/apis/gitops/v1alpha1"
	"github.com/jenkins-x-plugins/jx-slack/pkg/slacker/fakeslack"
	"github.com/jenkins-x-plugins/jx-slack/pkg/testpipelines"
	fakescm "github.com/jenkins-x/go-scm/scm/driver/fake"
	fakejx "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/ghodss/yaml"

	"github.com/pkg/errors"

	"github.com/jenkins-x/jx-logging/v3/pkg/log"

	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	"github.com/slack-go/slack"

	"github.com/stretchr/testify/assert"
)

var (
	// generateTestOutput enable to regenerate the expected output
	generateTestOutput = false
)

func TestPipelineMessages(t *testing.T) {
	ns := "jx"
	owner := "myorg"
	repo := "myrepo"
	branch := "main"
	channel := v1alpha1.DefaultSlackChannel

	testCases := []struct {
		name             string
		kind             v1alpha1.NotifyKind
		expected         []int
		expectedMessages []string
	}{
		{
			name:     "fail-or-success",
			kind:     v1alpha1.NotifyKindFailureOrFirstSuccess,
			expected: []int{1, 1},
		},
		{
			name:     "failure",
			kind:     v1alpha1.NotifyKindFailure,
			expected: []int{1, 0},
		},
	}

	testDir := filepath.Join("test_data", "messages")

	for _, tc := range testCases {
		name := tc.name
		t.Logf("running test %s with kind %s\n", name, string(tc.kind))

		scmClient, _ := fakescm.NewDefault()
		slackClient := fakeslack.NewFakeSlack()

		sourceConfig := &v1alpha1.SourceConfig{
			Spec: v1alpha1.SourceConfigSpec{
				Groups: []v1alpha1.RepositoryGroup{
					{
						Provider: "https://fake.git",
						Owner:    owner,
						Repositories: []v1alpha1.Repository{
							{
								Name: repo,
								Slack: &v1alpha1.SlackNotify{
									Channel:  v1alpha1.DefaultSlackChannel,
									Kind:     tc.kind,
									Pipeline: v1alpha1.PipelineKindAll,
								},
							},
						},
					},
				},
			},
		}

		pa1 := testpipelines.CreateTestPipelineActivity(ns, owner, repo, branch, "release", "1", jenkinsv1.ActivityStatusTypeFailed)
		pa2 := testpipelines.CreateTestPipelineActivity(ns, owner, repo, branch, "release", "2", jenkinsv1.ActivityStatusTypeSucceeded)

		jxClient := fakejx.NewSimpleClientset(pa1, pa2)

		o := &Options{
			KubeClient:    fake.NewSimpleClientset(),
			JXClient:      jxClient,
			ScmClient:     scmClient,
			SlackClient:   slackClient,
			SourceConfigs: sourceConfig,
		}
		o.Namespace = ns
		o.MessageFormat.DashboardURL = "https://dashboard-jx.dev.jenkins-x.me"

		err := o.PipelineMessage(pa1)
		require.NoError(t, err, "failed to process pipeline %s for test %s", pa1.Name, name)

		expectedDir := filepath.Join(testDir, name)
		expectedCount := 0
		if len(tc.expected) > 0 {
			expectedCount = tc.expected[0]
		}
		slackClient.AssertMessageCount(t, channel, expectedCount, expectedDir, "pa1", generateTestOutput, "for activity "+pa1.Name+" for test "+name)
		slackClient.Messages = nil

		err = o.PipelineMessage(pa2)
		require.NoError(t, err, "failed to process pipeline %s for test %s", pa2.Name)

		expectedCount = 0
		if len(tc.expected) > 1 {
			expectedCount = tc.expected[1]
		}
		slackClient.AssertMessageCount(t, channel, expectedCount, expectedDir, "pa2", generateTestOutput, "for activity "+pa2.Name+" for test "+name)
	}
}

func TestSlackBotOptions_createAttachments(t *testing.T) {
	o := &Options{}
	type fields struct {
		filename string
	}
	tests := []struct {
		name              string
		fields            fields
		wantNumberOfSteps int
		want              []slack.Attachment
	}{
		{name: "multi_step_stage", fields: struct{ filename string }{filename: "stage_multiple_steps.yaml"}, wantNumberOfSteps: 6, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			act, err := getPipelineActivity(tt.fields.filename)
			assert.NoError(t, err, "failed to read files")

			var attachments []slack.Attachment
			for _, step := range act.Spec.Steps {
				stepAttachments := o.createAttachments(act, &step)
				if len(stepAttachments) > 0 {
					attachments = append(attachments, stepAttachments...)
				}
			}

			if tt.wantNumberOfSteps != len(attachments) {
				t.Errorf("createAttachments() number of steps = %v, want %v", len(attachments), tt.wantNumberOfSteps)
			}

			// lets print all the steps as it is nice to see in the test logs what we get
			for _, i2 := range attachments {
				log.Logger().Infof("%s", i2.Text)
			}

		})
	}
}

func getPipelineActivity(filename string) (*jenkinsv1.PipelineActivity, error) {
	testData := path.Join("test_data", "bot")
	testfile, err := ioutil.ReadFile(path.Join(testData, filename))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read file %s", path.Join(testData, filename))
	}
	act := &jenkinsv1.PipelineActivity{}
	err = yaml.Unmarshal(testfile, &act)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal testfile %s", testfile)
	}
	return act, nil
}

func TestIsUserPipelineStep(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "build", want: true},
		{name: "build something", want: true},
		{name: "setup something", want: true},
		{name: "setVersion something", want: true},
		{name: "preBuild something", want: true},
		{name: "postBuild something", want: true},
		{name: "promote something", want: true},
		{name: "pipeline something", want: true},
		{name: "Credential Initializer", want: false},
		{name: "Working Dir Initializer", want: false},
		{name: "Place Tools", want: false},
		{name: "Git Source", want: false},
		{name: "Git Merge", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUserPipelineStep(tt.name); got != tt.want {
				t.Errorf("isUserPipelineStep() %s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

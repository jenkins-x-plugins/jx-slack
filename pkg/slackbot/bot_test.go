package slackbot

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/ghodss/yaml"

	"github.com/pkg/errors"

	"github.com/jenkins-x/jx/pkg/log"

	"gotest.tools/assert"

	jenkinsv1 "github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"
	"github.com/slack-go/slack"
)

func TestSlackBotOptions_createAttachments(t *testing.T) {

	o := &SlackBotOptions{}
	type fields struct {
		filename string
	}
	tests := []struct {
		name              string
		fields            fields
		wantNumberOfSteps int
		want              []slack.Attachment
	}{
		{name: "multi_step_stage", fields: struct{ filename string }{filename: "stage_multiple_steps.yaml"}, wantNumberOfSteps: 11, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			act, err := getPipelineActivity(tt.fields.filename)
			assert.NilError(t, err, "failed to read files")

			attachments := []slack.Attachment{}
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

func Test_isUserPipelineStep(t *testing.T) {
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

package slackbot

import (
	jenkinsv1 "github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"
	jxkube "github.com/jenkins-x/jx/pkg/kube"
)

type byBuildNumber []jenkinsv1.PipelineActivity

func (s byBuildNumber) Len() int {
	return len(s)
}

func (s byBuildNumber) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byBuildNumber) getBuildNumber(activity jenkinsv1.PipelineActivity) string {
	details := jxkube.CreatePipelineDetails(&activity)
	return details.Build
}

func (s byBuildNumber) Less(i, j int) bool {
	return s.getBuildNumber(s[i]) < s.getBuildNumber(s[j])
}

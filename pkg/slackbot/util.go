package slackbot

import (
	"strings"

	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
)

type byBuildNumber []jenkinsv1.PipelineActivity

func (s byBuildNumber) Len() int {
	return len(s)
}

func (s byBuildNumber) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s *byBuildNumber) getBuildNumber(activity *jenkinsv1.PipelineActivity) string {
	details := CreatePipelineDetails(activity)
	return details.Build
}

func (s byBuildNumber) Less(i, j int) bool {
	return s.getBuildNumber(&s[i]) < s.getBuildNumber(&s[j])
}

func containsIgnoreCase(s []string, e string) bool {
	for _, a := range s {
		if strings.EqualFold(a, e) {
			return true
		}
	}
	return false
}

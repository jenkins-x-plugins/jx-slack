package slackbot

import (
	"fmt"

	jenkinsv1 "github.com/jenkins-x/jx/v2/pkg/apis/jenkins.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *GlobalClients) getPipelineActivities(org string, repo string, prn int) (*jenkinsv1.PipelineActivityList, error) {
	return c.JXClient.JenkinsV1().PipelineActivities(c.Namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("owner=%s, branch=PR-%d, repository=%s", org, prn, repo),
	})
}

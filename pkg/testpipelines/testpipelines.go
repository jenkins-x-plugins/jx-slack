package testpipelines

import (
	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/naming"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// CreateTestPipelineActivity creates a PipelineActivity with the given arguments
func CreateTestPipelineActivity(ns, owner, repo, branch, context, build string, status jenkinsv1.ActivityStatusType) *jenkinsv1.PipelineActivity {
	name := owner + "-" + repo + "-" + branch
	if context != "" {
		name += "-" + context
	}
	name += "-" + build
	name = naming.ToValidName(name)
	gitURL := "https://fake.git/" + owner + "/" + repo + ".git"

	return &jenkinsv1.PipelineActivity{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			CreationTimestamp: metav1.Time{
				Time: time.Now(),
			},
		},
		Spec: jenkinsv1.PipelineActivitySpec{
			Build:         build,
			Status:        status,
			GitURL:        gitURL,
			GitRepository: repo,
			GitOwner:      owner,
			GitBranch:     branch,
			Context:       context,
			Pipeline:      owner + "/" + name + "/" + branch,
		},
	}
}

package slackbot

import (
	"fmt"

	"github.com/nlopes/slack"

	jenkinsv1 "github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"

	jenkninsv1client "github.com/jenkins-x/jx/pkg/client/clientset/versioned"
)

// SlackUserResolver allows slack users to be converted to Jenkins X users
type SlackUserResolver struct {
	SlackClient *slack.Client
	JXClient    jenkninsv1client.Interface
	Namespace   string
}

// SlackUserLogin returns the login for the slack provider, or an empty string if not found
func (r *SlackUserResolver) SlackUserLogin(user *jenkinsv1.User) (string, error) {
	for _, a := range user.Spec.Accounts {
		if a.Provider == r.SlackProviderKey() {
			return a.ID, nil
		}
	}
	if user.Spec.Email != "" {
		// Attempt to lookup by email and associate
		slackUser, err := r.SlackClient.GetUserByEmail(user.Spec.Email)
		if err != nil {
			if err.Error() == "users_not_found" {
				// Ignore users_not_found as this just means we return an empty string
				return "", nil
			}
			return "", err
		}
		user.Spec.Accounts = append(user.Spec.Accounts, jenkinsv1.AccountReference{
			Provider: r.SlackProviderKey(),
			ID:       slackUser.ID,
		})
		_, err = r.JXClient.JenkinsV1().Users(r.Namespace).Update(user)
		return slackUser.ID, nil
	}
	return "", nil
}

// SlackProviderKey returns the provider key for this SlackUserResolver
func (r *SlackUserResolver) SlackProviderKey() string {
	return fmt.Sprintf("slack.apps.jenkins-x.com/userid")
}

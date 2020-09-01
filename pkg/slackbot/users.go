package slackbot

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jenkins-x/jx-logging/pkg/log"

	"github.com/pkg/errors"

	"github.com/slack-go/slack"

	jenkinsv1 "github.com/jenkins-x/jx/v2/pkg/apis/jenkins.io/v1"

	jenkninsv1client "github.com/jenkins-x/jx/v2/pkg/client/clientset/versioned"
)

const (
	userMappingfile = "/secrets/users/mapping.txt"
)

// SlackUserResolver allows slack users to be converted to Jenkins X users
type SlackUserResolver struct {
	SlackClient  *slack.Client
	JXClient     jenkninsv1client.Interface
	Namespace    string
	UserMappings map[string]string
}

// NewSlackUserResolver creates a new struct to work with resolving slack user details
func NewSlackUserResolver(slackClient *slack.Client, jenkinsClient jenkninsv1client.Interface, namespace string) SlackUserResolver {
	return SlackUserResolver{
		SlackClient: slackClient,
		JXClient:    jenkinsClient,
		Namespace:   namespace,
	}
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
		email, err := r.getSlackEmailFromMapping(user.Spec.Email, userMappingfile)
		if err != nil {
			// user may have the same email address in both git and slack to try that if no explicit mapping
			email = user.Spec.Email
			log.Logger().Warnf("no mapped email address so using git user email %s to find id in slack", email)
		}
		slackUser, err := r.SlackClient.GetUserByEmail(email)
		if err != nil {
			return "", errors.Wrapf(err, "could not find Slack ID using email %s", email)
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

func (r *SlackUserResolver) getSlackEmailFromMapping(gitUserEmail, fileLocation string) (string, error) {
	if gitUserEmail == "" {
		return "", errors.New("no git user email")
	}
	if fileLocation == "" {
		return "", errors.New("no user mapping file location")
	}

	if r.UserMappings == nil {
		r.UserMappings = make(map[string]string)
	}

	if len(r.UserMappings) == 0 {
		f, err := os.Open(fileLocation)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s", fileLocation)
		}
		defer func() {
			if err = f.Close(); err != nil {
				log.Logger().Errorf("failed to close file %s", fileLocation)
			}
		}()
		s := bufio.NewScanner(f)
		for s.Scan() {

			emails := strings.Split(s.Text(), ":")
			if len(emails) != 2 {
				return "", fmt.Errorf("line should contain two parts GIT_USER_EMAIL:SLACK_USER_EMAIL %s", s.Text())
			}
			if r.UserMappings[emails[0]] != "" {
				return "", fmt.Errorf("duplicate mapping found for git user email %s", emails[0])
			}
			r.UserMappings[emails[0]] = emails[1]

		}
		err = s.Err()
		if err != nil {
			return "", errors.Wrapf(err, "failed scanning lines from file %s", fileLocation)
		}
	}
	if r.UserMappings[gitUserEmail] == "" {
		return "", fmt.Errorf("no slack email found for git user email %s", gitUserEmail)
	}
	return r.UserMappings[gitUserEmail], nil
}

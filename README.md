# slack - NOT YET READY FOR USE

WARNING: This is an experimental Jenkins X Labs project, it is likely to have significant change and APIs may break while
we gather feedback and aim to get it into an alpha state.

The Slack app for Jenkins X provides integration with Jenkins X and Slack, originally authored at CloudBees by Pete Muir

## Features

* Sends a message when a build starts on any of the projects inside the cluster updating the message in real time as your pipeline progresses. Works for both pipelines and releases. Can be a DM or to a room.
* Sends a message when a Pull Request is created, CC'ing the reviewers allocated and updates the message as the PR gets approved/merged. Can be a DM or to a room. Message gets updated as PR status changes (e.g. builds passing, merged etc.)

## Future ideas
* Can be configured to send failure messages when release pipelines fail
* Integrate with traceability

## Feedback

Got any great ideas we can add to the Slack App? If so [Raise a issue here](https://github.com/jenkins-x-labs/issues)

## Install the app

`jx add app slack`

The installer will ask you to enter:

* the namespace to watch
* the channel to post to
* the token to use

## Configure the app

The app is configured using a custom resources of kind `SlackBot`. For example:

```
   kind: SlackBot
   apiVersion: slack.app.jenkins-x.io/v1alpha1
   metadata:
     name: test-slack-bot
   spec:
     pipelines:
     - directMessage: true
       orgs:
       - name: cheese
       - name: meat
     pullRequests:
     - channel: vegetables
       notifyReviewers: true
       orgs:
       - name: vegetables
         repos:
         - carrots
     - channel: brassicas
       notifyReviewers: true
       orgs:
       - name: vegetables
         repos:
         - cabbage
         - brussel_sprouts
     namespace: jx
     tokenReference:
       kind: Secret
       name: test-slack-bot-secret
     
```

Each integration is configured in a separate custom resource. To add a new integration, create
a new custom resource.

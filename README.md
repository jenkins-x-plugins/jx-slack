# slack 

The Slack app for Jenkins X provides integration with Jenkins X and Slack

This has been developed and tested so far using pipelines triggered by commits from GitHub and deployed using Jenkins X on Google Container Engine.

## Features

* Sends a message when a build starts on any of the projects inside the cluster updating the message in real time as your pipeline progresses. Works for both pipelines and releases. Can be a DM or to a room.

![](./docs/images/room.png)

* Sends a message when a Pull Request is created, CC'ing the reviewers allocated and updates the message as the PR gets approved/merged. Can be a DM or to a room. Message gets updated as PR status changes (e.g. builds passing, merged etc.)

![](./docs/images/dm.png)

## Feedback

Got any great ideas we can add to the Slack App? If so [Raise a issue here](https://github.com/jenkins-x-plugins/jx-slack/issues)

## Install the app

See the [Install Guide](https://jenkins-x.io/v3/develop/ui/slack/#creating-the-slack-app)

## Development

The slack app was developed against a cluster using Helm 3, for faster iterations you can run...
```bash
./build-deploy.sh
```

_Note_ this is just for testing as it does not integrate with Jenkins X GitOps

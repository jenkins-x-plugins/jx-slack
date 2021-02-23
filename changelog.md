### Linux

```shell
curl -L https://github.com/jenkins-x-plugins/jx-slack/releases/download/v0.0.55/jx-slack-linux-amd64.tar.gz | tar xzv 
sudo mv slack /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x-plugins/jx-slack/releases/download/v0.0.55/jx-slack-darwin-amd64.tar.gz | tar xzv
sudo mv slack /usr/local/bin
```
## Changes

### Bug Fixes

* pollish PR messages (James Strachan)
* use nicer PR message (James Strachan)
* lint the code a bit more (James Strachan)
* use nicer annotation for message ids (James Strachan)

### Chores

* fmt (James Strachan)

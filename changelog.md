### Linux

```shell
curl -L https://github.com/jenkins-x-plugins/jx-slack/releases/download/v0.0.56/jx-slack-linux-amd64.tar.gz | tar xzv 
sudo mv slack /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x-plugins/jx-slack/releases/download/v0.0.56/jx-slack-darwin-amd64.tar.gz | tar xzv
sudo mv slack /usr/local/bin
```
## Changes

### Bug Fixes

* pollish PR messages (James Strachan)

### Chores

* fix chart (James Strachan)
* upgrade deps (James Strachan)

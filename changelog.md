### Linux

```shell
curl -L https://github.com/jenkins-x-plugins/jx-slack/releases/download/v0.1.4/jx-slack-linux-amd64.tar.gz | tar xzv 
sudo mv slack /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x-plugins/jx-slack/releases/download/v0.1.4/jx-slack-darwin-amd64.tar.gz | tar xzv
sudo mv slack /usr/local/bin
```
## Changes

### Bug Fixes

* edit lint file to lint.yaml as in triggers (JordanGoasdoue)

### Chores

* replace strings.Title with cases.Title (JordanGoasdoue)
* Filter groups before getting repo form it (JordanGoasdoue)

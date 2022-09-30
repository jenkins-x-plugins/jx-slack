### Linux

```shell
curl -L https://github.com/jenkins-x-plugins/jx-slack/releases/download/v0.2.1/jx-slack-linux-amd64.tar.gz | tar xzv 
sudo mv slack /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x-plugins/jx-slack/releases/download/v0.2.1/jx-slack-darwin-amd64.tar.gz | tar xzv
sudo mv slack /usr/local/bin
```
## Changes

### Code Refactoring

* upgrade go version to 1.18.6 (ankitm123)

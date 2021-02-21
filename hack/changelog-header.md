### Linux

```shell
curl -L https://github.com/jenkins-x-plugins/slack/releases/download/v{{.Version}}/jx-slack-linux-amd64.tar.gz | tar xzv 
sudo mv slack /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x-plugins/slack/releases/download/v{{.Version}}/jx-slack-darwin-amd64.tar.gz | tar xzv
sudo mv slack /usr/local/bin
```

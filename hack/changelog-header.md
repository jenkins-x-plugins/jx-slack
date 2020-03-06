### Linux

```shell
curl -L https://github.com/jenkins-x-labs/jxl/releases/download/v{{.Version}}/slack-linux-amd64.tar.gz | tar xzv 
sudo mv slack /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x-labs/jxl/releases/download/v{{.Version}}/slack-darwin-amd64.tar.gz | tar xzv
sudo mv slack /usr/local/bin
```

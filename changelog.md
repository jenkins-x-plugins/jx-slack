### Linux

```shell
curl -L https://github.com/jenkins-x-labs/jxl/releases/download/v0.0.37/slack-linux-amd64.tar.gz | tar xzv 
sudo mv slack /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x-labs/jxl/releases/download/v0.0.37/slack-darwin-amd64.tar.gz | tar xzv
sudo mv slack /usr/local/bin
```
## Changes

### Bug Fixes

* chart (James Strachan)
* lets discover the git URL if none is supplied (James Strachan)
* polish chart (James Strachan)
* triggers (James Strachan)
* added pipeline (James Strachan)
* upgrade dependencies (James Strachan)
* better links to logs (James Strachan)
* use better links to build logs (James Strachan)
* added better unit testing of message generation (James Strachan)
* add better tests of the various predicates (James Strachan)
* migrate to v3 (James Strachan)

### Chores

* polish readme (James Strachan)
* added fake slack (James Strachan)
* first spike of finding previous pipeline (James Strachan)
* better testing (James Strachan)
* refactor a bit more (James Strachan)
* refactored a bit more (James Strachan)
* remove old CRD code (James Strachan)
* better validation code (James Strachan)
* switch to latest jx-gitops (James Strachan)
* switch to using the gitops config (James Strachan)

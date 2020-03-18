#!/usr/bin/env bash
set -e
set -x

if [ "$1" = "" ]; then
    echo "pass a version as an argument"
    exit -1
fi

VERSION=$1

make linux && docker build -t rawlingsj80/slack:$VERSION . && \
  docker push rawlingsj80/slack:$VERSION

# helm upgrade slack jx-labs/slack
helm upgrade --set googleSecretsManager=true --set image.repository=rawlingsj80/slack --set image.tag=$VERSION slack jx-labs/slack

#kubectl apply -f  ~/.jx/localSecrets/slack.yaml

sleep 10

kubectl logs -f deploy/slack-slack
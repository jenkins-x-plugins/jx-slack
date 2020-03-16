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

helm uninstall slack || true
cd charts/slack && helm install --set googleSecretsManager=true --set image.repository=rawlingsj80/slack --set image.tag=$VERSION slack .

#kubectl apply -f  ~/.jx/localSecrets/slack.yaml

# hack until we watch for new SlackBot kinds being added
#kubectl scale deploy slack-slack --replicas 0
#kubectl scale deploy slack-slack --replicas 1

sleep 10

kubectl logs -f deploy/slack-slack
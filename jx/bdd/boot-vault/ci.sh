#!/usr/bin/env bash
set -e
set -x

export GHE_USERNAME="dev1"
export GHE_EMAIL="jenkins-x@googlegroups.com"

# fix broken `BUILD_NUMBER` env var
export BUILD_NUMBER="$BUILD_ID"

JX_HOME="/tmp/jxhome"
KUBECONFIG="/tmp/jxhome/config"

# lets avoid the git/credentials causing confusion during the test
export XDG_CONFIG_HOME=$JX_HOME

mkdir -p $JX_HOME/git

jx --version

# replace the credentials file with a single user entry
echo "https://$GHE_USERNAME:$GHE_ACCESS_TOKEN@github.beescloud.com" > $JX_HOME/git/credentials

gcloud auth activate-service-account --key-file $GKE_SA

# lets setup git
git config --global --add user.name JenkinsXBot
git config --global --add user.email jenkins-x@googlegroups.com

echo "running the BDD tests with JX_HOME = $JX_HOME"

sed -e s/\$VERSION/${VERSION_PREFIX}${VERSION}/g -e s/\$CODECOV_TOKEN/${CODECOV_TOKEN}/g boot-vault.platform.yaml.template > boot-vault.platform.yaml
sed -e s/\$VERSION/${VERSION_PREFIX}${VERSION}/g -e s/\$CODECOV_TOKEN/${CODECOV_TOKEN}/g boot-vault.prow.yaml.template > boot-vault.prow.yaml

# setup jx boot parameters
export JX_VALUE_ADMINUSER_PASSWORD="$JENKINS_PASSWORD"
export JX_VALUE_PIPELINEUSER_USERNAME="$GHE_USERNAME"
export JX_VALUE_PIPELINEUSER_EMAIL="$GHE_EMAIL"
export JX_VALUE_PIPELINEUSER_TOKEN="$GHE_ACCESS_TOKEN"
export JX_VALUE_PROW_HMACTOKEN="$GHE_ACCESS_TOKEN"

# TODO temporary hack until the batch mode in jx is fixed...
export JX_BATCH_MODE="true"

git clone https://github.com/jenkins-x/jenkins-x-boot-config.git boot-source
cp jx/bdd/boot-vault/jx-requirements.yml boot-source
cp jx/bdd/boot-vault/parameters.yaml boot-source/env
cd boot-source

cp env/jenkins-x-platform/values.tmpl.yaml tmp.yaml
cat tmp.yaml ../boot-vault.platform.yaml > env/jenkins-x-platform/values.tmpl.yaml
rm tmp.yaml

# TODO hack until we fix boot to do this too!
helm init --client-only
helm repo add jenkins-x https://storage.googleapis.com/chartmuseum.jenkins-x.io


jx step bdd \
    --versions-repo https://github.com/jenkins-x/jenkins-x-versions.git \
    --config ../jx/bdd/boot-vault/cluster.yaml \
    --gopath /tmp \
    --git-provider=github \
    --git-provider-url https://github.beescloud.com \
    --git-username $GHE_USERNAME \
    --git-api-token $GHE_ACCESS_TOKEN \
    --default-admin-password $JENKINS_PASSWORD \
    --no-delete-app \
    --no-delete-repo \
    --tests install \
    --tests test-quickstart-golang-http \
    --tests test-app-lifecycle

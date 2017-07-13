#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOPATH=$(go env GOPATH)
SRC=$GOPATH/src
BIN=$GOPATH/bin
ROOT=$GOPATH
REPO_ROOT=$GOPATH/src/github.com/appscode/kubed

source "$REPO_ROOT/hack/libbuild/common/lib.sh"
source "$REPO_ROOT/hack/libbuild/common/public_image.sh"

APPSCODE_ENV=${APPSCODE_ENV:-dev}
IMG=kubed
OSM_VER=${OSM_VER:-0.5.0}

DIST=$REPO_ROOT/dist
mkdir -p $DIST
if [ -f "$DIST/.tag" ]; then
	export $(cat $DIST/.tag | xargs)
fi

clean() {
    pushd $REPO_ROOT/hack/docker
	rm -rf kubed
	popd
}

build_binary() {
	pushd $REPO_ROOT
	./hack/builddeps.sh
    ./hack/make.py build kubed
	detect_tag $REPO_ROOT/dist/.tag

    # Download restic
    rm -rf $DIST/osm
    mkdir $DIST/osm
    cd $DIST/osm
    wget https://cdn.appscode.com/binaries/osm/${OSM_VER}/osm-linux-amd64
	popd
}

build_docker() {
	pushd $REPO_ROOT/hack/docker
	cp $DIST/kubed/kubed-linux-amd64 kubed
	cp $DIST/osm/osm-linux-amd64 osm
	chmod 755 kubed

	cat >Dockerfile <<EOL
FROM alpine

RUN set -x \
  && apk add --update --no-cache ca-certificates

COPY kubed /kubed
COPY osm /osm
ENTRYPOINT ["/kubed"]
EOL
	local cmd="docker build -t appscode/$IMG:$TAG ."
	echo $cmd; $cmd

	rm kubed Dockerfile
	popd
}

build() {
	build_binary
	build_docker
}

docker_push() {
	if [ "$APPSCODE_ENV" = "prod" ]; then
		echo "Nothing to do in prod env. Are you trying to 'release' binaries to prod?"
		exit 0
	fi

    if [[ "$(docker images -q appscode/$IMG:$TAG 2> /dev/null)" != "" ]]; then
        docker_up $IMG:$TAG
    fi
}

docker_release() {
	if [ "$APPSCODE_ENV" != "prod" ]; then
		echo "'release' only works in PROD env."
		exit 1
	fi
	if [ "$TAG_STRATEGY" != "git_tag" ]; then
		echo "'apply_tag' to release binaries and/or docker images."
		exit 1
	fi

    if [[ "$(docker images -q appscode/$IMG:$TAG 2> /dev/null)" != "" ]]; then
        docker push appscode/$IMG:$TAG
    fi
}

source_repo $@

#!/bin/bash

set -e
set -x

CURDIR="$(dirname $BASH_SOURCE)"

. $CURDIR/_build-lib.sh

# Build the images based on arguments provided.
# It supports debian and rhel image. Debian image for development environment and travis CI. While RHEL image is a UBI
# and can be used for testing and gitlab CI.
# Apart from it in case of debian and debug image it also create a f5-ipam-controller-builder image which is based on debian

# For a debian image it uses:  Dockerfile-debian-builder and Dockerfile-debian-runtime
# For a debian based debug image it uses:  Dockerfile-debian-builder and Dockerfile-debian-debug
# For RHEL it uses single multi-stage Dockerfile: Dockerfile-rhel-multistage

WKDIR=$(mktemp -d /tmp/docker-build.XXXX)
VERSION_BUILD_ARGS=$(${CURDIR}/version-tool.py docker-build-args)
VERSION_INFO=$(${CURDIR}/version-tool.py version)

cp -rf $CURDIR/../../f5-ipam-controller $WKDIR/

get_attributions(){
  id=$(docker create $1)
    docker cp $id:/go/out/bin/all_attributions.txt .
    docker rm -v $id
}

NO_CACHE_ARGS=""
if $CLEAN_BUILD; then
  NO_CACHE_ARGS="--no-cache"
  docker rmi $BUILD_IMG_TAG || true
  if [[ $BASE_OS != "rhel" ]]; then
    docker rmi f5-ipam-controller-builder || true
  fi
  if [[ $DEBUG == 0 ]]; then
    docker rmi $BUILD_IMG_TAG-debug || true
  fi
fi

if [[ $BASE_OS == "rhel" ]]; then
  docker build --pull --force-rm ${NO_CACHE_ARGS} \
  -t $IMG_TAG \
  -f $WKDIR/f5-ipam-controller/build-tools/Dockerfile-rhel-multistage \
  --build-arg COVERALLS_TOKEN=${COVERALLS_TOKEN:-false} \
  --build-arg RUN_TESTS=${RUN_TESTS:-false} \
  --build-arg BUILD_VERSION=${BUILD_VERSION} \
  --build-arg BUILD_INFO=${BUILD_INFO} \
  --build-arg VERSION_INFO=${VERSION_INFO} \
  --build-arg LICENSE=${LICENSE} \
  --label BUILD_STAMP=$BUILD_STAMP \
  ${VERSION_BUILD_ARGS} \
  $WKDIR
else
  docker build --force-rm ${NO_CACHE_ARGS} \
  -t f5-ipam-controller-builder \
  -f $WKDIR/f5-ipam-controller/build-tools/Dockerfile-debian-builder \
  --build-arg COVERALLS_TOKEN=${COVERALLS_TOKEN:-false} \
  --build-arg RUN_TESTS=${RUN_TESTS:-false} \
  --build-arg BUILD_VERSION=${BUILD_VERSION} \
  --build-arg BUILD_INFO=${BUILD_INFO} \
  --build-arg LICENSE=${LICENSE} \
  --label BUILD_STAMP=$BUILD_STAMP \
  ${VERSION_BUILD_ARGS} \
  $WKDIR

  if [ $DEBUG == 0 ]; then
    docker build --force-rm ${NO_CACHE_ARGS} \
    -t $IMG_TAG-debug \
    -f $WKDIR/f5-ipam-controller/build-tools/Dockerfile-debian-debug \
    --build-arg BUILD_INFO=${BUILD_INFO} \
    --build-arg VERSION_INFO=${VERSION_INFO} \
    --label BUILD_STAMP=$BUILD_STAMP \
    ${VERSION_BUILD_ARGS} \
    $WKDIR
  else
    docker build --force-rm ${NO_CACHE_ARGS} \
    -t $IMG_TAG \
    -f $WKDIR/f5-ipam-controller/build-tools/Dockerfile-debian-runtime \
    --build-arg BUILD_INFO=${BUILD_INFO} \
    --build-arg VERSION_INFO=${VERSION_INFO} \
    --label BUILD_STAMP=$BUILD_STAMP \
    ${VERSION_BUILD_ARGS} \
    $WKDIR
  fi
  if $LICENSE; then
    get_attributions f5-ipam-controller-builder
  fi
fi

rm -rf /tmp/docker-build.????

if [[ $DEBUG == 0 ]]; then
  docker history $IMG_TAG-debug
  echo "Built docker image $IMG_TAG-debug"
else
  docker history $IMG_TAG
  echo "Built docker image $IMG_TAG"
fi

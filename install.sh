#!/usr/bin/env bash
# Inspired by: github.com/moovweb/gvm

{ # this ensures the entire script is downloaded #

  display_error() {
    tput sgr0
    tput setaf 1
    echo "ERROR: $1"
    tput sgr0
    exit 1
  }


  MACHINE_TYPE=`uname -m`
  if [ ${MACHINE_TYPE} != 'x86_64' ]; then
    display_error "32 bits architecture is not supported"
  fi

  OS=`uname | tr '[:upper:]' '[:lower:]'`

  REPO="https://github.com/manifoldco/manifold-cli"

  LATEST_RELEASE=$(curl -L -s -H 'Accept: application/json' ${REPO}/releases/latest)

  # The releases are returned in the format {"id":3622206,"tag_name":"hello-1.0.0.11",...}, we have to extract the tag_name.
  LATEST_VERSION=$(echo $LATEST_RELEASE | sed -e 's/.*"tag_name":"\([^"]*\)".*/\1/')

  FILENAME="manifold-cli_${LATEST_VERSION:1}_${OS}_amd64.zip"
  ARTIFACT_URL="${REPO}/releases/download/${LATEST_VERSION}/${FILENAME}"

  echo "Downloading latest release (${LATEST_VERSION}) for manifold-cli"
  curl --compressed -L -q $ARTIFACT_URL --output $FILENAME

}

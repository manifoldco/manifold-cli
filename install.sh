#!/usr/bin/env bash
# Inspired by:
# - https://github.com/creationix/nvm
# - https://github.com/moovweb/gvm
# - https://blog.markvincze.com/download-artifacts-from-a-latest-github-release-in-sh-and-powershell/

{ # this ensures the entire script is downloaded #

  display_error() {
    tput sgr0
    tput setaf 1
    echo "ERROR: $1"
    tput sgr0
    exit 1
  }

  try_profile() {
    if [ -z "${1-}" ] || [ ! -f "${1}" ]; then
      return 1
    fi
    echo "${1}"
  }

  #
  # Detect profile file if not specified as environment variable
  # (eg: PROFILE=~/.myprofile)
  # The echo'ed path is guaranteed to be an existing file
  # Otherwise, an empty string is returned
  #
  detect_profile() {
    if [ -n "${PROFILE}" ] && [ -f "${PROFILE}" ]; then
      echo "${PROFILE}"
      return
    fi

    local DETECTED_PROFILE
    DETECTED_PROFILE=''
    local SHELLTYPE
    SHELLTYPE="$(basename "/$SHELL")"

    if [ "$SHELLTYPE" = "bash" ]; then
      if [ -f "$HOME/.bashrc" ]; then
        DETECTED_PROFILE="$HOME/.bashrc"
      elif [ -f "$HOME/.bash_profile" ]; then
        DETECTED_PROFILE="$HOME/.bash_profile"
      fi
    elif [ "$SHELLTYPE" = "zsh" ]; then
      DETECTED_PROFILE="$HOME/.zshrc"
    fi

    if [ -z "$DETECTED_PROFILE" ]; then
      for EACH_PROFILE in ".profile" ".bashrc" ".bash_profile" ".zshrc"
      do
        if DETECTED_PROFILE="$(try_profile "${HOME}/${EACH_PROFILE}")"; then
          break
        fi
      done
    fi

    if [ ! -z "$DETECTED_PROFILE" ]; then
      echo "$DETECTED_PROFILE"
    fi
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

  # Create filename accordingly to manifoldco/promulgate structure
  FILENAME="manifold-cli_${LATEST_VERSION:1}_${OS}_amd64.zip"

  ARTIFACT_URL="${REPO}/releases/download/${LATEST_VERSION}/${FILENAME}"

  DESTINATION_DIR="${HOME}/.manifold/bin"

  echo "Creating directory at $DESTINATION_DIR"
  mkdir -p $DESTINATION_DIR

  pushd $DESTINATION_DIR > /dev/null

  echo "Downloading latest release (${LATEST_VERSION}) for manifold-cli"

  curl --compressed -L -q $ARTIFACT_URL --output $FILENAME

  echo "Extracting file"
  unzip $FILENAME

  rm $FILENAME

  popd > /dev/null

  PROFILE=`detect_profile`

  PATH_CHANGE="export PATH=\"\$PATH:$DESTINATION_DIR\""

  if [ "$PROFILE" == "" ]; then
    echo "Unable to locate profile settings file(Something like $HOME/.bashrc or $HOME/.bash_profile)"
    echo
    echo "You will have to manually add the following line:"
    echo
    echo "  $PATH_CHANGE"
    echo
  else
    echo $PATH_CHANGE >> $PROFILE
    echo "Your $PROFILE has changed to include $DESTINATION_DIR into your PATH"
  fi

  echo
  echo "All done!"
  echo "Please restart or resource your terminal session."
  echo "Happy manifold use! :)"
}

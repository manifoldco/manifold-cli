#!/bin/sh
# Inspired by:
# - https://github.com/creationix/nvm
# - https://github.com/moovweb/gvm
# - https://blog.markvincze.com/download-artifacts-from-a-latest-github-release-in-sh-and-powershell/

{ # this ensures the entire script is downloaded #

  error_exit() {
    tput sgr0
    tput setaf 1
    echo "ERROR: $1"
    tput sgr0
    exit 1
  }

  success_msg() {
    command printf "["
    tput sgr0
    tput setaf 2
    command printf "OK"
    tput sgr0
    command printf "] ${1}\\n"
  }

  warning_msg() {
    command printf "["
    tput sgr0
    tput setaf 3
    command printf "!!"
    tput sgr0
    command printf "] ${1}\\n"
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
    error_exit "32 bits architecture is not supported"
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

  mkdir -p $DESTINATION_DIR

  pushd $DESTINATION_DIR > /dev/null

  #curl --compressed -L -q $ARTIFACT_URL --output $FILENAME
  success_msg "Latest version ($LATEST_VERSION) downloaded"

  #unzip -o $FILENAME
  success_msg "Binary installed at $DESTINATION_DIR"

  #rm $FILENAME

  popd > /dev/null

  PROFILE=`detect_profile`

  PATH_CHANGE="export PATH=\"\$PATH:$DESTINATION_DIR\""

  if [[ ":$PATH:" == *":$DESTINATION_DIR:"* ]]; then
    success_msg "\$PATH already contained $DESTINATION_DIR. Skipping..."
  elif [ "$PROFILE" == "" ]; then
    warning_msg "Unable to locate profile settings file (something like $HOME/.bashrc or $HOME/.bash_profile)"
    warning_msg "You will have to manually add the following line:"
    echo
    command printf "\\t$PATH_CHANGE\\n"
  else
    echo $PATH_CHANGE >> $PROFILE
    success_msg "Your $PROFILE has changed to include $DESTINATION_DIR into your PATH"
    warning_msg "Please restart or re-source your terminal session."
  fi

  echo
  success_msg "All done. Happy manifold use! :)"
}

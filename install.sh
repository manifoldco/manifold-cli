#!/bin/sh
# Inspired by:
# - https://github.com/creationix/nvm
# - https://github.com/moovweb/gvm
# - https://blog.markvincze.com/download-artifacts-from-a-latest-github-release-in-sh-and-powershell/

{ # this ensures the entire script is downloaded #

  is_installed() {
    type "$1" > /dev/null 2>&1
  }

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

  if [ "$OS" != "linux" ] && [ "$OS" != "darwin" ]; then
    error_exit "Operational System $OS not supported."
  fi

  if ! is_installed tar; then
    error_exit "You must have 'tar' installed before proceeding."
  fi

  if is_installed manifold; then
    warning_msg "Previous installation detected: v`manifold -v`"
  fi

  REPO="https://github.com/manifoldco/manifold-cli"

  if [ -z "$MANIFOLD_VERSION" ]; then
    MANIFOLD_VERSION="0.7.0"
  fi

  if [ -z "$MANIFOLD_DIR" ]; then
    MANIFOLD_DIR="${HOME}/.manifold/bin"
  fi

  mkdir -p $MANIFOLD_DIR

  pushd $MANIFOLD_DIR > /dev/null

  # Create filename accordingly to manifoldco/promulgate structure
  FILENAME="manifold-cli_${MANIFOLD_VERSION}_${OS}_amd64.tar.gz"

  ARTIFACT_URL="${REPO}/releases/download/v${MANIFOLD_VERSION}/${FILENAME}"

  curl -sS --compressed -L -q $ARTIFACT_URL --output $FILENAME
  success_msg "Version ($MANIFOLD_VERSION) downloaded"

  tar xvf $FILENAME 1> /dev/null

  rm $FILENAME

  popd > /dev/null

  if is_installed manifold; then
    PREVIOUS_INSTALLATION=`which manifold`
    if [ "$PREVIOUS_INSTALLATION" == "$MANIFOLD_DIR/manifold" ]; then
      success_msg "Binary updated at $MANIFOLD_DIR"
    else
      success_msg "Binary installed at $MANIFOLD_DIR"
      warning_msg "Another executable detected: `type manifold`"
      warning_msg "To manually run this current installation: .$MANIFOLD_DIR/manifold"
    fi
  else
    success_msg "Binary installed at $MANIFOLD_DIR"
  fi

  PROFILE=`detect_profile`

  PATH_CHANGE="export PATH=\"\$PATH:$MANIFOLD_DIR\""

  if [[ ":$PATH:" == *":$MANIFOLD_DIR:"* ]]; then
    success_msg "\$PATH already contained $MANIFOLD_DIR. Skipping..."
  elif [ "$PROFILE" == "" ]; then
    warning_msg "Unable to locate profile settings file (something like $HOME/.bashrc or $HOME/.bash_profile)"
    warning_msg "You will have to manually add the following line:"
    echo
    command printf "\\t$PATH_CHANGE\\n"
  else
    echo $PATH_CHANGE >> $PROFILE
    success_msg "Your $PROFILE has changed to include $MANIFOLD_DIR into your PATH"
    warning_msg "Please restart or re-source your terminal session."
  fi

  echo
  success_msg "All done. Happy manifold use! :)"
}

#!/bin/sh

{ # this ensures the entire script is downloaded #

  is_installed() {
    type "$1" > /dev/null 2>&1
  }

  error_exit() {
    if is_installed tput; then
      tput sgr0
      tput setaf 1
      echo "ERROR: $1"
      tput sgr0
    else
      echo "ERROR: $1"
    fi
    exit 1
  }

  success_msg() {
    if is_installed tput; then
      command printf "["
      tput sgr0
      tput setaf 2
      command printf "OK"
      tput sgr0
      command printf "] ${1}\\n"
    else
      echo "[OK]: $1"
    fi
  }

  warning_msg() {
    if is_installed tput; then
      command printf "["
      tput sgr0
      tput setaf 3
      command printf "!!"
      tput sgr0
      command printf "] ${1}\\n"
    else
      echo "[!!]: $1"
    fi
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

  autocomplete() {
    local DETECTED_PROFILE
    AUTOCOMPLETE=''

    local SHELLTYPE
    SHELLTYPE="$(basename "/$SHELL")"

    if [ "$SHELLTYPE" = "bash" ]; then
      AUTOCOMPLETE=$(cat <<'EOF'
#! /bin/bash
export MANIFOLD_AUTOCOMPLETE=true
_manifold_bash_autocomplete() {
  local cur opts base
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
  COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
  return 0
}
complete -F _manifold_bash_autocomplete manifold
EOF
)
    fi

    if [ ! -z "$AUTOCOMPLETE" ]; then
      echo "$AUTOCOMPLETE"
    fi
  }

  PROFILE=`detect_profile`
  AUTOCOMPLETE=`autocomplete`

  if [ -z "$MANIFOLD_DIR" ]; then
    MANIFOLD_DIR="${HOME}/.manifold/bin"
  fi

  mkdir -p $MANIFOLD_DIR

  if [ "$PROFILE" = "" ]; then
    error_exit "Unable to locate profile settings file (something like $HOME/.bashrc or $HOME/.bash_profile)"
  elif [ "$MANIFOLD_AUTOCOMPLETE" = "" ]; then
    echo "$AUTOCOMPLETE" > "$MANIFOLD_DIR/.manifold_completion"
    echo "source $MANIFOLD_DIR/.manifold_completion" >> $PROFILE

    success_msg "Your $PROFILE has changed to manifold autocomplete"
    warning_msg "Please restart or re-source your terminal session."
  fi

}

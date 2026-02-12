#!/usr/bin/env bash

######################################################################
#   _
#  | |    __ _  ___  _ __  _ __ ___   ___ _ __ ___   __ _ _ __
# / __)  / _` |/ _ \| '_ \| '__/ _ \ / __| '_ ` _ \ / _` | '__|
# \__ \ | (_| | (_) | |_) | | | (_) | (__| | | | | | (_| | |
# (   /  \__, |\___/| .__/|_|  \___/ \___|_| |_| |_|\__, |_|
#  |_|   |___/      |_|                             |___/
#
# [goprocmgr] This program is a configuration manager and process
#             runner for servers, it has an http API to manage and
#             retrieve the configuration. It also provides a CLI
#             client for interacting with the API.
#
# https://github.com/etu/goprocmgr
######################################################################

# Main completion function
_goprocmgr() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    opts="-config -serve -list -list-format -add -remove -start -stop -logs -version"

    # Case handling based on the previous word
    case "${prev}" in
        -list-format)
            mapfile -t COMPREPLY < <(compgen -W "table csv" -- "${cur}")
            return 0
            ;;
        -remove)
            mapfile -t COMPREPLY < <(compgen -W "$(__goprocmgr_get_names)" -- "${cur}")
            return 0
            ;;
        -start)
            mapfile -t COMPREPLY < <(compgen -W "$(__goprocmgr_get_stopped_names)" -- "${cur}")
            return 0
            ;;
        -stop|-logs)
            mapfile -t COMPREPLY < <(compgen -W "$(__goprocmgr_get_running_names)" -- "${cur}")
            return 0
            ;;
    esac

    # Generic option completion
    if [[ ${cur} == -* ]] ; then
        mapfile -t COMPREPLY < <(compgen -W "${opts}" -- "${cur}")
        return 0
    fi
}

# Helper function to get all process names
__goprocmgr_get_names() {
    goprocmgr -list -list-format csv 2>/dev/null | tail -n +2 | cut -d ',' -f 1
}

# Helper function to get running process names
__goprocmgr_get_running_names() {
    goprocmgr -list -list-format csv 2>/dev/null | awk -F ',' '$2 == "true" {print $1}'
}

# Helper function to get stopped process names
__goprocmgr_get_stopped_names() {
    goprocmgr -list -list-format csv 2>/dev/null | awk -F ',' '$2 == "false" {print $1}'
}

# Registering the completion function for goprocmgr
complete -F _goprocmgr goprocmgr

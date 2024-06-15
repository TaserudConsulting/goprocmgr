#!/usr/bin/env fish

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
# https://github.com/TaserudConsulting/goprocmgr
######################################################################

function __goprocmgr_get_names
    # Get the list of server names
    goprocmgr -list -list-format csv 2> /dev/null | tail -n +2 | cut -d ',' -f 1
end

function __goprocmgr_get_running_names
    # Get the list of running server names
    goprocmgr -list -list-format csv 2> /dev/null | awk -F ',' '$2 == "true" {print $1}'
end

function __goprocmgr_get_stopped_names
    # Get the list of stopped server names
    goprocmgr -list -list-format csv 2> /dev/null | awk -F ',' '$2 == "false" {print $1}'
end

# Set known action flags to be able to make completions not complete
# two different actions at once.
set -l actions '-serve -list -add -remove -start -stop -logs -version'

# Complete main options
complete --command goprocmgr --condition "not __fish_seen_subcommand_from -config"  --old-option config --require-parameter --force-files                                  --description 'Specify the configuration file'
complete --command goprocmgr --condition '__fish_seen_subcommand_from -list'        --old-option list-format --exclusive           --arguments 'table csv'                 --description 'Specify the list format (table, csv)'
complete --command goprocmgr --condition "not __fish_seen_subcommand_from $actions" --old-option serve  --no-files                                                         --description 'Run the serve command (start the web server)'
complete --command goprocmgr --condition "not __fish_seen_subcommand_from $actions" --old-option list   --no-files                                                         --description 'List the stored servers'
complete --command goprocmgr --condition "not __fish_seen_subcommand_from $actions" --old-option add    --require-parameter                                                --description 'Add a new server'
complete --command goprocmgr --condition "not __fish_seen_subcommand_from $actions" --old-option remove --exclusive         --arguments '(__goprocmgr_get_names)'          --description 'Remove an existing server by its name'
complete --command goprocmgr --condition "not __fish_seen_subcommand_from $actions" --old-option start  --exclusive         --arguments '(__goprocmgr_get_stopped_names)'  --description 'Start an existing server by its name'
complete --command goprocmgr --condition "not __fish_seen_subcommand_from $actions" --old-option stop   --exclusive         --arguments '(__goprocmgr_get_running_names)'  --description 'Stop an existing server by its name'
complete --command goprocmgr --condition "not __fish_seen_subcommand_from $actions" --old-option logs   --exclusive         --arguments '(__goprocmgr_get_running_names)'  --description 'Tail the logs from an existing server by its name'
complete --command goprocmgr --condition "not __fish_seen_subcommand_from $actions" --old-option version --no-files                                                        --description 'Print version'

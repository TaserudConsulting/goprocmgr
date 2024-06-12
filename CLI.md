% goprocmgr(1) Version %undefined-version% | User Manual
# NAME
goprocmgr - A tool to manage servers and their processes.

# SYNOPSIS
**goprocmgr** [OPTIONS]

# DESCRIPTION
`goprocmgr` is a command-line utility for managing servers and their
processes. It provides commands to *serve*, *list*, *add*, *remove*,
*start*, *stop*, and tail logs of servers.

# OPTIONS
**-config** *file*
: Specify the configuration file. This can be used with any command
: since the config defines how to connect to the API. Default is
: `~/.config/goprocmgr.json`.

**-serve**
: Run the serve command (start the web server). Default is true.

**-list**
: List the stored servers.

**-list-format** *format*
: Specify the list format (*table*, *csv*) when using the list
: command. Default is *table*.

**-add** *command*
: Add a new server, capturing the current directory and environment,
: and then takes the command as an argument.

**-remove** *name*
: Remove an existing server by its name.

**-start** *name*
: Start an existing server by its name.

**-stop** *name*
: Stop an existing server by its name.

**-logs** *name*
: Tail the logs from an existing server by its name.

**-version**
: Print the version of the utility.

# EXAMPLES
Run the web server:
: goprocmgr -serve

List stored servers in CSV format:
: goprocmgr -list -list-format csv

Add a new server (*name* is picked by the current working directory name):
: goprocmgr -add "start-command for server"

Remove a server:
: goprocmgr -remove *name*

Start a server:
: goprocmgr -start *name*

Stop a server:
: goprocmgr -stop *name*

Tail the logs of a server:
: goprocmgr -logs *name*

Print version:
: goprocmgr -version

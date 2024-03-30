// JS File
'use strict';

const App = () => {
    // Store state for the server list
    const serverListState = van.state([])

    // Store state for selected server in the UI.
    const selectedServerState = van.state(localStorage.getItem('selectedServerState') ?? null)

    // Save the selected server state to local storage when it's updated.
    van.derive(() => {
        localStorage.setItem('selectedServerState', selectedServerState.val)
    });

    // This loads the current configured servers and their running state, then it updates
    // the serverListState to update the rendered list.
    const loadServers = (async () => {
        const configs = await (await fetch('/api/config')).json()
        const runners = await (await fetch('/api/runner')).json()

        const tmpServerListState = []

        for (const serverName in configs.servers) {
            tmpServerListState.push({
                name: serverName,
                running: (serverName in runners),
                stdout: runners[serverName] ? runners[serverName].stdout : [],
                stderr: runners[serverName] ? runners[serverName].stderr : [],
            })
        }

        serverListState.val = tmpServerListState
    })

    // Actually load the state
    loadServers()

    // And refresh the state every second
    setInterval(loadServers, 1000)

    // Update serverListState for the object with the name of `name` to running state of `state`
    const setServerListStateFor = (name, state) => {
        const tmpServerListState = []

        for (const item of serverListState.val) {
            if (item.name === name) {
                item.running = state
            }

            tmpServerListState.push(item)
        }

        serverListState.val = tmpServerListState
    }

    // Get the running state of the server with the name of `name`
    const getServerListStateFor = (name) => {
        for (const item of serverListState.val) {
            if (item.name === name) {
                return item.running
            }
        }

        return false
    }

    // Object to render the actual items in the server list
    const ServerItem = (name) => {
        const toggleServer = async () => {
            if (getServerListStateFor(name)) {
                await fetch(`/api/runner/${name}`, { method: 'DELETE' })
                setServerListStateFor(name, false)
                return
            }

            await fetch(`/api/runner/${name}`, { method: 'POST' })
            setServerListStateFor(name, true)
        }

        return van.tags.li(
            {
                class: () => selectedServerState.val === name ? 'server-item selected' : 'server-item',
                onclick: () => { selectedServerState.val = name }
            },
            name,
            van.tags.label(
                { class: 'switch', for: 'toggle-' + name },
                van.tags.input({
                    type: 'checkbox',
                    id: 'toggle-' + name,
                    checked: getServerListStateFor(name),
                    onclick: () => toggleServer(),
                }),
                van.tags.div({ class: 'slider' })
            )
        )
    }

    // Derive the server list state into a list of items to render
    const serverList = van.derive(() => van.tags.ul(
        { class: 'server-list' },
        serverListState.val.map((item) => ServerItem(item.name))
    ))

    // Derive the selection state to render the main viewer
    const mainViewer = van.derive(() => {
        // Render the stderr of the selected server, however it will render the newest
        // lines at the top of the viewer.
        const stderrViewer = (name) => {
            let stderrLogLines = []

            // Loop through serverListState until you find the server with the name of `name`
            for (const item of serverListState.val) {
                if (item.name === name && item.stderr !== null) {
                    stderrLogLines = item.stderr

                    break
                }
            }

            return van.tags.div(
                { id: 'stderr-wrapper' },
                van.tags.div(
                    { id: 'stderr' },
                    stderrLogLines.reverse().map((line) => van.tags.div(line))
                )
            )
        }

        // Render the stdout of the selected server, however it will render the newest
        // lines at the top of the viewer.
        const stdoutViewer = (name) => {
            let stdoutLogLines = []

            // Loop through serverListState until you find the server with the name of `name`
            for (const item of serverListState.val) {
                if (item.name === name && item.stdout !== null) {
                    stdoutLogLines = item.stdout

                    break
                }
            }

            return van.tags.div(
                { id: 'stdout-wrapper' },
                van.tags.div(
                    { id: 'stdout' },
                    stdoutLogLines.reverse().map((line) => van.tags.div(line))
                )
            )
        }

        // Select a welcome message based on if there's servers or not.
        const welcomeMessage = (serverListState.val.length === 0) ? 'No servers configured yet :)' : 'Select a server to view its logs :)'

        return van.tags.div(
            { id: 'content' },
            (!selectedServerState.val || selectedServerState.val === 'null') ? van.tags.div({ id: 'frontpage' }, welcomeMessage) : null,
            (!getServerListStateFor(selectedServerState.val)) ? van.tags.div({ id: 'frontpage' }, 'Server "', selectedServerState.val, '" is currently not started :)') : null,
            (getServerListStateFor(selectedServerState.val) && !!selectedServerState.val) ? stderrViewer(selectedServerState.val) : null,
            (getServerListStateFor(selectedServerState.val) && !!selectedServerState.val) ? stdoutViewer(selectedServerState.val) : null,
        )
    })

    return van.tags.div(
        { id: 'wrapper' },
        van.tags.nav(
            { id: 'nav' },
            van.tags.h1(
                { onclick: () => { selectedServerState.val = null } },
                'goprocmgr'
            ),
            serverList
        ),
        mainViewer,
    )
}

van.add(document.getElementById('app'), App());

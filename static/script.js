// JS File
'use strict';

const App = () => {
    // State to pause the refresh of the server list
    const pauseRefresh = van.state(false)

    // Store state for the server list
    const serverListState = van.state([])

    // Store state for selected server in the UI.
    const selectedServerState = van.state(localStorage.getItem('selectedServerState') ?? null)

    // Save the selected server state to local storage when it's updated.
    van.derive(() => {
        localStorage.setItem('selectedServerState', selectedServerState.val)
    })

    // This loads the current configured servers and their running state, then it updates
    // the serverListState to update the rendered list.
    const loadServers = async () => {
        if (pauseRefresh.val) {
            return
        }

        const configs = await (await fetch('/api/config')).json()
        const runners = await (await fetch('/api/runner')).json()

        serverListState.val = Object.keys(configs.servers).map(serverName => ({
            name: serverName,
            port: runners[serverName]?.port ?? 0,
            running: (serverName in runners),
            stdout: runners[serverName]?.stdout || [],
            stderr: runners[serverName]?.stderr || [],
        }))
    }

    // Actually load the state
    loadServers()

    // And refresh the state every second
    setInterval(loadServers, 1000)

    // Update serverListState for the object with the name of `name` to running state of `state`
    const setServerListStateFor = (name, state) => {
        serverListState.val = serverListState.val.map(item => {
            if (item.name === name) {
                return { ...item, running: state }
            }
            return item
        })
    }

    // Get the running state of the server with the name of `name`
    const getServerListStateFor = name => {
        const server = serverListState.val.find(item => item.name === name);

        return server ? server.running : false;
    }

    // Toggle the server with the name of `name`
    const toggleServer = async (name) => {
        if (getServerListStateFor(name)) {
            await fetch(`/api/runner/${name}`, { method: 'DELETE' })
            setServerListStateFor(name, false)
            return
        }

        await fetch(`/api/runner/${name}`, { method: 'POST' })
        setServerListStateFor(name, true)
    }

    // Object to render the actual items in the server list
    const ServerItem = name => van.tags.li(
        {
            class: () => selectedServerState.val === name ? 'server-item selected' : 'server-item',
            onclick: () => { selectedServerState.val = name }
        },
        getServerListStateFor(name) ? van.tags.a(
            { target: '_blank', href: `http://${window.location.hostname}:${serverListState.val.find(item => item.name === name)?.port ?? 0}` },
            name,
        ) : name,
        getServerListStateFor(name) ? van.tags.span(
            { class: 'log-item-count' },
            ' (',
            van.tags.span({ class: 'stderr' }, serverListState.val.find(item => item.name === name)?.stderr?.length ?? 0),
            '/',
            van.tags.span({ class: 'stdout' }, serverListState.val.find(item => item.name === name)?.stdout?.length ?? 0),
            ')',
        ) : null,
        van.tags.label(
            { class: 'switch', for: 'toggle-' + name },
            van.tags.input({
                type: 'checkbox',
                id: 'toggle-' + name,
                checked: getServerListStateFor(name),
                onclick: () => toggleServer(name),
            }),
            van.tags.div({ class: 'slider' })
        )
    )


    // Derive the server list state into a list of items to render
    const serverList = van.derive(() => van.tags.ul(
        { class: 'server-list' },
        van.tags.li(
            { class: 'server-item refresh-toggle' },
            'auto refresh',
            van.tags.label(
                { class: 'switch', for: 'refresh-toggle' },
                van.tags.input({
                    type: 'checkbox',
                    id: 'refresh-toggle',
                    checked: () => !pauseRefresh.val ? 'checked' : null,
                    onclick: () => { pauseRefresh.val = !pauseRefresh.val },
                }),
                van.tags.div({ class: 'slider' })
            )
        ),
        serverListState.val.map(item => ServerItem(item.name))
    ))

    // Derive the selection state to render the main viewer
    const mainViewer = van.derive(() => {
        // Render the stderr of the selected server, however it will render the newest
        // lines at the top of the viewer.
        const stderrViewer = name => {
            const server = serverListState.val.find(item => item.name === name);
            const stderrLogLines = server?.stderr || [];

            return van.tags.div(
                { id: 'stderr-wrapper' },
                van.tags.div(
                    { id: 'stderr' },
                    stderrLogLines.reverse().map(line => van.tags.div(line))
                )
            );
        }

        // Render the stdout of the selected server, however it will render the newest
        // lines at the top of the viewer.
        const stdoutViewer = name => {
            const server = serverListState.val.find(item => item.name === name && item.stdout !== null);
            const stdoutLogLines = server?.stdout || [];

            return van.tags.div(
                { id: 'stdout-wrapper' },
                van.tags.div(
                    { id: 'stdout' },
                    stdoutLogLines.reverse().map(line => van.tags.div(line))
                )
            );
        }

        // Select a welcome message based on if there's servers or not.
        const welcomeMessage = (serverListState.val.length === 0) ? 'No servers configured yet :)' : 'Select a server to view its logs :)'

        return van.tags.div(
            { id: 'content' },
            (!selectedServerState.val || selectedServerState.val === 'null') ? van.tags.div({ id: 'frontpage' }, welcomeMessage) : null,
            (!getServerListStateFor(selectedServerState.val) && selectedServerState.val !== 'null' && selectedServerState.val !== null) ?
                van.tags.div({ id: 'frontpage' }, 'Server "', selectedServerState.val, '" is currently not started :)') : null,
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
            serverList,
            van.tags.aside(
                { class: 'bottom-nav' },
                van.tags.a(
                    { href: 'https://github.com/TaserudConsulting/goprocmgr', target: '_blank' },
                    'GitHub'
                ),
            )
        ),
        mainViewer,
    )
}

van.add(document.getElementById('app'), App());

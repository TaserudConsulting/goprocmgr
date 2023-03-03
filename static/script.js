// JS File
'use strict';

(function () {
    const root = document.getElementById('root')
    const nav = root.appendChild(document.createElement(`nav`))
    const main = root.appendChild(document.createElement(`div`))

    //
    // This function will query the config and runner API's to get the
    // configured servers and then to determine which ones are running
    // or not. It should be used to render a list of processes and
    // display if it's running or not.
    //
    const fetchServers = async() => {
        const response = {};

        const configs = await (await fetch('/api/config')).json()
        const runners = await (await fetch('/api/runner')).json()

        // Go through configs to build response containing some info
        // about the service and it's current status.
        for (const serverName in configs.servers) {
            response[serverName] = {
                name: serverName,
                cmd: configs.servers[serverName].cmd,
                cwd: configs.servers[serverName].cwd,
                isRunning: serverName in runners,
            }
        }

        return response
    }

    nav.textContent = "navigation"
    nav.id = "nav"
    main.textContent = "main viewer"
    main.id = "content"
})()

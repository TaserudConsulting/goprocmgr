// JS File
'use strict';

(function () {
    const root = document.getElementById('root')
    const nav = root.appendChild(document.createElement(`nav`))
    const navH1 = nav.appendChild(document.createElement(`h1`))
    const navUl = nav.appendChild(document.createElement(`ul`))
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

    //
    // This function actually clears the UL tag from items and then
    // creates new list items to update the menu.
    //
    const renderMenu = async() => {
        const servers = await fetchServers()

        if (navUl.textContent === "🔃") {
            // Empty the navigation list before we re-render it.
            navUl.textContent = ""
        }

        // Go through all servers and add them to the navigation list.
        for (const serverName in servers) {
            // Try to select the previously selecting element
            let li = document.getElementById("li-server-" + serverName)

            // If it doesn't exist, create it.
            if (!li) {
                li = navUl.appendChild(document.createElement(`li`))
                li.id = "li-server-" + serverName
            }

            // Prepare checked="" attribute for input type checkbox.
            const checkStatus = servers[serverName].isRunning ? 'checked=""' : ''

            // Update the content
            li.innerHTML = `
              <span>${serverName}</span> - <input type="checkbox" ${checkStatus} />
            `
        }
    }

    nav.id = "nav"
    navH1.textContent = "navigation"
    navUl.textContent = "🔃"
    main.textContent = "main viewer"
    main.id = "content"

    // Render the menu and keep updating it every now and then.
    renderMenu()
    setInterval(renderMenu, 5000)
})()

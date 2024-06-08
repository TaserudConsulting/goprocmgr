'use strict'

document.addEventListener('alpine:init', () => {
    Alpine.data('app', () => ({
        // Application state
        serverList: [],

        // The selected server, this is used to show the logs for a specific server.
        selectedServer: localStorage.getItem('selectedServer') === 'null' ? null : localStorage.getItem('selectedServer'),

        // The key event, this is used to listen for key events and trigger actions.
        keyEvent: new KeyboardEvent("keydown"),

        // Show the keybinds, this is used to show the keybinds on the page.
        showKeybinds: false,

        // The WebSocket connection, this is used to get data from the server.
        ws: null,

        // Initialize the application
        init() {
            // Setup the WebSocket connection to get data from the server.
            this.setupWebSocket()

            // Listen for keydown events on the document.
            document.addEventListener('keydown', (event) => {
                this.keyEvent = event
                this.handleKeyEvents()
            })

            // Watch for changes to the selected server and save it to local storage
            // so we can remember the selected server when the page is reloaded.
            this.$watch('selectedServer', (value) => {
                localStorage.setItem('selectedServer', value)
            })
        },

        // Setup the WebSocket connection to get data from the server.
        setupWebSocket() {
            // Create a new WebSocket connection to the server.
            this.ws = new WebSocket('ws://' + window.location.host + '/api/ws')

            // On message, we parse the JSON data and update the server list.
            this.ws.onmessage = (event) => {
                const data = JSON.parse(event.data)

                // Update the server list with the data from the server.
                this.serverList = Object.keys(data.configs.servers).map(serverName => ({
                    name: serverName,
                    port: data.runners[serverName]?.port ?? 0,
                    running: (serverName in data.runners),
                    logs: data.runners[serverName]?.logs || [],
                }))
            }

            // Reconnect on close, we also wipe the web socket instance
            // and the contents of the application. Then we wait a bit
            // before trying to reconnect.
            this.ws.onclose = () => {
                this.ws = null
                this.serverList = []

                setTimeout(() => {
                    this.setupWebSocket()
                }, 1000)
            }
        },

        // Toggle the server state, if it's running, stop it, if it's stopped, start it.
        async toggleServer(name) {
            await fetch(`/api/runner/${name}`, {
                method: this.getServer(name).running ? 'DELETE' : 'POST',
            })
        },

        // Get the server by name
        getServer(name) {
            return this.serverList.find(item => item.name === name) || {}
        },

        // Get the count of logs by output
        countLogsByOutput(name, output) {
            return this.getServer(name).logs.filter(log => log.output === output).length
        },

        // Format a timestamp to HH:MM:SS
        formatTimestamp(timestamp) {
            return new Date(timestamp).toLocaleTimeString([], {
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit',
                hour12: false,
            })
        },

        // Handle key events for keyboard shortcuts
        handleKeyEvents() {
            if (this.keyEvent.key === 'Escape') {
                this.selectedServer = null
            }

            if (this.keyEvent.key === 't' && this.selectedServer) {
                this.keyEvent = new KeyboardEvent("keydown")
                this.toggleServer(this.selectedServer)
            }

            if (this.keyEvent.key === 'n') {
                const currentIndex = this.serverList.findIndex(item => item.name === this.selectedServer)
                const nextIndex = currentIndex + 1

                if (nextIndex < this.serverList.length) {
                    this.selectedServer = this.serverList[nextIndex].name
                }
            }

            if (this.keyEvent.key === 'p') {
                const currentIndex = this.serverList.findIndex(item => item.name === this.selectedServer)
                const previousIndex = currentIndex - 1

                if (previousIndex >= 0) {
                    this.selectedServer = this.serverList[previousIndex].name
                }
            }

            if (this.keyEvent.key === 'h') {
                this.showKeybinds = !this.showKeybinds
            }
        },
    }))
})

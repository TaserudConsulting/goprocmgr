'use strict'

document.addEventListener('alpine:init', () => {
    Alpine.data('app', () => ({
        // Application state
        serverList: [],
        serverLogs: [],

        // The selected server, this is used to show the logs for a specific server.
        selectedServer: localStorage.getItem('selectedServer') === 'null' ? null : localStorage.getItem('selectedServer') || null,

        // The key event, this is used to listen for key events and trigger actions.
        keyEvent: new KeyboardEvent("keydown"),
        keyEventHandled: false,

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
                this.serverLogs = []
                this.subscribeToServer(value)
            })
        },

        // Setup the WebSocket connection to get data from the server.
        setupWebSocket() {
            // Create a new WebSocket connection to the server.
            this.ws = new WebSocket(`ws://${window.location.host}/api/ws`)

            // On open, subscribe to the currently selected server.
            this.ws.onopen = () => {
                this.subscribeToServer(this.selectedServer)
            }

            // On message, we parse the JSON data and update the server list.
            this.ws.onmessage = (event) => {
                const data = JSON.parse(event.data)

                // Check if it's a full state or a specific server update
                if (data.servers) {
                    this.serverList = Object.values(data.servers)

                    // Count servers that has the is_running flag set to true
                    document.title = 'goprocmgr (' + this.serverList.filter(item => item.is_running).length + ')'
                } else if (data.server && data.logs) {
                    // Specific server update
                    this.serverLogs = data.logs
                }
            }

            // Reconnect on close, we also wipe the web socket instance
            // and the contents of the application. Then we wait a bit
            // before trying to reconnect.
            this.ws.onclose = () => {
                this.ws = null
                this.serverList = []
                this.serverLogs = []

                setTimeout(() => {
                    this.setupWebSocket()
                }, 1000)
            }
        },

        // Subscribe to updates for a specific server
        subscribeToServer(serverName) {
            if (this.ws && this.ws.readyState === WebSocket.OPEN && serverName) {
                this.ws.send(JSON.stringify({ name: serverName }))
            }
        },

        // Toggle the server state, if it's running, stop it, if it's stopped, start it.
        async toggleServer(name) {
            await fetch(`/api/runner/${name}`, {
                method: this.getServer(name).is_running ? 'DELETE' : 'POST',
            })
        },

        // Get the server by name
        getServer(name) {
            return this.serverList.find(item => item.name === name) || {}
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
            if (this.keyEventHandled) return
            this.keyEventHandled = true

            setTimeout(() => {
                this.keyEventHandled = false
            }, 100)

            if (this.keyEvent.key === 'Escape') {
                this.selectedServer = null
            }

            if (this.keyEvent.key === 't' && this.selectedServer) {
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

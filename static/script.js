'use strict'

document.addEventListener('alpine:init', () => {
    Alpine.data('app', () => ({
        // Application state
        serverList: [],
        serverLogs: [],
        serverLogsOffset: 0, // Track the current offset for pagination
        previousStdoutCount: 0, // Track previous stdout count for restart detection
        previousStderrCount: 0, // Track previous stderr count for restart detection

        // The selected server, this is used to show the logs for a specific server.
        selectedServer: localStorage.getItem('selectedServer') === 'null' ? null : localStorage.getItem('selectedServer') || null,

        // The key event, this is used to listen for key events and trigger actions.
        keyEvent: new KeyboardEvent("keydown"),
        keyEventHandled: false,

        // Show the keybinds, this is used to show the keybinds on the page.
        showKeybinds: false,

        // The WebSocket connection, this is used to get data from the server.
        ws: null,

        // Allow auto scrolling
        autoScroll: true,

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
                this.serverLogsOffset = 0 // Reset offset when changing servers
                this.previousStdoutCount = 0 // Reset count tracking when changing servers
                this.previousStderrCount = 0 // Reset count tracking when changing servers
                this.subscribeToServer(value)
                this.scrollServerItemIntoViewIfNeeded(value)
            })

            // Watch for changes in the serverLogs array
            this.$watch('serverLogs', (_) => {
                this.$nextTick(() => {
                    if (this.autoScroll) {
                        this.scrollToBottom()
                    }
                })
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

                    // Scroll into view.
                    if (this.selectedServer) {
                        this.scrollServerItemIntoViewIfNeeded(this.selectedServer)
                    }
                } else if (data.server && data.logs !== undefined) {
                    // Specific server update with pagination
                    
                    // Check if server was restarted (logs were cleared on server side)
                    // This happens when:
                    // 1. total_count is less than our current offset
                    // 2. we receive offset 0 with logs while we already have logs
                    // 3. stdout_count or stderr_count decreased from previous value
                    // 4. both counts went to 0 (server stopped) and we have existing logs
                    const currentStdoutCount = data.server.stdout_count || 0
                    const currentStderrCount = data.server.stderr_count || 0
                    const stdoutDecreased = this.previousStdoutCount > 0 && currentStdoutCount < this.previousStdoutCount
                    const stderrDecreased = this.previousStderrCount > 0 && currentStderrCount < this.previousStderrCount
                    const serverStopped = currentStdoutCount === 0 && currentStderrCount === 0 && this.serverLogs.length > 0
                    
                    if (data.total_count < this.serverLogsOffset || 
                        (data.offset === 0 && this.serverLogs.length > 0 && data.logs.length > 0) ||
                        stdoutDecreased || stderrDecreased || serverStopped) {
                        // Server was restarted or stopped, clear client logs and reset offset
                        this.serverLogs = []
                        this.serverLogsOffset = 0
                    }
                    
                    // Update previous counts for next comparison
                    this.previousStdoutCount = currentStdoutCount
                    this.previousStderrCount = currentStderrCount
                    
                    // Append new logs to existing logs efficiently
                    // Add unique IDs to each log entry for proper rendering
                    // Use the offset from the server response as the base for IDs
                    const logsWithIds = data.logs.map((log, idx) => ({
                        ...log,
                        _id: `${data.offset + idx}`
                    }))
                    this.serverLogs.push(...logsWithIds)
                    // Update our offset to match what we've received
                    this.serverLogsOffset = data.offset + data.logs.length
                }
            }

            // Reconnect on close, we also wipe the web socket instance
            // and the contents of the application. Then we wait a bit
            // before trying to reconnect.
            this.ws.onclose = () => {
                this.ws = null
                this.serverList = []
                this.serverLogs = []
                this.serverLogsOffset = 0 // Reset offset on reconnect
                this.previousStdoutCount = 0 // Reset count tracking on reconnect
                this.previousStderrCount = 0 // Reset count tracking on reconnect

                setTimeout(() => {
                    this.setupWebSocket()
                }, 1000)
            }
        },

        // Method to check scroll position to enable or disable autoScroll
        checkScrollPosition() {
            const logsWrapper = this.$refs.logsWrapper

            // If the user is near the bottom, enable auto-scroll
            this.autoScroll = logsWrapper.scrollTop + logsWrapper.clientHeight >= logsWrapper.scrollHeight - 10
        },

        // Scroll to the bottom of the logs
        scrollToBottom() {
            if (this.$refs.logsWrapper) {
                this.$refs.logsWrapper.scrollTop = this.$refs.logsWrapper.scrollHeight
            }
        },

        // Scroll the server item into view
        scrollServerItemIntoViewIfNeeded(serverName) {
            this.$nextTick(() => {
                // Use querySelector to find the element with the matching data-list-item-server-name attribute
                const serverItem = document.querySelector(`[data-list-item-server-name="${serverName}"]`);

                if (serverItem) {
                    serverItem.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'nearest' })
                }
            })
        },

        // Subscribe to updates for a specific server
        subscribeToServer(serverName) {
            if (this.ws && this.ws.readyState === WebSocket.OPEN && serverName) {
                this.ws.send(JSON.stringify({ name: serverName, offset: this.serverLogsOffset }))
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

            if (this.keyEvent.key === 'e') {
                this.scrollToBottom()
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

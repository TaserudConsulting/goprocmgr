'use strict'

document.addEventListener('alpine:init', () => {
    Alpine.data('app', () => ({
        serverList: [],
        selectedServer: localStorage.getItem('selectedServer') === 'null' ? null : localStorage.getItem('selectedServer'),
        keyEvent: new KeyboardEvent("keydown"),
        showKeybinds: false,
        ws: null,
        init() {
            this.setupWebSocket()

            document.addEventListener('keydown', (event) => {
                this.keyEvent = event
                this.handleKeyEvents()
            })

            this.$watch('selectedServer', (value) => {
                localStorage.setItem('selectedServer', value)
            })
        },
        setupWebSocket() {
            this.ws = new WebSocket('ws://' + window.location.host + '/api/ws')

            this.ws.onmessage = (event) => {
                const data = JSON.parse(event.data)

                this.serverList = Object.keys(data.configs.servers).map(serverName => ({
                    name: serverName,
                    port: data.runners[serverName]?.port ?? 0,
                    running: (serverName in data.runners),
                    logs: data.runners[serverName]?.logs || [],
                }))
            }
        },
        async toggleServer(name) {
            if (this.getServer(name).running) {
                await fetch(`/api/runner/${name}`, { method: 'DELETE' })
                this.setServerState(name, false)
            } else {
                await fetch(`/api/runner/${name}`, { method: 'POST' })
                this.setServerState(name, true)
            }
        },
        setServerState(name, state) {
            this.serverList = this.serverList.map(item => {
                if (item.name === name) {
                    return { ...item, running: state }
                }
                return item
            })
        },
        getServer(name) {
            return this.serverList.find(item => item.name === name) || {}
        },
        countLogsByOutput(name, output) {
            return this.getServer(name).logs.filter(log => log.output === output).length
        },
        formatTimestamp(timestamp) {
            // Format the timestamp to HH:MM:SS
            return new Date(timestamp).toLocaleTimeString([], {
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit',
                hour12: false,
            })
        },
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

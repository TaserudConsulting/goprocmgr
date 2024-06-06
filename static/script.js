'use strict';

document.addEventListener('alpine:init', () => {
    Alpine.data('app', () => ({
        pauseRefresh: false,
        serverList: [],
        selectedServer: localStorage.getItem('selectedServer') ?? null,
        keyEvent: new KeyboardEvent("keydown"),
        showKeybinds: false,
        init() {
            this.loadServers();
            setInterval(() => this.loadServers(), 1000);

            document.addEventListener('keydown', (event) => {
                this.keyEvent = event;
                this.handleKeyEvents();
            });

            this.$watch('selectedServer', (value) => {
                localStorage.setItem('selectedServer', value);
            });
        },
        async loadServers() {
            if (this.pauseRefresh) return;

            const configs = await (await fetch('/api/config')).json();
            const runners = await (await fetch('/api/runner')).json();

            this.serverList = Object.keys(configs.servers).map(serverName => ({
                name: serverName,
                port: runners[serverName]?.port ?? 0,
                running: (serverName in runners),
                stdout: runners[serverName]?.stdout || [],
                stderr: runners[serverName]?.stderr || [],
            }));
        },
        async toggleServer(name) {
            if (this.getServer(name).running) {
                await fetch(`/api/runner/${name}`, { method: 'DELETE' });
                this.setServerState(name, false);
            } else {
                await fetch(`/api/runner/${name}`, { method: 'POST' });
                this.setServerState(name, true);
            }
        },
        setServerState(name, state) {
            this.serverList = this.serverList.map(item => {
                if (item.name === name) {
                    return { ...item, running: state };
                }
                return item;
            });
        },
        getServer(name) {
            return this.serverList.find(item => item.name === name) || {};
        },
        handleKeyEvents() {
            if (this.keyEvent.key === 'Escape') {
                this.selectedServer = null;
            }

            if (this.keyEvent.key === 't' && this.selectedServer) {
                this.keyEvent = new KeyboardEvent("keydown");
                this.toggleServer(this.selectedServer);
            }

            if (this.keyEvent.key === 'n') {
                const currentIndex = this.serverList.findIndex(item => item.name === this.selectedServer);
                const nextIndex = currentIndex + 1;

                if (nextIndex < this.serverList.length) {
                    this.selectedServer = this.serverList[nextIndex].name;
                }
            }

            if (this.keyEvent.key === 'p') {
                const currentIndex = this.serverList.findIndex(item => item.name === this.selectedServer);
                const previousIndex = currentIndex - 1;

                if (previousIndex >= 0) {
                    this.selectedServer = this.serverList[previousIndex].name;
                }
            }

            if (this.keyEvent.key === 'r') {
                this.pauseRefresh = !this.pauseRefresh;
            }

            if (this.keyEvent.key === 'h') {
                this.showKeybinds = !this.showKeybinds;
            }
        },
    }));
});

const reopenTimeouts = [500, 1000, 1500, 2000, 5000, 10000, 30000, 60000]

const firWindow = typeof window !== 'undefined' ? window : null
const firDocument = typeof document !== 'undefined' ? document : null

export default websocket = (url, socketOptions, dispatchServerEvents) => {
    let socket, openPromise, reopenTimeoutHandler
    let reopenCount = 0
    let pendingHeartbeat = false

    // modified from https://github.com/arlac77/svelte-websocket-store/blob/master/src/index.mjs
    // thank you https://github.com/arlac77 !!
    function reopenTimeout() {
        const n = reopenCount
        reopenCount++
        return reopenTimeouts[
            n >= reopenTimeouts.length - 1 ? reopenTimeouts.length - 1 : n
        ]
    }

    function closeSocket() {
        if (reopenTimeoutHandler) {
            clearTimeout(reopenTimeoutHandler)
        }

        if (socket && socket.readyState === WebSocket.CLOSED) {
            socket = undefined
        }

        if (socket && socket.readyState == WebSocket.CONNECTING) {
            setTimeout(() => {
                socket.close()
                socket = undefined
            }, 1000)
        }

        if (socket && socket.readyState == WebSocket.OPEN) {
            socket.close()
            socket = undefined
        }
    }

    if (firWindow && firWindow.addEventListener) {
        firWindow.addEventListener('pagehide', () => {
            if (socket && socket.readyState == WebSocket.OPEN) {
                socket.close()
                socket = undefined
            }
            if (socket && socket.readyState == WebSocket.CONNECTING) {
                setTimeout(() => {
                    socket.close()
                    socket = undefined
                }, 1000)
            }
        })
    }

    if (firWindow && firWindow.addEventListener) {
        firWindow.addEventListener('pageshow', () => {
            reopenSocket()
        })
    }

    if (firDocument && firDocument.addEventListener) {
        firDocument.addEventListener('visibilitychange', () => {
            if (firDocument.visibilityState === 'visible') {
                sendAndAckHeartbeat()
            }
        })
    }

    function sendAndAckHeartbeat() {
        if (!socket || socket.readyState !== WebSocket.OPEN) {
            return
        }
        socket.send(JSON.stringify({ event_id: 'heartbeat' }))
        pendingHeartbeat = true
        setTimeout(() => {
            if (pendingHeartbeat) {
                pendingHeartbeat = false
                reopenSocket()
            }
        }, 500)
    }

    function reopenSocket() {
        if (
            socket &&
            (socket.readyState === WebSocket.CONNECTING ||
                socket.readyState === WebSocket.OPEN)
        ) {
            return
        }
        closeSocket()
        reopenTimeoutHandler = setTimeout(() => {
            openSocket()
                .then(() => {
                    pendingHeartbeat = false
                })
                .catch((e) => {
                    console.error(e)
                })
        }, reopenTimeout())
    }

    async function openSocket() {
        if (reopenTimeoutHandler) {
            clearTimeout(reopenTimeoutHandler)
            reopenTimeoutHandler = undefined
        }

        // we are still in the opening phase
        if (openPromise) {
            return openPromise
        }

        try {
            socket = new WebSocket(url, socketOptions)
        } catch (e) {
            console.error("can't create socket", e)
        }

        socket.onclose = (event) => {
            console.warn('socket closed', event)
            if (event.code == 4001) {
                console.warn(`socket closed by server: unauthorized`)
                if (event.reason) {
                    window.location.href = event.reason
                }
                closeSocket()
                return
            }
            return reopenSocket()
        }
        socket.onmessage = (event) => {
            try {
                const serverEvents = JSON.parse(event.data)
                if (serverEvents.event_id === 'heartbeat_ack') {
                    pendingHeartbeat = false
                    return
                }
                dispatchServerEvents(serverEvents)
            } catch (e) {}
        }

        socket.onerror = (error) => {
            console.warn('socket error', error)
            return reopenSocket()
        }

        openPromise = new Promise((resolve, reject) => {
            socket.onerror = (error) => {
                console.error('socket error on connect', error)
                reject(error)
                openPromise = undefined
            }
            socket.onopen = (event) => {
                reopenCount = 0
                resolve()
                openPromise = undefined
            }
        })
        return openPromise
    }

    openSocket()
        .then(() => {
            pendingHeartbeat = false
        })
        .catch((e) => console.error(e))

    return {
        emit(value) {
            if (socket && socket.readyState === WebSocket.OPEN) {
                socket.send(JSON.stringify(value))
                return true
            } else {
                return false
            }
        },
    }
}

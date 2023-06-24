const reopenTimeouts = [500, 1000, 1500, 2000, 5000, 10000, 30000, 60000]

const firWindow = typeof window !== 'undefined' ? window : null

export default websocket = (url, socketOptions, dispatchServerEvents) => {
    let socket, openPromise, reopenTimeoutHandler
    let reopenCount = 0

    // socket code copied from https://github.com/arlac77/svelte-websocket-store/blob/master/src/index.mjs
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
            reOpenSocket()
        })
    }

    function reOpenSocket() {
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
                    //console.log("socket connected");
                    // location.reload(true)
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
            // console.log("socket disconnected")
        }

        socket.onclose = (event) => {
            if (event.code == 4001) {
                console.log(`socket closed by server: unauthorized`)
                if (event.reason) {
                    window.location.href = event.reason
                }
                closeSocket()
                return
            }
            return reOpenSocket()
        }
        socket.onmessage = (event) => {
            try {
                const serverEvents = JSON.parse(event.data)
                dispatchServerEvents(serverEvents)
            } catch (e) {}
        }

        openPromise = new Promise((resolve, reject) => {
            socket.onerror = (error) => {
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
        .then(() => {})
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

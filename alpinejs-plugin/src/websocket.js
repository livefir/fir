import { getSessionIDFromCookie } from './utils' // Adjust path if necessary

const REOPEN_TIMEOUTS = [500, 1000, 1500, 2000, 5000, 10000, 30000, 60000]
const MAX_RECONNECT_ATTEMPTS = 10
const HEARTBEAT_TIMEOUT = 500
const firDocument = typeof document !== 'undefined' ? document : null

/**
 * Creates a WebSocket connection with automatic reconnection and heartbeat functionality
 * @param {string} url - WebSocket URL
 * @param {string|string[]} socketOptions - WebSocket protocol options
 * @param {Function} processServerEventsCallback - Callback to handle and dispatch server events (e.g., processAndDispatchServerEvents)
 * @returns {Object} WebSocket API
 */
export default function createWebSocket(
    url,
    socketOptions,
    processServerEventsCallback // Renamed parameter for clarity
) {
    console.log('createWebSocket called')
    let socket = null
    let openPromise = null
    let reopenTimeoutHandler = null
    let reconnectAttempts = 0
    let pendingHeartbeat = false
    let visibilityHandler = null
    let isClosedByUser = false

    /**
     * Calculate exponential backoff time for reconnection attempts
     * @returns {number} Timeout in milliseconds
     */
    function getReconnectTimeout() {
        const index = Math.min(reconnectAttempts, REOPEN_TIMEOUTS.length - 1)
        reconnectAttempts++
        return REOPEN_TIMEOUTS[index]
    }

    /**
     * Properly close the WebSocket connection
     */
    function closeSocket() {
        if (reopenTimeoutHandler) {
            clearTimeout(reopenTimeoutHandler)
            reopenTimeoutHandler = null
        }

        if (!socket) return

        try {
            if (
                socket.readyState === WebSocket.OPEN ||
                socket.readyState === WebSocket.CONNECTING
            ) {
                socket.close()
            }
        } catch (err) {
            console.error('Error closing socket:', err)
        } finally {
            socket = null
        }
    }

    /**
     * Send heartbeat and verify server response
     */
    function sendAndAckHeartbeat() {
        if (!socket || socket.readyState !== WebSocket.OPEN) {
            return
        }

        try {
            socket.send(JSON.stringify({ event_id: 'heartbeat' }))
            pendingHeartbeat = true

            setTimeout(() => {
                if (pendingHeartbeat && !isClosedByUser) {
                    console.warn('Heartbeat timeout - reconnecting')
                    pendingHeartbeat = false
                    closeSocket()
                    reopenSocket()
                }
            }, HEARTBEAT_TIMEOUT)
        } catch (err) {
            console.error('Error sending heartbeat:', err)
            pendingHeartbeat = false
        }
    }

    /**
     * Schedule a socket reconnection
     */
    function reopenSocket() {
        if (isClosedByUser) return

        if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
            console.error(
                `Maximum reconnection attempts (${MAX_RECONNECT_ATTEMPTS}) reached`
            )
            return
        }

        if (socket && socket.readyState === WebSocket.CONNECTING) {
            return
        }

        if (socket && socket.readyState === WebSocket.OPEN) {
            closeSocket()
        }

        const timeout = getReconnectTimeout()
        console.log(
            `Reconnecting in ${timeout}ms (attempt ${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS})`
        )

        reopenTimeoutHandler = setTimeout(() => {
            openSocket()
                .then(() => {
                    pendingHeartbeat = false
                })
                .catch((e) => {
                    console.error('Reconnection failed:', e)
                    reopenSocket()
                })
        }, timeout)
    }

    /**
     * Open a new WebSocket connection
     * @returns {Promise} Promise resolving when connection is established
     */
    async function openSocket() {
        if (reopenTimeoutHandler) {
            clearTimeout(reopenTimeoutHandler)
            reopenTimeoutHandler = null
        }

        // Return existing promise if we're already connecting
        if (openPromise) {
            return openPromise
        }

        try {
            socket = new WebSocket(url, socketOptions)
        } catch (err) {
            console.error('Failed to create WebSocket:', err)
            throw err
        }

        // Set up socket event handlers
        socket.onclose = (event) => {
            console.warn('WebSocket closed', event)

            if (event.code === 4001) {
                console.warn('Socket closed by server: unauthorized')
                if (event.reason && typeof window !== 'undefined') {
                    window.location.href = event.reason
                }
                return
            }

            if (!isClosedByUser) {
                reopenSocket()
            }
        }

        socket.onmessage = (event) => {
            try {
                const serverEvents = JSON.parse(event.data)
                if (serverEvents.event_id === 'heartbeat_ack') {
                    pendingHeartbeat = false
                    return
                }
                // Use the renamed callback parameter
                processServerEventsCallback(serverEvents)
            } catch (err) {
                console.error('Error processing message:', err)
            }
        }

        socket.onerror = (error) => {
            console.warn('WebSocket error', error)
        }

        // Create promise for connection establishment
        openPromise = new Promise((resolve, reject) => {
            const errorHandler = (error) => {
                console.error('WebSocket connection error:', error)
                socket.onopen = null
                socket.onerror = socket.onerror || (() => {})
                reject(error)
                openPromise = null
            }

            const openHandler = () => {
                console.log('WebSocket connected')
                reconnectAttempts = 0
                socket.onerror = (error) => {
                    console.warn('WebSocket error:', error)
                }
                resolve()
                openPromise = null
            }

            socket.onerror = errorHandler
            socket.onopen = openHandler

            // Timeout if connection takes too long
            setTimeout(() => {
                if (openPromise) {
                    errorHandler(new Error('Connection timeout'))
                }
            }, 5000)
        })

        return openPromise
    }

    // Set up visibility change handler
    if (firDocument) {
        visibilityHandler = () => {
            if (firDocument.visibilityState === 'visible') {
                if (!socket || socket.readyState !== WebSocket.OPEN) {
                    openSocket().catch((err) => console.error(err))
                } else {
                    sendAndAckHeartbeat()
                }
            }
        }
        firDocument.addEventListener('visibilitychange', visibilityHandler)
    }

    // Initial connection
    openSocket()
        .then(() => {
            pendingHeartbeat = false
        })
        .catch((err) => console.error('Initial connection failed:', err))

    // Public API
    return {
        /**
         * Send a message through the WebSocket
         * @param {any} value - Data to send (will be JSON stringified)
         * @returns {boolean} Success status
         */
        emit(value) {
            if (socket && socket.readyState === WebSocket.OPEN) {
                try {
                    socket.send(JSON.stringify(value))
                    return true
                } catch (err) {
                    console.error('Error sending message:', err)
                    return false
                }
            }
            return false
        },

        /**
         * Close the WebSocket connection and clean up resources
         */
        close() {
            isClosedByUser = true

            if (firDocument && visibilityHandler) {
                firDocument.removeEventListener(
                    'visibilitychange',
                    visibilityHandler
                )
                visibilityHandler = null
            }

            closeSocket()
        },

        /**
         * Get the current state of the WebSocket connection
         * @returns {number|null} WebSocket readyState or null if not initialized
         */
        getState() {
            return socket ? socket.readyState : null
        },
    }
}

export const setupWebSocketConnection = async (
    processEventsFn,
    fetchFn = fetch,
    windowLocation = window.location,
    // Use the default export 'createWebSocket' as the default factory
    wsFactory = createWebSocket
) => {
    console.log('Setting up WebSocket connection...')
    let connectURL = `ws://${windowLocation.host}${windowLocation.pathname}`
    if (windowLocation.protocol === 'https:') {
        connectURL = `wss://${windowLocation.host}${windowLocation.pathname}`
    }

    if (!getSessionIDFromCookie()) {
        console.error('No session ID found in cookie. WebSocket disabled.')
        return null
    }

    try {
        const response = await fetchFn(windowLocation.href, { method: 'HEAD' })
        if (response.headers.get('X-FIR-WEBSOCKET-ENABLED') === 'true') {
            console.log('WebSocket enabled, attempting connection...')
            // The parameters (connectURL, [], callback) align with createWebSocket's signature
            // (url, socketOptions, processServerEventsCallback)
            return wsFactory(connectURL, [], (events) =>
                processEventsFn(events)
            )
        } else {
            console.log('WebSocket not enabled by server.')
            return null
        }
    } catch (error) {
        console.error('Error checking WebSocket status:', error)
        return null
    }
}

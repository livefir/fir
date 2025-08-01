import websocket, { setupWebSocketConnection } from './websocket' // Import both default and named export
import morph from '@alpinejs/morph'
import {
    // createCustomEvent, // Only if used directly in index.js, otherwise remove
    // dispatchEventOnIdTarget, // Only if used directly in index.js, otherwise remove
    // dispatchEventOnClassTarget, // Only if used directly in index.js, otherwise remove
    // isValidFirEvent, // Only if used directly in index.js, otherwise remove
    FIR_PREFIX,
    // ID_SELECTOR_PREFIX, // Only if used directly in index.js, otherwise remove
    // CLASS_SELECTOR_PREFIX, // Only if used directly in index.js, otherwise remove
    dispatchSingleServerEvent,
    processAndDispatchServerEvents,
    postEvent, // Import postEvent
} from './eventDispatcher'
// Import from new utility modules
// import { getSessionIDFromCookie } from './utils' // Only if used directly in index.js, otherwise remove
import { firMutationObserverDirective } from './firMutationObserverDirective'
import { createFirMagicFunctions } from './magicFunctions'

// Constants
// FIR_PREFIX, ID_SELECTOR_PREFIX, CLASS_SELECTOR_PREFIX have been moved to eventDispatcher.js

// Define postEvent function outside Plugin // MOVED to eventDispatcher.js
// export const postEvent = async ( ... ) => { ... }

// Main Alpine Plugin Function
const Plugin = (Alpine) => {
    // Initialize global $fir object
    window.$fir = window.$fir || {}

    // Initialize WebSocket connection asynchronously
    let socket = null
    setupWebSocketConnection(
        processAndDispatchServerEvents,
        fetch,
        window.location,
        websocket
    )
        .then((wsInstance) => {
            socket = wsInstance
            // Expose WebSocket instance globally for connection status checking
            window.$fir.ws = socket
            if (socket) {
                console.log('WebSocket connection established.')
            }
        })
        .catch((err) => {
            console.error('WebSocket setup failed:', err)
            window.$fir.ws = null
        })

    // Global event listener for reloads
    window.addEventListener(`${FIR_PREFIX}reload`, () => {
        window.location.reload()
    })

    // Register fir-mutation-observer directive
    firMutationObserverDirective(Alpine)

    // Register $fir magic helper
    Alpine.magic('fir', (el, { Alpine }) => {
        const postFnWrapper = (firEvent) => {
            // Call the imported postEvent
            postEvent(
                firEvent,
                socket,
                processAndDispatchServerEvents,
                dispatchSingleServerEvent,
                fetch,
                window.location
            )
        }
        return createFirMagicFunctions(el, Alpine, postFnWrapper)
    })

    Alpine.plugin(morph)
}

export default Plugin

// Import Alpine
import Alpine from 'alpinejs'
import FirPlugin from './src/index'

// Make Alpine available globally
window.Alpine = Alpine

// Set up mock cookies
document.cookie = '_fir_session_=test-session-id; _fir_route_=test-route-id'

// Global setup for testing
// Mock missing browser APIs
if (typeof window !== 'undefined') {
    // Mock HTMLFormElement.prototype.requestSubmit if it doesn't exist
    if (!HTMLFormElement.prototype.requestSubmit) {
        HTMLFormElement.prototype.requestSubmit = function (submitter) {
            // Create and dispatch a submit event
            const submitEvent = new Event('submit', {
                bubbles: true,
                cancelable: true,
            })

            // Add submitter to the event if provided
            if (submitter) {
                submitEvent.submitter = submitter
            }

            this.dispatchEvent(submitEvent)
        }
    }
}

// Mock HTMLFormElement.prototype.requestSubmit
if (typeof document !== 'undefined') {
    HTMLFormElement.prototype.requestSubmit =
        HTMLFormElement.prototype.requestSubmit ||
        function (submitter) {
            // Create a submit event
            const submitEvent = new Event('submit', {
                bubbles: true,
                cancelable: true,
            })

            // If submitter is provided, add it to the event
            if (submitter) {
                Object.defineProperty(submitEvent, 'submitter', {
                    value: submitter,
                    enumerable: true,
                })
            }

            // Dispatch the event
            this.dispatchEvent(submitEvent)
            return true
        }
}

// Mock global functions used by the plugin
window.morphElement = jest.fn()
window.appendElement = jest.fn()
window.prependElement = jest.fn()
window.removeElement = jest.fn()
window.removeParentElement = jest.fn()
window.dispatchServerEvents = jest.fn()
window.post = jest.fn()

// Global mocks for functions used in tests
global.morphElement = jest.fn()
global.appendElement = jest.fn()
global.prependElement = jest.fn()
global.removeElement = jest.fn()
global.removeParentElement = jest.fn()
global.dispatchServerEvents = jest.fn()
global.post = jest.fn()

// Global mocks for fetch API
global.fetch = jest.fn().mockImplementation(() =>
    Promise.resolve({
        headers: {
            get: jest.fn().mockReturnValue('true'),
        },
        json: jest.fn().mockResolvedValue([]),
        redirected: false,
    })
)

// Mock websocket
global.WebSocket = class MockWebSocket {
    constructor() {
        this.onmessage = null
        this.onopen = null
        this.onclose = null
        this.onerror = null
        this.close = jest.fn()
        // Simulate a successful connection
        setTimeout(() => {
            if (this.onopen) this.onopen({ target: this })
        }, 0)
    }
    send(data) {
        return true
    }
}

// Initialize any global variables needed for tests
window.isObject = (obj) => {
    return Object.prototype.toString.call(obj) === '[object Object]'
}

// Initialize Alpine with your plugin
console.error = jest.fn() // Temporarily silence console errors during setup
Alpine.plugin(FirPlugin)
console.error = console.error // Restore console.error

// Any other global test setup code

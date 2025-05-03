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

// Common mocks for all tests
global.post = jest.fn()
window.post = jest.fn()

// Initialize Alpine with your plugin
console.error = jest.fn() // Temporarily silence console errors during setup
Alpine.plugin(FirPlugin)
console.error = console.error // Restore console.error

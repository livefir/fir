import Alpine from 'alpinejs'
import FirPlugin from '../src/index'

// Polyfill fetch for Node.js environment
if (typeof global.fetch !== 'function') {
    global.fetch = jest.fn().mockImplementation(() =>
        Promise.resolve({
            headers: {
                get: jest.fn().mockReturnValue('true'),
            },
            json: jest.fn().mockResolvedValue([]),
            redirected: false,
        })
    )
}

// Mock other browser APIs that may be missing
if (typeof window.TextEncoder === 'undefined') {
    global.TextEncoder = class TextEncoder {
        encode(text) {
            return Buffer.from(text)
        }
    }
}

if (typeof window.TextDecoder === 'undefined') {
    global.TextDecoder = class TextDecoder {
        decode(buffer) {
            return Buffer.from(buffer).toString()
        }
    }
}

// Initialize Alpine only once for all tests
let isAlpineInitialized = false

export function initializeAlpineOnce() {
    if (!isAlpineInitialized) {
        // Mock needed global functions before initializing Alpine
        window.post = window.post || jest.fn()
        window.morphElement = window.morphElement || jest.fn()
        window.appendElement = window.appendElement || jest.fn()
        window.prependElement = window.prependElement || jest.fn()
        window.removeElement = window.removeElement || jest.fn()
        window.removeParentElement = window.removeParentElement || jest.fn()
        window.dispatchServerEvents = window.dispatchServerEvents || jest.fn()

        // Add mocked websocket
        window.WebSocket =
            window.WebSocket ||
            class MockWebSocket {
                constructor() {
                    this.onmessage = null
                    this.onopen = null
                    this.onclose = null
                    this.onerror = null
                    this.close = jest.fn()

                    // Simulate connection
                    setTimeout(() => {
                        if (this.onopen) this.onopen({ target: this })
                    }, 0)
                }
                send(data) {
                    return true
                }
            }

        Alpine.plugin(FirPlugin)
        Alpine.start()
        isAlpineInitialized = true
    }
    return Alpine
}

// Add this function to the file

export function createTestElement(html) {
    // Create a container
    const container = document.createElement('div')
    document.body.appendChild(container)

    // Set the inner HTML
    container.innerHTML = html

    // Initialize Alpine for this element
    Alpine.initTree(container)

    return {
        container,
        // Helper to find elements
        find: (selector) => container.querySelector(selector),
        // Helper to clean up
        cleanup: () => document.body.removeChild(container),
    }
}

import Alpine from 'alpinejs'
import firPlugin from '../src/index' // Adjust path if needed
import morph from '@alpinejs/morph'

// Mock the fetch function globally
global.fetch = jest.fn()

// Helper to dispatch events easily
const dispatch = (type, detail) => {
    window.dispatchEvent(
        new CustomEvent(type, {
            detail,
            bubbles: true,
            composed: true,
            cancelable: true,
        })
    )
}

describe('Alpine.js $fir Magic E2E Tests', () => {
    beforeAll(() => {
        // Mock the initial HEAD request *before* registering the plugin
        global.fetch.mockResolvedValueOnce({
            ok: true,
            headers: new Headers({ 'X-FIR-WEBSOCKET-ENABLED': 'false' }), // Mock response for HEAD check
        })

        // Register morph plugin globally for all tests in this file
        Alpine.plugin(morph)
        // Register your fir plugin globally (now fetch is mocked for its init)
        Alpine.plugin(firPlugin)

        // Explicitly start Alpine *once* after plugins are registered
        Alpine.start()
    })

    beforeEach(() => {
        // Reset mocks and setup default implementation
        global.fetch.mockClear()
        global.fetch.mockImplementation(async (url, options) => {
            if (options?.method === 'HEAD') {
                // console.log('Mock Fetch HEAD:', url); // Optional: debug log
                return {
                    ok: true,
                    headers: new Headers({
                        'X-FIR-WEBSOCKET-ENABLED': 'false',
                    }),
                }
            }
            // Default for POST or other methods
            // console.log('Mock Fetch POST/Other:', url, options); // Optional: debug log
            return {
                ok: true,
                redirected: false,
                headers: new Headers({ 'Content-Type': 'application/json' }),
                json: async () => [],
            }
        })

        document.body.innerHTML = '' // Clear previous test DOM

        // Mock minimal cookie setup
        Object.defineProperty(document, 'cookie', {
            writable: true,
            value: '_fir_session_=test-session-id',
        })
    })

    test('$fir.emit sends correct fetch request', async () => {
        // Arrange: Set up DOM
        document.body.innerHTML = `
            <div x-data>
                <button id="myButton" @click="$fir.emit('buttonClick', { value: 123 })">Click Me</button>
            </div>
        `
        // Wait for Alpine to initialize the new DOM content
        await Alpine.nextTick()

        // Act: Simulate click
        document.getElementById('myButton').click()

        // Wait for potential async operations within the event handler/post function
        await Alpine.nextTick() // Add another tick just in case

        // Assert: Check if fetch was called correctly
        expect(global.fetch).toHaveBeenCalledTimes(1) // Should be 1 POST call now

        // Find the actual POST call (ignoring potential HEAD calls)
        const fetchCall = global.fetch.mock.calls.find(
            (call) => call[1]?.method === 'POST'
        )
        expect(fetchCall).toBeDefined()

        const [url, options] = fetchCall
        expect(url).toBe(window.location.pathname) // Should post to the current path
        expect(options.method).toBe('POST')
        expect(options.headers['Content-Type']).toBe('application/json')
        expect(options.headers['X-FIR-MODE']).toBe('event')

        const body = JSON.parse(options.body)
        expect(body.event_id).toBe('buttonClick')
        expect(body.params).toEqual({ value: 123 })
        expect(body.session_id).toBe('test-session-id')
        expect(body.element_key).toBeNull() // No fir-key on the button
        expect(body.target).toBeUndefined()
        expect(body).toHaveProperty('ts')
    })

    test('$fir.submit sends correct fetch request for POST form', async () => {
        // Arrange: Set up DOM
        document.body.innerHTML = `
            <div x-data>
                <form id="myForm" @submit.prevent="$fir.submit()">
                    <input type="text" name="username" value="testuser" />
                    <button type="submit">Submit</button>
                </form>
            </div>
        `
        // Wait for Alpine to initialize the new DOM content
        await Alpine.nextTick()

        // Act: Simulate form submission
        document
            .getElementById('myForm')
            .dispatchEvent(
                new Event('submit', { bubbles: true, cancelable: true })
            )

        // Wait for potential async operations
        await Alpine.nextTick()

        // Assert: Check fetch call
        expect(global.fetch).toHaveBeenCalledTimes(1) // Should be 1 POST call

        const [url, options] = global.fetch.mock.calls[0] // Assuming only one call
        expect(url).toBe(window.location.pathname)
        expect(options.method).toBe('POST')
        expect(options.headers['Content-Type']).toBe('application/json')
        expect(options.headers['X-FIR-MODE']).toBe('event')

        const body = JSON.parse(options.body)
        expect(body.event_id).toBe('myForm') // Event ID defaults to form ID
        expect(body.params).toEqual({ username: ['testuser'] }) // Form data format
        expect(body.is_form).toBe(true)
        expect(body.session_id).toBe('test-session-id')
    })

    test('$fir.replaceEl updates element on server event', async () => {
        // Arrange: Set up DOM with listener
        document.body.innerHTML = `
            <div x-data>
                <div id="targetElement" @fir:updatecontent:ok.window="$fir.replaceEl()">
                    Initial Content
                </div>
            </div>
        `
        // Wait for Alpine to initialize the new DOM content
        await Alpine.nextTick()

        const targetElement = document.getElementById('targetElement')
        expect(targetElement.textContent.trim()).toBe('Initial Content')

        // Act: Simulate the server sending back an event
        const serverEventDetail = {
            html: '<div id="targetElement">Updated Content</div>',
        }
        dispatch('fir:updatecontent:ok', serverEventDetail)

        // Assert: Check if the element content was replaced
        await Alpine.nextTick() // Wait for listener and morph

        expect(targetElement.textContent.trim()).toBe('Updated Content')
        expect(document.getElementById('targetElement')).not.toBeNull()
    })

    test('$fir.prependEl updates element on server event', async () => {
        // Arrange: Set up DOM with listener
        document.body.innerHTML = `
            <div x-data>
                <ul id="targetList" @fir:additem:ok.window="$fir.prependEl()">
                    <li>Existing Item</li>
                </ul>
            </div>
        `
        // Wait for Alpine to initialize the new DOM content
        await Alpine.nextTick()

        const targetElement = document.getElementById('targetList')
        expect(targetElement.children.length).toBe(1)
        expect(targetElement.firstElementChild.textContent.trim()).toBe(
            'Existing Item'
        )

        // Act: Simulate the server sending back an event to prepend an item
        const serverEventDetail = { html: '<li>Prepended Item</li>' }
        dispatch('fir:additem:ok', serverEventDetail)

        // Assert: Check if the element content was prepended
        await Alpine.nextTick() // Wait for listener and morph

        expect(targetElement.children.length).toBe(2)
        expect(targetElement.firstElementChild.textContent.trim()).toBe(
            'Prepended Item'
        )
        expect(targetElement.lastElementChild.textContent.trim()).toBe(
            'Existing Item'
        )
    })

    // Add more tests for other magic functions:
    // - $fir.submit with GET
    // - $fir.appendEl
    // - $fir.afterEl
    // - $fir.beforeEl
    // - $fir.removeEl
    // - $fir.removeParentEl
    // - $fir.reset
    // - $fir.toggleDisabled (checking attributes on pending/done events)
})

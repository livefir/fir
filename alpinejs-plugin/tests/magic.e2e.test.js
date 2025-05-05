import Alpine from 'alpinejs'
import morph from '@alpinejs/morph'

// Import the plugin itself
import firPlugin from '../src/index'

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

// *** NEW: Helper function to wait for a DOM condition ***
async function waitForDOMUpdate(
    checkFn,
    { timeout = 200, interval = 20 } = {}
) {
    const endTime = Date.now() + timeout
    while (Date.now() < endTime) {
        await Alpine.nextTick() // Allow Alpine to process updates
        if (checkFn()) {
            return // Condition met
        }
        // Wait briefly before checking again
        await new Promise((resolve) => setTimeout(resolve, interval))
    }
    throw new Error(`waitForDOMUpdate timed out after ${timeout}ms`)
}
// *** End new helper ***

// Mock history API
let mockPushState

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

        mockPushState = jest.spyOn(window.history, 'pushState')
    })

    beforeEach(() => {
        // Reset mocks and setup default implementation for fetch
        global.fetch.mockClear()
        global.fetch.mockImplementation(async (url, options) => {
            if (options?.method === 'HEAD') {
                return {
                    ok: true,
                    headers: new Headers({
                        'X-FIR-WEBSOCKET-ENABLED': 'false',
                    }),
                }
            }
            // Default for POST or other methods
            return {
                ok: true,
                redirected: false,
                headers: new Headers({ 'Content-Type': 'application/json' }),
                json: async () => [], // Default empty array response
            }
        })

        document.body.innerHTML = '' // Clear previous test DOM

        // Mock minimal cookie setup
        Object.defineProperty(document, 'cookie', {
            writable: true,
            value: '_fir_session_=test-session-id',
        })

        mockPushState.mockClear() // Clear history mock calls
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

    test('$fir.submit handles GET form and updates history', async () => {
        document.body.innerHTML = `
            <div x-data>
                <form id="searchForm" method="GET" @submit.prevent="$fir.submit()">
                    <input type="text" name="query" value="alpine" />
                    <button type="submit">Search</button>
                </form>
            </div>
        `
        await Alpine.nextTick()
        document
            .getElementById('searchForm')
            .dispatchEvent(
                new Event('submit', { bubbles: true, cancelable: true })
            )
        await Alpine.nextTick()
        expect(global.fetch).toHaveBeenCalledTimes(1)
        expect(mockPushState).toHaveBeenCalledTimes(1)
        // ... rest of assertions ...
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

    test('$fir.appendEl updates element on server event', async () => {
        document.body.innerHTML = `
            <div x-data>
                <ul id="targetList" @fir:additem:ok.window="$fir.appendEl()">
                    <li>Existing Item</li>
                </ul>
            </div>
        `
        await Alpine.nextTick()
        const targetElement = document.getElementById('targetList')
        const serverEventDetail = { html: '<li>Appended Item</li>' }
        dispatch('fir:additem:ok', serverEventDetail)
        await Alpine.nextTick()
        expect(targetElement.children.length).toBe(2)
        expect(targetElement.lastElementChild.textContent.trim()).toBe(
            'Appended Item'
        )
    })

    // Re-enable the test by removing .skip
    test('$fir.afterEl inserts element on server event', async () => {
        // Arrange
        document.body.innerHTML = `
            <div x-data>
                <div id="container">
                    <span id="marker" @fir:insertafter:ok.window="$fir.afterEl()">Marker</span>
                </div>
            </div>
        `
        await Alpine.nextTick()
        const container = document.getElementById('container')
        const marker = document.getElementById('marker') // Get reference to the marker
        expect(container.children.length).toBe(1)

        // Act
        const serverEventDetail = { html: '<div>Inserted After</div>' }
        dispatch('fir:insertafter:ok', serverEventDetail)

        // Wait for the DOM update using the helper, checking the condition
        await waitForDOMUpdate(() => container.children.length === 2, {
            timeout: 500,
        }) // Increased timeout slightly just in case

        // Assert
        expect(container.children.length).toBe(2)
        expect(container.children[0].id).toBe('marker') // Check marker is still first
        expect(container.children[1].textContent.trim()).toBe('Inserted After') // Check new element is second
    })

    test('$fir.beforeEl inserts element on server event', async () => {
        document.body.innerHTML = `
            <div x-data>
                <div id="container">
                    <span id="marker" @fir:insertbefore:ok.window="$fir.beforeEl()">Marker</span>
                </div>
            </div>
        `
        await Alpine.nextTick()
        const container = document.getElementById('container')
        const serverEventDetail = { html: '<div>Inserted Before</div>' }
        dispatch('fir:insertbefore:ok', serverEventDetail)
        await Alpine.nextTick()
        expect(container.children.length).toBe(2)
        expect(container.children[0].textContent.trim()).toBe('Inserted Before')
    })

    test('$fir.removeEl removes element on server event', async () => {
        document.body.innerHTML = `
            <div x-data>
                <div id="removable" @fir:removeme:ok.window="$fir.removeEl()">Remove Me</div>
            </div>
        `
        await Alpine.nextTick()
        expect(document.getElementById('removable')).not.toBeNull()
        dispatch('fir:removeme:ok', {})
        await Alpine.nextTick()
        expect(document.getElementById('removable')).toBeNull()
    })

    test('$fir.removeParentEl removes parent element on server event', async () => {
        document.body.innerHTML = `
            <div x-data>
                <div id="parentContainer">
                    <button id="remover" @fir:removeparent:ok.window="$fir.removeParentEl()">Remove Parent</button>
                </div>
            </div>
        `
        await Alpine.nextTick()
        expect(document.getElementById('parentContainer')).not.toBeNull()
        dispatch('fir:removeparent:ok', {})
        await Alpine.nextTick()
        expect(document.getElementById('parentContainer')).toBeNull()
    })

    test('$fir.reset resets form on server event', async () => {
        document.body.innerHTML = `
            <div x-data>
                <form id="testForm" @fir:submitted:ok.window="$fir.reset()">
                    <input type="text" name="message" value="Initial Value">
                </form>
            </div>
        `
        await Alpine.nextTick()
        const form = document.getElementById('testForm')
        const input = form.elements['message']
        input.value = 'Changed Value'
        dispatch('fir:submitted:ok', {})
        await Alpine.nextTick()
        expect(input.value).toBe('Initial Value')
    })

    test('$fir.toggleDisabled toggles attributes on events', async () => {
        document.body.innerHTML = `
            <div x-data>
                <button id="myBtn"
                    @click="$fir.emit('process')"
                    @fir:process:pending.window="$fir.toggleDisabled()"
                    @fir:process:done.window="$fir.toggleDisabled()">
                    Process
                </button>
            </div>
        `
        await Alpine.nextTick()
        const button = document.getElementById('myBtn')
        dispatch('fir:process:pending', {})
        await Alpine.nextTick()
        expect(button.hasAttribute('disabled')).toBe(true)
        dispatch('fir:process:done', {})
        await Alpine.nextTick()
        expect(button.hasAttribute('disabled')).toBe(false)
    })

    // --- Test using multi-statement handler (trailing () needed) ---
    test('$fir.emit works with multiple statements', async () => {
        document.body.innerHTML = `
            <div x-data="{ clicked: false }">
                <button id="myButton" @click="clicked = true; $fir.emit('buttonClick', { value: 456 })()">Click Me</button>
                <span x-show="clicked">Clicked!</span>
            </div>
        `
        await Alpine.nextTick()
        const spanElement = document.querySelector('span') // Get reference to the span
        expect(spanElement.style.display).toBe('none')

        document.getElementById('myButton').click()

        // *** CHANGE: Use waitForDOMUpdate instead of fixed nextTicks ***
        // Wait specifically for the span's display style to change from 'none'
        await waitForDOMUpdate(() => spanElement.style.display !== 'none')
        // *** End Change ***

        // Check Alpine state change (now that we waited for the DOM update)
        expect(spanElement.style.display).not.toBe('none')

        // Check fetch call
        // Add a minimal wait to ensure async fetch mock logic completes if needed
        await new Promise((resolve) => setTimeout(resolve, 0))
        expect(global.fetch).toHaveBeenCalledTimes(1)
        const fetchCall = global.fetch.mock.calls.find(
            (call) => call[1]?.method === 'POST'
        )
        expect(fetchCall).toBeDefined()
        const body = JSON.parse(fetchCall[1].body)
        expect(body.event_id).toBe('buttonClick')
        expect(body.params).toEqual({ value: 456 })
    })

    test('$fir.emit should use element ID if provided ID is null/undefined', async () => {
        document.body.innerHTML = `
            <div x-data>
                <button id="fallbackBtn" @click="$fir.emit()">Click</button>
            </div>
        `
        await Alpine.nextTick()
        document.getElementById('fallbackBtn').click()
        await Alpine.nextTick()

        expect(global.fetch).toHaveBeenCalledTimes(1)
        const fetchCall = global.fetch.mock.calls.find(
            (call) => call[1]?.method === 'POST'
        )
        expect(fetchCall).toBeDefined()
        const body = JSON.parse(fetchCall[1].body)
        expect(body.event_id).toBe('fallbackBtn') // Uses element ID
    })

    test('$fir.emit should error if no ID provided and element has no ID', async () => {
        const consoleSpy = jest
            .spyOn(console, 'error')
            .mockImplementation(() => {})
        document.body.innerHTML = `
            <div x-data>
                <button @click="$fir.emit()">Click</button> {/* No button ID */}
            </div>
        `
        await Alpine.nextTick()
        document.querySelector('button').click()
        await Alpine.nextTick()

        expect(consoleSpy).toHaveBeenCalledWith(
            "event id is empty and element id is not set. can't emit event"
        )
        expect(global.fetch).not.toHaveBeenCalled() // Should not attempt fetch
        consoleSpy.mockRestore()
    })

    // Skip this test in JSDOM: event.submitter is not populated correctly.
    // This behavior needs to be verified via browser-based E2E tests.
    test.skip('fir.submit should use event from submitter formaction', async () => {
        document.body.innerHTML = `
            <div x-data>
                <form id="testForm" @submit.prevent="$fir.submit()">
                    <input name="data" value="abc"/>
                    <button type="submit" formaction="/?event=fromButton">Submit</button>
                </form>
            </div>
        `
        await Alpine.nextTick()

        const form = document.getElementById('testForm')
        const button = form.querySelector('button')

        button.click()

        await Alpine.nextTick()
        await new Promise(process.nextTick)

        expect(global.fetch).toHaveBeenCalledTimes(1)
        const fetchCall = global.fetch.mock.calls.find(
            (call) => call[1]?.method === 'POST'
        )
        expect(fetchCall).toBeDefined()
        const body = JSON.parse(fetchCall[1].body)
        expect(body.event_id).toBe('fromButton')
    })
}) // End describe block

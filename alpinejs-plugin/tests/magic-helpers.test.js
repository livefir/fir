import Alpine from 'alpinejs'
import FirPlugin, { createFirMagicFunctions } from '../src/index'
import { initializeAlpineOnce } from './test-helpers' // Import from the helper file

describe('Fir Magic Helpers', () => {
    let container

    beforeEach(() => {
        // Set up a container for our tests
        container = document.createElement('div')
        document.body.appendChild(container)

        // Setup mock cookies - include both session and route cookies
        Object.defineProperty(document, 'cookie', {
            writable: true,
            value: '_fir_session_=test-session-id; _fir_route_=test-route-id',
        })

        // Initialize Alpine once
        initializeAlpineOnce()

        // Define a mock post for this test suite if needed, or use the one from test-helpers
        window.post = window.post || jest.fn()
    })

    afterEach(() => {
        document.body.removeChild(container)
        jest.clearAllMocks()
    })

    describe('reset() helper', () => {
        test('should call reset() on form elements', () => {
            // Setup
            const form = document.createElement('form')
            document.body.appendChild(form)

            // Spy on the form's reset method
            const resetSpy = jest
                .spyOn(form, 'reset')
                .mockImplementation(() => {})

            // Get the magic function directly
            const magicFunctions = createFirMagicFunctions(
                form,
                Alpine,
                window.post
            )
            const resetFn = magicFunctions.reset()

            // Call the function directly
            resetFn({})

            // Assert
            expect(resetSpy).toHaveBeenCalled()

            // Clean up
            document.body.removeChild(form)
        })

        test('should log error when used on non-form elements', () => {
            // Setup
            const div = document.createElement('div')
            document.body.appendChild(div)

            // Spy on console.error
            const consoleSpy = jest
                .spyOn(console, 'error')
                .mockImplementation(() => {})

            // Create the magic functions
            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                window.post
            )
            const resetFn = magicFunctions.reset()

            // Call the function directly
            resetFn({})

            // Assert
            expect(consoleSpy).toHaveBeenCalledWith(
                '$fir.reset() can only be used on form elements'
            )
        })
    })

    describe('toggleDisabled() helper', () => {
        test('should disable element on pending state event', () => {
            // Setup
            const button = document.createElement('button')
            document.body.appendChild(button)

            // Create the magic functions
            const magicFunctions = createFirMagicFunctions(
                button,
                Alpine,
                window.post
            )
            const toggleDisabledFn = magicFunctions.toggleDisabled()

            // Call the function with a pending event
            toggleDisabledFn({ type: 'fir:action:pending' })

            // Assert
            expect(button.hasAttribute('disabled')).toBe(true)
            expect(button.getAttribute('aria-disabled')).toBe('true')
        })

        test('should enable element on ok state event', () => {
            // Setup
            const button = document.createElement('button')
            button.setAttribute('disabled', '')
            button.setAttribute('aria-disabled', 'true')
            document.body.appendChild(button)

            // Create the magic functions
            const magicFunctions = createFirMagicFunctions(
                button,
                Alpine,
                window.post
            )
            const toggleDisabledFn = magicFunctions.toggleDisabled()

            // Call the function with an ok event
            toggleDisabledFn({ type: 'fir:action:ok' })

            // Assert
            expect(button.hasAttribute('disabled')).toBe(false)
            expect(button.hasAttribute('aria-disabled')).toBe(false)
        })

        test('should error on unsupported elements', () => {
            // Setup
            const div = document.createElement('div')
            document.body.appendChild(div)

            // Spy on console.error
            const consoleSpy = jest
                .spyOn(console, 'error')
                .mockImplementation(() => {})

            // Create the magic functions
            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                window.post
            )
            const toggleDisabledFn = magicFunctions.toggleDisabled()

            // Call the function with a pending event
            toggleDisabledFn({ type: 'fir:action:pending' })

            // Assert
            expect(consoleSpy).toHaveBeenCalledWith(
                '$fir.toggleDisabled() cannot be used on <div> elements'
            )
        })
    })

    describe('DOM Manipulation Edge Cases', () => {
        test('replaceEl should handle empty HTML string', () => {
            const el = document.createElement('div')
            el.id = 'original'
            el.innerHTML = '<span>Initial Content</span>' // Give it some content initially
            document.body.appendChild(el)
            const magicFunctions = createFirMagicFunctions(
                el,
                Alpine,
                jest.fn()
            )
            const replaceElFn = magicFunctions.replaceEl()
            const event = { detail: { html: '' } } // Empty HTML

            // Spy on morph to ensure it's NOT called
            const morphSpy = jest.spyOn(Alpine, 'morph')

            replaceElFn(event)

            // Assert that morph was NOT called for an empty string
            expect(morphSpy).not.toHaveBeenCalled()
            // Assert that the element's content was cleared
            expect(el.innerHTML).toBe('')

            morphSpy.mockRestore() // Restore the spy
        })

        test('appendEl should handle empty HTML string', () => {
            const el = document.createElement('div')
            el.innerHTML = '<span>Hi</span>'
            document.body.appendChild(el)
            const magicFunctions = createFirMagicFunctions(
                el,
                Alpine,
                jest.fn()
            )
            const appendElFn = magicFunctions.appendEl()
            const event = { detail: { html: '' } } // Empty HTML

            const initialHTML = el.innerHTML
            appendElFn(event)

            // Content should remain unchanged
            expect(el.innerHTML).toBe(initialHTML)
        })

        // Add similar tests for prependEl, afterEl, beforeEl with empty HTML
    })

    describe('toggleDisabled() edge cases', () => {
        test('should not change state for non-standard event states', () => {
            const button = document.createElement('button')
            document.body.appendChild(button)
            const magicFunctions = createFirMagicFunctions(
                button,
                Alpine,
                jest.fn()
            )
            const toggleDisabledFn = magicFunctions.toggleDisabled()

            // Call with an unknown state
            toggleDisabledFn({ type: 'fir:action:unknown' })
            expect(button.hasAttribute('disabled')).toBe(false)

            // Call with 'done' state (should enable)
            button.setAttribute('disabled', '')
            toggleDisabledFn({ type: 'fir:action:done' })
            expect(button.hasAttribute('disabled')).toBe(false)

            // Call with 'error' state (should enable)
            button.setAttribute('disabled', '')
            toggleDisabledFn({ type: 'fir:action:error' })
            expect(button.hasAttribute('disabled')).toBe(false)
        })
    })

    // You can add more tests for other magic helpers here
})

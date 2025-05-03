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
    })

    afterEach(() => {
        document.body.removeChild(container)
        jest.clearAllMocks()
    })

    describe('reset() helper', () => {
        test('should call reset() on form elements', async () => {
            // Setup - we'll use direct DOM element testing without Alpine
            const form = document.createElement('form')
            document.body.appendChild(form)

            // Spy on the form's reset method
            const resetSpy = jest
                .spyOn(form, 'reset')
                .mockImplementation(() => {})

            // Get the magic function directly (this is the most reliable approach)
            const magicFunctions = createFirMagicFunctions(form, Alpine)
            const resetFn = magicFunctions.reset()

            // Call the function directly (simulating what would happen when Alpine calls it)
            resetFn({})

            // Assert
            expect(resetSpy).toHaveBeenCalled()

            // Clean up
            document.body.removeChild(form)
        })

        test('should log error when used on non-form elements', () => {
            // Setup
            container.innerHTML = `
        <div x-data>
            <button id="resetBtn" @click="$fir.reset()">Reset</button>
        </div>
    `

            // Initialize Alpine for this element
            Alpine.plugin(FirPlugin)
            Alpine.initTree(container)

            const resetBtn = container.querySelector('#resetBtn')
            const consoleSpy = jest.spyOn(console, 'error')

            // Act - click the button directly
            resetBtn.click()

            // Assert
            expect(consoleSpy).toHaveBeenCalledWith(
                '$fir.reset() can only be used on form elements'
            )
        })

        test('should call reset() on form elements (direct test)', () => {
            // Create form element
            const form = document.createElement('form')

            // Spy on form.reset
            const resetSpy = jest
                .spyOn(form, 'reset')
                .mockImplementation(() => {})

            // Create the magic functions using real Alpine
            const magicFunctions = createFirMagicFunctions(form, Alpine)

            // Get the reset function
            const resetFn = magicFunctions.reset()

            // Call the function directly
            resetFn({})

            // Check that form.reset was called
            expect(resetSpy).toHaveBeenCalled()
        })

        test('should log error when used on non-form elements (direct test)', () => {
            // Create div element (not a form)
            const div = document.createElement('div')

            // Spy on console.error
            const consoleSpy = jest
                .spyOn(console, 'error')
                .mockImplementation(() => {})

            // Create the magic functions using real Alpine
            const magicFunctions = createFirMagicFunctions(div, Alpine)

            // Get the reset function
            const resetFn = magicFunctions.reset()

            // Call the function directly
            resetFn({})

            // Check that error was logged
            expect(consoleSpy).toHaveBeenCalledWith(
                '$fir.reset() can only be used on form elements'
            )

            // Restore original console.error
            consoleSpy.mockRestore()
        })
    })

    describe('toggleDisabled() helper', () => {
        test('should disable element on pending state event', () => {
            // Setup
            container.innerHTML = `
        <div x-data>
            <button id="testBtn">Test</button>
        </div>
      `

            const button = container.querySelector('#testBtn')

            // Create the magic functions directly
            const magicFunctions = createFirMagicFunctions(button, Alpine)

            // Get the toggleDisabled function
            const toggleDisabledFn = magicFunctions.toggleDisabled()

            // Call the function directly with a mock event
            toggleDisabledFn({ type: 'fir:action:pending' })

            // Assert
            expect(button.hasAttribute('disabled')).toBe(true)
            expect(button.getAttribute('aria-disabled')).toBe('true')
        })

        test('should enable element on ok state event', () => {
            // Setup
            container.innerHTML = `
        <div x-data>
            <button id="testBtn" disabled aria-disabled="true">Test</button>
        </div>
      `

            const button = container.querySelector('#testBtn')

            // Create the magic functions directly
            const magicFunctions = createFirMagicFunctions(button, Alpine)

            // Get the toggleDisabled function
            const toggleDisabledFn = magicFunctions.toggleDisabled()

            // Call the function directly with a mock event
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

            // Create the magic functions directly
            const magicFunctions = createFirMagicFunctions(div, Alpine)

            // Get the toggleDisabled function
            const toggleDisabledFn = magicFunctions.toggleDisabled()

            // Call the function with a mock event
            toggleDisabledFn({ type: 'fir:action:pending' })

            // Assert
            expect(consoleSpy).toHaveBeenCalledWith(
                '$fir.toggleDisabled() cannot be used on <div> elements'
            )

            // Clean up
            consoleSpy.mockRestore()
        })

        test('should disable element on pending state event (direct test)', () => {
            // Create button element
            const button = document.createElement('button')

            // Create the magic functions using real Alpine
            const magicFunctions = createFirMagicFunctions(button, Alpine)

            // Get the toggleDisabled function
            const toggleDisabledFn = magicFunctions.toggleDisabled()

            // Call the function with a pending event
            toggleDisabledFn({ type: 'fir:action:pending' })

            // Check that the button was disabled
            expect(button.hasAttribute('disabled')).toBe(true)
            expect(button.getAttribute('aria-disabled')).toBe('true')
        })

        test('should error on unsupported elements (direct test)', () => {
            // Create div element (doesn't support disabled)
            const div = document.createElement('div')

            // Spy on console.error
            const consoleSpy = jest
                .spyOn(console, 'error')
                .mockImplementation(() => {})

            // Create the magic functions using real Alpine
            const magicFunctions = createFirMagicFunctions(div, Alpine)

            // Get the toggleDisabled function
            const toggleDisabledFn = magicFunctions.toggleDisabled()

            // Call the function with a pending event
            toggleDisabledFn({ type: 'fir:action:pending' })

            // Check that error was logged with exact message
            expect(consoleSpy).toHaveBeenCalledWith(
                '$fir.toggleDisabled() cannot be used on <div> elements'
            )

            // Restore original console.error
            consoleSpy.mockRestore()
        })
    })

    // You can add more tests for other magic helpers here
})

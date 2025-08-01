import Alpine from 'alpinejs'
import FirPlugin, { createFirMagicFunctions } from '../src/magicFunctions'
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
        test('should toggle disabled state from enabled to disabled', () => {
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

            // Initial state should be enabled
            expect(button.hasAttribute('disabled')).toBe(false)

            // Call the function to toggle to disabled
            toggleDisabledFn({ type: 'fir:action:any' })

            // Assert it's now disabled
            expect(button.hasAttribute('disabled')).toBe(true)
            expect(button.getAttribute('aria-disabled')).toBe('true')
        })

        test('should toggle disabled state from disabled to enabled', () => {
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

            // Initial state should be disabled
            expect(button.hasAttribute('disabled')).toBe(true)

            // Call the function to toggle to enabled
            toggleDisabledFn({ type: 'fir:action:any' })

            // Assert it's now enabled
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
        test('should always toggle regardless of event type', () => {
            const button = document.createElement('button')
            document.body.appendChild(button)
            const magicFunctions = createFirMagicFunctions(
                button,
                Alpine,
                jest.fn()
            )
            const toggleDisabledFn = magicFunctions.toggleDisabled()

            // Initially enabled, should toggle to disabled
            expect(button.hasAttribute('disabled')).toBe(false)
            toggleDisabledFn({ type: 'fir:action:unknown' })
            expect(button.hasAttribute('disabled')).toBe(true)

            // Now disabled, should toggle to enabled regardless of event type
            toggleDisabledFn({ type: 'fir:action:done' })
            expect(button.hasAttribute('disabled')).toBe(false)

            // Toggle again with different event type
            toggleDisabledFn({ type: 'fir:action:error' })
            expect(button.hasAttribute('disabled')).toBe(true)
        })
    })

    describe('replace() helper', () => {
        test('should replace element content with morphing', () => {
            // Setup
            const div = document.createElement('div')
            div.innerHTML = '<span>Original</span>'
            document.body.appendChild(div)

            // Mock Alpine.morph
            const morphSpy = jest
                .spyOn(Alpine, 'morph')
                .mockImplementation(() => {})

            // Create the magic functions
            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                window.post
            )
            const replaceFn = magicFunctions.replace()

            // Call the function with new HTML
            replaceFn({ detail: { html: '<span>New Content</span>' } })

            // Assert that morph was called with correct parameters
            expect(morphSpy).toHaveBeenCalled()

            morphSpy.mockRestore()
        })

        test('should handle empty HTML in replace', () => {
            const div = document.createElement('div')
            div.innerHTML = '<span>Original</span>'
            document.body.appendChild(div)

            const morphSpy = jest
                .spyOn(Alpine, 'morph')
                .mockImplementation(() => {})

            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                window.post
            )
            const replaceFn = magicFunctions.replace()

            replaceFn({ detail: { html: '' } })

            expect(morphSpy).toHaveBeenCalled()

            morphSpy.mockRestore()
        })
    })

    describe('prependEl() helper', () => {
        test('should prepend HTML to element', () => {
            // Setup
            const div = document.createElement('div')
            div.innerHTML = '<span>Existing</span>'
            document.body.appendChild(div)

            // Create the magic functions
            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                window.post
            )
            const prependElFn = magicFunctions.prependEl()

            // Call the function with new HTML
            prependElFn({ detail: { html: '<p>Prepended</p>' } })

            // Assert that content was prepended
            expect(div.innerHTML).toBe('<p>Prepended</p><span>Existing</span>')
        })

        test('should handle empty HTML in prependEl', () => {
            const div = document.createElement('div')
            div.innerHTML = '<span>Existing</span>'
            document.body.appendChild(div)

            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                window.post
            )
            const prependElFn = magicFunctions.prependEl()

            const initialHTML = div.innerHTML
            prependElFn({ detail: { html: '' } })

            // Content should remain unchanged
            expect(div.innerHTML).toBe(initialHTML)
        })
    })

    describe('afterEl() helper', () => {
        test('should insert HTML after element', () => {
            // Setup
            const container = document.createElement('div')
            const target = document.createElement('span')
            target.textContent = 'Target'
            container.appendChild(target)
            document.body.appendChild(container)

            // Create the magic functions
            const magicFunctions = createFirMagicFunctions(
                target,
                Alpine,
                window.post
            )
            const afterElFn = magicFunctions.afterEl()

            // Call the function with new HTML
            afterElFn({ detail: { html: '<p>After</p>' } })

            // Assert that content was inserted after
            expect(container.innerHTML).toBe('<span>Target</span><p>After</p>')
        })

        test('should handle empty HTML in afterEl', () => {
            const container = document.createElement('div')
            const target = document.createElement('span')
            target.textContent = 'Target'
            container.appendChild(target)
            document.body.appendChild(container)

            const magicFunctions = createFirMagicFunctions(
                target,
                Alpine,
                window.post
            )
            const afterElFn = magicFunctions.afterEl()

            const initialHTML = container.innerHTML
            afterElFn({ detail: { html: '' } })

            // Content should remain unchanged
            expect(container.innerHTML).toBe(initialHTML)
        })
    })

    describe('beforeEl() helper', () => {
        test('should insert HTML before element', () => {
            // Setup
            const container = document.createElement('div')
            const target = document.createElement('span')
            target.textContent = 'Target'
            container.appendChild(target)
            document.body.appendChild(container)

            // Create the magic functions
            const magicFunctions = createFirMagicFunctions(
                target,
                Alpine,
                window.post
            )
            const beforeElFn = magicFunctions.beforeEl()

            // Call the function with new HTML
            beforeElFn({ detail: { html: '<p>Before</p>' } })

            // Assert that content was inserted before
            expect(container.innerHTML).toBe('<p>Before</p><span>Target</span>')
        })

        test('should handle empty HTML in beforeEl', () => {
            const container = document.createElement('div')
            const target = document.createElement('span')
            target.textContent = 'Target'
            container.appendChild(target)
            document.body.appendChild(container)

            const magicFunctions = createFirMagicFunctions(
                target,
                Alpine,
                window.post
            )
            const beforeElFn = magicFunctions.beforeEl()

            const initialHTML = container.innerHTML
            beforeElFn({ detail: { html: '' } })

            // Content should remain unchanged
            expect(container.innerHTML).toBe(initialHTML)
        })
    })

    describe('removeEl() helper', () => {
        test('should remove the element from DOM', () => {
            // Setup
            const container = document.createElement('div')
            const target = document.createElement('span')
            target.textContent = 'To Remove'
            container.appendChild(target)
            document.body.appendChild(container)

            // Create the magic functions
            const magicFunctions = createFirMagicFunctions(
                target,
                Alpine,
                window.post
            )
            const removeElFn = magicFunctions.removeEl()

            // Verify element exists initially
            expect(container.contains(target)).toBe(true)

            // Call the function
            removeElFn({})

            // Assert that element was removed
            expect(container.contains(target)).toBe(false)
            expect(container.innerHTML).toBe('')
        })
    })

    describe('removeParentEl() helper', () => {
        test('should remove the parent element from DOM', () => {
            // Setup
            const grandparent = document.createElement('div')
            const parent = document.createElement('div')
            const target = document.createElement('span')
            target.textContent = 'Target'

            parent.appendChild(target)
            grandparent.appendChild(parent)
            document.body.appendChild(grandparent)

            // Create the magic functions
            const magicFunctions = createFirMagicFunctions(
                target,
                Alpine,
                window.post
            )
            const removeParentElFn = magicFunctions.removeParentEl()

            // Verify parent exists initially
            expect(grandparent.contains(parent)).toBe(true)

            // Call the function
            removeParentElFn({})

            // Assert that parent was removed
            expect(grandparent.contains(parent)).toBe(false)
            expect(grandparent.innerHTML).toBe('')
        })

        test('should handle element with no parent gracefully', () => {
            const target = document.createElement('span')

            const magicFunctions = createFirMagicFunctions(
                target,
                Alpine,
                window.post
            )
            const removeParentElFn = magicFunctions.removeParentEl()

            // Should not throw error when element has no parent
            expect(() => removeParentElFn({})).not.toThrow()
        })
    })

    describe('toggleClass() helper', () => {
        test('should toggle classes when element does not have them', () => {
            // Setup
            const div = document.createElement('div')
            document.body.appendChild(div)

            // Create the magic functions
            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                window.post
            )
            const toggleClassFn = magicFunctions.toggleClass(
                'loading',
                'active'
            )

            // Initial state - no classes
            expect(div.classList.contains('loading')).toBe(false)
            expect(div.classList.contains('active')).toBe(false)

            // Call toggle function
            toggleClassFn({ type: 'fir:action:any' })

            // Assert classes were added
            expect(div.classList.contains('loading')).toBe(true)
            expect(div.classList.contains('active')).toBe(true)
        })

        test('should toggle classes when element already has them', () => {
            // Setup
            const div = document.createElement('div')
            div.classList.add('loading', 'active')
            document.body.appendChild(div)

            // Create the magic functions
            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                window.post
            )
            const toggleClassFn = magicFunctions.toggleClass(
                'loading',
                'active'
            )

            // Initial state - has classes
            expect(div.classList.contains('loading')).toBe(true)
            expect(div.classList.contains('active')).toBe(true)

            // Call toggle function
            toggleClassFn({ type: 'fir:action:any' })

            // Assert classes were removed
            expect(div.classList.contains('loading')).toBe(false)
            expect(div.classList.contains('active')).toBe(false)
        })

        test('should toggle mixed class states individually', () => {
            const div = document.createElement('div')
            div.classList.add('loading') // Has 'loading', doesn't have 'active'
            document.body.appendChild(div)

            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                window.post
            )
            const toggleClassFn = magicFunctions.toggleClass(
                'loading',
                'active'
            )

            // Initial state
            expect(div.classList.contains('loading')).toBe(true)
            expect(div.classList.contains('active')).toBe(false)

            toggleClassFn({ type: 'fir:action:any' })

            // 'loading' should be removed (was present), 'active' should be added (was absent)
            expect(div.classList.contains('loading')).toBe(false)
            expect(div.classList.contains('active')).toBe(true)
        })

        test('should handle no class names with error', () => {
            const div = document.createElement('div')
            document.body.appendChild(div)

            const consoleSpy = jest
                .spyOn(console, 'error')
                .mockImplementation(() => {})

            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                window.post
            )
            const toggleClassFn = magicFunctions.toggleClass()

            toggleClassFn({ type: 'fir:action:any' })

            expect(consoleSpy).toHaveBeenCalledWith(
                '$fir.toggleClass() requires at least one class name'
            )

            consoleSpy.mockRestore()
        })

        test('should handle non-string class names with error', () => {
            const div = document.createElement('div')
            document.body.appendChild(div)

            const consoleSpy = jest
                .spyOn(console, 'error')
                .mockImplementation(() => {})

            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                window.post
            )
            const toggleClassFn = magicFunctions.toggleClass(
                'valid',
                123,
                'another'
            )

            toggleClassFn({ type: 'fir:action:any' })

            expect(consoleSpy).toHaveBeenCalledWith(
                'Class name must be a string, got: number'
            )

            consoleSpy.mockRestore()
        })
    })

    describe('redirect() helper', () => {
        test('should redirect to specified URL', () => {
            // Mock window.location.href
            delete window.location
            window.location = { href: '' }

            const div = document.createElement('div')
            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                window.post
            )
            const redirectFn = magicFunctions.redirect('/dashboard')

            redirectFn({})

            expect(window.location.href).toBe('/dashboard')
        })

        test('should redirect to default URL when no parameter provided', () => {
            delete window.location
            window.location = { href: '' }

            const div = document.createElement('div')
            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                window.post
            )
            const redirectFn = magicFunctions.redirect()

            redirectFn({})

            expect(window.location.href).toBe('/')
        })

        test('should use URL from event detail if provided', () => {
            delete window.location
            window.location = { href: '' }

            const div = document.createElement('div')
            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                window.post
            )
            const redirectFn = magicFunctions.redirect('/default')

            redirectFn({ detail: { url: '/from-event' } })

            expect(window.location.href).toBe('/from-event')
        })

        test('should handle invalid URL with error', () => {
            const consoleSpy = jest
                .spyOn(console, 'error')
                .mockImplementation(() => {})

            const div = document.createElement('div')
            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                window.post
            )
            const redirectFn = magicFunctions.redirect(123) // Invalid URL

            redirectFn({})

            expect(consoleSpy).toHaveBeenCalledWith(
                '$fir.redirect() requires a valid URL string'
            )

            consoleSpy.mockRestore()
        })

        test('should fallback to default URL when event detail URL is null', () => {
            delete window.location
            window.location = { href: '' }

            const div = document.createElement('div')
            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                window.post
            )
            const redirectFn = magicFunctions.redirect('/default')

            redirectFn({ detail: { url: null } }) // null URL from event should fallback

            expect(window.location.href).toBe('/default')
        })
    })

    // You can add more tests for other magic helpers here
})

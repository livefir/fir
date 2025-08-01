import Alpine from 'alpinejs'
import morph from '@alpinejs/morph'
import { createFirMagicFunctions } from '../src/magicFunctions'

// Don't mock Alpine or morph
jest.unmock('alpinejs')
jest.unmock('@alpinejs/morph')

describe('Direct Tests for createFirMagicFunctions', () => {
    // Setup for each test
    beforeEach(() => {
        document.body.innerHTML = ''

        // Setup document cookie for testing
        Object.defineProperty(document, 'cookie', {
            writable: true,
            value: '_fir_session_=test-session-id',
        })

        // Setup global post function
        global.post = jest.fn()

        // Initialize Alpine with the morph plugin
        Alpine.plugin(morph)
    })

    afterEach(() => {
        jest.clearAllMocks()
    })

    describe('replace()', () => {
        test('should replace element content when event has HTML', () => {
            // Setup
            const el = document.createElement('div')
            el.innerHTML = 'Original content'
            document.body.appendChild(el)

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                el,
                Alpine,
                global.post
            )
            const replaceFn = magicFunctions.replace()

            // Create mock event with HTML content
            const event = {
                detail: {
                    html: '<span>New content</span>',
                },
            }

            // Act
            replaceFn(event)

            // Assert - should contain the new content
            expect(el.textContent).toContain('New content')
        })
    })

    describe('replaceEl()', () => {
        test('should replace entire element with event HTML', () => {
            // Setup
            const container = document.createElement('div')
            const el = document.createElement('div')
            el.id = 'original'
            el.innerHTML = 'Original element'
            container.appendChild(el)
            document.body.appendChild(container)

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                el,
                Alpine,
                global.post
            )
            const replaceElFn = magicFunctions.replaceEl()

            // Create mock event with HTML content that includes an ID
            const event = {
                detail: {
                    html: '<span id="replaced">Replaced element</span>',
                },
            }

            // Act
            replaceElFn(event)

            // Assert - original should be gone and new one should exist
            expect(document.getElementById('original')).toBeNull()
            expect(document.getElementById('replaced')).not.toBeNull()
        })
    })

    describe('appendEl()', () => {
        test('should append HTML to element', () => {
            // Setup
            const el = document.createElement('div')
            el.innerHTML = 'Original content'
            document.body.appendChild(el)

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                el,
                Alpine,
                global.post
            )
            const appendElFn = magicFunctions.appendEl()

            // Create mock event with HTML content
            const event = {
                detail: {
                    html: '<span class="appended">Appended content</span>',
                },
            }

            // Act
            appendElFn(event)

            // Assert
            expect(el.textContent).toContain('Original content')
            expect(el.textContent).toContain('Appended content')
            expect(el.querySelector('.appended')).not.toBeNull()
        })
    })

    describe('prependEl()', () => {
        test('should prepend HTML to element', () => {
            // Setup
            const el = document.createElement('div')
            el.innerHTML = 'Original content'
            document.body.appendChild(el)

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                el,
                Alpine,
                global.post
            )
            const prependElFn = magicFunctions.prependEl()

            // Create mock event with HTML content
            const event = {
                detail: {
                    html: '<span class="prepended">Prepended content</span>',
                },
            }

            // Act
            prependElFn(event)

            // Assert
            expect(el.textContent).toContain('Original content')
            expect(el.textContent).toContain('Prepended content')
            expect(el.querySelector('.prepended')).not.toBeNull()
        })
    })

    describe('afterEl()', () => {
        test('should insert HTML after element', () => {
            // Setup - create the container first
            const container = document.createElement('div')
            document.body.appendChild(container)

            // Then create the test element inside the container
            const el = document.createElement('div')
            el.innerHTML = 'Original content'
            container.appendChild(el)

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                el,
                Alpine,
                global.post
            )
            const afterElFn = magicFunctions.afterEl()

            // Create mock event with HTML content
            const event = {
                detail: {
                    html: '<span class="after">After content</span>',
                },
            }

            // Act - this will use the real Alpine.morph internally
            afterElFn(event)

            // Assert - find the element by class
            const afterElement = container.querySelector('.after')
            expect(afterElement).not.toBeNull()
            expect(afterElement.textContent).toBe('After content')

            // Clean up
            document.body.removeChild(container)
        })
    })

    describe('beforeEl()', () => {
        test('should insert HTML before element', () => {
            // Create a container first to provide proper parent context
            const container = document.createElement('div')
            document.body.appendChild(container)

            // Then create the test element inside the container
            const el = document.createElement('div')
            el.innerHTML = 'Original content'
            container.appendChild(el)

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                el,
                Alpine,
                global.post
            )
            const beforeElFn = magicFunctions.beforeEl()

            // Create a mock event with properly structured HTML
            const event = {
                detail: {
                    html: '<span class="before">Before content</span>',
                },
            }

            // Act - this will use the real Alpine.morph internally
            beforeElFn(event)

            // Assert - find the element by class
            const beforeElement = container.querySelector('.before')
            expect(beforeElement).not.toBeNull()
            expect(beforeElement.textContent).toBe('Before content')

            // Clean up
            document.body.removeChild(container)
        })
    })

    describe('removeEl()', () => {
        test('should remove element from DOM', () => {
            // Setup
            const container = document.createElement('div')
            const el = document.createElement('div')
            el.id = 'element-to-remove'
            container.appendChild(el)
            document.body.appendChild(container)

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                el,
                Alpine,
                global.post
            )
            const removeElFn = magicFunctions.removeEl()

            // Act
            removeElFn({})

            // Assert
            expect(document.getElementById('element-to-remove')).toBeNull()
        })
    })

    describe('removeParentEl()', () => {
        test('should remove parent element from DOM', () => {
            // Setup
            const grandparent = document.createElement('div')
            grandparent.id = 'grandparent'

            const parent = document.createElement('div')
            parent.id = 'parent'
            grandparent.appendChild(parent)

            const el = document.createElement('div')
            el.id = 'child'
            parent.appendChild(el)

            document.body.appendChild(grandparent)

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                el,
                Alpine,
                global.post
            )
            const removeParentElFn = magicFunctions.removeParentEl()

            // Act
            removeParentElFn({})

            // Assert
            expect(document.getElementById('parent')).toBeNull()
            expect(document.getElementById('child')).toBeNull()
            expect(document.getElementById('grandparent')).not.toBeNull()
        })
    })

    describe('reset()', () => {
        test('should reset a form element', () => {
            // Setup
            const form = document.createElement('form')
            const input = document.createElement('input')
            input.name = 'test'
            input.value = 'test value'
            form.appendChild(input)
            document.body.appendChild(form)

            form.reset = jest.fn()

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                form,
                Alpine,
                global.post
            )
            const resetFn = magicFunctions.reset()

            // Act
            resetFn({})

            // Assert
            expect(form.reset).toHaveBeenCalled()
        })

        test('should log error when used on non-form element', () => {
            // Setup
            const div = document.createElement('div')
            document.body.appendChild(div)

            console.error = jest.fn()

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                global.post
            )
            const resetFn = magicFunctions.reset()

            // Act
            resetFn({})

            // Assert
            expect(console.error).toHaveBeenCalledWith(
                '$fir.reset() can only be used on form elements'
            )
        })
    })

    describe('toggleDisabled()', () => {
        test('should toggle disabled state from enabled to disabled', () => {
            // Setup
            const button = document.createElement('button')
            document.body.appendChild(button)

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                button,
                Alpine,
                global.post
            )
            const toggleDisabledFn = magicFunctions.toggleDisabled()

            // Initial state should be enabled
            expect(button.hasAttribute('disabled')).toBe(false)

            // Act
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

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                button,
                Alpine,
                global.post
            )
            const toggleDisabledFn = magicFunctions.toggleDisabled()

            // Initial state should be disabled
            expect(button.hasAttribute('disabled')).toBe(true)

            // Act
            toggleDisabledFn({ type: 'fir:action:any' })

            // Assert it's now enabled
            expect(button.hasAttribute('disabled')).toBe(false)
            expect(button.hasAttribute('aria-disabled')).toBe(false)
        })

        test('should log error when used on unsupported element', () => {
            // Setup
            const div = document.createElement('div')
            document.body.appendChild(div)

            console.error = jest.fn()

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                div,
                Alpine,
                global.post
            )
            const toggleDisabledFn = magicFunctions.toggleDisabled()

            // Act
            toggleDisabledFn({ type: 'fir:action:any' })

            // Assert
            expect(console.error).toHaveBeenCalledWith(
                '$fir.toggleDisabled() cannot be used on <div> elements'
            )
        })
    })

    describe('emit()', () => {
        test('should call post with correct parameters', () => {
            // Setup
            const el = document.createElement('div')
            el.id = 'test-id'
            document.body.appendChild(el)

            global.post = jest.fn()

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                el,
                Alpine,
                global.post
            )
            const emitFn = magicFunctions.emit(
                'test-event',
                { foo: 'bar' },
                '#target'
            )

            // Act
            emitFn({})

            // Assert
            expect(global.post).toHaveBeenCalledWith({
                event_id: 'test-event',
                params: { foo: 'bar' },
                target: '#target',
                element_key: null,
                session_id: 'test-session-id',
            })
        })

        test('should use element ID when no event ID provided', () => {
            // Setup
            const el = document.createElement('div')
            el.id = 'test-id'
            document.body.appendChild(el)

            global.post = jest.fn()

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                el,
                Alpine,
                global.post
            )
            const emitFn = magicFunctions.emit()

            // Act
            emitFn({})

            // Assert
            expect(global.post).toHaveBeenCalledWith(
                expect.objectContaining({
                    event_id: 'test-id',
                })
            )
        })

        test('should validate parameters', () => {
            // Setup
            const el = document.createElement('div')
            el.id = 'test-id'
            document.body.appendChild(el)

            console.error = jest.fn()
            global.post = jest.fn()

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                el,
                Alpine,
                global.post
            )

            // Test invalid ID
            const emitWithInvalidId = magicFunctions.emit(123)
            emitWithInvalidId({})
            expect(console.error).toHaveBeenCalledWith(
                'id 123 is not a string.'
            )

            // Test invalid params
            console.error.mockClear()
            const emitWithInvalidParams = magicFunctions.emit(
                'test-event',
                'not-an-object'
            )
            emitWithInvalidParams({})
            expect(console.error).toHaveBeenCalledWith(
                'params not-an-object is not an object.'
            )

            // Test invalid target
            console.error.mockClear()
            const emitWithInvalidTarget = magicFunctions.emit(
                'test-event',
                {},
                'invalid-target'
            )
            emitWithInvalidTarget({})
            expect(console.error).toHaveBeenCalledWith(
                'target must start with # or .'
            )
        })
    })

    describe('submit()', () => {
        test('should gather form data and submit', () => {
            // Setup
            const form = document.createElement('form')
            form.id = 'test-form'
            form.method = 'post'
            form.action = '/?event=test-submit'

            const input = document.createElement('input')
            input.name = 'testInput'
            input.value = 'testValue'
            form.appendChild(input)

            document.body.appendChild(form)

            global.post = jest.fn()

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                form,
                Alpine,
                global.post
            )
            const submitFn = magicFunctions.submit()

            // Create a submit event
            const submitEvent = new Event('submit', {
                bubbles: true,
                cancelable: true,
            })

            // Act
            submitFn(submitEvent)

            // Assert
            expect(global.post).toHaveBeenCalledWith(
                expect.objectContaining({
                    event_id: 'test-submit',
                    is_form: true,
                    params: expect.any(Object), // Changed from form_data to params
                })
            )

            // Clean up
            document.body.removeChild(form)
        })

        test('should extract event ID from form action URL', () => {
            // Setup
            const form = document.createElement('form')
            form.method = 'post'
            form.action = 'https://example.com/?event=from-url'
            document.body.appendChild(form)

            global.post = jest.fn()

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                form,
                Alpine,
                global.post
            )
            const submitFn = magicFunctions.submit()

            // Create a submit event
            const submitEvent = new Event('submit', {
                bubbles: true,
                cancelable: true,
            })

            // Act
            submitFn(submitEvent)

            // Assert
            expect(global.post).toHaveBeenCalledWith(
                expect.objectContaining({
                    event_id: 'from-url',
                })
            )
        })

        test('should override with options if provided', () => {
            // Setup
            const form = document.createElement('form')
            form.id = 'test-form'
            document.body.appendChild(form)

            global.post = jest.fn()

            // Use real Alpine instance
            const magicFunctions = createFirMagicFunctions(
                form,
                Alpine,
                global.post
            )
            const submitFn = magicFunctions.submit({
                event: 'override-event',
                params: { extra: 'param' },
            })

            // Create a submit event
            const submitEvent = new Event('submit', {
                bubbles: true,
                cancelable: true,
            })

            // Act
            submitFn(submitEvent)

            // Assert
            expect(global.post).toHaveBeenCalledWith(
                expect.objectContaining({
                    event_id: 'override-event',
                    params: expect.objectContaining({
                        extra: 'param',
                    }),
                })
            )
        })
    })
})

import {
    eventHTML,
    toElements,
    toElement,
    morphElement,
    afterElement,
    beforeElement,
    appendElement,
    prependElement,
    removeElement,
    removeParentElement,
} from '../src/domUtils'

// Helper to set up a basic DOM environment for tests
const setupDOM = (html = '') => {
    document.body.innerHTML = html
    return document.body
}

describe('domUtils', () => {
    let originalConsoleError
    let originalConsoleLog

    beforeEach(() => {
        originalConsoleError = console.error
        originalConsoleLog = console.log
        console.error = jest.fn()
        console.log = jest.fn()
        // Clear DOM before each test
        document.body.innerHTML = ''
    })

    afterEach(() => {
        console.error = originalConsoleError
        console.log = originalConsoleLog
        jest.clearAllMocks()
    })

    describe('eventHTML', () => {
        test('should extract HTML from event.detail.html', () => {
            const event = { detail: { html: '<p>Test</p>' } }
            expect(eventHTML(event)).toBe('<p>Test</p>')
        })

        test('should return empty string if event.detail.html is missing', () => {
            const event = { detail: { data: 'something' } }
            expect(eventHTML(event)).toBe('')
        })

        test('should return empty string if event.detail.html is null or undefined', () => {
            expect(eventHTML({ detail: { html: null } })).toBe('')
            expect(eventHTML({ detail: { html: undefined } })).toBe('')
        })

        test('should return empty string if event.detail is missing', () => {
            const event = {}
            expect(eventHTML(event)).toBe('')
        })

        test('should return empty string if event is null or undefined', () => {
            expect(eventHTML(null)).toBe('')
            expect(eventHTML(undefined)).toBe('')
        })

        test('should convert non-string event.detail.html to string', () => {
            expect(eventHTML({ detail: { html: 123 } })).toBe('123')
            expect(eventHTML({ detail: { html: true } })).toBe('true')
        })
    })

    describe('toElements', () => {
        test('should convert HTML string to NodeList', () => {
            const html = '<span>Hello</span><p>World</p>'
            const nodes = toElements(html)
            expect(nodes.length).toBe(2)
            expect(nodes[0].tagName).toBe('SPAN')
            expect(nodes[1].tagName).toBe('P')
        })

        test('should return empty NodeList for empty string', () => {
            const nodes = toElements('')
            expect(nodes.length).toBe(0)
        })
    })

    describe('toElement', () => {
        test('should convert HTML string to a single Node', () => {
            const html = '<div>Hello</div>'
            const node = toElement(html)
            expect(node.tagName).toBe('DIV')
            expect(node.textContent).toBe('Hello')
        })

        test('should return first Node if multiple top-level elements', () => {
            const html = '<span>First</span><div>Second</div>'
            const node = toElement(html)
            expect(node.tagName).toBe('SPAN')
        })

        test('should return null for empty string', () => {
            const node = toElement('')
            expect(node).toBeNull()
        })

        test('should return a text node if HTML is just text', () => {
            const node = toElement('Just text')
            expect(node.nodeType).toBe(Node.TEXT_NODE)
            expect(node.textContent).toBe('Just text')
        })
    })

    describe('morphElement', () => {
        let mockAlpine
        let targetElement

        beforeEach(() => {
            mockAlpine = {
                morph: jest.fn(),
            }
            setupDOM('<div id="target">Initial</div>')
            targetElement = document.getElementById('target')
        })

        test('should call Alpine.morph with element and value', () => {
            const newValue = '<p>New Content</p>'
            morphElement(targetElement, newValue, mockAlpine)
            expect(mockAlpine.morph).toHaveBeenCalledWith(
                targetElement,
                newValue,
                expect.any(Object)
            )
        })

        test('should set innerHTML to empty if value is empty string', () => {
            morphElement(targetElement, '', mockAlpine)
            expect(targetElement.innerHTML).toBe('')
            expect(mockAlpine.morph).not.toHaveBeenCalled()
        })

        test('should log error and not call morph if value is null', () => {
            morphElement(targetElement, null, mockAlpine)
            expect(console.error).toHaveBeenCalledWith(
                'morph value is null or undefined'
            )
            expect(mockAlpine.morph).not.toHaveBeenCalled()
        })

        test('should log error and not call morph if value is undefined', () => {
            morphElement(targetElement, undefined, mockAlpine)
            expect(console.error).toHaveBeenCalledWith(
                'morph value is null or undefined'
            )
            expect(mockAlpine.morph).not.toHaveBeenCalled()
        })

        test('should use fir-key in morph options', () => {
            targetElement.setAttribute('fir-key', 'myKey')
            const newValue = '<p>New</p>'
            morphElement(targetElement, newValue, mockAlpine)
            const morphOptions = mockAlpine.morph.mock.calls[0][2]
            const keyFn = morphOptions.key
            expect(keyFn(targetElement)).toBe('myKey')
            const divWithoutKey = document.createElement('div')
            expect(keyFn(divWithoutKey)).toBeNull()
        })
    })

    describe('afterElement', () => {
        test('should insert HTML after the element', () => {
            setupDOM(
                // Call setupDOM to set document.body.innerHTML
                '<div id="parent"><span id="ref">Ref</span></div>'
            )
            const actualParentDiv = document.getElementById('parent') // Get the actual parent div
            const refEl = document.getElementById('ref')
            afterElement(refEl, '<p>New</p>')
            expect(actualParentDiv.innerHTML).toBe(
                '<span id="ref">Ref</span><p>New</p>'
            )
        })

        test('should log error if element has no parent', () => {
            const el = document.createElement('span') // No parent
            afterElement(el, '<p>New</p>')
            expect(console.error).toHaveBeenCalledWith(
                'Element has no parent, cannot insert after'
            )
        })
    })

    describe('beforeElement', () => {
        test('should insert HTML before the element', () => {
            setupDOM(
                // Call setupDOM to set document.body.innerHTML
                '<div id="parent"><span id="ref">Ref</span></div>'
            )
            const actualParentDiv = document.getElementById('parent') // Get the actual parent div
            const refEl = document.getElementById('ref')
            beforeElement(refEl, '<p>New</p>')
            expect(actualParentDiv.innerHTML).toBe(
                '<p>New</p><span id="ref">Ref</span>'
            )
        })

        test('should log error if element has no parent', () => {
            const el = document.createElement('span') // No parent
            beforeElement(el, '<p>New</p>')
            expect(console.error).toHaveBeenCalledWith(
                'Element has no parent, cannot insert before'
            )
        })
    })

    describe('appendElement', () => {
        let mockAlpine
        let targetElement

        beforeEach(() => {
            mockAlpine = {
                morph: jest.fn((el, newEl) => {
                    // Simple mock for morph
                    el.innerHTML = newEl.innerHTML
                }),
            }
            setupDOM('<div id="target">Initial</div>')
            targetElement = document.getElementById('target')
        })

        test('should append HTML to the element and morph', () => {
            appendElement(targetElement, '<p>Appended</p>', mockAlpine)
            // Check based on the simple morph mock
            expect(targetElement.innerHTML).toContain('Initial')
            expect(targetElement.innerHTML).toContain('<p>Appended</p>')
            // Verify morphElement (which calls Alpine.morph) was involved indirectly
            // This requires morphElement to be non-mocked or spied upon if we want to check its call.
            // For simplicity, we check the outcome.
        })
    })

    describe('prependElement', () => {
        let mockAlpine
        let targetElement

        beforeEach(() => {
            mockAlpine = {
                morph: jest.fn((el, newEl) => {
                    // Simple mock for morph
                    el.innerHTML = newEl.innerHTML
                }),
            }
            setupDOM('<div id="target">Initial</div>')
            targetElement = document.getElementById('target')
        })

        test('should prepend HTML to the element and morph', () => {
            prependElement(targetElement, '<p>Prepended</p>', mockAlpine)
            expect(targetElement.innerHTML).toContain('Initial')
            expect(targetElement.innerHTML).toContain('<p>Prepended</p>')
            expect(targetElement.innerHTML.startsWith('<p>Prepended</p>')).toBe(
                true
            )
        })
    })

    describe('removeElement', () => {
        test('should remove the element from DOM', () => {
            setupDOM('<div id="parent"><span id="child">Child</span></div>')
            const childEl = document.getElementById('child')
            expect(document.getElementById('child')).not.toBeNull()
            removeElement(childEl)
            expect(document.getElementById('child')).toBeNull()
        })
    })

    describe('removeParentElement', () => {
        test('should remove the parent element from DOM', () => {
            setupDOM(
                '<div id="grandparent"><div id="parent"><span id="child">Child</span></div></div>'
            )
            const childEl = document.getElementById('child')
            expect(document.getElementById('parent')).not.toBeNull()
            removeParentElement(childEl)
            expect(document.getElementById('parent')).toBeNull()
            expect(document.getElementById('grandparent')).not.toBeNull() // Grandparent should still exist
        })

        test('should log error if element has no parent', () => {
            const el = document.createElement('span') // No parent
            removeParentElement(el)
            expect(console.error).toHaveBeenCalledWith(
                'Element has no parent element to remove.'
            )
        })
    })
})

// ... other imports ...

// REMOVE or COMMENT OUT the global jest.mock if its sole purpose was to mock dispatchEventOnElement
// for the tests of dispatchEventOnIdTarget/ClassTarget.
// If it's used for other functions (like dispatchSingleServerEvent's old mocking style),
// you might need to adjust it or refactor those tests too.
/*
jest.mock('../src/eventDispatcher', () => {
    const originalModule = jest.requireActual('../src/eventDispatcher');
    return {
        __esModule: true,
        ...originalModule,
        // dispatchEventOnElement: jest.fn(), // This was the problematic part for DI
    };
});
*/

// Get the actual module before mocking
const actualEventDispatcher = jest.requireActual('../src/eventDispatcher')

jest.mock('../src/eventDispatcher', () => {
    const original = jest.requireActual('../src/eventDispatcher')
    return {
        ...original,
        generateFirEventNames: jest.fn(), // Inline mock
    }
})

// Import from the (now partially) mocked module.
// generateFirEventNames will be the jest.fn() created in the mock factory.
// Other functions will be the originals.
import {
    FIR_PREFIX,
    ID_SELECTOR_PREFIX,
    CLASS_SELECTOR_PREFIX,
    createCustomEvent,
    isListenableElement,
    dispatchEventOnIdTarget,
    dispatchEventOnClassTarget,
    isValidFirEvent,
    dispatchSingleServerEvent,
    processAndDispatchServerEvents,
    postEvent,
    generateFirEventNames as mockGenerateFirEventNames, // This is the mock function
} from '../src/eventDispatcher'

// Helper to set up DOM elements for testing
const setupDOM = (html) => {
    document.body.innerHTML = html
}

describe('eventDispatcher', () => {
    let originalConsoleWarn
    let originalConsoleError
    let originalConsoleLog

    beforeEach(() => {
        // Mock console methods
        originalConsoleWarn = console.warn
        originalConsoleError = console.error
        originalConsoleLog = console.log
        console.warn = jest.fn()
        console.error = jest.fn()
        console.log = jest.fn()
    })

    afterEach(() => {
        // Restore console methods
        console.warn = originalConsoleWarn
        console.error = originalConsoleError
        console.log = originalConsoleLog
        // Clean up DOM
        document.body.innerHTML = ''
        jest.clearAllMocks()
    })

    // ... tests for createCustomEvent, isListenableElement, dispatchEventOnElement ...
    // ... dispatchEventOnIdTarget, dispatchEventOnClassTarget, isValidFirEvent ...
    // ... dispatchSingleServerEvent, processAndDispatchServerEvents ...
    // (These tests remain the same as previously provided)

    describe('createCustomEvent', () => {
        test('should create a CustomEvent with correct type and detail', () => {
            const type = 'test:event'
            const detail = { data: 'payload' }
            const event = createCustomEvent(type, detail)

            expect(event).toBeInstanceOf(CustomEvent)
            expect(event.type).toBe(type)
            expect(event.detail).toEqual(detail)
            expect(event.bubbles).toBe(true)
            expect(event.composed).toBe(true)
            expect(event.cancelable).toBe(true)
        })
    })

    describe('isListenableElement', () => {
        const eventType = 'fir:myevent'
        beforeEach(() => {
            setupDOM(`
                <div id="el1" @${eventType}="handle"></div>
                <div id="el2" x-on:${eventType}="handle"></div>
                <div id="el3" @${eventType}="handle" x-on:${eventType}="handleAlso"></div>
                <div id="el4"></div>
                <div id="el5" @fir:otherevent="handle"></div>
            `)
        })

        test('should return true for element with @event listener', () => {
            const el = document.getElementById('el1')
            expect(isListenableElement(el, eventType)).toBe(true)
        })

        test('should return true for element with x-on:event listener', () => {
            const el = document.getElementById('el2')
            expect(isListenableElement(el, eventType)).toBe(true)
        })

        test('should return true for element with both listeners', () => {
            const el = document.getElementById('el3')
            expect(isListenableElement(el, eventType)).toBe(true)
        })

        test('should return false for element with no matching listeners', () => {
            const el = document.getElementById('el4')
            expect(isListenableElement(el, eventType)).toBe(false)
        })

        test('should return false for element with different event listener', () => {
            const el = document.getElementById('el5')
            expect(isListenableElement(el, eventType)).toBe(false)
        })
    })

    describe('dispatchEventOnElement', () => {
        const eventType = 'fir:testevent'
        const eventDetail = { message: 'hello' }
        let mockElement
        let customEvent
        let actualDispatchEventOnElement // To store and use the original function

        beforeAll(() => {
            // Get the ACTUAL implementation for this specific test suite
            actualDispatchEventOnElement = jest.requireActual(
                '../src/eventDispatcher'
            ).dispatchEventOnElement
        })

        beforeEach(() => {
            mockElement = document.createElement('div')
            mockElement.dispatchEvent = jest.fn()
            customEvent = new CustomEvent(eventType, {
                detail: eventDetail,
                bubbles: true,
                composed: true,
                cancelable: true,
            })
        })

        test('should dispatch event if element is listenable', () => {
            const localIsListenableElement = jest.fn().mockReturnValue(true)
            actualDispatchEventOnElement(
                mockElement,
                customEvent,
                localIsListenableElement
            )
            expect(localIsListenableElement).toHaveBeenCalledWith(
                mockElement,
                eventType
            )
            expect(mockElement.dispatchEvent).toHaveBeenCalledTimes(1)
            const dispatchedEvent = mockElement.dispatchEvent.mock.calls[0][0]
            expect(dispatchedEvent).toBeInstanceOf(CustomEvent)
            expect(dispatchedEvent.type).toBe(eventType)
            expect(dispatchedEvent.detail).toEqual(eventDetail)
        })

        test('should not dispatch event if element is not listenable', () => {
            const localIsListenableElement = jest.fn().mockReturnValue(false)
            actualDispatchEventOnElement(
                mockElement,
                customEvent,
                localIsListenableElement
            )
            expect(localIsListenableElement).toHaveBeenCalledWith(
                mockElement,
                eventType
            )
            expect(mockElement.dispatchEvent).not.toHaveBeenCalled()
        })

        test('should not dispatch event if element is null', () => {
            const localIsListenableElement = jest.fn()
            actualDispatchEventOnElement(
                null,
                customEvent,
                localIsListenableElement
            )
            expect(localIsListenableElement).not.toHaveBeenCalled()
            expect(mockElement.dispatchEvent).not.toHaveBeenCalled()
        })

        test('should not attempt to check listenable if element is null', () => {
            const localIsListenableElement = jest.fn().mockReturnValue(true)
            actualDispatchEventOnElement(
                null,
                customEvent,
                localIsListenableElement
            )
            expect(localIsListenableElement).not.toHaveBeenCalled()
            expect(mockElement.dispatchEvent).not.toHaveBeenCalled()
        })

        test('should create a new event instance when dispatching', () => {
            const localIsListenableElement = jest.fn().mockReturnValue(true)
            actualDispatchEventOnElement(
                mockElement,
                customEvent,
                localIsListenableElement
            )
            expect(mockElement.dispatchEvent).toHaveBeenCalledTimes(1)
            const dispatchedEvent = mockElement.dispatchEvent.mock.calls[0][0]
            expect(dispatchedEvent).not.toBe(customEvent)
            expect(dispatchedEvent.type).toBe(customEvent.type)
            expect(dispatchedEvent.detail).toEqual(customEvent.detail)
            expect(dispatchedEvent.bubbles).toBe(customEvent.bubbles)
            expect(dispatchedEvent.composed).toBe(customEvent.composed)
            expect(dispatchedEvent.cancelable).toBe(customEvent.cancelable)
        })
    })

    describe('dispatchEventOnIdTarget', () => {
        const eventType = 'fir:idtargetevent'
        const eventDetail = { id: 1 }
        let customEvent
        let mockDoc
        let mockTargetElement
        let mockSibling1, mockSibling2
        let mockInjectedDispatchEventOnElement // Our specific mock for these tests

        beforeEach(() => {
            customEvent = createCustomEvent(eventType, eventDetail)
            mockTargetElement = document.createElement('div')
            mockTargetElement.id = 'target'
            mockSibling1 = document.createElement('div')
            mockSibling1.id = 'sibling1'
            mockSibling2 = document.createElement('div')
            mockSibling2.id = 'sibling2'

            const parent = document.createElement('div')
            parent.appendChild(mockTargetElement)
            parent.appendChild(mockSibling1)
            parent.appendChild(mockSibling2)

            mockDoc = {
                getElementById: jest.fn().mockImplementation((id) => {
                    if (id === 'target') return mockTargetElement
                    return null
                }),
            }
            mockInjectedDispatchEventOnElement = jest.fn()
        })

        test('should dispatch on target and listening siblings if element found', () => {
            const mockIsListenableElement = jest.fn().mockReturnValue(true)
            dispatchEventOnIdTarget(
                customEvent,
                'target',
                mockDoc,
                mockIsListenableElement,
                mockInjectedDispatchEventOnElement // Inject mock
            )
            expect(mockDoc.getElementById).toHaveBeenCalledWith('target')
            expect(mockInjectedDispatchEventOnElement).toHaveBeenCalledTimes(3)
            expect(mockInjectedDispatchEventOnElement).toHaveBeenCalledWith(
                mockTargetElement,
                customEvent,
                mockIsListenableElement
            )
            expect(mockInjectedDispatchEventOnElement).toHaveBeenCalledWith(
                mockSibling1,
                customEvent,
                mockIsListenableElement
            )
            expect(mockInjectedDispatchEventOnElement).toHaveBeenCalledWith(
                mockSibling2,
                customEvent,
                mockIsListenableElement
            )
        })

        test('should warn and not dispatch if element not found', () => {
            mockDoc.getElementById.mockReturnValue(null)
            const mockIsListenableElement = jest.fn()
            dispatchEventOnIdTarget(
                customEvent,
                'nonexistent',
                mockDoc,
                mockIsListenableElement,
                mockInjectedDispatchEventOnElement // Inject mock
            )
            expect(mockDoc.getElementById).toHaveBeenCalledWith('nonexistent')
            expect(console.warn).toHaveBeenCalledWith(
                `Target element with ID "nonexistent" not found for event "${eventType}".`
            )
            expect(mockInjectedDispatchEventOnElement).not.toHaveBeenCalled()
        })
    })

    describe('dispatchEventOnClassTarget', () => {
        const eventType = 'fir:classevent'
        const eventDetail = { class: 'test' }
        let customEvent
        let mockDoc
        let mockElement1, mockElement2
        let mockInjectedDispatchEventOnElement

        beforeEach(() => {
            customEvent = createCustomEvent(eventType, eventDetail)
            mockElement1 = document.createElement('div')
            mockElement2 = document.createElement('div')
            mockDoc = {
                getElementsByClassName: jest
                    .fn()
                    .mockReturnValue([mockElement1, mockElement2]),
            }
            // mockInjectedDispatchEventOnElement is created for each test
            mockInjectedDispatchEventOnElement = jest.fn()
            // REMOVED: dispatchEventOnElement.mockClear() - as dispatchEventOnElement is not a global mock here
        })

        test('should dispatch on elements with targetClass', () => {
            dispatchEventOnClassTarget(
                customEvent,
                'my-class',
                null,
                mockDoc,
                mockInjectedDispatchEventOnElement // Inject mock
            )
            expect(mockDoc.getElementsByClassName).toHaveBeenCalledWith(
                'my-class'
            )
            expect(mockInjectedDispatchEventOnElement).toHaveBeenCalledWith(
                mockElement1,
                customEvent
            )
            expect(mockInjectedDispatchEventOnElement).toHaveBeenCalledWith(
                mockElement2,
                customEvent
            )
        })

        test('should dispatch on elements with targetClass and key', () => {
            dispatchEventOnClassTarget(
                customEvent,
                'my-class',
                'my-key',
                mockDoc,
                mockInjectedDispatchEventOnElement // Inject mock
            )
            expect(mockDoc.getElementsByClassName).toHaveBeenCalledWith(
                'my-class--my-key'
            )
            expect(mockInjectedDispatchEventOnElement).toHaveBeenCalledWith(
                mockElement1,
                customEvent
            )
            expect(mockInjectedDispatchEventOnElement).toHaveBeenCalledWith(
                mockElement2,
                customEvent
            )
        })

        test('should not throw if no elements match', () => {
            mockDoc.getElementsByClassName.mockReturnValue([])
            expect(() =>
                dispatchEventOnClassTarget(
                    customEvent,
                    'no-such-class',
                    null,
                    mockDoc,
                    mockInjectedDispatchEventOnElement // Inject mock
                )
            ).not.toThrow()
            expect(mockInjectedDispatchEventOnElement).not.toHaveBeenCalled()
        })
    })

    describe('isValidFirEvent', () => {
        test('should return true for a valid event', () => {
            expect(isValidFirEvent({ type: 'fir:myevent', detail: {} })).toBe(
                true
            )
            expect(
                isValidFirEvent({ type: 'fir:myevent:subevent', detail: {} })
            ).toBe(true)
        })
        test('should return false and log error for null event', () => {
            expect(isValidFirEvent(null)).toBe(false)
            expect(console.error).toHaveBeenCalledWith(
                'Server event is null or undefined.'
            )
        })
        test('should return false and log error for undefined event', () => {
            expect(isValidFirEvent(undefined)).toBe(false)
            expect(console.error).toHaveBeenCalledWith(
                'Server event is null or undefined.'
            )
        })
        test('should return false and log error for missing type', () => {
            const event = { detail: {} }
            expect(isValidFirEvent(event)).toBe(false)
            expect(console.error).toHaveBeenCalledWith(
                'Server event type is missing or invalid:',
                event
            )
        })
        test('should return false and log error for invalid type (not string)', () => {
            const event = { type: 123, detail: {} }
            expect(isValidFirEvent(event)).toBe(false)
            expect(console.error).toHaveBeenCalledWith(
                'Server event type is missing or invalid:',
                event
            )
        })
        test('should return false and log error for type not starting with fir:', () => {
            const event = { type: 'my:event', detail: {} }
            expect(isValidFirEvent(event)).toBe(false)
            expect(console.error).toHaveBeenCalledWith(
                `Server event type "${event.type}" is invalid. Must start with "fir:" and have at least two parts.`
            )
        })
        test('should return false and log error for type with less than two parts', () => {
            const event = { type: 'fir:', detail: {} }
            expect(isValidFirEvent(event)).toBe(false)
            expect(console.error).toHaveBeenCalledWith(
                `Server event type "${event.type}" is invalid. Must start with "fir:" and have at least two parts.`
            )
        })
    })

    // Tests for dispatchSingleServerEvent, processAndDispatchServerEvents, postEvent
    // might need review if their mocking strategy relied on the global jest.mock
    // for dispatchEventOnElement, dispatchEventOnIdTarget, or dispatchEventOnClassTarget.
    // The require-based mocking within their describe blocks should still work for now,
    // as it targets the module exports directly.

    describe('dispatchSingleServerEvent', () => {
        let mockWindow
        let mockDoc
        let serverEvent
        // let spiedDispatchEventOnElement; // REMOVE this spy
        let testableDispatcherModule

        beforeEach(() => {
            jest.resetModules()
            mockWindow = { dispatchEvent: jest.fn() }
            mockDoc = {
                getElementById: jest.fn(),
                getElementsByClassName: jest.fn(),
            }
            // serverEvent.type will be 'fir:test:event'
            serverEvent = { type: 'fir:test:event', detail: { data: 'test' } }

            testableDispatcherModule = require('../src/eventDispatcher')

            // REMOVE Spy for dispatchEventOnElement
            // spiedDispatchEventOnElement = jest
            //     .spyOn(testableDispatcherModule, 'dispatchEventOnElement')
            //     .mockImplementation(jest.fn());
        })

        afterEach(() => {
            jest.restoreAllMocks()
        })

        test('should dispatch event on window', () => {
            testableDispatcherModule.dispatchSingleServerEvent(
                serverEvent,
                mockWindow,
                mockDoc
            )

            expect(mockWindow.dispatchEvent).toHaveBeenCalledTimes(1)
            const dispatchedEvent = mockWindow.dispatchEvent.mock.calls[0][0]
            expect(dispatchedEvent).toBeInstanceOf(CustomEvent)
            expect(dispatchedEvent.type).toBe(serverEvent.type)
            expect(dispatchedEvent.detail).toEqual(serverEvent.detail)
        })

        test('should dispatch to ID target and its listenable siblings if specified', () => {
            serverEvent.target = '#myId'
            const mockTargetElement = document.createElement('div')
            mockTargetElement.id = 'myId'
            mockTargetElement.setAttribute(
                `x-on:${serverEvent.type}`,
                'handler1'
            ) // Make listenable
            mockTargetElement.dispatchEvent = jest.fn() // Mock its own dispatchEvent

            const mockSibling1 = document.createElement('div')
            mockSibling1.setAttribute(`x-on:${serverEvent.type}`, 'handler2') // Make listenable
            mockSibling1.dispatchEvent = jest.fn() // Mock its own dispatchEvent

            const mockSibling2NonListenable = document.createElement('div') // Not listenable
            mockSibling2NonListenable.dispatchEvent = jest.fn()

            const parent = document.createElement('div')
            parent.appendChild(mockTargetElement)
            parent.appendChild(mockSibling1)
            parent.appendChild(mockSibling2NonListenable)

            mockDoc.getElementById.mockReturnValue(mockTargetElement)

            testableDispatcherModule.dispatchSingleServerEvent(
                serverEvent,
                mockWindow,
                mockDoc
            )

            expect(mockDoc.getElementById).toHaveBeenCalledWith('myId')

            // Check window dispatch
            expect(mockWindow.dispatchEvent).toHaveBeenCalledTimes(1)
            const eventDispatchedToWindow =
                mockWindow.dispatchEvent.mock.calls[0][0]

            // Check target element dispatch
            expect(mockTargetElement.dispatchEvent).toHaveBeenCalledTimes(1)
            const eventOnTarget =
                mockTargetElement.dispatchEvent.mock.calls[0][0]
            expect(eventOnTarget.type).toBe(serverEvent.type)
            expect(eventOnTarget.detail).toEqual(serverEvent.detail)
            expect(eventOnTarget).not.toBe(eventDispatchedToWindow) // dispatchEventOnElement creates a new event

            // Check listenable sibling dispatch
            expect(mockSibling1.dispatchEvent).toHaveBeenCalledTimes(1)
            const eventOnSibling1 = mockSibling1.dispatchEvent.mock.calls[0][0]
            expect(eventOnSibling1.type).toBe(serverEvent.type)
            expect(eventOnSibling1.detail).toEqual(serverEvent.detail)

            // Check non-listenable sibling (should not have dispatch called by dispatchEventOnElement)
            expect(
                mockSibling2NonListenable.dispatchEvent
            ).not.toHaveBeenCalled()
        })

        test('should dispatch to class target elements if specified', () => {
            serverEvent.target = '.myClass'
            serverEvent.key = 'myKey'

            const mockElement1 = document.createElement('div')
            mockElement1.setAttribute(`x-on:${serverEvent.type}`, 'handler1') // Make listenable
            mockElement1.dispatchEvent = jest.fn()

            const mockElement2 = document.createElement('div') // Not listenable for this event
            mockElement2.setAttribute(`x-on:fir:other-event`, 'handler2')
            mockElement2.dispatchEvent = jest.fn()

            const mockElement3 = document.createElement('div')
            mockElement3.setAttribute(`x-on:${serverEvent.type}`, 'handler3') // Make listenable
            mockElement3.dispatchEvent = jest.fn()

            mockDoc.getElementsByClassName.mockReturnValue([
                mockElement1,
                mockElement2,
                mockElement3,
            ])

            testableDispatcherModule.dispatchSingleServerEvent(
                serverEvent,
                mockWindow,
                mockDoc
            )

            expect(mockDoc.getElementsByClassName).toHaveBeenCalledWith(
                'myClass--myKey'
            )

            const eventDispatchedToWindow =
                mockWindow.dispatchEvent.mock.calls[0][0]

            expect(mockElement1.dispatchEvent).toHaveBeenCalledTimes(1)
            const eventOnElement1 = mockElement1.dispatchEvent.mock.calls[0][0]
            expect(eventOnElement1.type).toBe(serverEvent.type)
            expect(eventOnElement1.detail).toEqual(serverEvent.detail)

            expect(mockElement2.dispatchEvent).not.toHaveBeenCalled() // Not listenable for serverEvent.type

            expect(mockElement3.dispatchEvent).toHaveBeenCalledTimes(1)
            const eventOnElement3 = mockElement3.dispatchEvent.mock.calls[0][0]
            expect(eventOnElement3.type).toBe(serverEvent.type)
            expect(eventOnElement3.detail).toEqual(serverEvent.detail)
        })

        test('should not dispatch to target elements if target is not specified', () => {
            // Elements that might exist but shouldn't be targeted
            const mockElement = document.createElement('div')
            mockElement.dispatchEvent = jest.fn()
            mockDoc.getElementById.mockReturnValue(mockElement)
            mockDoc.getElementsByClassName.mockReturnValue([mockElement])

            testableDispatcherModule.dispatchSingleServerEvent(
                serverEvent,
                mockWindow,
                mockDoc
            )
            expect(mockDoc.getElementById).not.toHaveBeenCalled()
            expect(mockDoc.getElementsByClassName).not.toHaveBeenCalled()
            expect(mockElement.dispatchEvent).not.toHaveBeenCalled() // Ensure no accidental dispatch
        })

        test('should warn if target format is invalid and not dispatch to elements', () => {
            serverEvent.target = 'invalidTarget'
            const mockElement = document.createElement('div')
            mockElement.dispatchEvent = jest.fn()
            mockDoc.getElementById.mockReturnValue(mockElement) // Set up just in case

            testableDispatcherModule.dispatchSingleServerEvent(
                serverEvent,
                mockWindow,
                mockDoc
            )
            expect(console.warn).toHaveBeenCalledWith(
                `Invalid target format "invalidTarget" for event "fir:test:event". Target must start with # or .`
            )
            expect(mockDoc.getElementById).not.toHaveBeenCalled()
            expect(mockDoc.getElementsByClassName).not.toHaveBeenCalled()
            expect(mockElement.dispatchEvent).not.toHaveBeenCalled()
        })
    })

    describe('processAndDispatchServerEvents', () => {
        let mockDispatchFn
        let testableDispatcherModule // Keep if you want to call processAndDispatchServerEvents from a specific instance

        beforeEach(() => {
            jest.resetModules() // Good practice if requiring the module
            testableDispatcherModule = require('../src/eventDispatcher')
            mockDispatchFn = jest.fn()
            // The specific mock setup for isValidFirEvent is removed from here.
            // console.error is already mocked in the outer describe block.
        })

        afterEach(() => {
            // No specific restoration for isValidFirEvent needed here anymore.
            // jest.restoreAllMocks() in the outer afterEach will handle console spies.
        })

        test('should do nothing for empty or non-array serverEvents', () => {
            // Call the function from the imported module or a fresh require if preferred
            processAndDispatchServerEvents(null, mockDispatchFn)
            expect(mockDispatchFn).not.toHaveBeenCalled()
            processAndDispatchServerEvents([], mockDispatchFn)
            expect(mockDispatchFn).not.toHaveBeenCalled()
        })

        test('should filter invalid events and dispatch valid ones with :done events', () => {
            const validEvent1 = { type: 'fir:event1', detail: { d: 1 } }
            const validEvent2 = { type: 'fir:event2:sub', detail: { d: 2 } }
            const invalidEvent = { type: 'bad:event' } // This will be handled by the real isValidFirEvent
            const serverEvents = [validEvent1, invalidEvent, validEvent2]

            // No need to mock isValidFirEvent's implementation here.

            processAndDispatchServerEvents(serverEvents, mockDispatchFn)

            // The real isValidFirEvent will be called internally.
            // We check the side-effects:
            expect(console.error).toHaveBeenCalledWith(
                // From the real isValidFirEvent
                `Server event type "${invalidEvent.type}" is invalid. Must start with "fir:" and have at least two parts.`
            )
            expect(mockDispatchFn).toHaveBeenCalledTimes(4) // 2 valid events + 2 done events
            expect(mockDispatchFn).toHaveBeenCalledWith(
                validEvent1,
                expect.any(Number),
                expect.any(Array)
            )
            expect(mockDispatchFn).toHaveBeenCalledWith(
                validEvent2,
                expect.any(Number),
                expect.any(Array)
            )
            expect(mockDispatchFn).toHaveBeenCalledWith(
                expect.objectContaining({
                    type: 'fir:event1:done',
                    target: '.fir-event1-done',
                    detail: validEvent1.detail,
                }),
                expect.any(Number),
                expect.any(Array)
            )
            expect(mockDispatchFn).toHaveBeenCalledWith(
                expect.objectContaining({
                    type: 'fir:event2:done',
                    target: '.fir-event2-done',
                    detail: validEvent2.detail,
                }),
                expect.any(Number),
                expect.any(Array)
            )
        })

        test('should not add :done for onevent or onload', () => {
            const onevent = { type: 'fir:onevent', detail: {} }
            const onload = { type: 'fir:onload', detail: {} }

            // The real isValidFirEvent will validate these.
            processAndDispatchServerEvents([onevent, onload], mockDispatchFn)

            expect(mockDispatchFn).toHaveBeenCalledTimes(2)
            // Check arguments of each call precisely
            expect(mockDispatchFn.mock.calls[0][0]).toEqual(onevent)
            expect(mockDispatchFn.mock.calls[1][0]).toEqual(onload)
            // console.error should not have been called if these are valid
            expect(console.error).not.toHaveBeenCalled()
        })

        test('should not add duplicate :done events for same base event name', () => {
            const eventA1 = { type: 'fir:eventA:part1', detail: { a: 1 } }
            const eventA2 = { type: 'fir:eventA:part2', detail: { a: 2 } }

            // The real isValidFirEvent will validate these.
            processAndDispatchServerEvents([eventA1, eventA2], mockDispatchFn)

            expect(mockDispatchFn).toHaveBeenCalledTimes(3) // eventA1, eventA2, one :done for eventA
            expect(mockDispatchFn.mock.calls[0][0]).toEqual(eventA1)
            expect(mockDispatchFn.mock.calls[1][0]).toEqual(eventA2)
            expect(mockDispatchFn.mock.calls[2][0]).toEqual(
                expect.objectContaining({
                    type: 'fir:eventA:done',
                    detail: eventA1.detail, // The :done event should carry the detail of the first event in the series
                })
            )
            // console.error should not have been called if these are valid
            expect(console.error).not.toHaveBeenCalled()
        })
    })

    describe('generateFirEventNames', () => {
        test('should generate correct names for PascalCase event_id', () => {
            const names = actualEventDispatcher.generateFirEventNames(
                'MyEvent',
                'pending'
            )
            expect(names.eventIdLower).toBe('myevent')
            expect(names.eventIdKebab).toBe('my-event')
            expect(names.typeLower).toBe('fir:myevent:pending')
            expect(names.targetLower).toBe('.fir-myevent-pending')
            expect(names.typeKebab).toBe('fir:my-event:pending')
            expect(names.targetKebab).toBe('.fir-my-event-pending')
        })

        test('should generate correct names for kebab-case event_id', () => {
            const names = actualEventDispatcher.generateFirEventNames(
                'my-event',
                'error'
            )
            expect(names.eventIdLower).toBe('my-event')
            expect(names.eventIdKebab).toBe('my-event')
            expect(names.typeLower).toBe('fir:my-event:error')
            expect(names.targetLower).toBe('.fir-my-event-error')
            expect(names.typeKebab).toBeNull()
            expect(names.targetKebab).toBeNull()
        })

        test('should generate correct names for lowercase event_id', () => {
            const names = actualEventDispatcher.generateFirEventNames(
                'myevent',
                'done'
            )
            expect(names.eventIdLower).toBe('myevent')
            expect(names.eventIdKebab).toBe('myevent')
            expect(names.typeLower).toBe('fir:myevent:done')
            expect(names.targetLower).toBe('.fir-myevent-done')
            expect(names.typeKebab).toBeNull()
            expect(names.targetKebab).toBeNull()
        })

        test('should handle event_id with numbers', () => {
            const names = actualEventDispatcher.generateFirEventNames(
                'Event123Test',
                'pending'
            )
            expect(names.eventIdLower).toBe('event123test')
            expect(names.eventIdKebab).toBe('event123-test')
            expect(names.typeLower).toBe('fir:event123test:pending')
            expect(names.targetLower).toBe('.fir-event123test-pending')
            expect(names.typeKebab).toBe('fir:event123-test:pending')
            expect(names.targetKebab).toBe('.fir-event123-test-pending')
        })
    })

    describe('postEvent', () => {
        let mockSocket
        let mockProcessEventsFn
        let mockDispatchSingleEventFn
        let mockFetchFn
        let mockWindowLocation
        let firEvent
        let mockNowFn
        const fixedTimestamp = 1234567890
        // mockGenerateFirEventNames is already imported from the top-level mock

        beforeEach(() => {
            mockSocket = { emit: jest.fn() }
            mockProcessEventsFn = jest.fn()
            mockDispatchSingleEventFn = jest.fn()
            mockFetchFn = jest.fn()
            mockWindowLocation = { pathname: '/testpath', href: '' }
            firEvent = { event_id: 'TestEvent', params: { data: 'value' } }
            mockNowFn = jest.fn().mockReturnValue(fixedTimestamp)

            // Clear and set default implementation for the mock from the top
            mockGenerateFirEventNames.mockClear()
            mockGenerateFirEventNames.mockImplementation((baseId, suffix) => {
                const eventIdLower = baseId.toLowerCase()
                const eventIdKebab = baseId
                    .replace(/([a-z0-9]|(?=[A-Z]))([A-Z])/g, '$1-$2')
                    .toLowerCase()
                const typeLower = `${actualEventDispatcher.FIR_PREFIX}${eventIdLower}:${suffix}`
                const targetLower = `.${typeLower.replace(/:/g, '-')}`
                let typeKebab = null,
                    targetKebab = null
                if (eventIdLower !== eventIdKebab) {
                    typeKebab = `${actualEventDispatcher.FIR_PREFIX}${eventIdKebab}:${suffix}`
                    targetKebab = `.${typeKebab.replace(/:/g, '-')}`
                }
                return {
                    eventIdLower,
                    eventIdKebab,
                    typeLower,
                    targetLower,
                    typeKebab,
                    targetKebab,
                }
            })
        })

        // afterEach can remain as is, jest.clearAllMocks() in global afterEach will clear mockGenerateFirEventNames

        test('should log error and return if event_id is missing', async () => {
            await postEvent(
                { params: {} },
                mockSocket,
                mockProcessEventsFn,
                mockDispatchSingleEventFn,
                mockFetchFn,
                mockWindowLocation,
                mockNowFn,
                mockGenerateFirEventNames // Pass the mock here
            )
            expect(console.error).toHaveBeenCalledWith(
                "event id is empty and element id is not set. can't emit event"
            )
            // ... other assertions ...
            expect(mockGenerateFirEventNames).not.toHaveBeenCalled()
        })

        test('should call nowFn and dispatch pending events, then send via WebSocket if successful', async () => {
            mockSocket.emit.mockReturnValue(true)

            await postEvent(
                firEvent,
                mockSocket,
                mockProcessEventsFn,
                mockDispatchSingleEventFn,
                mockFetchFn,
                mockWindowLocation,
                mockNowFn,
                mockGenerateFirEventNames // Pass the mock here
            )

            expect(mockNowFn).toHaveBeenCalledTimes(1)
            expect(mockGenerateFirEventNames).toHaveBeenCalledWith(
                'TestEvent',
                'pending'
            )
            const pendingNames = mockGenerateFirEventNames.mock.results[0].value
            // ... rest of assertions for this test ...
            expect(mockDispatchSingleEventFn).toHaveBeenCalledWith({
                type: pendingNames.typeLower,
                target: pendingNames.targetLower,
                detail: { params: firEvent.params },
            })
            if (pendingNames.typeKebab) {
                expect(mockDispatchSingleEventFn).toHaveBeenCalledWith({
                    type: pendingNames.typeKebab,
                    target: pendingNames.targetKebab,
                    detail: { params: firEvent.params },
                })
            }
            expect(mockSocket.emit).toHaveBeenCalledWith({
                ...firEvent,
                ts: fixedTimestamp,
            })
            expect(mockFetchFn).not.toHaveBeenCalled()
        })

        describe('Fetch Fallback Scenarios (WebSocket emit fails or socket is null)', () => {
            beforeEach(() => {
                if (mockSocket) mockSocket.emit.mockReturnValue(false)
                // mockGenerateFirEventNames is cleared and reset in the outer beforeEach
            })

            test('should fallback to Fetch, process successful JSON response', async () => {
                const mockServerEvents = [{ type: 'fir:response', detail: {} }]
                mockFetchFn.mockResolvedValue({
                    ok: true,
                    redirected: false,
                    headers: {
                        get: jest.fn().mockReturnValue('application/json'),
                    },
                    json: jest.fn().mockResolvedValue(mockServerEvents),
                })
                await postEvent(
                    firEvent,
                    mockSocket,
                    mockProcessEventsFn,
                    mockDispatchSingleEventFn,
                    mockFetchFn,
                    mockWindowLocation,
                    mockNowFn,
                    mockGenerateFirEventNames // Pass the mock
                )
                // ... assertions ...
                expect(mockGenerateFirEventNames).toHaveBeenCalledWith(
                    'TestEvent',
                    'pending'
                )
            })

            test('should handle fetch redirect', async () => {
                mockFetchFn.mockResolvedValue({
                    ok: true,
                    redirected: true,
                    url: '/newlocation',
                })
                await postEvent(
                    firEvent,
                    mockSocket,
                    mockProcessEventsFn,
                    mockDispatchSingleEventFn,
                    mockFetchFn,
                    mockWindowLocation,
                    mockNowFn,
                    mockGenerateFirEventNames // Pass the mock
                )
                // ... assertions ...
                expect(mockGenerateFirEventNames).toHaveBeenCalledWith(
                    'TestEvent',
                    'pending'
                )
            })

            test('should handle fetch non-JSON response', async () => {
                mockFetchFn.mockResolvedValue({
                    ok: true,
                    redirected: false,
                    headers: { get: jest.fn().mockReturnValue('text/html') },
                })
                await postEvent(
                    firEvent,
                    mockSocket,
                    mockProcessEventsFn,
                    mockDispatchSingleEventFn,
                    mockFetchFn,
                    mockWindowLocation,
                    mockNowFn,
                    mockGenerateFirEventNames // Pass the mock
                )
                // ... assertions ...
                expect(mockGenerateFirEventNames).toHaveBeenCalledWith(
                    'TestEvent',
                    'pending'
                )
            })

            test('should handle fetch HTTP error and dispatch error events', async () => {
                mockFetchFn.mockResolvedValue({
                    ok: false,
                    status: 500,
                    headers: { get: jest.fn() },
                })
                // mockGenerateFirEventNames.mockClear() // Cleared in outer beforeEach

                await postEvent(
                    firEvent,
                    mockSocket,
                    mockProcessEventsFn,
                    mockDispatchSingleEventFn,
                    mockFetchFn,
                    mockWindowLocation,
                    mockNowFn,
                    mockGenerateFirEventNames // Pass the mock
                )
                // ... assertions ...
                expect(mockGenerateFirEventNames).toHaveBeenCalledWith(
                    'TestEvent',
                    'pending'
                ) // For initial dispatch
                expect(mockGenerateFirEventNames).toHaveBeenCalledWith(
                    'TestEvent',
                    'error'
                ) // For error dispatch
                const errorNames = mockGenerateFirEventNames.mock.results.find(
                    (r) => r.value.typeLower.endsWith(':error')
                ).value
                // ... rest of error dispatch assertions
                expect(mockDispatchSingleEventFn).toHaveBeenCalledWith(
                    expect.objectContaining({
                        type: errorNames.typeLower,
                        detail: { error: 'HTTP error! status: 500' },
                    })
                )
                if (errorNames.typeKebab) {
                    expect(mockDispatchSingleEventFn).toHaveBeenCalledWith(
                        expect.objectContaining({
                            type: errorNames.typeKebab,
                            detail: { error: 'HTTP error! status: 500' },
                        })
                    )
                }
            })

            test('should handle fetch network error and dispatch error events', async () => {
                const networkError = new Error('Network failure')
                mockFetchFn.mockRejectedValue(networkError)
                // mockGenerateFirEventNames.mockClear() // Cleared in outer beforeEach

                await postEvent(
                    firEvent,
                    mockSocket,
                    mockProcessEventsFn,
                    mockDispatchSingleEventFn,
                    mockFetchFn,
                    mockWindowLocation,
                    mockNowFn,
                    mockGenerateFirEventNames // Pass the mock
                )
                // ... assertions ...
                expect(mockGenerateFirEventNames).toHaveBeenCalledWith(
                    'TestEvent',
                    'pending'
                )
                expect(mockGenerateFirEventNames).toHaveBeenCalledWith(
                    'TestEvent',
                    'error'
                )
                const errorNames = mockGenerateFirEventNames.mock.results.find(
                    (r) => r.value.typeLower.endsWith(':error')
                ).value
                // ... rest of error dispatch assertions
                expect(mockDispatchSingleEventFn).toHaveBeenCalledWith(
                    expect.objectContaining({
                        type: errorNames.typeLower,
                        detail: { error: networkError.message },
                    })
                )
                if (errorNames.typeKebab) {
                    expect(mockDispatchSingleEventFn).toHaveBeenCalledWith(
                        expect.objectContaining({
                            type: errorNames.typeKebab,
                            detail: { error: networkError.message },
                        })
                    )
                }
            })
        })

        test('should use null socket and fallback to Fetch', async () => {
            const mockServerEvents = [{ type: 'fir:response', detail: {} }]
            mockFetchFn.mockResolvedValue({
                ok: true,
                redirected: false,
                headers: { get: jest.fn().mockReturnValue('application/json') },
                json: jest.fn().mockResolvedValue(mockServerEvents),
            })
            await postEvent(
                firEvent,
                null, // Null socket
                mockProcessEventsFn,
                mockDispatchSingleEventFn,
                mockFetchFn,
                mockWindowLocation,
                mockNowFn,
                mockGenerateFirEventNames // Pass the mock
            )
            // ... assertions ...
            expect(mockGenerateFirEventNames).toHaveBeenCalledWith(
                'TestEvent',
                'pending'
            )
        })

        test('should dispatch only one set of pending/error events if event_id is already kebab-case', async () => {
            const kebabEvent = { event_id: 'test-event', params: {} }
            mockSocket.emit.mockReturnValue(true)
            // mockGenerateFirEventNames is cleared and reset in beforeEach.
            // The default mockImplementation should handle kebab-case correctly by returning null for typeKebab.

            await postEvent(
                kebabEvent,
                mockSocket,
                mockProcessEventsFn,
                mockDispatchSingleEventFn,
                mockFetchFn,
                mockWindowLocation,
                mockNowFn,
                mockGenerateFirEventNames // Pass the mock
            )
            expect(mockGenerateFirEventNames).toHaveBeenCalledWith(
                'test-event',
                'pending'
            )
            // Check that dispatchSingleEventFn was called only once for pending
            // because typeKebab should be null from the mock implementation
            const pendingNamesResult =
                mockGenerateFirEventNames.mock.results.find((r) =>
                    r.value.typeLower.endsWith(':pending')
                ).value
            expect(pendingNamesResult.typeKebab).toBeNull()

            let pendingDispatchCount = 0
            mockDispatchSingleEventFn.mock.calls.forEach((callArgs) => {
                if (callArgs[0].type === pendingNamesResult.typeLower)
                    pendingDispatchCount++
            })
            expect(pendingDispatchCount).toBe(1)

            // Simulate fetch error
            mockDispatchSingleEventFn.mockClear()
            mockSocket.emit.mockReturnValue(false)
            mockFetchFn.mockRejectedValue(new Error('fail'))
            // mockGenerateFirEventNames is cleared and reset in beforeEach for the next "call" to postEvent implicitly

            await postEvent(
                kebabEvent,
                mockSocket,
                mockProcessEventsFn,
                mockDispatchSingleEventFn,
                mockFetchFn,
                mockWindowLocation,
                mockNowFn,
                mockGenerateFirEventNames // Pass the mock
            )
            expect(mockGenerateFirEventNames).toHaveBeenCalledWith(
                'test-event',
                'error'
            )
            const errorNamesResult =
                mockGenerateFirEventNames.mock.results.find((r) =>
                    r.value.typeLower.endsWith(':error')
                ).value
            expect(errorNamesResult.typeKebab).toBeNull()

            let errorDispatchCount = 0
            mockDispatchSingleEventFn.mock.calls.forEach((callArgs) => {
                if (callArgs[0].type === errorNamesResult.typeLower)
                    errorDispatchCount++
            })
            expect(errorDispatchCount).toBe(1)
        })
    })
})

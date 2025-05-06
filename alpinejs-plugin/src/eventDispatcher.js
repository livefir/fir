export const FIR_PREFIX = 'fir:'
export const ID_SELECTOR_PREFIX = '#'
export const CLASS_SELECTOR_PREFIX = '.'
const ALPINE_EVENT_PREFIX = '@'
const ALPINE_XON_PREFIX = 'x-on:'

/**
 * Creates a standard CustomEvent object for dispatching.
 * @param {string} type - The event type (e.g., "fir:myevent:pending").
 * @param {any} detail - The event detail payload.
 * @returns {CustomEvent} The configured CustomEvent object.
 */
export const createCustomEvent = (type, detail) => {
    const opts = {
        detail: detail,
        bubbles: true,
        composed: true, // Allows events to pass the shadow DOM barrier.
        cancelable: true,
    }
    return new CustomEvent(type, opts)
}

/**
 * Checks if an element has Alpine listeners for a specific event type.
 * @param {Element} element - The DOM element to check.
 * @param {string} eventType - The event type (e.g., "fir:myevent:pending").
 * @returns {boolean} True if the element listens for the event, false otherwise.
 */
export const isListenableElement = (element, eventType) => {
    return (
        element.hasAttribute(`${ALPINE_EVENT_PREFIX}${eventType}`) ||
        element.hasAttribute(`${ALPINE_XON_PREFIX}${eventType}`)
    )
}

/**
 * Dispatches an event on a specific element if it's listening.
 * @param {Element} element - The target DOM element.
 * @param {CustomEvent} event - The event object to dispatch.
 */
export const dispatchEventOnElement = (element, event) => {
    // Only dispatch if the element exists and is listening for this specific event type
    if (element && isListenableElement(element, event.type)) {
        // Dispatch a new instance of the event for this specific target
        element.dispatchEvent(
            new CustomEvent(event.type, {
                detail: event.detail,
                bubbles: event.bubbles,
                composed: event.composed,
                cancelable: event.cancelable,
            })
        )
    }
}

/**
 * Dispatches an event to an element specified by ID and its listening siblings.
 * @param {CustomEvent} event - The event object to dispatch.
 * @param {string} targetId - The ID of the target element (without '#').
 */
export const dispatchEventOnIdTarget = (event, targetId, doc = document) => {
    const element = doc.getElementById(targetId)
    if (!element) {
        console.warn(
            `Target element with ID "${targetId}" not found for event "${event.type}".`
        )
        return
    }

    // Dispatch on the element itself
    dispatchEventOnElement(element, event)

    // Dispatch on siblings listening for the event
    if (element.parentNode) {
        const siblings = [...element.parentNode.children].filter(
            (node) => node !== element
        )
        siblings.forEach((sibling) => dispatchEventOnElement(sibling, event))
    }
}

/**
 * Dispatches an event to elements specified by class name.
 * @param {CustomEvent} event - The event object to dispatch.
 * @param {string} targetClass - The class name of the target elements (without '.').
 * @param {string|null} key - An optional key to append to the class name (e.g., "myclass--mykey").
 */
export const dispatchEventOnClassTarget = (
    event,
    targetClass,
    key,
    doc = document
) => {
    let className = targetClass
    // Append key if provided, creating a more specific class selector
    if (key) {
        className += '--' + key
    }
    const elements = doc.getElementsByClassName(className)
    // It's okay if no elements match, just proceed.
    Array.from(elements).forEach((element) =>
        dispatchEventOnElement(element, event)
    )
}

/**
 * Validates a server event object structure and type format.
 * @param {any} serverEvent - The event object to validate.
 * @returns {boolean} True if the event is valid, false otherwise.
 */
export const isValidFirEvent = (serverEvent) => {
    if (!serverEvent) {
        console.error('Server event is null or undefined.')
        return false
    }
    if (!serverEvent.type || typeof serverEvent.type !== 'string') {
        console.error('Server event type is missing or invalid:', serverEvent)
        return false
    }
    const parts = serverEvent.type.split(':')
    // Ensure it starts with 'fir:' and has at least 'fir' and 'eventName' parts
    if (parts.length < 2 || parts[0] !== FIR_PREFIX.slice(0, -1)) {
        console.error(
            `Server event type "${serverEvent.type}" is invalid. Must start with "fir:" and have at least two parts.`
        )
        return false
    }
    return true
}

// Dispatches a single server event to the window and specific targets (ID or class)
export const dispatchSingleServerEvent = (
    serverEvent,
    windowObj = window,
    doc = document
) => {
    const event = createCustomEvent(serverEvent.type, serverEvent.detail)
    windowObj.dispatchEvent(event)

    const target = serverEvent.target
    if (target) {
        if (target.startsWith(ID_SELECTOR_PREFIX)) {
            const targetId = target.substring(1)
            dispatchEventOnIdTarget(event, targetId, doc)
        } else if (target.startsWith(CLASS_SELECTOR_PREFIX)) {
            const targetClass = target.substring(1)
            dispatchEventOnClassTarget(event, targetClass, serverEvent.key, doc)
        } else {
            console.warn(
                `Invalid target format "${target}" for event "${serverEvent.type}". Target must start with # or .`
            )
        }
    }
}

// Processes an array of server events, validates them, adds ':done' events, and dispatches them
export const processAndDispatchServerEvents = (
    serverEvents,
    dispatchFn = dispatchSingleServerEvent // Now refers to dispatchSingleServerEvent in this module
) => {
    if (!Array.isArray(serverEvents) || serverEvents.length === 0) {
        return
    }

    const validEvents = serverEvents.filter(isValidFirEvent) // Assumes isValidFirEvent is available in this module
    const eventsToDispatch = [...validEvents]
    const processedEventNames = new Set()

    validEvents.forEach((serverEvent) => {
        const parts = serverEvent.type.split(':')
        if (parts.length < 2) return // Malformed event type

        const eventName = parts[1]

        // Skip special or already processed events
        if (
            eventName === 'onevent' ||
            eventName === 'onload' ||
            processedEventNames.has(eventName)
        ) {
            return
        }

        processedEventNames.add(eventName)

        // Add a corresponding ':done' event
        const doneEventType = `${FIR_PREFIX}${eventName}:done` // FIR_PREFIX is now local
        eventsToDispatch.push({
            type: doneEventType,
            target: `.${doneEventType.replaceAll(':', '-')}`, // Target class based on done event type
            detail: serverEvent.detail, // Carry over details from original event
        })
    })

    eventsToDispatch.forEach(dispatchFn)
}

export const postEvent = async (
    firEvent,
    socket, // WebSocket instance, passed in
    processEventsFn = processAndDispatchServerEvents, // Default to the one in this module
    dispatchSingleEventFn = dispatchSingleServerEvent, // Default to the one in this module
    fetchFn = fetch, // Default to global fetch
    windowLocation = window.location // Default to global window.location
) => {
    if (!firEvent.event_id) {
        console.error(
            "event id is empty and element id is not set. can't emit event"
        )
        return
    }

    Object.assign(firEvent, { ts: Date.now() })

    // Dispatch pending events immediately
    const eventIdLower = firEvent.event_id.toLowerCase()
    const eventIdKebab = firEvent.event_id
        .replace(/([a-z0-9]|(?=[A-Z]))([A-Z])/g, '$1-$2')
        .toLowerCase()

    let eventTypeLower = `${FIR_PREFIX}${eventIdLower}:pending` // FIR_PREFIX is from this module
    dispatchSingleEventFn({
        // dispatchSingleEventFn defaults to dispatchSingleServerEvent from this module
        type: eventTypeLower,
        target: `.${eventTypeLower.replaceAll(':', '-')}`,
    })

    if (eventIdLower !== eventIdKebab) {
        let eventTypeKebab = `${FIR_PREFIX}${eventIdKebab}:pending` // FIR_PREFIX is from this module
        dispatchSingleEventFn({
            // dispatchSingleEventFn defaults to dispatchSingleServerEvent from this module
            type: eventTypeKebab,
            target: `.${eventTypeKebab.replaceAll(':', '-')}`,
        })
    }

    // Try sending via WebSocket if available and connected
    if (socket && socket.emit(firEvent)) {
        return // Event sent via WebSocket
    }

    // Fallback to Fetch
    const body = JSON.stringify(firEvent)
    try {
        const response = await fetchFn(windowLocation.pathname, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-FIR-MODE': 'event',
            },
            body: body,
        })

        if (response.redirected) {
            window.location.href = response.url // Use injected window object if needed
            return
        }

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`)
        }

        const contentType = response.headers.get('content-type')
        if (contentType && contentType.includes('application/json')) {
            const serverEvents = await response.json()
            if (serverEvents) {
                processEventsFn(serverEvents) // processEventsFn defaults to processAndDispatchServerEvents from this module
            }
        } else {
            // Handle non-JSON responses if necessary, or just ignore
            console.log('Received non-JSON response, ignoring.')
        }
    } catch (error) {
        console.error(
            `${firEvent.event_id} fetch error: ${error}, request body: ${body}`
        )
        // Optionally dispatch error events here
        const errorEventTypeLower = `${FIR_PREFIX}${eventIdLower}:error` // FIR_PREFIX is from this module
        dispatchSingleEventFn({
            // dispatchSingleEventFn defaults to dispatchSingleServerEvent from this module
            type: errorEventTypeLower,
            target: `.${errorEventTypeLower.replaceAll(':', '-')}`,
            detail: { error: error.message },
        })
        if (eventIdLower !== eventIdKebab) {
            const errorEventTypeKebab = `${FIR_PREFIX}${eventIdKebab}:error` // FIR_PREFIX is from this module
            dispatchSingleEventFn({
                // dispatchSingleEventFn defaults to dispatchSingleServerEvent from this module
                type: errorEventTypeKebab,
                target: `.${errorEventTypeKebab.replaceAll(':', '-')}`,
                detail: { error: error.message },
            })
        }
    }
}

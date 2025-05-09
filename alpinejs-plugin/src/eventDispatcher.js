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
    return new CustomEvent(type, {
        detail,
        bubbles: true,
        composed: true,
        cancelable: true,
    })
}

/**
 * Checks if an element has Alpine listeners for a specific event type.
 * @param {Element} element - The DOM element to check.
 * @param {string} eventType - The event type (e.g., "fir:myevent:pending").
 * @returns {boolean} True if the element listens for the event, false otherwise.
 */
export const isListenableElement = (element, eventType) => {
    if (!element || typeof element.hasAttribute !== 'function') {
        return false
    }
    return (
        element.hasAttribute(ALPINE_EVENT_PREFIX + eventType) ||
        element.hasAttribute(ALPINE_XON_PREFIX + eventType)
    )
}

/**
 * Dispatches an event on a specific element if it's listening.
 * @param {Element} element - The target DOM element.
 * @param {CustomEvent} event - The event object to dispatch.
 */
export const dispatchEventOnElement = (
    element,
    event,
    isListenable = isListenableElement
) => {
    if (element && isListenable(element, event.type)) {
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
export const dispatchEventOnIdTarget = (
    event,
    targetId,
    doc = document,
    isListenable = isListenableElement,
    // Add dispatchEventElementFunc as a parameter with a default
    dispatchEventElementFunc = dispatchEventOnElement
) => {
    const element = doc.getElementById(targetId)
    if (!element) {
        console.warn(
            `Target element with ID "${targetId}" not found for event "${event.type}".`
        )
        return
    }

    // Use the passed-in or default function
    dispatchEventElementFunc(element, event, isListenable)

    if (element.parentNode) {
        const siblings = [...element.parentNode.children].filter(
            (node) => node !== element
        )
        siblings.forEach((sibling) =>
            // Use the passed-in or default function
            dispatchEventElementFunc(sibling, event, isListenable)
        )
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
    doc = document,
    // Add dispatchEventElementFunc as a parameter with a default
    dispatchEventElementFunc = dispatchEventOnElement
) => {
    let className = targetClass
    if (key) {
        className += '--' + key
    }
    const elements = doc.getElementsByClassName(className)
    Array.from(elements).forEach((element) =>
        // Use the passed-in or default function
        // Note: original dispatchEventOnClassTarget calls dispatchEventOnElement with 2 args
        dispatchEventElementFunc(element, event)
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
    if (
        typeof serverEvent.type !== 'string' ||
        !serverEvent.type.startsWith(FIR_PREFIX) ||
        serverEvent.type.split(':').length < 2 ||
        serverEvent.type.split(':')[1] === ''
    ) {
        if (
            typeof serverEvent.type !== 'string' ||
            serverEvent.type.trim() === ''
        ) {
            console.error(
                'Server event type is missing or invalid:',
                serverEvent
            )
        } else {
            console.error(
                `Server event type "${serverEvent.type}" is invalid. Must start with "fir:" and have at least two parts.`
            )
        }
        return false
    }
    return true
}

// Dispatches a single server event to the window and specific targets (ID or class)
export const dispatchSingleServerEvent = (
    serverEvent,
    windowObj = window,
    doc = document
    // If dispatchSingleServerEvent calls dispatchEventOnIdTarget or dispatchEventOnClassTarget,
    // and you want to control their internal calls to dispatchEventOnElement for tests of
    // dispatchSingleServerEvent, you might need to pass the mock down through here too,
    // or rely on those functions using their defaults (which would be the actual dispatchEventOnElement).
    // For now, this example focuses on testing dispatchEventOnIdTarget/ClassTarget directly.
) => {
    const event = createCustomEvent(serverEvent.type, serverEvent.detail)

    if (windowObj && typeof windowObj.dispatchEvent === 'function') {
        windowObj.dispatchEvent(event)
    } else {
        console.error(
            'Error: windowObj.dispatchEvent is not a function. Current windowObj:',
            windowObj
        )
        // Optionally, you could try to fall back to the global window if windowObj is problematic
        // if (window && typeof window.dispatchEvent === 'function') {
        //     console.warn('Falling back to global window.dispatchEvent');
        //     window.dispatchEvent(event);
        // } else {
        //     console.error('Global window.dispatchEvent is also not a function.');
        // }
    }

    const target = serverEvent.target
    if (target) {
        if (target.startsWith(ID_SELECTOR_PREFIX)) {
            const targetId = target.substring(1)
            // It will use its default for dispatchEventElementFunc unless specified
            dispatchEventOnIdTarget(event, targetId, doc)
        } else if (target.startsWith(CLASS_SELECTOR_PREFIX)) {
            const targetClass = target.substring(1)
            // It will use its default for dispatchEventElementFunc unless specified
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

    eventsToDispatch.forEach((eventToDispatch) => {
        dispatchFn(eventToDispatch) // Call with only the event argument
    })
}

/**
 * Generates standard event names (lowercase, kebab-case) and types/targets for a given suffix.
 * @param {string} baseEventId - The base event ID (e.g., "MyEvent" or "my-event").
 * @param {string} suffix - The suffix to append (e.g., "pending", "error").
 * @returns {object} An object containing { eventIdLower, eventIdKebab, typeLower, targetLower, typeKebab, targetKebab }.
 */
export const generateFirEventNames = (baseEventId, suffix) => {
    const eventIdLower = baseEventId.toLowerCase()
    const eventIdKebab = baseEventId
        .replace(/([a-z0-9])([A-Z])/g, '$1-$2')
        .replace(/([A-Z])([A-Z][a-z])/g, '$1-$2') // Handles cases like "MyEVENT" -> "my-event"
        .toLowerCase()

    const typeLower = `${FIR_PREFIX}${eventIdLower}:${suffix}`
    const targetLower = `.${typeLower.replace(/:/g, '-')}` // Replace all colons for class selector

    let typeKebab = null
    let targetKebab = null

    if (eventIdLower !== eventIdKebab) {
        typeKebab = `${FIR_PREFIX}${eventIdKebab}:${suffix}`
        targetKebab = `.${typeKebab.replace(/:/g, '-')}` // Replace all colons for class selector
    }

    return {
        eventIdLower,
        eventIdKebab,
        typeLower,
        targetLower,
        typeKebab,
        targetKebab,
    }
}

/**
 * Handles the Fetch fallback logic for postEvent.
 * @private
 */
async function _handleFetchFallback(
    firEvent,
    fetchFn,
    windowLocation,
    processEventsFn,
    dispatchSingleEventFn,
    now // Current timestamp, passed from postEvent
) {
    const { eventIdLower, eventIdKebab } = generateFirEventNames(
        firEvent.event_id,
        '_temp_suffix_for_ids_only_'
    ) // Suffix doesn't matter here
    const body = JSON.stringify({ ...firEvent, ts: now }) // Ensure timestamp is in the body

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
            windowLocation.href = response.url // Note: direct assignment
            return
        }

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`)
        }

        const contentType = response.headers.get('content-type')
        if (contentType && contentType.includes('application/json')) {
            const serverEvents = await response.json()
            if (serverEvents) {
                processEventsFn(serverEvents)
            }
        } else {
            console.log('Received non-JSON response, ignoring.')
        }
    } catch (error) {
        console.error(
            `${firEvent.event_id} fetch error: ${error}, request body: ${body}`
        )
        const errorNames = generateFirEventNames(firEvent.event_id, 'error')
        dispatchSingleEventFn({
            type: errorNames.typeLower,
            target: errorNames.targetLower,
            detail: { error: error.message },
        })
        if (errorNames.typeKebab) {
            dispatchSingleEventFn({
                type: errorNames.typeKebab,
                target: errorNames.targetKebab,
                detail: { error: error.message },
            })
        }
    }
}

export const postEvent = async (
    firEvent,
    socket,
    processEvents = processAndDispatchServerEvents,
    dispatchSingleEvent = dispatchSingleServerEvent,
    fetchFn = fetch,
    windowLocation = window.location,
    now = Date.now,
    generateFirEventNamesFn = generateFirEventNames // Add as parameter with default
) => {
    if (!firEvent.event_id && !firEvent.element_id) {
        console.error(
            "event id is empty and element id is not set. can't emit event"
        )
        return
    }

    const eventId = firEvent.event_id || firEvent.element_id
    const timestamp = now()

    const dispatchFirEvent = (suffix, detail) => {
        const names = generateFirEventNamesFn(eventId, suffix) // Use the injected/defaulted function
        dispatchSingleEvent({
            type: names.typeLower,
            target: names.targetLower,
            detail,
        })
        if (names.typeKebab) {
            dispatchSingleEvent({
                type: names.typeKebab,
                target: names.targetKebab,
                detail,
            })
        }
    }

    dispatchFirEvent('pending', { params: firEvent.params })

    const payload = { ...firEvent, ts: timestamp }

    if (socket && socket.emit && socket.emit(payload)) {
        return
    }

    try {
        const response = await fetchFn(windowLocation.pathname, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-FIR-MODE': 'event',
            },
            body: JSON.stringify(payload),
        })

        if (response.redirected) {
            windowLocation.href = response.url
            return
        }

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`)
        }

        const contentType = response.headers.get('content-type')
        if (contentType && contentType.includes('application/json')) {
            const serverEvents = await response.json()
            processEvents(serverEvents)
        } else {
            console.log('Received non-JSON response, ignoring.')
        }
    } catch (error) {
        console.error(
            `${eventId} fetch error: ${error}, request body: ${JSON.stringify(
                payload
            )}`
        )
        dispatchFirEvent('error', { error: error.message }) // This will also use generateFirEventNamesFn
    }
}

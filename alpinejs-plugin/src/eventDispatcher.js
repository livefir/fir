// filepath: /Users/adnaan/code/livefir/fir/alpinejs-plugin/src/eventDispatcher.js
const FIR_PREFIX = 'fir:'
const ID_SELECTOR_PREFIX = '#'
const CLASS_SELECTOR_PREFIX = '.'
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
export const dispatchEventOnIdTarget = (event, targetId) => {
    const element = document.getElementById(targetId)
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
export const dispatchEventOnClassTarget = (event, targetClass, key) => {
    let className = targetClass
    // Append key if provided, creating a more specific class selector
    if (key) {
        className += '--' + key
    }
    const elements = document.getElementsByClassName(className)
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

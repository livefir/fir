import websocket from './websocket'
import morph from '@alpinejs/morph'
import {
    createCustomEvent,
    dispatchEventOnIdTarget,
    dispatchEventOnClassTarget,
    isValidFirEvent,
} from './eventDispatcher'

// Utility functions
const isObject = (obj) => {
    return Object.prototype.toString.call(obj) === '[object Object]'
}

// Define magic functions outside of Alpine.magic for better testability
export const createFirMagicFunctions = (el, Alpine, postFn) => {
    // Add postFn parameter
    // Helper functions
    const eventHTML = (event) => {
        let html = ''
        if (event?.detail) {
            html = event.detail.html ? event.detail.html : ''
        }
        return html
    }

    const morphElement = (el, value) => {
        // Allow empty strings, only return if null or undefined
        if (value == null) {
            console.error(`morph value is null or undefined`)
            return
        }
        // If the value is an empty string, clear the element instead of morphing
        if (value === '') {
            el.innerHTML = ''
            return
        }

        Alpine.morph(el, value, {
            key(el) {
                return el.getAttribute('fir-key')
            },
        })
    }

    const toElements = (htmlString) => {
        var template = document.createElement('template')
        template.innerHTML = htmlString
        return template.content.childNodes
    }

    const toElement = (htmlString) => {
        var template = document.createElement('template')
        template.innerHTML = htmlString
        return template.content.firstChild
    }

    const afterElement = (el, value) => {
        // Alternative implementation:
        if (el.parentNode) {
            // Insert the new element before the element that comes *after* el
            el.parentNode.insertBefore(toElement(value), el.nextSibling)
        } else {
            console.error('Element has no parent, cannot insert after')
        }
    }

    const beforeElement = (el, value) => {
        if (el.parentNode) {
            el.parentNode.insertBefore(toElement(value), el)
        } else {
            console.error('Element has no parent, cannot insert before')
        }
    }

    const appendElement = (el, value) => {
        let clonedEl = el.cloneNode(true)
        clonedEl.append(...toElements(value))
        morphElement(el, clonedEl)
    }

    const prependElement = (el, value) => {
        let clonedEl = el.cloneNode(true)
        clonedEl.prepend(...toElements(value))
        morphElement(el, clonedEl)
    }

    const removeElement = (el) => {
        el.remove()
    }

    const removeParentElement = (el) => {
        el.parentElement.remove()
    }

    const getSessionIDFromCookie = () => {
        return document.cookie
            .split('; ')
            .find((row) => row.startsWith('_fir_session_='))
            ?.substring(14)
    }

    // Define individual magic functions
    const replace = () => {
        return function (event) {
            let toHTML = el.cloneNode(false)
            toHTML.innerHTML = eventHTML(event).trim()
            morphElement(el, toHTML.outerHTML)
        }
    }

    const replaceEl = () => {
        return function (event) {
            morphElement(el, eventHTML(event))
        }
    }

    const appendEl = () => {
        return function (event) {
            appendElement(el, eventHTML(event))
        }
    }

    const prependEl = () => {
        return function (event) {
            prependElement(el, eventHTML(event))
        }
    }

    const afterEl = () => {
        return function (event) {
            afterElement(el, eventHTML(event))
        }
    }

    const beforeEl = () => {
        return function (event) {
            beforeElement(el, eventHTML(event))
        }
    }

    const removeEl = () => {
        return function (event) {
            removeElement(el)
        }
    }

    const removeParentEl = () => {
        return function (event) {
            removeParentElement(el)
        }
    }

    const reset = () => {
        return function (event) {
            if (el instanceof HTMLFormElement) {
                el.reset()
            } else {
                console.error('$fir.reset() can only be used on form elements')
            }
        }
    }

    const toggleDisabled = () => {
        return function (event) {
            // Elements that typically support the disabled attribute
            const supportsDisabled = [
                'button',
                'fieldset',
                'input',
                'optgroup',
                'option',
                'select',
                'textarea',
                'command',
                'keygen',
                'progress',
            ]

            // Check if element type supports disabled
            const tagName = el.tagName.toLowerCase()
            if (!supportsDisabled.includes(tagName)) {
                console.error(
                    `$fir.toggleDisabled() cannot be used on <${tagName}> elements`
                )
                return
            }

            // Extract state from the event type
            const eventParts = event.type.split(':')
            const state = eventParts.length >= 3 ? eventParts[2] : ''

            // Determine if we should disable based on the state
            // Disable on 'pending', enable on 'ok', 'error', 'done'
            const shouldDisable = state === 'pending'

            // Apply the disabled state
            if (shouldDisable) {
                el.setAttribute('disabled', '')
                el.setAttribute('aria-disabled', 'true')
            } else {
                el.removeAttribute('disabled')
                el.removeAttribute('aria-disabled')
            }
        }
    }

    const emit = (id, params, target) => {
        return function (event) {
            if (id) {
                if (typeof id !== 'string') {
                    console.error(`id ${id} is not a string.`)
                    return
                }
            } else {
                if (!el.getAttribute('id')) {
                    console.error(
                        `event id is empty and element id is not set. can't emit event`
                    )
                    return
                }
                id = el.getAttribute('id')
            }
            if (params) {
                if (!isObject(params)) {
                    console.error(`params ${params} is not an object.`)
                    return
                }
            } else {
                params = {}
            }

            if (target && !target.startsWith('#') && !target.startsWith('.')) {
                console.error('target must start with # or .')
                return
            }

            // Use the passed-in postFn
            postFn({
                event_id: id,
                params: params,
                target: target,
                element_key: el.getAttribute('fir-key'),
                session_id: getSessionIDFromCookie(),
            })
        }
    }

    const submit = (opts) => {
        // Implementation remains the same
        return function (event) {
            if (event.type !== 'submit' && !(el instanceof HTMLFormElement)) {
                console.error(
                    `event type ${event.type} is not submit nor the element is an instance of HTMLFormElement.
                     $fir.submit() can only be used on forms.`
                )
                return
            }

            let form
            if (el instanceof HTMLFormElement) {
                form = el
            } else {
                form = event.target
            }

            if (
                (!form.getAttribute('id') &&
                    !form.action &&
                    !event.submitter) ||
                (event.submitter && !event.submitter.formAction)
            ) {
                console.error(`event id is empty, form element id is not set, form action is not set,
                or it wasn't sumbmitted by a button with formaction set. can't submit form`)
                return
            }

            let formMethod = form.getAttribute('method')
            if (!formMethod) {
                formMethod = 'get'
            }

            let formData = new FormData(form)
            let eventID

            if (form.getAttribute('id')) {
                eventID = form.getAttribute('id')
            }
            if (form.action) {
                const url = new URL(form.action)
                if (url.searchParams.get('event')) {
                    eventID = url.searchParams.get('event')
                }
            }
            if (event.submitter && event.submitter.formAction) {
                const url = new URL(event.submitter.formAction)
                if (url.searchParams.get('event')) {
                    eventID = url.searchParams.get('event')
                }
            }

            if (event.submitter && event.submitter.name) {
                formData.append(event.submitter.name, event.submitter.value)
            }
            let params = {}
            formData.forEach((value, key) => (params[key] = new Array(value)))
            let target = ''

            if (opts) {
                if (opts.event) {
                    eventID = opts.event
                }
                if (opts.params) {
                    params = opts.params
                }
                if (opts.target) {
                    target = opts.target
                }
            }

            if (target && !target.startsWith('#') && !target.startsWith('.')) {
                console.error('target must start with # or .')
                return
            }

            if (!eventID) {
                console.error(
                    `event id is empty and element id is not set. can't emit event`
                )
                return
            }

            // Use the passed-in postFn
            postFn({
                event_id: eventID,
                params: params,
                is_form: true,
                target: target,
                element_key: el.getAttribute('fir-key'),
                session_id: getSessionIDFromCookie(),
            })

            if (formMethod.toLowerCase() === 'get') {
                const url = new URL(window.location)
                formData.forEach((value, key) => {
                    if (value) {
                        url.searchParams.set(key, value)
                    } else {
                        url.searchParams.delete(key)
                    }
                })

                Object.keys(params).forEach((key) => {
                    if (params[key]) {
                        url.searchParams.set(key, params[key])
                    } else {
                        url.searchParams.delete(key)
                    }
                })

                url.searchParams.forEach((value, key) => {
                    if (!formData.has(key) && !params.hasOwnProperty(key)) {
                        url.searchParams.delete(key)
                    }
                })
                window.history.pushState({}, '', url)
            }
            return
        }
    }

    // Return all the magic functions
    return {
        replace,
        replaceEl,
        appendEl,
        prependEl,
        afterEl,
        beforeEl,
        removeEl,
        removeParentEl,
        reset,
        toggleDisabled,
        emit,
        submit,
    }
}

// Define postEvent function outside Plugin
/**
 * Handles sending the firEvent via WebSocket or Fetch.
 * @param {object} firEvent - The event object to send.
 * @param {Function} dispatchSingleServerEvent - Function to dispatch pending events.
 * @param {Function} processAndDispatchServerEvents - Function to process fetch responses.
 * @param {object|null} socket - The WebSocket instance (or null).
 * @param {Function} fetchFn - The fetch function (usually window.fetch).
 */
export const postEvent = (
    firEvent,
    dispatchSingleServerEvent,
    processAndDispatchServerEvents,
    socket,
    fetchFn = fetch // Default to global fetch
) => {
    if (!firEvent.event_id) {
        // Use console.error for actual errors
        console.error(
            "event id is empty and element id is not set. can't emit event"
        )
        return // Stop execution if no ID
    }

    Object.assign(firEvent, { ts: Date.now() })

    const eventIdLower = firEvent.event_id.toLowerCase()
    const eventIdKebab = firEvent.event_id
        .replace(/([a-z0-9]|(?=[A-Z]))([A-Z])/g, '$1-$2')
        .toLowerCase()

    // Dispatch pending events using the provided dispatcher
    let eventTypeLower = `fir:${eventIdLower}:pending`
    dispatchSingleServerEvent({
        type: eventTypeLower,
        target: `.${eventTypeLower.replaceAll(':', '-')}`,
    })

    if (eventIdLower !== eventIdKebab) {
        let eventTypeKebab = `fir:${eventIdKebab}:pending`
        dispatchSingleServerEvent({
            type: eventTypeKebab,
            target: `.${eventTypeKebab.replaceAll(':', '-')}`,
        })
    }

    // Use the provided socket instance
    if (socket && socket.emit(firEvent)) {
        // Event sent via websocket
    } else {
        // Event sent via HTTP POST using the provided fetch function
        const body = JSON.stringify(firEvent)
        fetchFn(window.location.pathname, {
            // Use fetchFn
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-FIR-MODE': 'event',
            },
            body: body,
        })
            .then((response) => {
                if (response.redirected) {
                    window.location.href = response.url
                    return null // Stop promise chain on redirect
                }
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`)
                }
                const contentType = response.headers.get('content-type')
                if (
                    contentType &&
                    contentType.indexOf('application/json') !== -1
                ) {
                    return response.json()
                } else {
                    return null // Handle non-JSON responses
                }
            })
            .then((serverEvents) => {
                if (serverEvents) {
                    // Use the provided processor function for HTTP responses
                    processAndDispatchServerEvents(serverEvents)
                }
            })
            .catch((error) => {
                console.error(
                    `${firEvent.event_id} fetch error: ${error}, request body: ${body}`
                )
            })
    }
}

const Plugin = (Alpine) => {
    // --- Start Refactored Dispatch Logic ---

    const FIR_PREFIX = 'fir:'
    const ID_SELECTOR_PREFIX = '#'
    const CLASS_SELECTOR_PREFIX = '.'

    /**
     * Dispatches a single server event to the appropriate targets (window, ID, class).
     * Assumes the input serverEvent object is valid.
     */
    const dispatchSingleServerEvent = (serverEvent) => {
        // Use directly imported function
        const event = createCustomEvent(serverEvent.type, serverEvent.detail)

        // Always dispatch on window first
        window.dispatchEvent(event)

        // Dispatch based on target selector
        const target = serverEvent.target
        if (target) {
            if (target.startsWith(ID_SELECTOR_PREFIX)) {
                const targetId = target.substring(1)
                // Use directly imported function
                dispatchEventOnIdTarget(event, targetId)
            } else if (target.startsWith(CLASS_SELECTOR_PREFIX)) {
                const targetClass = target.substring(1)
                // Use directly imported function
                dispatchEventOnClassTarget(event, targetClass, serverEvent.key)
            } else {
                console.warn(
                    `Invalid target format "${target}" for event "${serverEvent.type}". Target must start with # or .`
                )
            }
        }
    }

    /**
     * Processes an array of server events, validates them, adds corresponding ':done' events,
     * and dispatches all resulting events.
     */
    const processAndDispatchServerEvents = (serverEvents) => {
        if (!Array.isArray(serverEvents) || serverEvents.length === 0) {
            return
        }

        // Use directly imported function
        const validEvents = serverEvents.filter(isValidFirEvent)
        const eventsToDispatch = [...validEvents]
        const processedEventNames = new Set()

        validEvents.forEach((serverEvent) => {
            const parts = serverEvent.type.split(':')
            const eventName = parts[1]

            if (
                eventName === 'onevent' ||
                eventName === 'onload' ||
                processedEventNames.has(eventName)
            ) {
                return
            }

            processedEventNames.add(eventName)

            const doneEventType = `${FIR_PREFIX}${eventName}:done`
            eventsToDispatch.push({
                type: doneEventType,
                target: `.${doneEventType.replaceAll(':', '-')}`,
                detail: serverEvent.detail,
            })
        })

        // Use dispatchSingleServerEvent which internally uses the imported functions now
        eventsToDispatch.forEach(dispatchSingleServerEvent)
    }

    // --- End Refactored Dispatch Logic ---

    const getSessionIDFromCookie = () => {
        return document.cookie
            .split('; ')
            .find((row) => row.startsWith('_fir_session_='))
            ?.substring(14)
    }

    // connect to websocket
    let connectURL = `ws://${window.location.host}${window.location.pathname}`
    if (window.location.protocol === 'https:') {
        connectURL = `wss://${window.location.host}${window.location.pathname}`
    }

    let socket
    if (getSessionIDFromCookie()) {
        // fetch HEAD request to check if websocket is enabled
        fetch(window.location.href, {
            method: 'HEAD',
        })
            .then((response) => {
                if (
                    response.headers.get('X-FIR-WEBSOCKET-ENABLED') === 'true'
                ) {
                    // Pass the orchestrator function as the callback
                    socket = websocket(connectURL, [], (events) =>
                        processAndDispatchServerEvents(events)
                    )
                }
            })
            .catch((error) => {
                console.error(error)
            })
    } else {
        console.error('no route id found in cookie. websocket disabled')
    }

    window.addEventListener('fir:reload', () => {
        window.location.reload()
    })

    // source from https://dev.to/iamcherta/hotwire-empty-states-with-alpinejs-4gpo
    /** fir-mutation-observer implements the https://developer.mozilla.org/en-US/docs/Web/API/MutationObserver as an
     * alpine directive. It allows you to observe changes to the DOM and react to them.
     * @example
     * e.g.x-fir-mutation-observer.child-list.subtree="if ($el.children.length === 0) { empty = true } else { empty = false }"
     */
    Alpine.directive(
        'fir-mutation-observer',
        (el, { expression, modifiers }, { evaluateLater, cleanup }) => {
            let callback = evaluateLater(expression)

            callback()

            let observer = new MutationObserver(() => {
                callback()
            })

            observer.observe(el, {
                childList: modifiers.includes('child-list'),
                attributes: modifiers.includes('attributes'),
                subtree: modifiers.includes('subtree'),
            })

            cleanup(() => {
                observer.disconnect()
            })
        }
    )

    // Register the fir magic helper
    Alpine.magic('fir', (el, { Alpine }) => {
        // Create a wrapper function to pass to createFirMagicFunctions
        // This wrapper calls the external postEvent with dependencies from this scope
        const postFnWrapper = (firEvent) => {
            postEvent(
                firEvent,
                dispatchSingleServerEvent, // Pass the function from Plugin scope
                processAndDispatchServerEvents, // Pass the function from Plugin scope
                socket, // Pass the socket instance from Plugin scope
                fetch // Pass the global fetch (or could be a mock in tests)
            )
        }

        // Create the magic functions, passing the wrapper function
        const magicFunctions = createFirMagicFunctions(
            el,
            Alpine,
            postFnWrapper
        )

        // Return an object with all the magic functions
        return {
            replace: magicFunctions.replace,
            replaceEl: magicFunctions.replaceEl,
            appendEl: magicFunctions.appendEl,
            prependEl: magicFunctions.prependEl,
            afterEl: magicFunctions.afterEl,
            beforeEl: magicFunctions.beforeEl,
            removeEl: magicFunctions.removeEl,
            removeParentEl: magicFunctions.removeParentEl,
            reset: magicFunctions.reset,
            toggleDisabled: magicFunctions.toggleDisabled,
            emit: magicFunctions.emit,
            submit: magicFunctions.submit,
        }
    })

    // The old dispatchServerEvents and dispatchServerEvent functions are now replaced
    // by the refactored helper functions (imported) and processAndDispatchServerEvents (local).

    Alpine.plugin(morph)
}

export default Plugin

import websocket from './websocket'
import morph from '@alpinejs/morph'
import {
    createCustomEvent,
    dispatchEventOnIdTarget,
    dispatchEventOnClassTarget,
    isValidFirEvent,
} from './eventDispatcher' // Assuming this file exists and is correct
// Import from new utility modules
import { isObject, getSessionIDFromCookie } from './utils' // Import utils
import * as domUtils from './domUtils' // Import all DOM utils under a namespace

// Define magic functions outside of Alpine.magic for better testability
export const createFirMagicFunctions = (el, Alpine, postFn) => {
    // Define individual magic functions using imported helpers
    const replace = () => {
        return function (event) {
            let toHTML = el.cloneNode(false)
            const html = domUtils.eventHTML(event) // Use imported domUtils.eventHTML
            toHTML.innerHTML = html // Assign potentially untrimmed HTML if original did
            // Use imported domUtils.morphElement, passing Alpine
            domUtils.morphElement(el, toHTML.outerHTML, Alpine)
        }
    }

    const replaceEl = () => {
        return function (event) {
            const html = domUtils.eventHTML(event) // Use imported domUtils.eventHTML
            // Use imported domUtils.morphElement, passing Alpine
            domUtils.morphElement(el, html, Alpine)
        }
    }

    const appendEl = () => {
        return function (event) {
            const html = domUtils.eventHTML(event) // Use imported domUtils.eventHTML
            // Use imported domUtils.appendElement, passing Alpine
            domUtils.appendElement(el, html, Alpine)
        }
    }

    const prependEl = () => {
        return function (event) {
            const html = domUtils.eventHTML(event) // Use imported domUtils.eventHTML
            // Use imported domUtils.prependElement, passing Alpine
            domUtils.prependElement(el, html, Alpine)
        }
    }

    const afterEl = () => {
        return function (event) {
            const html = domUtils.eventHTML(event) // Use imported domUtils.eventHTML
            // Use imported domUtils.afterElement
            domUtils.afterElement(el, html)
        }
    }

    const beforeEl = () => {
        return function (event) {
            const html = domUtils.eventHTML(event) // Use imported domUtils.eventHTML
            // Use imported domUtils.beforeElement
            domUtils.beforeElement(el, html)
        }
    }

    const removeEl = () => {
        return function (event) {
            // Use imported domUtils.removeElement
            domUtils.removeElement(el)
        }
    }

    const removeParentEl = () => {
        return function (event) {
            // Use imported domUtils.removeParentElement
            domUtils.removeParentElement(el)
        }
    }

    // reset and toggleDisabled implementations remain the same
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

    // emit and submit use imported utils
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

// Define postEvent function outside Plugin (as before)
export const postEvent = (
    firEvent,
    dispatchSingleServerEvent,
    processAndDispatchServerEvents,
    socket,
    fetchFn = fetch
) => {
    if (!firEvent.event_id) {
        console.error(
            "event id is empty and element id is not set. can't emit event"
        )
        return
    }

    Object.assign(firEvent, { ts: Date.now() })

    const eventIdLower = firEvent.event_id.toLowerCase()
    const eventIdKebab = firEvent.event_id
        .replace(/([a-z0-9]|(?=[A-Z]))([A-Z])/g, '$1-$2')
        .toLowerCase()

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

    if (socket && socket.emit(firEvent)) {
    } else {
        const body = JSON.stringify(firEvent)
        fetchFn(window.location.pathname, {
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
                    return null
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
                    return null
                }
            })
            .then((serverEvents) => {
                if (serverEvents) {
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
    const FIR_PREFIX = 'fir:'
    const ID_SELECTOR_PREFIX = '#'
    const CLASS_SELECTOR_PREFIX = '.'

    const dispatchSingleServerEvent = (serverEvent) => {
        const event = createCustomEvent(serverEvent.type, serverEvent.detail)
        window.dispatchEvent(event)

        const target = serverEvent.target
        if (target) {
            if (target.startsWith(ID_SELECTOR_PREFIX)) {
                const targetId = target.substring(1)
                dispatchEventOnIdTarget(event, targetId)
            } else if (target.startsWith(CLASS_SELECTOR_PREFIX)) {
                const targetClass = target.substring(1)
                dispatchEventOnClassTarget(event, targetClass, serverEvent.key)
            } else {
                console.warn(
                    `Invalid target format "${target}" for event "${serverEvent.type}". Target must start with # or .`
                )
            }
        }
    }

    const processAndDispatchServerEvents = (serverEvents) => {
        if (!Array.isArray(serverEvents) || serverEvents.length === 0) {
            return
        }

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

        eventsToDispatch.forEach(dispatchSingleServerEvent)
    }

    let connectURL = `ws://${window.location.host}${window.location.pathname}`
    if (window.location.protocol === 'https:') {
        connectURL = `wss://${window.location.host}${window.location.pathname}`
    }
    let socket
    if (getSessionIDFromCookie()) {
        fetch(window.location.href, { method: 'HEAD' })
            .then((response) => {
                if (
                    response.headers.get('X-FIR-WEBSOCKET-ENABLED') === 'true'
                ) {
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

    Alpine.magic('fir', (el, { Alpine }) => {
        const postFnWrapper = (firEvent) => {
            postEvent(
                firEvent,
                dispatchSingleServerEvent,
                processAndDispatchServerEvents,
                socket,
                fetch
            )
        }
        const magicFunctions = createFirMagicFunctions(
            el,
            Alpine,
            postFnWrapper
        )

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

    Alpine.plugin(morph)
}

export default Plugin

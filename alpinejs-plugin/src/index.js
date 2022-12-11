import { Iodine } from '@kingshott/iodine'
import websocket from './websocket'
import morph from '@alpinejs/morph'

const Plugin = (Alpine) => {
    const iodine = new Iodine()

    // init default store
    Alpine.store('fir', {})
    const updateStore = (storeName, data) => {
        if (!isObject(data)) {
            Alpine.store(storeName, data)
            return
        }
        const prevStore = Object.assign({}, Alpine.store(storeName))
        const nextStore = { ...prevStore, ...data }
        Alpine.store(storeName, nextStore)
    }

    // connect to websocket
    let connectURL = `ws://${window.location.host}${window.location.pathname}`
    if (window.location.protocol === 'https:') {
        connectURL = `wss://${window.location.host}${window.location.pathname}`
    }
    const socket = websocket(
        connectURL,
        [],
        (patchOperation) => operations[patchOperation.op](patchOperation),
        updateStore
    )

    Alpine.magic('fir', (el, { Alpine }) => {
        return {
            emit(id, params) {
                return function (event) {
                    if (!(el instanceof HTMLFormElement)) {
                        if (!id && el.id) {
                            id = el.id
                        }

                        if (!id) {
                            console.error(
                                `element ${el} has niether id set nor emit was called with an id.`
                            )
                            return
                        }

                        if (typeof id !== 'string') {
                            console.error(`id ${id} is not a string.`)
                            return
                        }

                        if (event && event.target && event.target.value) {
                            let key = 'targetValue'
                            if (event.target.name) {
                                key = event.target.name
                            }
                            if (!params) {
                                params = { key: event.target.value }
                            } else {
                                params[key] = event.target.value
                            }
                        }

                        post(el, id, params, false)
                        return
                    }

                    if (
                        !id &&
                        !el.id &&
                        !event.submitter &&
                        !event.submitter.formAction
                    ) {
                        console.error(`event id is empty, the form element id is not set, 
                        or it wasn't sumbmitted by a button with formaction set. can't emit event`)
                        return
                    }

                    let inputs = [...el.querySelectorAll('input[data-rules]')]
                    let formErrors = { errors: {} }
                    inputs.map((input) => {
                        const rules = JSON.parse(input.dataset.rules)
                        const isValid = iodine.is(input.value, rules)
                        if (isValid !== true) {
                            formErrors.errors[input.getAttribute('name')] =
                                iodine.getErrorMessage(isValid)
                        }
                    })

                    let formMethod = el.getAttribute('method')
                    if (!formMethod) {
                        formMethod = 'get'
                    }

                    // update form errors store
                    const prevStore = Object.assign({}, Alpine.store(el.id))
                    const nextStore = { ...prevStore, ...formErrors }
                    Alpine.store(el.id, nextStore)
                    if (Object.keys(formErrors.errors).length == 0) {
                        let formData = new FormData(el)
                        let eventID = id
                        if (!eventID) {
                            if (el.id) {
                                eventID = el.id
                            }
                            if (el.action) {
                                const url = new URL(event.submitter.formAction)
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
                        } else {
                            if (typeof eventID !== 'string') {
                                console.error(`id ${eventID} is not a string.`)
                                return
                            }
                        }
                        if (event.submitter && event.submitter.name) {
                            formData.append(
                                event.submitter.name,
                                event.submitter.value
                            )
                        }
                        let params = {}
                        formData.forEach(
                            (value, key) => (params[key] = new Array(value))
                        )
                        post(el, eventID, params, true)
                        if (formMethod.toLowerCase() === 'get') {
                            const url = new URL(window.location)
                            formData.forEach((value, key) =>
                                url.searchParams.set(key, value)
                            )
                            window.history.pushState({}, '', url)
                        }
                        return
                    }
                }
            },
        }
    })

    const selectAll = (operation, callbackfn) => {
        const prevFocusElement = document.activeElement
        const elements = document.querySelectorAll(operation.selector)
        elements.forEach((el) => el && callbackfn(el, operation.value))
        const currFocusElement = document.activeElement
        if (
            prevFocusElement &&
            prevFocusElement.focus &&
            prevFocusElement !== currFocusElement
        ) {
            prevFocusElement.focus()
        }
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

    const operations = {
        replace: (operation) =>
            selectAll(operation, (el, value) => {
                Alpine.morph(el, value, {
                    key(el) {
                        return el.id
                    },
                })
            }),
        after: (operation) =>
            selectAll(operation, (el, value) => {
                el.insertBefore(toElement(value), el.nextSibling)
            }),
        before: (operation) =>
            selectAll(operation, (el, value) => {
                el.insertBefore(toElement(value), el)
            }),
        append: (operation) =>
            selectAll(operation, (el, value) => {
                el.append(...toElements(value))
            }),
        prepend: (operation) =>
            selectAll(operation, (el, value) => {
                el.prepend(...toElements(value))
            }),
        remove: (operation) =>
            selectAll(operation, (el, value) => {
                el.remove()
            }),
        reload: () => window.location.reload(),
        resetForm: (operation) =>
            selectAll(operation, (el, value) => {
                if (el instanceof HTMLFormElement) {
                    el.reset()
                }
            }),
        navigate: (operation) => window.location.replace(operation.selector),
        store: (operation) => updateStore(operation.selector, operation.value),
    }

    const post = (el, eventId, params, isForm) => {
        if (!eventId) {
            throw new Error('event id is required.')
        }
        let detail = {
            elementId: el.id,
            eventId: eventId,
            params: params,
        }

        let eventName = 'fir:emit'
        let startEventName = 'fir:emit-start'
        let startEventNameCamel = 'fir:emitStart'
        let endEventName = 'fir:emit-end'
        let endEventNameCamel = 'fir:emitEnd'

        const options = {
            detail,
            bubbles: true,
            // Allows events to pass the shadow DOM barrier.
            composed: true,
            cancelable: true,
        }

        const eventIdlower = eventId.toLowerCase()
        // camel to kebab case
        const eventIdKebab = eventId
            .replace(/([a-z0-9]|(?=[A-Z]))([A-Z])/g, '$1-$2')
            .toLowerCase()

        el.dispatchEvent(new CustomEvent(startEventName, options))
        el.dispatchEvent(
            new CustomEvent(`${startEventName}:${eventIdlower}`, options)
        )
        el.dispatchEvent(
            new CustomEvent(`${startEventNameCamel}:${eventIdlower}`, options)
        )
        el.dispatchEvent(
            new CustomEvent(`${startEventName}:${eventIdKebab}`, options)
        )
        el.dispatchEvent(
            new CustomEvent(`${eventName}:${eventIdKebab}`, options)
        )

        el.dispatchEvent(
            new CustomEvent(`${eventName}:${eventIdlower}`, options)
        )

        const event = {
            event_id: eventId,
            params: params,
            is_form: isForm,
        }
        if (socket.emit(event)) {
            el.dispatchEvent(new CustomEvent(endEventName, options))
            el.dispatchEvent(
                new CustomEvent(`${endEventName}:${eventIdlower}`, options)
            )
            el.dispatchEvent(
                new CustomEvent(`${eventName}:${eventIdlower}`, options)
            )
            el.dispatchEvent(
                new CustomEvent(`${endEventNameCamel}:${eventIdlower}`, options)
            )
            el.dispatchEvent(
                new CustomEvent(`${endEventName}:${eventIdKebab}`, options)
            )
            el.dispatchEvent(
                new CustomEvent(`${eventName}:${eventIdKebab}`, options)
            )
        } else {
            fetch(window.location.pathname, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-FIR-MODE': 'event',
                },
                body: JSON.stringify(event),
            })
                .then((response) => response.json())
                .then((patchOperations) => {
                    patchOperations.forEach((patchOperation) => {
                        operations[patchOperation.op](patchOperation)
                    })
                })
                .catch((error) => {
                    console.error(
                        `${endEventName} error: ${error}, detail: ${detail}`,
                        error
                    )
                })
                .finally(() => {
                    el.dispatchEvent(new CustomEvent(endEventName, options))
                    el.dispatchEvent(
                        new CustomEvent(
                            `${endEventName}:${eventIdlower}`,
                            options
                        )
                    )
                    el.dispatchEvent(
                        new CustomEvent(`${eventName}:${eventIdlower}`, options)
                    )
                    el.dispatchEvent(
                        new CustomEvent(
                            `${endEventNameCamel}:${eventIdlower}`,
                            options
                        )
                    )
                    el.dispatchEvent(
                        new CustomEvent(
                            `${endEventName}:${eventIdKebab}`,
                            options
                        )
                    )
                    el.dispatchEvent(
                        new CustomEvent(`${eventName}:${eventIdKebab}`, options)
                    )
                })
        }
    }

    Alpine.plugin(morph)
}

const isObject = (obj) => {
    return Object.prototype.toString.call(obj) === '[object Object]'
}

export default Plugin

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
            replace(patch) {
                return function (event) {
                    post(el, buildFirEvent(el, event, patch, 'replace'))
                }
            },
            append(patch) {
                return function (event) {
                    post(el, buildFirEvent(el, event, patch, 'append'))
                }
            },
            prepend(patch) {
                return function (event) {
                    post(el, buildFirEvent(el, event, patch, 'prepend'))
                }
            },
            after(patch) {
                return function (event) {
                    post(el, buildFirEvent(el, event, patch, 'after'))
                }
            },
            before(patch) {
                return function (event) {
                    post(el, buildFirEvent(el, event, patch, 'before'))
                }
            },
            remove(patch) {
                return function (event) {
                    post(el, buildFirEvent(el, event, patch, 'remove'))
                }
            },
            emit(id, params) {
                return function (event) {
                    if (id) {
                        if (typeof id !== 'string') {
                            console.error(`id ${id} is not a string.`)
                            return
                        }
                    } else {
                        if (!el.id) {
                            console.error(
                                `event id is empty and element id is not set. can't emit event`
                            )
                            return
                        }
                        id = el.id
                    }
                    if (params) {
                        if (!isObject(params)) {
                            console.error(`params ${params} is not an object.`)
                            return
                        }
                    } else {
                        params = {}
                    }
                    post(el, {
                        event_id: id,
                        params: params,
                    })
                }
            },
            submit(id) {
                return function (event) {
                    if (event.type !== 'submit') {
                        console.error(
                            `event type ${event.type} is not submit. $fir.submit() can only be used on forms.`
                        )
                        return
                    }

                    const form = event.target

                    if (
                        !id &&
                        !form.id &&
                        !event.submitter &&
                        !event.submitter.formAction
                    ) {
                        console.error(`event id is empty, the form element id is not set, 
                        or it wasn't sumbmitted by a button with formaction set. can't emit event`)
                        return
                    }

                    let inputs = [...form.querySelectorAll('input[data-rules]')]
                    let formErrors = { errors: {} }
                    inputs.map((input) => {
                        const rules = JSON.parse(input.dataset.rules)
                        const isValid = iodine.is(input.value, rules)
                        if (isValid !== true) {
                            formErrors.errors[input.getAttribute('name')] =
                                iodine.getErrorMessage(isValid)
                        }
                    })

                    let formMethod = form.getAttribute('method')
                    if (!formMethod) {
                        formMethod = 'get'
                    }

                    // update form errors store
                    const prevStore = Object.assign({}, Alpine.store(form.id))
                    const nextStore = { ...prevStore, ...formErrors }
                    Alpine.store(form.id, nextStore)
                    if (Object.keys(formErrors.errors).length == 0) {
                        let formData = new FormData(form)
                        let eventID = id
                        if (!eventID) {
                            if (form.id) {
                                eventID = form.id
                            }
                            if (form.action) {
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

                        const detail = {
                            isForm: true,
                            form_id: form.id,
                            form_data: params,
                        }

                        const options = {
                            detail,
                            bubbles: true,
                            // Allows events to pass the shadow DOM barrier.
                            composed: true,
                            cancelable: true,
                        }

                        const submitCustomEvent = new CustomEvent(
                            eventID.toLowerCase(),
                            options
                        )

                        el.dispatchEvent(submitCustomEvent)

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

    const buildFirEvent = (el, event, patch, action) => {
        const buildPatch = (patch) => {
            if (!patch.selector) {
                throw new Error('patchset requires a selector')
            }

            if (!operations[patch.op]) {
                throw new Error(`patchset op ${patch.op} not supported`)
            }
            let value = el.dataset['firBlock']
            if (patch.block) {
                value = patch.block
            } else if (patch.template) {
                value = patch.template
            }
            if (value == undefined) {
                value = el.id
            }

            if (typeof value !== 'string') {
                throw new Error(
                    'patch requires either block or template to be a string'
                )
            }

            return {
                selector: patch.selector,
                op: patch.op,
                value: value,
            }
        }

        if (!patch) {
            patch = {
                op: action,
                selector: `#${el.id}`,
            }
        } else {
            if (typeof patch !== 'object') {
                console.error(
                    `arg to ${action} must be an object. e.g. {selector: '#id', block: 'id'}`
                )
                return {}
            }
            patch.op = action
            if (!patch.selector) {
                patch.selector = `#${el.id}`
            }
        }
        if (event.detail && event.detail.isForm) {
            return {
                event_id: event.type,
                params: event.detail.form_data,
                patchset: [buildPatch(patch)],
                form_id: event.detail.form_id,
            }
        }
        return {
            event_id: event.type,
            params: event.detail,
            patchset: [buildPatch(patch)],
        }
    }

    const post = (el, firEvent) => {
        if (!firEvent.event_id) {
            throw new Error('event id is required.')
        }

        let detail = {
            eventId: firEvent.event_id,
            params: firEvent.params,
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

        const eventIdlower = firEvent.event_id.toLowerCase()
        // camel to kebab case
        const eventIdKebab = firEvent.event_id
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

        if (socket.emit(firEvent)) {
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
                body: JSON.stringify(firEvent),
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

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
        (ev) => dispatchServerEvent(ev),
        updateStore
    )

    window.addEventListener('fir:reload', () => {
        window.location.reload()
    })

    Alpine.directive('fir-store', (el, { expression }, { evaluate }) => {
        const val = evaluate(expression)
        Alpine.store('fir', val)
    })

    Alpine.magic('fir', (el, { Alpine }) => {
        return {
            replace() {
                return function (event) {
                    //console.log('=====================>')
                    //console.log(el)
                    //console.log(event)
                    //console.log(event.detail)
                    let toHTML = el.cloneNode(false)
                    toHTML.innerHTML = event.detail.trim()
                    morphElement(el, toHTML.outerHTML)
                }
            },
            replaceEl() {
                return function (event) {
                    morphElement(el, event.detail)
                }
            },
            appendEl() {
                return function (event) {
                    appendElement(el, event.detail)
                }
            },
            prependEl() {
                return function (event) {
                    prependElement(el, event.detail)
                }
            },
            afterEl() {
                return function (event) {
                    afterElement(el, event.detail)
                }
            },
            beforeEl() {
                return function (event) {
                    beforeElement(el, event.detail)
                }
            },
            removeEl() {
                return function (event) {
                    removeElement(el, event.detail)
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
                        target: el.id,
                    })
                }
            },
            submit(id) {
                return function (event) {
                    if (
                        event.type !== 'submit' &&
                        !(el instanceof HTMLFormElement)
                    ) {
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
                        (!id && !form.id && !form.action && !event.submitter) ||
                        (event.submitter && !event.submitter.formAction)
                    ) {
                        console.error(`event id is empty, form element id is not set, form action is not set,
                        or it wasn't sumbmitted by a button with formaction set. can't submit form`)
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

                        let redirect = false
                        if (formMethod.toLowerCase() === 'post') {
                            redirect = true
                        }
                        // post event to server
                        post(el, {
                            event_id: eventID,
                            params: params,
                            form_id: form.id,
                            target: el.id,
                            redirect: redirect,
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
                            url.searchParams.forEach((value, key) => {
                                if (!formData.has(key)) {
                                    url.searchParams.delete(key)
                                }
                            })
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

    const morphElement = (el, value) => {
        Alpine.morph(el, value, {
            key(el) {
                return el.id
            },
        })
    }

    const afterElement = (el, value) => {
        el.insertBefore(toElement(value), el.nextSibling)
    }

    const beforeElement = (el, value) => {
        el.insertBefore(toElement(value), el)
    }

    const appendElement = (el, value) => {
        el.append(...toElements(value))
    }

    const prependElement = (el, value) => {
        el.prepend(...toElements(value))
    }

    const removeElement = (el) => {
        el.remove()
    }

    const dispatchServerEvent = (serverEvent) => {
        const customEvent = new CustomEvent(serverEvent.type, {
            detail: serverEvent.detail,
            bubbles: true,
            // Allows events to pass the shadow DOM barrier.
            composed: true,
            cancelable: true,
        })
        if (!serverEvent.target) {
            //document.dispatchEvent(event)
            window.dispatchEvent(customEvent)
            return
        }

        const elem = document.getElementById(serverEvent.target)
        const getSiblings = (elm) =>
            elm &&
            elm.parentNode &&
            [...elm.parentNode.children].filter(
                (node) =>
                    node != elm &&
                    (node.hasAttribute(`@${serverEvent.type}`) ||
                        node.hasAttribute(`x-on:${serverEvent.type}`))
            )

        const sibs = getSiblings(elem)
        sibs.forEach((sib) => sib.dispatchEvent(customEvent))
        elem.dispatchEvent(customEvent)
    }

    const post = (el, firEvent) => {
        if (!firEvent.event_id) {
            throw new Error('event id is required.')
        }

        let detail = {
            eventId: firEvent.event_id,
            params: firEvent.params,
        }

        const options = {
            detail,
            bubbles: true,
            // Allows events to pass the shadow DOM barrier.
            composed: true,
            cancelable: true,
        }

        const eventIdLower = firEvent.event_id.toLowerCase()
        // camel to kebab case
        const eventIdKebab = firEvent.event_id
            .replace(/([a-z0-9]|(?=[A-Z]))([A-Z])/g, '$1-$2')
            .toLowerCase()

        el.dispatchEvent(
            new CustomEvent(`fir:${eventIdLower}:pending`, options)
        )
        if (eventIdLower !== eventIdKebab) {
            el.dispatchEvent(
                new CustomEvent(`fir:${eventIdKebab}:pending`, options)
            )
        }

        if (socket.emit(firEvent)) {
        } else {
            const body = JSON.stringify(firEvent)
            fetch(window.location.pathname, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-FIR-MODE': 'event',
                },
                body: body,
            })
                .then((response) => response.json())
                .then((serverEvents) => {
                    serverEvents.forEach((ev) => {
                        dispatchServerEvent(ev)
                    })
                })
                .catch((error) => {
                    console.error(
                        `${eventIdLower} error: ${error}, request body: ${body}`,
                        error
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

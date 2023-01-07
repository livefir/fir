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

    const getRouteIDFromCookie = () => {
        return document.cookie
            .split('; ')
            .find((row) => row.startsWith('_fir_route_id='))
            ?.split('=')[1]
    }

    // connect to websocket
    let connectURL = `ws://${window.location.host}${window.location.pathname}`
    if (window.location.protocol === 'https:') {
        connectURL = `wss://${window.location.host}${window.location.pathname}`
    }

    let socket
    if (getRouteIDFromCookie()) {
        socket = websocket(
            connectURL,
            [],
            (events) => dispatchServerEvents(events),
            updateStore
        )
    } else {
        console.error('no route id found in cookie. websocket disabled')
    }

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
                    post(el, {
                        event_id: id,
                        params: params,
                        target: el.getAttribute('id'),
                        route_id: getRouteIDFromCookie(),
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
                        (!id &&
                            !form.getAttribute('id') &&
                            !form.action &&
                            !event.submitter) ||
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
                    const prevStore = Object.assign(
                        {},
                        Alpine.store(form.getAttribute('id'))
                    )
                    const nextStore = { ...prevStore, ...formErrors }
                    Alpine.store(form.getAttribute('id'), nextStore)
                    if (Object.keys(formErrors.errors).length == 0) {
                        let formData = new FormData(form)
                        let eventID = id
                        if (!eventID) {
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
                            is_form: true,
                            target: el.getAttribute('id'),
                            redirect: redirect,
                            route_id: getRouteIDFromCookie(),
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

    const dispatchServerEvents = (serverEvents) => {
        if (!serverEvents) {
            console.error(`server events is empty`)
            return
        }
        if (serverEvents.length == 0) {
            console.error(`server events is empty`)
            return
        }

        const doneEvents = new Set()
        serverEvents.forEach((serverEvent) => {
            if (!serverEvent) {
                console.error(`server event is empty`)
                return
            }
            if (serverEvent && serverEvent.type == '') {
                console.error(`server event type is empty`)
                return
            }

            const parts = serverEvent.type.split(':')
            if (parts.length < 2) {
                console.error(
                    `server event type ${serverEvent.type} is invalid`
                )
                return
            }

            if (parts[0] != 'fir') {
                console.error(
                    `server event type ${serverEvent.type} is invalid`
                )
                return
            }
            const eventName = parts[1]

            if (eventName === 'onevent' || eventName === 'onload') {
                return
            }
            if (doneEvents.has(eventName)) {
                return
            }
            doneEvents.add(eventName)
            serverEvents.push({
                type: `fir:${eventName}:done`,
                target: serverEvent.target,
                detail: serverEvent.detail,
            })
        })

        for (const doneEvent of doneEvents) {
            serverEvents.push(doneEvent)
        }
        serverEvents.forEach((serverEvent) => {
            dispatchServerEvent(serverEvent)
        })
    }

    const dispatchServerEvent = (serverEvent) => {
        const opts = {
            detail: serverEvent.detail,
            bubbles: true,
            // Allows events to pass the shadow DOM barrier.
            composed: true,
            cancelable: true,
        }
        const renderEvent = new CustomEvent(serverEvent.type, opts)
        if (!serverEvent.target) {
            //document.dispatchEvent(event)
            window.dispatchEvent(renderEvent)
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
        sibs.forEach((sib) => {
            sib.dispatchEvent(renderEvent)
        })
        elem.dispatchEvent(renderEvent)
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

        if (socket && socket.emit(firEvent)) {
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
                    dispatchServerEvents(serverEvents)
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

import websocket from './websocket'
import morph from '@alpinejs/morph'

const Plugin = (Alpine) => {
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
                    socket = websocket(connectURL, [], (events) =>
                        dispatchServerEvents(events)
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
        return {
            replace(detail) {
                return function (event) {
                    //console.log('=====================>')
                    //console.log(el)
                    //console.log(event)
                    //console.log(event.detail)
                    let eventDetail = event.detail
                    if (detail) {
                        eventDetail = detail
                    }
                    let toHTML = el.cloneNode(false)
                    toHTML.innerHTML = eventDetail.trim()
                    morphElement(el, toHTML.outerHTML)
                }
            },
            replaceEl(detail) {
                return function (event) {
                    let eventDetail = event.detail
                    if (detail) {
                        eventDetail = detail
                    }
                    morphElement(el, eventDetail)
                }
            },
            appendEl(detail) {
                return function (event) {
                    let eventDetail = event.detail
                    if (detail) {
                        eventDetail = detail
                    }
                    appendElement(el, eventDetail)
                }
            },
            prependEl(detail) {
                return function (event) {
                    let eventDetail = event.detail
                    if (detail) {
                        eventDetail = detail
                    }
                    prependElement(el, eventDetail)
                }
            },
            afterEl(detail) {
                return function (event) {
                    let eventDetail = event.detail
                    if (detail) {
                        eventDetail = detail
                    }
                    afterElement(el, eventDetail)
                }
            },
            beforeEl(detail) {
                return function (event) {
                    let eventDetail = event.detail
                    if (detail) {
                        eventDetail = detail
                    }
                    beforeElement(el, eventDetail)
                }
            },
            removeEl(detail) {
                return function (event) {
                    let eventDetail = event.detail
                    if (detail) {
                        eventDetail = detail
                    }
                    removeElement(el, eventDetail)
                }
            },
            removeParentEl(detail) {
                return function (event) {
                    let eventDetail = event.detail
                    if (detail) {
                        eventDetail = detail
                    }
                    removeParentElement(el, eventDetail)
                }
            },
            emit(id, params, target) {
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

                    if (
                        target &&
                        !target.startsWith('#') &&
                        !target.startsWith('.')
                    ) {
                        console.error('target must start with # or .')
                        return
                    }
                    post({
                        event_id: id,
                        params: params,
                        target: target,
                        element_key: el.getAttribute('fir-key'),
                        session_id: getSessionIDFromCookie(),
                    })
                }
            },
            submit(opts) {
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
                        formData.append(
                            event.submitter.name,
                            event.submitter.value
                        )
                    }
                    let params = {}
                    formData.forEach(
                        (value, key) => (params[key] = new Array(value))
                    )
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

                    if (
                        target &&
                        !target.startsWith('#') &&
                        !target.startsWith('.')
                    ) {
                        console.error('target must start with # or .')
                        return
                    }

                    if (!eventID) {
                        console.error(
                            `event id is empty and element id is not set. can't emit event`
                        )
                        return
                    }

                    // post event to server
                    post({
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
                            if (
                                !formData.has(key) &&
                                !params.hasOwnProperty(key)
                            ) {
                                url.searchParams.delete(key)
                            }
                        })
                        window.history.pushState({}, '', url)
                    }
                    return
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
        if (!value) {
            console.error(`morph value is null`)
            return
        }
        Alpine.morph(el, value, {
            key(el) {
                return el.getAttribute('fir-key')
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
            let eventType = `fir:${eventName}:done`
            serverEvents.push({
                type: eventType,
                target: `.${eventType.replaceAll(':', '-')}`,
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
        window.dispatchEvent(renderEvent)
        if (serverEvent.target && serverEvent.target.startsWith('#')) {
            const elem = document.getElementById(
                serverEvent.target.substring(1)
            )

            elem.dispatchEvent(renderEvent)
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
        }

        if (serverEvent.target && serverEvent.target.startsWith('.')) {
            const elems = Array.from(
                document.getElementsByClassName(serverEvent.target.substring(1))
            )
            for (let i = 0; i < elems.length; i++) {
                if (
                    elems[i].hasAttribute(`@${serverEvent.type}`) ||
                    elems[i].hasAttribute(`x-on:${serverEvent.type}`)
                ) {
                    elems[i].dispatchEvent(renderEvent)
                }
            }
            // targed with key
            if (serverEvent.key) {
                const elems = Array.from(
                    document.getElementsByClassName(
                        serverEvent.target.substring(1) + '--' + serverEvent.key
                    )
                )
                for (let i = 0; i < elems.length; i++) {
                    if (
                        elems[i].hasAttribute(`@${serverEvent.type}`) ||
                        elems[i].hasAttribute(`x-on:${serverEvent.type}`)
                    ) {
                        elems[i].dispatchEvent(renderEvent)
                    }
                }
            }
        }
    }

    const post = (firEvent) => {
        if (!firEvent.event_id) {
            throw new Error('event id is required.')
        }

        Object.assign(firEvent, { ts: Date.now() })

        const eventIdLower = firEvent.event_id.toLowerCase()
        // camel to kebab case
        const eventIdKebab = firEvent.event_id
            .replace(/([a-z0-9]|(?=[A-Z]))([A-Z])/g, '$1-$2')
            .toLowerCase()

        eventTypeLower = `fir:${eventIdLower}:pending`
        dispatchServerEvent({
            type: eventTypeLower,
            target: `.${eventTypeLower.replaceAll(':', '-')}`,
        })

        if (eventIdLower !== eventIdKebab) {
            eventTypeKebab = `fir:${eventIdKebab}:pending`
            dispatchServerEvent({
                type: eventTypeKebab,
                target: `.${eventTypeKebab.replaceAll(':', '-')}`,
            })
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
                .then((response) => {
                    if (response.redirected) {
                        window.location.href = response.url
                        return
                    }
                    return response.json()
                })
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

import * as domUtils from './domUtils' // Import all DOM utils under a namespace
import { isObject, getSessionIDFromCookie } from './utils' // Import utils

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

    const toggleClass = (...classNames) => {
        return function (event) {
            if (!classNames || classNames.length === 0) {
                console.error(
                    '$fir.toggleClass() requires at least one class name'
                )
                return
            }

            // Extract state from the event type
            const eventParts = event.type.split(':')
            const state = eventParts.length >= 3 ? eventParts[2] : ''

            // Toggle classes based on the state
            // Add classes on 'pending', remove on 'ok', 'error', 'done'
            const shouldAddClasses = state === 'pending'

            classNames.forEach((className) => {
                if (typeof className !== 'string') {
                    console.error(
                        `Class name must be a string, got: ${typeof className}`
                    )
                    return
                }

                if (shouldAddClasses) {
                    el.classList.add(className)
                } else {
                    el.classList.remove(className)
                }
            })
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
                    // isObject is now imported
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
                session_id: getSessionIDFromCookie(), // getSessionIDFromCookie is now imported
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
                session_id: getSessionIDFromCookie(), // getSessionIDFromCookie is now imported
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

    const redirect = (url = '/') => {
        return function (event) {
            // Allow the URL to be passed either as parameter or via event detail
            const targetUrl = event?.detail?.url || url

            if (typeof targetUrl !== 'string') {
                console.error('$fir.redirect() requires a valid URL string')
                return
            }

            // Redirect to the specified URL
            window.location.href = targetUrl
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
        toggleClass,
        emit,
        submit,
        redirect,
    }
}

import { Iodine } from '@kingshott/iodine';
import websocket from "./websocket";
import morph from "@alpinejs/morph";

const Plugin = (Alpine) => {
    const iodine = new Iodine();

    // init default store
    Alpine.store("fir", {})
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
    if (window.location.protocol === "https:") {
        connectURL = `wss://${window.location.host}${window.location.pathname}`
    }
    websocket(connectURL, [], (patchOperation) => operations[patchOperation.op](patchOperation), updateStore);

    Alpine.directive('fir-store', (el, { expression }, { evaluate }) => {
        const val = evaluate(expression)
        if (isObject(val)) {
            for (const [key, value] of Object.entries(val)) {
                Alpine.store(key, value)
            }
            return
        }
    })

    Alpine.magic('fir', (el, { Alpine }) => {
        return {
            emit(id, params) {
                post(el, id, params)
            },
            navigate(to) {
                if (!to) {
                    return
                }
                window.location.href = to;
            },
            submit(eventID) {
                if (!(el instanceof HTMLFormElement)) {
                    console.error("Element is not a form. Can't submit event. Please use $fir.emit for non-form elements")
                    return
                }
                let inputs = [...el.querySelectorAll("input[data-rules]")];
                let formErrors = { errors: {} }
                inputs.map((input) => {
                    const rules = JSON.parse(input.dataset.rules);
                    const isValid = iodine.is(input.value, rules);
                    if (isValid !== true) {
                        formErrors.errors[input.getAttribute("name")] = iodine.getErrorMessage(isValid)
                    }
                });

                const formName = el.getAttribute("name")
                // update form errors store
                const prevStore = Object.assign({}, Alpine.store(formName))
                const nextStore = { ...prevStore, ...formErrors }
                Alpine.store(formName, nextStore)
                if (Object.keys(formErrors.errors).length == 0) {
                    let formData = new FormData(el);
                    let params = {};
                    formData.forEach((value, key) => params[key] = value);
                    params["formName"] = formName;
                    post(el, eventID, params)
                    return;
                }
            }
        }
    })

    const selectAll = (operation, callbackfn) => {
        const prevFocusElement = document.activeElement;
        const elements = document.querySelectorAll(operation.selector);
        elements.forEach(el => el && callbackfn(el, operation.value));
        const currFocusElement = document.activeElement;
        if (prevFocusElement && prevFocusElement.focus && prevFocusElement !== currFocusElement) {
            prevFocusElement.focus();
        }
    }

    const toElements = (htmlString) => {
        var template = document.createElement('template');
        template.innerHTML = htmlString;
        return template.content.childNodes;
    }

    const toElement = (htmlString) => {
        var template = document.createElement('template');
        template.innerHTML = htmlString;
        return template.content.firstChild;
    }

    const operations = {
        morph: operation => selectAll(operation, (el, value) => {
            Alpine.morph(el, value, {
                key(el) {
                    return el.id
                }
            })
        }),
        after: operation => selectAll(operation, (el, value) => {
            el.insertBefore(toElement(value), el.nextSibling)
        }),
        before: operation => selectAll(operation, (el, value) => {
            el.insertBefore(toElement(value), el)
        }),
        append: operation => selectAll(operation, (el, value) => {
            el.append(...toElements(value))
        }),
        prepend: operation => selectAll(operation, (el, value) => {
            el.prepend(...toElements(value))
        }),
        remove: operation => selectAll(operation, (el, value) => {
            el.remove()
        }),
        reload: () => window.location.reload(),
        resetForm: operation => selectAll(operation, (el, value) => {
            if (el instanceof HTMLFormElement) {
                el.reset()
            }
        }),
        store: (operation) => updateStore(operation.selector, operation.value)
    }

    const post = (el, id, params) => {
        let detail = {
            id: el.id,
            eventId: id,
            params: params,
        }

        let startEventName = "fir:emit-start"
        let endEventName = "fir:emit-end"
        if (el instanceof HTMLFormElement) {
            detail['formName'] = el.getAttribute("name");
            startEventName = "fir:submit-start"
            endEventName = "fir:submit-end"
        }

        const options = {
            detail,
            bubbles: true,
            // Allows events to pass the shadow DOM barrier.
            composed: true,
            cancelable: true,
        }


        el.dispatchEvent(new CustomEvent(startEventName, options))
        if (detail.id) {
            el.dispatchEvent(new CustomEvent(`${startEventName}:${detail.id}`, options))
        }
        if (detail.formName) {
            el.dispatchEvent(new CustomEvent(`${startEventName}:${detail.formName}`, options))
        }

        fetch(window.location.pathname, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-FIR-MODE': 'event'
            },
            body: JSON.stringify({
                id: id,
                params: params,
            }),
        })
            .then(response => response.json())
            .then(patchOperations => {
                patchOperations.forEach(patchOperation => {
                    operations[patchOperation.op](patchOperation)
                });
            })
            .catch((error) => {
                console.error(`${endEventName} error: ${error}, detail: ${detail}`, error,);
            }).finally(() => {
                el.dispatchEvent(new CustomEvent(endEventName, options))
                if (detail.id) {
                    el.dispatchEvent(new CustomEvent(`${endEventName}:${detail.id}`, options))
                }
                if (detail.formName) {
                    el.dispatchEvent(new CustomEvent(`${endEventName}:${detail.formName}`, options))
                }
            });
    }

    Alpine.plugin(morph)
}


const isObject = (obj) => {
    return Object.prototype.toString.call(obj) === '[object Object]';
};


export default Plugin
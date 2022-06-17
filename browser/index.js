import Alpine from 'alpinejs';
import morph from '@alpinejs/morph';
import persist from '@alpinejs/persist'
import { Iodine } from '@kingshott/iodine';
import websocket from "./websocket";

const iodine = new Iodine();

const isObject = (obj) => {
    return Object.prototype.toString.call(obj) === '[object Object]';
};


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
        emit: post,
        navigate(to) {
            if (!to) {
                return
            }
            window.location.href = to;
        },
        submit(eventID) {

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
                post(eventID, params)
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

const operations = {
    morph: operation => selectAll(operation, (el, value) => {
        Alpine.morph(el, value, {
            key(el) {
                return el.id
            }
        })
    }),
    // browser
    reload: () => window.location.reload(),
    store: (operation) => updateStore(operation.selector, operation.value)
}


let connectURL = `ws://${window.location.host}${window.location.pathname}`
if (window.location.protocol === "https:") {
    connectURL = `wss://${window.location.host}${window.location.pathname}`
}

const post = (id, params) => {
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
            console.error('Error:', error);
        });
}

websocket(connectURL, [], (patchOperation) => operations[patchOperation.op](patchOperation), updateStore);

Alpine.plugin(morph)
Alpine.plugin(persist)
window.Alpine = Alpine
Alpine.start()


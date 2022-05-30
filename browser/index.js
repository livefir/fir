import operations from "./operations";
import { Iodine } from '@kingshott/iodine';

const iodine = new Iodine();

const isObject = (obj) => {
    return Object.prototype.toString.call(obj) === '[object Object]';
};
document.addEventListener('alpine:init', () => {
    let connectURL = `ws://${window.location.host}${window.location.pathname}`
    if (window.location.protocol === "https:") {
        connectURL = `wss://${window.location.host}${window.location.pathname}`
    }
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

    const dispatch = eventDispatcher(connectURL, [], (eventData) => operations[eventData.op](eventData), updateStore);
    dispatch("init", {})

    Alpine.magic('fir', (el, { Alpine }) => {
        return {
            dispatch: dispatch,
            submit(eventID) {

                let inputs = [...el.querySelectorAll("input[data-rules]")];
                let formErrors = { errors: {} }
                inputs.map((input) => {
                    const rules = JSON.parse(input.dataset.rules);
                    const isValid = iodine.is(input.value, rules);
                    console.log(input.value, rules, isValid)
                    if (isValid !== true) {
                        console.log("invalid", input.getAttribute("name"))
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
                    dispatch(eventID, params)
                    return;
                }
            }
        }
    })
})

const reopenTimeouts = [2000, 5000, 10000, 30000, 60000];

const eventDispatcher = (
    url,
    socketOptions,
    invokeOp,
    updateStore
) => {
    let socket, openPromise, reopenTimeoutHandler;
    let reopenCount = 0;

    // socket code copied from https://github.com/arlac77/svelte-websocket-store/blob/master/src/index.mjs
    // thank you https://github.com/arlac77 !!
    function reopenTimeout() {
        const n = reopenCount;
        reopenCount++;
        return reopenTimeouts[
            n >= reopenTimeouts.length - 1 ? reopenTimeouts.length - 1 : n
        ];
    }

    function closeSocket() {
        if (reopenTimeoutHandler) {
            clearTimeout(reopenTimeoutHandler);
        }

        if (socket) {
            socket.close();
            socket = undefined;
        }
    }

    function reOpenSocket() {
        closeSocket();
        reopenTimeoutHandler = setTimeout(() => {
            openSocket().then(() => {
                //console.log("socket connected");
                // location.reload(true)
            }).catch(e => {
                console.error(e)
            })
        },
            reopenTimeout());
    }

    async function openSocket() {
        if (reopenTimeoutHandler) {
            clearTimeout(reopenTimeoutHandler);
            reopenTimeoutHandler = undefined;
        }

        // we are still in the opening phase
        if (openPromise) {
            return openPromise;
        }

        try {
            socket = new WebSocket(url, socketOptions);
        } catch (e) {
            // console.log("socket disconnected")
        }


        socket.onclose = event => reOpenSocket();
        socket.onmessage = event => {
            try {
                const eventData = JSON.parse(event.data);
                if (eventData.op) {
                    if (eventData.op == "update-store") {
                        updateStore(eventData.selector, eventData.value)
                    } else {
                        invokeOp(eventData);
                    }

                }
            } catch (e) {
            }

        };

        openPromise = new Promise((resolve, reject) => {
            socket.onerror = error => {
                reject(error);
                openPromise = undefined;
            };
            socket.onopen = event => {
                reopenCount = 0;
                resolve();
                openPromise = undefined;
            };
        });
        return openPromise;
    }

    openSocket().then(() => { }).catch(e => console.error(e));
    return (id, params) => {
        if (!id) {
            throw 'event.id is required';
        }
        const event = {
            id: id,
            params: params,
        }
        const send = () => socket.send(JSON.stringify(event));
        if (!socket || socket && socket.readyState !== WebSocket.OPEN) openSocket().then(send).catch(e => console.error(e));
        else send();
    }
}
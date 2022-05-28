import morphdom from 'morphdom';

const selectAll = (operation, callbackfn) => {
    const prevFocusElement = document.activeElement;
    const elements = document.querySelectorAll(operation.selector);
    elements.forEach(el => el && callbackfn(el, operation.value));
    const currFocusElement = document.activeElement;
    if (prevFocusElement && prevFocusElement.focus && prevFocusElement !== currFocusElement) {
        prevFocusElement.focus();
    }
}

export default {
    // dom mutations
    morph: operation => selectAll(operation, (el, value) => {
        morphdom(el, value, morphOptions)
    }),
    // browser
    reload: operation => window.location.reload()
}


const morphOptions = {
    onBeforeElUpdated: function (fromEl, toEl) {
        // spec - https://dom.spec.whatwg.org/#concept-node-equals
        if (fromEl.isEqualNode(toEl)) {
            return false
        }

        return true
    },
    childrenOnly: false,
}
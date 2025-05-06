// --- Helper functions moved from index.js (implementations kept identical) ---

/**
 * Extracts HTML string from an event detail, defaulting to an empty string.
 * @param {CustomEvent|null|undefined} event - The event object.
 * @returns {string} The HTML string or an empty string.
 */
export const eventHTML = (event) => {
    let html = ''
    if (event?.detail) {
        // Keep original logic: check existence before accessing .html
        html = event.detail.html ? String(event.detail.html) : ''
    }
    // Keep original logic: return untrimmed HTML if that's what it did
    return html
}

/**
 * Converts an HTML string into a NodeList of child nodes.
 * @param {string} htmlString - The HTML string to parse.
 * @returns {NodeList} A list of nodes.
 */
export const toElements = (htmlString) => {
    var template = document.createElement('template')
    template.innerHTML = htmlString
    return template.content.childNodes
}

/**
 * Converts an HTML string into a single Element node (the first child).
 * @param {string} htmlString - The HTML string to parse.
 * @returns {Node|null} The first Element node or null.
 */
export const toElement = (htmlString) => {
    var template = document.createElement('template')
    template.innerHTML = htmlString
    return template.content.firstChild
}

/**
 * Morphs an element with new content using Alpine's morph plugin.
 * Requires the Alpine instance.
 * @param {Element} el - The target element to morph.
 * @param {string|Node} value - The new HTML content string or Node to morph to.
 * @param {object} Alpine - The Alpine instance (needed for Alpine.morph).
 */
export const morphElement = (el, value, Alpine) => {
    // Added Alpine parameter
    // Keep original logic for null/undefined check
    if (value == null) {
        console.error(`morph value is null or undefined`)
        return
    }
    // Keep original logic for empty string check
    if (value === '') {
        el.innerHTML = ''
        return
    }
    // Keep original Alpine.morph call
    Alpine.morph(el, value, {
        key(el) {
            return el.getAttribute('fir-key')
        },
    })
    // NOTE: Removed try/catch if it wasn't in the original working version
}

/**
 * Inserts HTML content after a given element. (Uses standard DOM)
 * @param {Element} el - The reference element.
 * @param {string} value - The HTML string to insert.
 */
export const afterElement = (el, value) => {
    // Keep original implementation using insertBefore and nextSibling
    if (el.parentNode) {
        // Use local toElement
        const newNode = toElement(value)
        // Check if newNode is valid before inserting (implicit in original?)
        if (newNode) {
            el.parentNode.insertBefore(newNode, el.nextSibling)
        }
    } else {
        console.error('Element has no parent, cannot insert after')
    }
}

/**
 * Inserts HTML content before a given element. (Uses standard DOM)
 * @param {Element} el - The reference element.
 * @param {string} value - The HTML string to insert.
 */
export const beforeElement = (el, value) => {
    // Keep original implementation using insertBefore
    if (el.parentNode) {
        // Use local toElement
        const newNode = toElement(value)
        if (newNode) {
            el.parentNode.insertBefore(newNode, el)
        }
    } else {
        console.error('Element has no parent, cannot insert before')
    }
}

/**
 * Appends HTML content to an element using cloning and morphing.
 * Requires the Alpine instance for morphElement.
 * @param {Element} el - The target element.
 * @param {string} value - The HTML string to append.
 * @param {object} Alpine - The Alpine instance.
 */
export const appendElement = (el, value, Alpine) => {
    // Added Alpine parameter
    // Keep original implementation using cloneNode, append, and morphElement
    let clonedEl = el.cloneNode(true)
    // Use local toElements
    clonedEl.append(...toElements(value))
    // Use local morphElement, passing Alpine
    morphElement(el, clonedEl, Alpine)
}

/**
 * Prepends HTML content to an element using cloning and morphing.
 * Requires the Alpine instance for morphElement.
 * @param {Element} el - The target element.
 * @param {string} value - The HTML string to prepend.
 * @param {object} Alpine - The Alpine instance.
 */
export const prependElement = (el, value, Alpine) => {
    // Added Alpine parameter
    // Keep original implementation using cloneNode, prepend, and morphElement
    let clonedEl = el.cloneNode(true)
    // Use local toElements
    clonedEl.prepend(...toElements(value))
    // Use local morphElement, passing Alpine
    morphElement(el, clonedEl, Alpine)
}

/**
 * Removes the given element from the DOM. (Uses standard DOM)
 * @param {Element} el - The element to remove.
 */
export const removeElement = (el) => {
    // Keep original implementation
    el.remove()
}

/**
 * Removes the parent element of the given element from the DOM. (Uses standard DOM)
 * @param {Element} el - The child element whose parent should be removed.
 */
export const removeParentElement = (el) => {
    // Keep original implementation
    if (el.parentElement) {
        el.parentElement.remove()
    } else {
        console.error('Element has no parent element to remove.')
    }
}

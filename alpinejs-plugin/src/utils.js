/**
 * Checks if a value is a plain JavaScript object.
 * @param {any} obj - The value to check.
 * @returns {boolean} True if the value is an object, false otherwise.
 */
export const isObject = (obj) => {
    return Object.prototype.toString.call(obj) === '[object Object]'
}

/**
 * Retrieves the session ID from the '_fir_session_' cookie.
 * (Assuming this logic was previously available in index.js scope)
 * @returns {string|undefined} The session ID or undefined if not found.
 */
export const getSessionIDFromCookie = () => {
    if (typeof document === 'undefined' || !document.cookie) {
        return undefined
    }
    return document.cookie
        .split('; ')
        .find((row) => row.startsWith('_fir_session_='))
        ?.substring(14)
}

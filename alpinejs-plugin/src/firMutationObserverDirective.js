export const firMutationObserverDirective = (Alpine) => {
    Alpine.directive(
        'fir-mutation-observer',
        (el, { expression, modifiers }, { evaluateLater, cleanup }) => {
            let callback = evaluateLater(expression)

            // Initial call
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
}

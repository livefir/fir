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

            // Build observer options from modifiers
            const options = {
                childList: modifiers.includes('child-list'),
                attributes: modifiers.includes('attributes'),
                subtree: modifiers.includes('subtree'),
                characterData: modifiers.includes('character-data'),
                attributeOldValue: modifiers.includes('attribute-old-value'),
                characterDataOldValue: modifiers.includes(
                    'character-data-old-value'
                ),
            }

            // Handle attributeFilter modifier (format: attribute-filter:attr1,attr2,attr3)
            const attributeFilterModifier = modifiers.find((mod) =>
                mod.startsWith('attribute-filter:')
            )
            if (attributeFilterModifier) {
                const filterValue = attributeFilterModifier.split(':')[1]
                if (filterValue) {
                    options.attributeFilter = filterValue
                        .split(',')
                        .map((attr) => attr.trim())
                }
            }

            observer.observe(el, options)

            cleanup(() => {
                observer.disconnect()
            })
        }
    )
}

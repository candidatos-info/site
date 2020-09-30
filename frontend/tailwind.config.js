module.exports = {
    future: {
        // removeDeprecatedGapUtilities: true,
        // purgeLayersByDefault: true,
    },
    purge: [],
    theme: {
        extend: {
            colors: {
                'flux-candidate': 'var(--color-flux-candidate)',
                'flux-visitor': 'var(--color-flux-visitor)',
                'primary-button': 'var(--color-primary-button)',
                'secondary-button': 'var(--color-secondary-button)',
                'disabled-button': 'var(--color-disabled-button)',
                'border-button': 'var(--color-border-button)',
                'disabled-field': 'var(--color-disabled-field)',
                'green-100': 'var(--color-green-100)',
                'green-50': 'var(--color-green-50)',
                'yellow-50': 'var(--color-yellow-50)',
                'text': 'var(--color-text)',
                'red': 'var(--color-red)',
                'orange-25': 'var(--color-orange-25)',
                'white': 'var(--color-white)',
            },
        }
    },
    variants: {},
    plugins: [],
}

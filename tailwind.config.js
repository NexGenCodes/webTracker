/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                primary: 'var(--color-primary)',
                accent: 'var(--color-accent)',
                'bg-dark': 'var(--color-bg-dark)',
                'main': 'var(--color-text-main)',
                'muted': 'var(--color-text-muted)',
                'heading': 'var(--color-heading)',
                'input-bg': 'var(--color-input-bg)',
                'input-border': 'var(--color-input-border)',
            }
        },
    },
    plugins: [],
}

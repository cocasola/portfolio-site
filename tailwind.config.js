const colors = require('tailwindcss/colors')

/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./view/**/*.{go,templ}"],
    theme: {
        extend: {
            colors: {
                primary: colors.sky[500],
                secondary: colors.red[400],
                backdrop: colors.neutral[950],
                faded: colors.neutral[700],
                fg: colors.gray[200],
                subtle: colors.gray[400],
                accent: colors.amber[300]
            }
        }
    },
    plugins: [],
}
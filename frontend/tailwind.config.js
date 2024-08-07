/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./src/**/*.{js,jsx,ts,tsx}"],
    theme: {
        extend: {
            fontFamily: {
                poppins: ["Poppins", "sans-serif"],
                montserrat: ["Montserrat", "sans-serif"], // Add Montserrat here
            },
            colors: {
                darkblue: "#003F87",
                lightblue: "#007AC8",
                darkgray: "#262B2F",
                white: "#FFFFFF",
                lightgray: "#667986",
            },
        },
    },
    plugins: [require("daisyui")],
    daisyui: {
        themes: ["light", "dark"], // Include any themes you want
    },
};

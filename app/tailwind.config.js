/** @type {import('tailwindcss').Config} */
export default {
  content: [],
  purge: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        accent: '#0A84FF',
        light: "#EAEEFF",
        half_light: "#BDC6EB",
        half_dark: "#989EA1",
        light_dark: "#F8F8F8"
      }
    },
  },
  plugins: [],
}


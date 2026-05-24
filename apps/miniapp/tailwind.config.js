/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        blue: '#4D7CFF',
        gold: '#FFC93D',
        green: '#39d98a',
        red: '#FF5C5C',
        ink: '#F4F3FA',
        mut: '#928FA6',
        bg: '#0E0D14',
        panel: '#1A1825',
        panel2: '#211E2D',
      },
      fontFamily: {
        sans: ['Manrope', 'sans-serif'],
        display: ['Unbounded', 'sans-serif'],
      },
    },
  },
  plugins: [],
};

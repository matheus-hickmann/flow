/**
 * Único ponto de definição das cores do tema (design tokens).
 * Use nas páginas as classes semânticas: primary, success, danger, secondary (e variantes -light).
 * Evite setar cores individuais nos componentes; referencie este tema.
 */
const themeColors = {
  primary: '#22c55e',
  'primary-light': '#dcfce7',
  secondary: '#1e3a5f',
  'secondary-light': '#dbeafe',
  success: '#22c55e',
  'success-light': '#dcfce7',
  danger: '#ef4444',
  'danger-light': '#ffe4e6',
};

const neutral = {
  50: '#f8fafc',
  100: '#f1f5f9',
  200: '#e2e8f0',
  300: '#cbd5e1',
  400: '#94a3b8',
  500: '#64748b',
  600: '#475569',
  700: '#334155',
  800: '#1e293b',
  900: '#0f172a',
};

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{html,ts}",
    "./projects/dashboard/src/**/*.{html,ts}",
    "./projects/transactions/src/**/*.{html,ts}",
    "./projects/accounts/src/**/*.{html,ts}",
    "./projects/reports/src/**/*.{html,ts}",
  ],
  theme: {
    extend: {
      colors: {
        ...themeColors,
        neutral,
        ledger: {
          blue: themeColors.secondary,
          green: themeColors.primary,
          'green-light': themeColors['primary-light'],
          'blue-light': themeColors['secondary-light'],
          'red-light': themeColors['danger-light'],
          salmon: '#fecaca',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
};

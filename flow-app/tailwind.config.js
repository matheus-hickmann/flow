/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./src/**/*.{html,ts}'],
  darkMode: 'class',
  theme: {
    extend: {
      fontFamily: {
        sans: ['"IBM Plex Sans"', 'system-ui', 'sans-serif'],
        serif: ['"Instrument Serif"', 'Times New Roman', 'serif'],
        mono: ['"IBM Plex Mono"', 'ui-monospace', 'monospace'],
      },
      colors: {
        // ─── Tokens semânticos preservados (compatibilidade com código atual) ──
        primary:           '#1d1a16',           // ink
        'primary-light':   '#f5f0e6',           // bg
        secondary:         '#6b665e',           // ink-soft
        'secondary-light': '#cfdfeb',           // sky pastel
        success:           '#2c6b3d',
        'success-light':   '#cce8d6',
        danger:            '#a84439',
        'danger-light':    '#edd1d6',
        gold:              '#6b5a16',
        'gold-light':      '#f3e9b9',

        // ─── Paleta Flow — pastel oklch ────────────────────────────────────────
        flow: {
          bg:        '#f5f0e6',
          paper:     '#fbf8f1',
          ink:       '#1d1a16',
          'ink-soft':'#6b665e',
          'ink-mute':'#9a9389',
          line:      '#e3ddcf',
          'line-soft':'#ecead8',
          mint:      '#cce8d6',
          'mint-ink':'#1f4f33',
          peach:     '#f4d8c7',
          'peach-ink':'#7a3a1d',
          butter:    '#f3e9b9',
          'butter-ink':'#6b5a16',
          lilac:     '#dcd2ec',
          'lilac-ink':'#4a3a73',
          rose:      '#edd1d6',
          'rose-ink':'#7a2a3a',
          sage:      '#cdd9be',
          'sage-ink':'#3d5526',
          sky:       '#cfdfeb',
          'sky-ink': '#1f4a6b',
          pos:       '#2c6b3d',
          neg:       '#a84439',
        },
      },
      letterSpacing: {
        'tight-display': '-0.02em',
      },
    },
  },
  plugins: [],
};

/** @type {import('tailwindcss').Config} */
//
// Tailwind config do Flow.
//
// Dark mode: usamos `class` em duas frentes complementares —
//   • Tailwind ativa `dark:` quando há `.dark` no <html>
//   • Os tokens CSS (`var(--flow-*)`) trocam quando há `data-theme="dark"`
//     no <body>. Os utilitários `flow-*` abaixo apontam para essas vars,
//     então respondem aos dois mecanismos automaticamente.
//
// Para usar cores pastel fixas (sem theming): use as classes legacy
// `bg-flow-static-mint`, `bg-flow-static-peach`, etc.
//
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
        // ─── Tokens semânticos preservados (compatibilidade) ──────────
        // Apontam para variáveis CSS — trocam automaticamente no dark.
        primary:           'var(--flow-ink)',
        'primary-light':   'var(--flow-bg)',
        secondary:         'var(--flow-ink-soft)',
        'secondary-light': 'var(--flow-sky)',
        success:           'var(--flow-pos)',
        'success-light':   'var(--flow-pos-bg)',
        danger:            'var(--flow-neg)',
        'danger-light':    'var(--flow-neg-bg)',
        gold:              'var(--flow-butter-ink)',
        'gold-light':      'var(--flow-butter)',

        // ─── Paleta Flow — themed via CSS vars ────────────────────────
        flow: {
          bg:           'var(--flow-bg)',
          paper:        'var(--flow-paper)',
          ink:          'var(--flow-ink)',
          'ink-soft':   'var(--flow-ink-soft)',
          'ink-mute':   'var(--flow-ink-mute)',
          line:         'var(--flow-line)',
          'line-soft':  'var(--flow-line-soft)',
          mint:         'var(--flow-mint)',
          'mint-ink':   'var(--flow-mint-ink)',
          peach:        'var(--flow-peach)',
          'peach-ink':  'var(--flow-peach-ink)',
          butter:       'var(--flow-butter)',
          'butter-ink': 'var(--flow-butter-ink)',
          lilac:        'var(--flow-lilac)',
          'lilac-ink':  'var(--flow-lilac-ink)',
          rose:         'var(--flow-rose)',
          'rose-ink':   'var(--flow-rose-ink)',
          sage:         'var(--flow-sage)',
          'sage-ink':   'var(--flow-sage-ink)',
          sky:          'var(--flow-sky)',
          'sky-ink':    'var(--flow-sky-ink)',
          pos:          'var(--flow-pos)',
          neg:          'var(--flow-neg)',
          band:         'var(--flow-band)',
          hairline:     'var(--flow-hairline)',
        },

        // ─── Pastéis estáticos (não trocam no dark) ────────────────────
        'flow-static': {
          mint:   '#cce8d6',
          peach:  '#f4d8c7',
          butter: '#f3e9b9',
          lilac:  '#dcd2ec',
          rose:   '#edd1d6',
          sage:   '#cdd9be',
          sky:    '#cfdfeb',
        },
      },
      boxShadow: {
        'flow-card': 'var(--flow-shadow-card)',
      },
      letterSpacing: {
        'tight-display': '-0.02em',
      },
    },
  },
  plugins: [],
};

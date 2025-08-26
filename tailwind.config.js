/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./templates/**/*.html",
    "./static/js/**/*.js"
  ],
  theme: {
    extend: {
      colors: {
        // Maritime minimalist color palette
        navy: {
          50: '#f0f4f8',
          100: '#d9e2ec',
          200: '#bcccdc',
          300: '#9fb3c8',
          400: '#829ab1',
          500: '#627d98',
          600: '#486581',
          700: '#334e68',
          800: '#243b53',
          900: '#1a365d', // Primary navy
        },
        slate: {
          50: '#f8fafc',
          100: '#f1f5f9',
          200: '#e2e8f0',
          300: '#cbd5e1',
          400: '#94a3b8',
          500: '#64748b', // Secondary slate
          600: '#475569',
          700: '#334155',
          800: '#1e293b',
          900: '#0f172a',
        },
        coral: {
          50: '#fef2f2',
          100: '#fee2e2',
          200: '#fecaca',
          300: '#fca5a5',
          400: '#f87171',
          500: '#f56565', // Accent coral
          600: '#dc2626',
          700: '#b91c1c',
          800: '#991b1b',
          900: '#7f1d1d',
        },
        background: '#fafafa', // Soft off-white
        surface: '#ffffff',     // Pure white
        text: {
          primary: '#1a202c',
          secondary: '#4a5568',
          tertiary: '#a0aec0',
        }
      },
      fontFamily: {
        'system': [
          '-apple-system',
          'BlinkMacSystemFont',
          '"Segoe UI"',
          'system-ui',
          'sans-serif'
        ]
      },
      spacing: {
        '18': '4.5rem',  // 72px
        '22': '5.5rem',  // 88px
        '88': '22rem',   // 352px
        '96': '24rem',   // 384px
      },
      minHeight: {
        'screen-75': '75vh',
        'screen-90': '90vh',
      },
      maxWidth: {
        'form': '400px',
        'content': '1200px',
      },
      borderRadius: {
        'minimal': '0.375rem', // 6px - consistent minimal radius
      },
      boxShadow: {
        'minimal': '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
        'card': '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
      },
      gridTemplateColumns: {
        'sidebar': '16rem 1fr',
        'dashboard': 'repeat(auto-fit, minmax(240px, 1fr))',
      }
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}
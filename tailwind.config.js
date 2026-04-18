module.exports = {
  content: [
    "./internal/ui/**/*.templ",
    "./internal/ui/**/*.go",
    "./web/template/**/*.html",
  ],
  theme: {
    extend: {
      colors: {
        background: "#f5efe2",
        foreground: "#221f1a",
        card: "#fffdf8",
        "card-foreground": "#221f1a",
        primary: "#b85c38",
        "primary-foreground": "#fffaf4",
        secondary: "#ead7b0",
        "secondary-foreground": "#3f3528",
        muted: "#efe4d2",
        "muted-foreground": "#6c6254",
        accent: "#d0b38a",
        "accent-foreground": "#2e2417",
        border: "#d8c6ab",
        ring: "#b85c38",
        input: "#fffaf4",
        destructive: "#b42318",
      },
      boxShadow: {
        xs: "0 1px 2px rgba(34, 31, 26, 0.08)",
      },
      fontFamily: {
        sans: ["IBM Plex Sans", "Avenir Next", "Segoe UI", "sans-serif"],
        mono: ["IBM Plex Mono", "SFMono-Regular", "monospace"],
      },
    },
  },
  plugins: [],
};

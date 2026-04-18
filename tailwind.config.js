module.exports = {
  content: [
    "./internal/ui/**/*.templ",
    "./internal/ui/**/*.go",
    "./web/template/**/*.html",
  ],
  theme: {
    extend: {
      colors: {
        background: "#ffffff",
        foreground: "#0a0a0a",
        card: "#ffffff",
        "card-foreground": "#0a0a0a",
        primary: "#171717",
        "primary-foreground": "#fafafa",
        secondary: "#f5f5f5",
        "secondary-foreground": "#171717",
        muted: "#f5f5f5",
        "muted-foreground": "#737373",
        accent: "#f5f5f5",
        "accent-foreground": "#171717",
        border: "#e5e5e5",
        ring: "#0a0a0a",
        input: "#e5e5e5",
        destructive: "#dc2626",
      },
      boxShadow: {
        xs: "0 1px 2px 0 rgb(0 0 0 / 0.04)",
      },
      fontFamily: {
        sans: ["Geist", "ui-sans-serif", "system-ui", "sans-serif"],
        mono: ["Geist Mono", "ui-monospace", "Menlo", "monospace"],
      },
    },
  },
  plugins: [],
};

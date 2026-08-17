import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        primary: "var(--color-primary)",
        background: "var(--color-bg)",
        text: "var(--color-text)",
        doing: "var(--color-warning)",
        done: "var(--color-success)"
      }
    }
  },
  plugins: []
};

export default config;

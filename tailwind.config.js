/** @type {import('tailwindcss').Config} */
import { iconsPlugin, getIconCollections } from "@egoist/tailwindcss-icons";

module.exports = {
  content: ["./templates/**/*.{html,js}"],
  theme: {
    extend: {
      gridTemplateColumns: {
        // all auto
        special: "repeat(6,auto)",
      },
    },
    fontFamily: {
      sans: [
        "ui-monospace",
        "SFMono-Regular",
        "Menlo",
        "Monaco",
        "Consolas",
        '"Liberation Mono"',
        '"Courier New"',
        "monospace",
      ],
    },
  },
  plugins: [
    require("@tailwindcss/typography"),
    require("daisyui"),
    iconsPlugin({
      // Select the icon collections you want to use
      collections: getIconCollections(["mdi", "svg-spinners"]),
    }),
  ],
  daisyui: {
    themes: [
      {
        light: {
          // ...require("daisyui/src/theming/themes")["cyberpunk"],
          ...require("daisyui/src/theming/themes")["nord"],
        },
      },
      {
        dark: {
          ...require("daisyui/src/theming/themes")["dracula"],

          // using colors from gruvbox
          primary: "#83a598", // blue
          secondary: "#d3869b", // purple
          accent: "#8ec07c", // aqua
          neutral: "#1d2021", // crust
          "neutral-content": "ebdbb2", // fg-color
          "base-100": "#282828", // base
          "base-content": "#ebdbb2", // base-content
          info: "#85c1dc", // sapphire
          success: "#b8bb26", // green
          warning: "#fabd2f", // yellow
          error: "#fb4936", // red
        },
      },
    ],
  },
};

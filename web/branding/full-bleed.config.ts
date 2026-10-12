import { defineConfig, minimal2023Preset } from "@vite-pwa/assets-generator/config"

const fullBleed = { padding: 0, resizeOptions: { background: "#0f172a" } }

export default defineConfig({
  preset: {
    ...minimal2023Preset,
    transparent: { ...minimal2023Preset.transparent, sizes: [], favicons: [] },
    maskable: { ...minimal2023Preset.maskable, ...fullBleed },
    apple: { ...minimal2023Preset.apple, ...fullBleed },
  },
  images: ["branding/logo-full.svg"],
})

import { defineConfig, minimal2023Preset } from "@vite-pwa/assets-generator/config"

export default defineConfig({
  preset: {
    ...minimal2023Preset,
    maskable: { ...minimal2023Preset.maskable, sizes: [] },
    apple: { ...minimal2023Preset.apple, sizes: [] },
  },
  images: ["public/logo.svg"],
})

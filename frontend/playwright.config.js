import { defineConfig } from '@playwright/test'

// End-to-end tests: a real Chrome clicks through the real app.
// Start the backend (go run .) and frontend (npm run dev) first, then:  npm run test:e2e
export default defineConfig({
  testDir: './e2e',
  timeout: 120_000,
  reporter: [['list'], ['json', { outputFile: 'e2e/results.json' }]],
  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:5173',
    channel: 'chrome', // uses your installed Google Chrome
    viewport: { width: 1280, height: 800 },
  },
})

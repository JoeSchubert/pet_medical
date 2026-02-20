#!/usr/bin/env node
/**
 * Captures screenshots for docs: all main pages and pet-detail tabs.
 * Run with: npm run capture-screenshots (from frontend/) while the app is running at http://localhost:8080.
 * Requires: npm install -D playwright && npx playwright install chromium
 */
import { chromium } from 'playwright'
import { fileURLToPath } from 'url'
import path from 'path'

const baseUrl = process.env.BASE_URL || 'http://localhost:8080'
const outDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../documentation/screenshots')

const wait = (ms) => new Promise((r) => setTimeout(r, ms))

async function main() {
  const browser = await chromium.launch({ headless: true })
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 } })
  try {
    // ---- Login page ----
    await page.goto(`${baseUrl}/login`, { waitUntil: 'networkidle' })
    await page.screenshot({ path: path.join(outDir, 'login.png') })
    console.log('Saved login.png')

    // ---- Login ----
    await page.fill('#login-email', 'admin@example.com')
    await page.fill('#login-password', 'admin123')
    await page.click('#login-submit')
    await page.waitForURL((url) => !url.pathname.includes('login'), { timeout: 15000 }).catch(() => {})
    if (page.url().includes('/login')) {
      console.warn('Login may have failed; check app and credentials')
    }

    // ---- Dashboard ----
    await page.goto(baseUrl + '/', { waitUntil: 'networkidle' })
    await page.waitForSelector('.page, .pet-grid, .card-panel', { timeout: 5000 }).catch(() => {})
    await wait(500)
    await page.screenshot({ path: path.join(outDir, 'dashboard.png') })
    console.log('Saved dashboard.png')

    // ---- Pet detail (and tabs) ----
    const petLink = page.locator('a.pet-card[href^="/pets/"], .pet-grid a[href^="/pets/"]').first()
    const petCount = await petLink.count()
    let petId = null
    if (petCount > 0) {
      const href = await petLink.getAttribute('href')
      const match = href && href.match(/\/pets\/([^/]+)/)
      if (match) petId = match[1]
      await page.goto(baseUrl + href, { waitUntil: 'networkidle' })
      await page.waitForSelector('.pet-detail-layout, .pet-hero', { timeout: 5000 }).catch(() => {})
      await wait(500)
      await page.screenshot({ path: path.join(outDir, 'pet-detail.png') })
      console.log('Saved pet-detail.png')

      // Vaccinations tab is default; capture Weights tab (give Recharts line time to draw)
      const weightsTab = page.locator('#pet-tab-btn-weights, button[role="tab"]').filter({ hasText: /weight/i }).first()
      if (await weightsTab.count() > 0) {
        await weightsTab.click()
        await page.waitForSelector('.recharts-wrapper, .recharts-surface', { timeout: 5000 }).catch(() => {})
        await wait(1500)
        await page.screenshot({ path: path.join(outDir, 'pet-detail-weights.png') })
        console.log('Saved pet-detail-weights.png')
      }
      // Documents tab
      const docsTab = page.locator('#pet-tab-btn-documents, button[role="tab"]').filter({ hasText: /document/i }).first()
      if (await docsTab.count() > 0) {
        await docsTab.click()
        await wait(400)
        await page.screenshot({ path: path.join(outDir, 'pet-detail-documents.png') })
        console.log('Saved pet-detail-documents.png')
      }
      // Photos tab
      const photosTab = page.locator('#pet-tab-btn-photos, button[role="tab"]').filter({ hasText: /photo/i }).first()
      if (await photosTab.count() > 0) {
        await photosTab.click()
        await wait(400)
        await page.screenshot({ path: path.join(outDir, 'pet-detail-photos.png') })
        console.log('Saved pet-detail-photos.png')
      }
    } else {
      console.warn('No pet link found; run with demo seed for pet-detail screenshots')
    }

    // ---- Add Pet form ----
    await page.goto(baseUrl + '/pets/new', { waitUntil: 'networkidle' })
    await page.waitForSelector('#pet-form-title, .form', { timeout: 5000 }).catch(() => {})
    await wait(400)
    await page.screenshot({ path: path.join(outDir, 'pet-form-add.png') })
    console.log('Saved pet-form-add.png')

    // ---- Edit Pet form ----
    if (petId) {
      await page.goto(baseUrl + `/pets/${petId}/edit`, { waitUntil: 'networkidle' })
      await page.waitForSelector('#pet-form-title, .form', { timeout: 5000 }).catch(() => {})
      await wait(400)
      await page.screenshot({ path: path.join(outDir, 'pet-form-edit.png') })
      console.log('Saved pet-form-edit.png')
    }

    // ---- Settings ----
    await page.goto(baseUrl + '/settings', { waitUntil: 'networkidle' })
    await page.waitForSelector('#settings-title, .settings-cards', { timeout: 5000 }).catch(() => {})
    await wait(400)
    await page.screenshot({ path: path.join(outDir, 'settings.png') })
    console.log('Saved settings.png')

    // ---- Users (admin) ----
    await page.goto(baseUrl + '/users', { waitUntil: 'networkidle' })
    await page.waitForSelector('#users-title, .users-layout', { timeout: 5000 }).catch(() => {})
    await wait(400)
    await page.screenshot({ path: path.join(outDir, 'users.png') })
    console.log('Saved users.png')

    // ---- Admin default options ----
    await page.goto(baseUrl + '/admin/options', { waitUntil: 'networkidle' })
    await page.waitForSelector('#admin-options-title, .admin-options-tabs', { timeout: 5000 }).catch(() => {})
    await wait(400)
    await page.screenshot({ path: path.join(outDir, 'admin-options.png') })
    console.log('Saved admin-options.png')
  } finally {
    await browser.close()
  }
}

main().catch((err) => {
  console.error(err)
  process.exit(1)
})

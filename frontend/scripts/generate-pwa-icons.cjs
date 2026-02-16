/**
 * Generates 192x192 and 512x512 PNG icons from public/vite.svg for PWA installability.
 * Chrome requires PNG icons at these sizes to offer "Install app" (standalone) instead of a shortcut.
 */
const sharp = require('sharp')
const fs = require('fs')
const path = require('path')

const publicDir = path.join(__dirname, '..', 'public')
const svgPath = path.join(publicDir, 'vite.svg')

if (!fs.existsSync(svgPath)) {
  console.error('scripts/generate-pwa-icons.cjs: public/vite.svg not found')
  process.exit(1)
}

const svg = fs.readFileSync(svgPath)

async function run() {
  await sharp(svg, { density: 300 })
    .resize(192, 192)
    .png()
    .toFile(path.join(publicDir, 'icon-192.png'))
  console.log('Generated public/icon-192.png')

  await sharp(svg, { density: 300 })
    .resize(512, 512)
    .png()
    .toFile(path.join(publicDir, 'icon-512.png'))
  console.log('Generated public/icon-512.png')
}

run().catch((err) => {
  console.error(err)
  process.exit(1)
})

import assert from 'node:assert/strict'
import { mkdir } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { join } from 'node:path'
import { spawn } from 'node:child_process'

const requireBrowser = createRequire(join(process.env.MOBILE_CARD_BROWSER_DIR, 'package.json'))
const { chromium } = requireBrowser('playwright')
const output = process.env.MOBILE_CARD_ARTIFACT_DIR
await mkdir(output, { recursive: true })
const server = spawn('pnpm', ['exec', 'vite', 'preview', '--config', 'tests/visual/vite.config.ts', '--host', '127.0.0.1', '--port', '4179', '--strictPort'], { stdio: 'inherit', detached: true })
let browser
try {
  const url = 'http://127.0.0.1:4179/tests/visual/index.html'
  let ready = false
  for (let attempt = 0; attempt < 60; attempt++) {
    try { ready = (await fetch(url)).ok } catch { /* preview is starting */ }
    if (ready) break
    await new Promise((resolve) => setTimeout(resolve, 500))
  }
  assert.ok(ready, 'fixture preview failed to start')
  browser = await chromium.launch()
  for (const width of [320, 390, 1280]) {
    const page = await browser.newPage({ viewport: { width, height: 900 }, deviceScaleFactor: 1 })
    const errors = []
    page.on('pageerror', (error) => errors.push(error.message))
    await page.route('**/*', (route) => new URL(route.request().url()).hostname === '127.0.0.1' ? route.continue() : route.abort())
    await page.goto(url, { waitUntil: 'networkidle' })
    await page.getByTestId('fixture-ready').waitFor()
    await page.evaluate(() => document.fonts.ready)
    await page.screenshot({ path: join(output, `cards-${width}.png`), fullPage: true })
    assert.deepEqual(errors, [], 'fixture has runtime errors')
    assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), `page overflows at ${width}`)
    for (const slug of ['available', 'sold-out', 'promotion']) {
      const card = page.getByTestId(slug).locator('a')
      const box = await card.boundingBox()
      assert.ok(box && box.x >= 0 && box.x + box.width <= width, `${slug} card overflows at ${width}`)
      const title = card.locator('h3')
      const metrics = await title.evaluate((element) => ({ height: element.getBoundingClientRect().height, line: parseFloat(getComputedStyle(element).lineHeight), clamp: getComputedStyle(element).webkitLineClamp, size: getComputedStyle(element).fontSize, weight: getComputedStyle(element).fontWeight, family: getComputedStyle(element).fontFamily }))
      if (width < 640) {
        assert.equal(metrics.clamp, '2', 'mobile title must clamp at two lines')
        assert.equal(metrics.size, '15px', 'mobile title type size changed')
        assert.equal(metrics.weight, '600', 'mobile title weight changed')
        assert.equal(metrics.line, 23, 'mobile title leading changed')
        assert.match(metrics.family, /PingFang SC/, 'mobile title must use the approved Chinese-first stack')
        const price = await card.locator('.product-price').evaluate((element) => ({ size: getComputedStyle(element).fontSize, weight: getComputedStyle(element).fontWeight }))
        assert.equal(price.size, '21px', 'mobile price hierarchy changed')
        assert.equal(price.weight, '700', 'mobile price weight changed')
        assert.ok(metrics.height <= metrics.line * 2 + 1, 'mobile title exceeds two lines')
      }
    }
    if (width === 1280 && process.env.MOBILE_CARD_BASE_SHA) {
      const before = await browser.newPage({ viewport: { width, height: 900 }, deviceScaleFactor: 1 })
      await before.route('**/*', (route) => new URL(route.request().url()).hostname === '127.0.0.1' ? route.continue() : route.abort())
      await before.goto(`${url}?before=1`, { waitUntil: 'networkidle' })
      await before.getByTestId('fixture-ready').waitFor()
      await before.evaluate(() => document.fonts.ready)
      await before.screenshot({ path: join(output, 'cards-1280-before.png'), fullPage: true })
      for (const slug of ['available', 'sold-out', 'promotion']) {
        const currentPNG = await page.getByTestId(slug).screenshot({ animations: 'disabled' })
        const beforePNG = await before.getByTestId(slug).screenshot({ animations: 'disabled' })
        assert.ok(currentPNG.equals(beforePNG), `desktop ${slug} differs visually from PR base`)
      }
      console.log('PASS desktop screenshots match PR base component exactly')
      await before.close()
    }
    const available = page.getByTestId('available')
    const buy = available.getByRole('button', { name: '快速购买' })
    const button = await buy.boundingBox()
    assert.ok(button)
    if (width < 640) {
      assert.ok(button.height >= 44, 'mobile purchase target is smaller than 44px')
      assert.match(await buy.innerText(), /快速购买/)
      assert.ok(await page.getByTestId('sold-out').getByRole('button').isDisabled(), 'sold-out purchase must be disabled')
      const title = await available.locator('h3').boundingBox()
      assert.ok(button.y > title.y + title.height, 'mobile action must sit below the title')
    } else {
      assert.equal(button.width, 36, 'desktop icon button width changed')
      assert.equal(button.height, 36, 'desktop icon button height changed')
      assert.equal((await buy.innerText()).trim(), '', 'desktop must keep icon-only action')
      assert.equal(await page.getByTestId('sold-out').getByRole('button').isVisible(), false)
    }
    await buy.click()
    await page.waitForFunction(() => document.querySelector('[data-testid="purchases"]').textContent === '1')
    assert.equal(await page.getByTestId('route').innerText(), '/', 'quick buy navigated to product detail')
    await available.locator('h3').click()
    await page.waitForFunction(() => document.querySelector('[data-testid="route"]').textContent === '/products/available')
    console.log(`PASS ${width}px: layout, title, purchase target, stock state, click isolation and product navigation`)
    await page.close()
  }
} finally {
  await browser?.close()
  if (server.pid) { try { process.kill(-server.pid, 'SIGTERM') } catch { /* server already stopped */ } }
}

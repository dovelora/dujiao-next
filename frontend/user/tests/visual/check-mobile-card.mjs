import assert from 'node:assert/strict'
import { mkdir, readFile, writeFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { join } from 'node:path'
import { spawn } from 'node:child_process'

const requireBrowser = createRequire(join(process.env.MOBILE_CARD_BROWSER_DIR, 'package.json'))
const { chromium } = requireBrowser('playwright')
const output = process.env.MOBILE_CARD_ARTIFACT_DIR
const origin = 'http://127.0.0.1:4179'
await mkdir(output, { recursive: true })
const server = spawn('pnpm', ['exec', 'vite', 'preview', '--config', 'tests/visual/vite.config.ts', '--host', '127.0.0.1', '--port', '4179', '--strictPort'], { stdio: 'inherit', detached: true })
let browser

const restrictRequests = async (page) => {
  const blocked = []
  await page.route('**/*', (route) => {
    const request = route.request()
    const url = new URL(request.url())
    if (url.origin === origin && ['GET', 'HEAD'].includes(request.method()) && !url.pathname.startsWith('/api/')) {
      return route.continue()
    }
    blocked.push(`${request.method()} ${url.origin}${url.pathname}`)
    return route.abort()
  })
  return blocked
}

try {
  const url = `${origin}/tests/visual/index.html`
  let ready = false
  for (let attempt = 0; attempt < 60; attempt++) {
    try { ready = (await fetch(url)).ok } catch { /* preview is starting */ }
    if (ready) break
    await new Promise((resolve) => setTimeout(resolve, 500))
  }
  assert.ok(ready, 'fixture preview failed to start')
  browser = await chromium.launch()
  for (const width of [320, 402, 1280, 1440]) {
    const page = await browser.newPage({ viewport: { width, height: width < 640 ? 874 : 1000 }, deviceScaleFactor: 1, serviceWorkers: 'block' })
    const errors = []
    page.on('pageerror', (error) => errors.push(error.message))
    const blocked = await restrictRequests(page)
    await page.goto(url, { waitUntil: 'networkidle' })
    await page.getByTestId('fixture-ready').waitFor()
    await page.evaluate(() => document.fonts.ready)
    await page.screenshot({ path: join(output, `cards-${width}.png`), fullPage: true })
    assert.deepEqual(errors, [], 'fixture has runtime errors')
    assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), `page overflows at ${width}`)
    const actionBoxes = []
    for (const slug of ['available', 'sold-out', 'promotion']) {
      const card = page.getByTestId(slug).locator('a')
      const box = await card.boundingBox()
      assert.ok(box && box.x >= 0 && box.x + box.width <= width, `${slug} card overflows at ${width}`)
      const metrics = await card.locator('h3').evaluate((element) => ({ height: element.getBoundingClientRect().height, line: parseFloat(getComputedStyle(element).lineHeight), clamp: getComputedStyle(element).webkitLineClamp, size: getComputedStyle(element).fontSize, weight: getComputedStyle(element).fontWeight, family: getComputedStyle(element).fontFamily }))
      const price = await card.locator('.product-price').evaluate((element) => ({ size: getComputedStyle(element).fontSize, weight: getComputedStyle(element).fontWeight }))
      assert.equal(metrics.clamp, '2', `${slug} title must clamp at two lines`)
      assert.equal(metrics.weight, '600', `${slug} title weight changed`)
      assert.match(metrics.family, /PingFang SC/, `${slug} title must use the approved Chinese-first stack`)
      assert.ok(metrics.height <= metrics.line * 2 + 1, `${slug} title exceeds two lines`)
      if (width < 640) {
        assert.equal(metrics.size, '15px', 'mobile title type size changed')
        assert.equal(metrics.line, 23, 'mobile title leading changed')
        assert.equal(price.size, '21px', 'mobile price hierarchy changed')
        assert.equal(price.weight, '700', 'mobile price weight changed')
      } else {
        assert.equal(metrics.size, '16px', 'desktop title type size changed')
        assert.equal(metrics.line, 24, 'desktop title leading changed')
        assert.equal(price.size, '18px', 'desktop price hierarchy changed')
        assert.equal(price.weight, '600', 'desktop price weight changed')
      }
      const action = card.getByRole('button')
      assert.ok(await action.isVisible(), `${slug} purchase action must remain visible`)
      assert.equal(await action.isDisabled(), slug === 'sold-out', `${slug} stock state is incorrect`)
      const actionBox = await action.boundingBox()
      assert.ok(actionBox)
      assert.ok(actionBox.x >= box.x && actionBox.x + actionBox.width <= box.x + box.width, `${slug} purchase action overflows its card`)
      actionBoxes.push(actionBox)
      if (width < 640) {
        assert.ok(actionBox.height >= 44, 'mobile purchase target is smaller than 44px')
        const titleBox = await card.locator('h3').boundingBox()
        assert.ok(actionBox.y > titleBox.y + titleBox.height, 'mobile action must sit below the title')
      } else {
        assert.equal(actionBox.width, 100, 'desktop purchase target width changed')
        assert.equal(actionBox.height, 40, 'desktop purchase target height changed')
      }
    }
    if (width >= 640) {
      assert.ok(actionBoxes.every(box => Math.abs(box.x - actionBoxes[0].x) <= 1), 'desktop purchase and sold-out actions are not aligned')
    }
    const available = page.getByTestId('available')
    const buy = available.getByRole('button', { name: '快速购买' })
    assert.match(await buy.innerText(), /快速购买/, 'purchase action needs a visible label')
    await buy.click()
    await page.waitForFunction(() => document.querySelector('[data-testid="purchases"]').textContent === '1')
    assert.equal(await page.getByTestId('route').innerText(), '/', 'quick buy navigated to product detail')
    await available.locator('h3').click()
    await page.waitForFunction(() => document.querySelector('[data-testid="route"]').textContent === '/products/available')
    assert.deepEqual(errors, [], 'card interaction has runtime errors')
    assert.deepEqual(blocked, [], 'card fixture attempted a non-local or API request')
    console.log(`PASS ${width}px: typography, layout, purchase target, stock state, click isolation and product navigation`)
    await page.close()
  }

  // Exercise the real critical HTML with the module deliberately unavailable.
  const initialIndex = await readFile(process.env.STOREFRONT_BOOT_INDEX || new URL('../../dist/index.html', import.meta.url), 'utf8')
  const initialHTML = initialIndex
    .replace(/<!--storefront-content-start-->[\s\S]*?<!--storefront-content-end-->/, '<main class="storefront-initial-content"><h1>Server-rendered product</h1></main>')
    .replace(/<script\b[^>]*type="module"[^>]*>[\s\S]*?<\/script>/g, '')
    .replace(/<link\b[^>]*rel="(?:stylesheet|modulepreload)"[^>]*>/g, '')
  for (const outcome of ['ready', 'failed', 'timeout', 'no-javascript']) {
    const page = await browser.newPage({ javaScriptEnabled: outcome !== 'no-javascript' })
    if (outcome === 'timeout') await page.clock.install()
    await restrictRequests(page)
    await page.route(`${origin}/startup-check`, route => route.fulfill({ contentType: 'text/html', body: initialHTML }))
    await page.goto(`${origin}/startup-check`)
    const content = page.locator('.storefront-initial-content')
    const overlay = page.locator('.storefront-boot')
    if (outcome !== 'no-javascript') {
      assert.equal(await content.isVisible(), false, 'raw SEO content flashes while the application loads')
      assert.equal(await overlay.isVisible(), true, 'initial styled loading state missing')
      if (outcome === 'timeout') await page.clock.fastForward(15000)
      else await page.evaluate(name => window.dispatchEvent(new Event(`storefront-${name}`)), outcome)
    }
    assert.equal(await content.isVisible(), true, `${outcome}: server-rendered fallback became inaccessible`)
    assert.equal(await overlay.isVisible(), false, `${outcome}: loading overlay did not clear`)
    await page.close()
    console.log(`PASS startup ${outcome}: initial content visibility and loading handoff`)
  }

  for (const fixture of [
    { file: 'checkout', resultKey: 'checkoutTestResult', minimumCases: 28 },
    { file: 'product-description', resultKey: 'descriptionTestResult', minimumCases: 27 },
  ]) {
    const page = await browser.newPage({ viewport: { width: 1280, height: 900 }, serviceWorkers: 'block' })
    const errors = []
    page.on('pageerror', error => errors.push(error.message))
    const blocked = await restrictRequests(page)
    await page.goto(`${origin}/tests/visual/${fixture.file}.html`, { waitUntil: 'networkidle' })
    await page.waitForFunction(() => ['passed', 'failed'].includes(document.documentElement.dataset.testStatus))
    const result = await page.evaluate(key => window[key], fixture.resultKey)
    await writeFile(join(output, `${fixture.file}-results.json`), JSON.stringify(result, null, 2))
    await page.screenshot({ path: join(output, `${fixture.file}-results.png`), fullPage: true })
    assert.ok(result && Array.isArray(result.results), `${fixture.file} did not return regression results`)
    assert.equal(result.failed, 0, `${fixture.file} failed: ${JSON.stringify(result.results.filter(entry => !entry.passed))}`)
    assert.ok(result.passed >= fixture.minimumCases, `${fixture.file} lost regression coverage`)
    assert.deepEqual(errors, [], `${fixture.file} has runtime errors`)
    assert.deepEqual(blocked, [], `${fixture.file} attempted a non-local or API request`)
    console.log(`PASS ${fixture.file}: ${result.passed} regression checks`)
    await page.close()
  }
} finally {
  await browser?.close()
  if (server.pid) { try { process.kill(-server.pid, 'SIGTERM') } catch { /* server already stopped */ } }
}

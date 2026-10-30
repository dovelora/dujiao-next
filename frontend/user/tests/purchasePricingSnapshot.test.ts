import test from 'node:test'
import assert from 'node:assert/strict'
import { resolvePurchasePricingSnapshot } from '../src/utils/purchasePricingSnapshot.ts'

test('purchase pricing snapshot replaces cached sku price and wholesale tiers', () => {
  const patch = resolvePurchasePricingSnapshot(
    {
      wholesale_prices: [{ sku_id: 12, min_quantity: 3, unit_price: '11.00' }],
    },
    {
      price_amount: '12.345',
    },
  )

  assert.equal(patch.priceAmount, '12.35')
  assert.deepEqual(patch.wholesalePrices, [
    { sku_id: 12, min_quantity: 3, unit_price: '11.00' },
  ])
})

test('purchase pricing snapshot preserves cached price when current sku price is invalid', () => {
  for (const priceAmount of [undefined, '', 'invalid', '0', '-1']) {
    const patch = resolvePurchasePricingSnapshot(
      { wholesale_prices: [] },
      { price_amount: priceAmount },
    )

    assert.equal(Object.hasOwn(patch, 'priceAmount'), false)
    assert.deepEqual(patch.wholesalePrices, [])
  }
})

test('purchase pricing snapshot clears cached wholesale tiers when the current product has none', () => {
  const patch = resolvePurchasePricingSnapshot(
    {},
    { price_amount: '10.00' },
  )

  assert.ok(Object.hasOwn(patch, 'wholesalePrices'))
  assert.equal(patch.wholesalePrices, undefined)
})

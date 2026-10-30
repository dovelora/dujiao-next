import test from 'node:test'
import assert from 'node:assert/strict'
import { catalogCanonicalPath } from '../src/utils/catalogSeo.ts'

test('catalog pagination has its own canonical without tracking or search parameters', () => {
  for (const path of ['/', '/products', '/categories/example', '/blog']) {
    assert.equal(catalogCanonicalPath(path, { page: '2', utm_source: 'test', q: 'search' }), `${path}?page=2`)
    for (const page of ['1', '0', '-1', '1.5', 'bad', '99999999999999999999', ['2', '3']]) {
      assert.equal(catalogCanonicalPath(path, { page }), path)
    }
  }
})

test('product detail and other pages ignore catalog pagination', () => {
  assert.equal(catalogCanonicalPath('/products/example', { page: '2' }), '/products/example')
  assert.equal(catalogCanonicalPath('/about', { page: '2' }), '/about')
})

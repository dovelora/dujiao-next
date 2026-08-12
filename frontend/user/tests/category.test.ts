import assert from 'node:assert/strict'
import test from 'node:test'

import { buildCategoryGroups } from '../src/utils/category.ts'

test('buildCategoryGroups keeps the hot-sales category first without disturbing other categories', () => {
  const groups = buildCategoryGroups([
    { id: 5, slug: 'ai', name: { 'zh-CN': 'AI 与效率' } },
    { id: 6, slug: 'cloud', name: { 'zh-CN': '云服务与开发' } },
    { id: 1, slug: 'BS', name: { 'zh-CN': '热销' } },
    { id: 2, slug: 'x-account', name: { 'zh-CN': 'X账号' } },
  ])

  assert.deepEqual(groups.map((group) => group.slug), ['BS', 'ai', 'cloud', 'x-account'])
})

test('buildCategoryGroups matches the hot-sales slug case-insensitively', () => {
  const groups = buildCategoryGroups([
    { id: 2, slug: 'other' },
    { id: 1, slug: ' bs ' },
  ])

  assert.deepEqual(groups.map((group) => group.id), [1, 2])
})

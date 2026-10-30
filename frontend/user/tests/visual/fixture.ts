import { createApp, h, ref } from 'vue'
import { createPinia } from 'pinia'
import { createHead } from '@unhead/vue/client'
import { createRouter, createMemoryHistory } from 'vue-router'
import i18n from '../../src/i18n'
import VaultProductListItem from '../../src/templates/vault/components/VaultProductListItem.vue'
import '../../src/style.css'
import '../../src/templates/vault/styles/vault.css'
import '@fontsource/rubik/latin-400.css'
import '@fontsource/rubik/latin-600.css'
import '@fontsource/rubik/latin-700.css'
import '@fontsource/nunito-sans/latin-400.css'
import '@fontsource/nunito-sans/latin-600.css'
import '@fontsource/nunito-sans/latin-700.css'

const products = [
  { slug: 'available', category: { name: { 'zh-CN': '数字服务' } }, title: { 'zh-CN': '数字服务组合套餐：包含完整交付说明、设备要求与适用范围的超长商品名称' }, images: ['/catalog-preview-v3/ai-workspace.png'], price_amount: '15.00', stock_status: 'in_stock', fulfillment_type: 'auto', purchase_type: 'member' },
  { slug: 'sold-out', title: { 'zh-CN': '已售罄商品：购买前请核对地区、使用要求及交付方式，长名称应保留两行' }, images: ['/catalog-preview-v3/cloud-vault.png'], price_amount: '39.00', stock_status: 'out_of_stock', fulfillment_type: 'auto', purchase_type: 'member' },
  { slug: 'promotion', title: { 'zh-CN': '限时促销与人工交付商品' }, images: ['/catalog-preview-v3/design-assets.png'], price_amount: '12345.67', promotion_price_amount: '9999.99', stock_status: 'low_stock', fulfillment_type: 'manual', purchase_type: 'guest', tags: ['包含一段非常长的商品附加说明用于检查换行', '适用范围请查看详情'] },
]
const router = createRouter({ history: createMemoryHistory(), routes: [
  { path: '/', component: { render: () => null } },
  { path: '/products/:slug', component: { render: () => null } },
] })
const app = createApp({
  setup() {
    const purchases = ref(0)
    return () => h('main', { class: 'vault-scope', style: 'padding:16px;min-height:100vh' }, [
      h('div', { class: 'catalog-page', style: 'max-width:1000px;margin:0 auto;display:grid;gap:12px' }, [
        h('p', { style: 'font-size:12px', 'data-testid': 'fixture-ready' }, '真实组件 · 固定测试数据'),
        ...products.map((product, index) => h('div', { 'data-testid': product.slug }, [h(VaultProductListItem, { product, index, onQuickBuy: () => purchases.value++ })])),
        h('output', { 'data-testid': 'purchases' }, String(purchases.value)),
        h('output', { 'data-testid': 'route' }, router.currentRoute.value.path),
      ]),
    ])
  },
})
app.use(createPinia()).use(createHead()).use(router).use(i18n)
await router.isReady()
app.mount('#app')

import { createApp, nextTick } from 'vue'
import { createPinia } from 'pinia'
import { createHead } from '@unhead/vue/client'
import './style.css'
import App from './App.vue'
import router from './router'
import i18n, { detectLocale, setI18nLocale, warmupLocaleMessages } from './i18n'
import { useTelegramMiniAppStore } from './stores/telegramMiniApp'
import { getActiveTemplate, initTemplateOverride } from './templates/registry'

// 预览用：?template=vault 持久化激活模板（站长正式切换走站点配置）
initTemplateOverride()

const app = createApp(App)
const head = createHead({
  hooks: {
    'dom:beforeRender': () => {
      // 服务端首屏标签由客户端 head 接管，避免导航后残留上一页的元信息。
      document.head.querySelectorAll('[data-storefront-seo]').forEach((tag) => tag.remove())
    },
  },
})
const pinia = createPinia()

app.use(pinia)
app.use(head)
app.use(router)
app.use(i18n)

// 非默认语言的语言包为懒加载 chunk，挂载前并行加载，避免首屏文案闪现兜底语言
Promise.all([
  router.isReady(),
  useTelegramMiniAppStore(pinia).init(),
  setI18nLocale(detectLocale()),
]).then(async () => {
  if (getActiveTemplate() === 'vault' && router.currentRoute.value.meta.resellerConsole !== true) {
    // Vite waits for this chunk's CSS before resolving the import.
    await import('./templates/vault/layout/VaultLayout.vue')
  }
  app.mount('#app')
  await nextTick()
  window.dispatchEvent(new Event('storefront-ready'))
}).catch((error) => {
  window.dispatchEvent(new Event('storefront-failed'))
  console.error('Storefront initialization failed:', error)
})

void router.isReady().then(() => {
    warmupLocaleMessages()
})

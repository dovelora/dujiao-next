<template>
  <footer class="login-footer">
    <nav :aria-label="t('footer.quickLinks')" class="flex flex-wrap items-center justify-center gap-x-4 gap-y-1">
      <RouterLink to="/terms" class="hover:text-foreground">{{ t('footer.terms') }}</RouterLink>
      <RouterLink to="/privacy" class="hover:text-foreground">{{ t('footer.privacy') }}</RouterLink>
      <a v-if="contact?.support_url" :href="contact.support_url" :target="contact.support_url.startsWith('https://') ? '_blank' : undefined" rel="noopener noreferrer" class="hover:text-foreground">{{ t('footer.supportUrl') }}</a>
      <a v-if="contact?.email" :href="`mailto:${contact.email}`" class="hover:text-foreground">{{ t('footer.email') }}</a>
      <a v-if="contact?.telegram" :href="contact.telegram" target="_blank" rel="noopener noreferrer" class="hover:text-foreground">Telegram</a>
      <a v-if="contact?.whatsapp" :href="contact.whatsapp" target="_blank" rel="noopener noreferrer" class="hover:text-foreground">WhatsApp</a>
      <a v-for="link in footerLinks" :key="link.name" :href="link.url || 'javascript:void(0)'" :target="link.url ? '_blank' : undefined" rel="noopener noreferrer" class="hover:text-foreground">{{ link.name }}</a>
    </nav>
    <p class="mt-2">© {{ year }} {{ brandName }}</p>
  </footer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '../../stores/app'
import { getLocalizedText } from '../../utils/resellerSiteConfig'

const { t } = useI18n()
const appStore = useAppStore()
const year = new Date().getFullYear()
const brandName = computed(() => String(appStore.config?.brand?.site_name || '').trim() || 'Dovelora')
const contact = computed(() => appStore.config?.contact as { support_url?: string; email?: string; telegram?: string; whatsapp?: string } | undefined)
const footerLinks = computed(() => {
  const links = appStore.config?.footer_links
  if (!Array.isArray(links)) return []
  return links.map((item: { name?: unknown; url?: unknown }) => ({
    name: typeof item?.name === 'string' ? item.name.trim() : getLocalizedText(item?.name as Record<string, string>, appStore.locale),
    url: String(item?.url || '').trim(),
  })).filter((item) => item.name)
})
</script>

<style scoped>
.login-footer {
  display: none;
}

@media (width < 640px) {
  .login-footer {
    display: block;
    flex-shrink: 0;
    width: 100%;
    max-width: 400px;
    font-size: 12px;
    line-height: 18px;
    text-align: center;
    color: var(--ui-text-muted);
  }
}
</style>

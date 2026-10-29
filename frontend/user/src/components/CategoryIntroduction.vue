<template>
  <div class="mb-6 space-y-3">
    <p v-if="description" class="max-w-3xl whitespace-pre-line text-sm leading-relaxed text-muted-foreground">{{ description }}</p>
    <RouterLink to="/blog" class="inline-flex text-sm font-medium text-primary underline-offset-4 hover:underline">{{ t('products.purchaseGuides') }}</RouterLink>
  </div>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'

defineProps<{ description?: string }>()
const { t } = useI18n()
</script>

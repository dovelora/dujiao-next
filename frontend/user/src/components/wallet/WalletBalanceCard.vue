<template>
  <section class="wallet-balance space-y-3">
    <header>
      <h2 class="text-xl font-semibold text-foreground">{{ t('personalCenter.wallet.title') }}</h2>
      <p class="mt-1 text-xs text-muted-foreground">{{ t('personalCenter.wallet.subtitle') }}</p>
    </header>

    <Alert v-if="alert" :variant="pageAlertVariant(alert.level)" :class="pageAlertToneClass(alert.level)">
      <AlertDescription>{{ alert.message }}</AlertDescription>
    </Alert>

    <div class="rounded-xl border bg-card p-4 sm:p-5">
      <p class="text-xs text-muted-foreground">{{ t('personalCenter.wallet.balanceLabel') }}</p>
      <p class="mt-2 break-words text-[28px] font-semibold leading-tight tabular-nums text-foreground">{{ balanceDisplay }}</p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { pageAlertVariant, pageAlertToneClass, type PageAlert } from '../../utils/alerts'
import { Alert, AlertDescription } from '@/components/ui/alert'

defineProps<{
  alert: PageAlert | null
  balanceDisplay: string
}>()

const { t } = useI18n()
</script>

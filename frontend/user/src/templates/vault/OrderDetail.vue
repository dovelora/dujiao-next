<template>
  <div class="purchase-support-page mx-auto w-full max-w-[1120px] px-4 pb-8 sm:px-6">
    <nav class="my-3.5 mt-[18px] flex flex-wrap items-center gap-2 text-[13px] text-muted-foreground">
      <RouterLink to="/" class="hover:text-primary">{{ t('nav.home') }}</RouterLink>
      <span class="text-hairline-strong">/</span>
      <RouterLink to="/me/orders" class="hover:text-primary">{{ t('orders.title') }}</RouterLink>
      <span class="text-hairline-strong">/</span>
      <span class="font-semibold text-foreground">{{ t('orderDetail.title') }}</span>
    </nav>

    <header class="mb-[18px]">
      <h1 class="mb-1.5 text-[28px] font-semibold leading-snug">{{ t('orderDetail.title') }}</h1>
      <p class="text-sm leading-6 text-muted-foreground">{{ t('orderDetail.subtitle') }}</p>
    </header>

    <!-- Loading -->
    <div v-if="loading" class="rounded-xl border bg-card p-4 lg:p-5">
      <div class="mb-4 h-5 w-[35%] rounded bg-secondary"></div>
      <div class="h-[200px] rounded-md bg-secondary"></div>
    </div>

    <!-- 不存在 -->
    <div v-else-if="!order" class="my-6 flex flex-col items-center gap-3 rounded-xl border border-dashed py-16 text-center text-muted-foreground">
      <AlertCircle class="h-10 w-10 opacity-60" />
      <p>{{ t('orderDetail.notFound') }}</p>
      <Button class="mt-2 rounded-full" size="sm" @click="debouncedLoadOrder()">{{ t('errorBoundary.retry') }}</Button>
    </div>

    <template v-else>
      <div class="mb-4 grid min-w-0 grid-cols-2 gap-4 rounded-xl border bg-card p-4 sm:grid-cols-2 lg:grid-cols-[minmax(0,1fr)_auto_auto] lg:items-center lg:gap-6 lg:p-5">
        <div class="col-span-2 min-w-0 lg:col-span-1">
          <div class="text-xs text-muted-foreground">{{ t('orders.orderNo') }}</div>
          <div class="mt-1 break-all font-semibold">{{ order.order_no }}</div>
          <div class="mt-1.5 text-[13px] text-muted-foreground">{{ t('orderDetail.createdAtLabel') }}：{{ formatDate(order.created_at) }}</div>
        </div>
        <div>
          <div class="text-xs text-muted-foreground">{{ t('orderDetail.amountTotal') }}</div>
          <div class="mt-1 support-amount break-words text-[24px] font-semibold tabular-nums">{{ formatMoney(order.total_amount, order.currency) }}</div>
        </div>
        <div class="flex flex-wrap items-center justify-end gap-2.5">
          <Badge :variant="statusVariant(order.status)" class="rounded-full">{{ statusLabel(order.status) }}</Badge>
          <Button v-if="order.status === 'pending_payment'" as-child size="sm" class="rounded-full">
            <RouterLink :to="`/pay?order_no=${order.order_no}`">{{ t('orderDetail.payNow') }}</RouterLink>
          </Button>
          <Button v-if="order.status === 'pending_payment'" type="button" variant="outline" size="sm" class="rounded-full border-destructive text-destructive hover:bg-destructive/10" @click="cancelOrder">{{ t('orderDetail.cancel') }}</Button>
        </div>
      </div>

      <VaultOrderBody
        :order="order"
        variant="user"
        :fulfillment-downloading="fulfillmentDownloading"
        @download="handleDownloadFulfillment"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { AlertCircle } from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import VaultOrderBody from './components/VaultOrderBody.vue'
import { useOrderDetail } from '../../composables/useOrderDetail'

const { t } = useI18n()

const {
  loading, order, debouncedLoadOrder, cancelOrder, fulfillmentDownloading, handleDownloadFulfillment,
  statusLabel, statusVariant, formatDate, formatMoney,
} = useOrderDetail()
</script>

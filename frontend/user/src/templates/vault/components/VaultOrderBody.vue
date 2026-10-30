<template>
  <div class="grid items-start gap-4 lg:grid-cols-[minmax(0,1fr)_320px] lg:gap-6">
    <div class="grid min-w-0 gap-4">
      <!-- 子订单（含卡密） -->
      <section v-if="order.children && order.children.length > 0" class="min-w-0 rounded-xl border bg-card p-4 lg:p-5">
        <h2 class="mb-3 text-base font-semibold">{{ t('orderDetail.childOrdersTitle') }}</h2>
        <div class="grid gap-4">
          <div v-for="child in order.children" :key="child.id" class="min-w-0 rounded-lg border p-4">
            <div class="grid min-w-0 gap-2 sm:flex sm:flex-wrap sm:items-start sm:justify-between sm:gap-3">
              <div class="min-w-0 flex-1">
                <div class="break-all text-[13px] text-muted-foreground">{{ t('orderDetail.childOrderNo') }}：{{ child.order_no }}</div>
                <div class="break-words text-[13px] text-muted-foreground">{{ t('orderDetail.childOrderAmount') }}：{{ formatMoney(child.total_amount, child.currency || order.currency) }}</div>
              </div>
              <Badge :variant="statusVariant(resolvedChildStatus(child))" size="sm" class="w-fit rounded-full">{{ statusLabel(resolvedChildStatus(child)) }}</Badge>
            </div>

            <h3 class="my-2.5 mt-4 text-sm font-semibold">{{ t('orderDetail.childItemsTitle') }}</h3>
            <div v-if="child.items && child.items.length" class="grid">
              <VaultOrderItem v-for="(item, cidx) in child.items" :key="cidx" :item="item" :currency="child.currency || order.currency" />
            </div>
            <div v-else class="text-[13px] text-muted-foreground">{{ t('orderDetail.noItems') }}</div>

            <div class="mt-3.5 border-t pt-3.5">
              <VaultOrderFulfillment
                v-if="child.fulfillment"
                :title="t('orderDetail.childFulfillmentTitle')"
                :fulfillment="child.fulfillment"
                :items="child.items"
                :order-no="child.order_no || order.order_no"
                :downloading="fulfillmentDownloading"
                @download="emit('download', $event)"
              />
              <div v-else class="text-[13px] text-muted-foreground">{{ t('orderDetail.childFulfillmentEmpty') }}</div>
            </div>
          </div>
        </div>
      </section>

      <!-- 主订单发货（含卡密） -->
      <section v-if="order.fulfillment" class="min-w-0 rounded-xl border bg-card p-4 lg:p-5">
        <VaultOrderFulfillment
          :fulfillment="order.fulfillment"
          :items="order.items"
          :order-no="order.order_no"
          :downloading="fulfillmentDownloading"
          @download="emit('download', $event)"
        >
          <template #title><h2 class="text-base font-semibold">{{ t('orderDetail.fulfillmentTitle') }}</h2></template>
        </VaultOrderFulfillment>
      </section>

      <!-- 商品 -->
      <section class="min-w-0 rounded-xl border bg-card p-4 lg:p-5">
        <h2 class="mb-3 text-base font-semibold">{{ t('orderDetail.itemsTitle') }}</h2>
        <div v-if="order.items && order.items.length > 0" class="grid">
          <VaultOrderItem v-for="(item, idx) in order.items" :key="idx" :item="item" :currency="order.currency" />
        </div>
        <div v-else class="text-[13px] text-muted-foreground">{{ t('orderDetail.noItems') }}</div>
      </section>

    </div>

    <div class="grid min-w-0 gap-4">
      <!-- 金额明细 -->
      <section class="min-w-0 rounded-xl border bg-card p-4 lg:p-5">
        <h2 class="mb-3 text-base font-semibold">{{ t('orderDetail.amountTitle') }}</h2>
        <div class="grid divide-y">
          <div class="flex items-start justify-between gap-3 py-3 first:pt-0 last:pb-0">
            <div class="text-[13px] text-muted-foreground">{{ t('orderDetail.amountOriginal') }}</div>
            <div class="support-amount min-w-0 break-words text-right text-[14px] font-semibold tabular-nums">{{ formatMoney(order.original_amount, order.currency) }}</div>
          </div>
          <div class="flex items-start justify-between gap-3 py-3 first:pt-0 last:pb-0">
            <div class="text-[13px] text-muted-foreground">{{ t('orderDetail.amountDiscount') }}</div>
            <div class="support-amount min-w-0 break-words text-right text-[14px] font-semibold tabular-nums" :class="{ 'text-destructive': hasDiscountAmount(order.discount_amount) }">{{ formatDiscountMoney(order.discount_amount, order.currency) }}</div>
          </div>
          <div class="flex items-start justify-between gap-3 py-3 first:pt-0 last:pb-0">
            <div class="text-[13px] text-muted-foreground">{{ t('orderDetail.amountTotal') }}</div>
            <div class="support-amount min-w-0 break-words text-right text-[14px] font-semibold tabular-nums">{{ formatMoney(order.total_amount, order.currency) }}</div>
          </div>
          <template v-if="variant === 'user'">
            <div v-if="hasAmount(order.wallet_paid_amount)" class="flex items-start justify-between gap-3 py-3 first:pt-0 last:pb-0">
              <div class="text-[13px] text-muted-foreground">{{ t('orderDetail.amountWalletPaid') }}</div>
              <div class="support-amount min-w-0 break-words text-right text-[14px] font-semibold tabular-nums">{{ formatMoney(order.wallet_paid_amount, order.currency) }}</div>
            </div>
            <div v-if="hasAmount(order.online_paid_amount)" class="flex items-start justify-between gap-3 py-3 first:pt-0 last:pb-0">
              <div class="text-[13px] text-muted-foreground">{{ t('orderDetail.amountOnlinePaid') }}</div>
              <div class="support-amount min-w-0 break-words text-right text-[14px] font-semibold tabular-nums">{{ formatMoney(order.online_paid_amount, order.currency) }}</div>
            </div>
            <div v-if="hasAmount(order.refunded_amount)" class="flex items-start justify-between gap-3 py-3 first:pt-0 last:pb-0">
              <div class="text-[13px] text-muted-foreground">{{ t('orderDetail.amountRefunded') }}</div>
              <div class="support-amount min-w-0 break-words text-right text-[14px] font-semibold tabular-nums">{{ formatMoney(order.refunded_amount, order.currency) }}</div>
            </div>
          </template>
          <div v-if="hasDiscountAmount(order.member_discount_amount)" class="flex items-start justify-between gap-3 py-3 text-[color:var(--gold-strong)]">
            <div class="text-[13px]">{{ t('orderDetail.amountMemberDiscount') }}</div>
            <div class="support-amount min-w-0 break-words text-right text-[14px] font-semibold tabular-nums">{{ formatDiscountMoney(order.member_discount_amount, order.currency) }}</div>
          </div>
          <div v-if="hasDiscountAmount(order.wholesale_discount_amount)" class="flex items-start justify-between gap-3 py-3 text-[color:var(--teal-strong)]">
            <div class="text-[13px]">{{ t('orderDetail.amountWholesaleDiscount') }}</div>
            <div class="support-amount min-w-0 break-words text-right text-[14px] font-semibold tabular-nums">{{ formatDiscountMoney(order.wholesale_discount_amount, order.currency) }}</div>
          </div>
        </div>
      </section>

      <!-- 退款记录 -->
      <section v-if="showRefundRecordsCard" class="min-w-0 rounded-xl border bg-card p-4 lg:p-5">
        <h2 class="mb-3 text-base font-semibold">{{ t('orderDetail.refundRecordsTitle') }}</h2>
        <div v-if="refundRecords.length > 0" class="grid grid-cols-[minmax(0,1fr)] divide-y">
          <div v-for="(record, idx) in refundRecords" :key="`refund-${idx}`" class="grid min-w-0 grid-cols-[minmax(0,1fr)] gap-2 py-3 first:pt-0 last:pb-0 text-[13px]">
            <div class="flex flex-wrap items-start justify-between gap-2">
              <span class="text-muted-foreground"><span class="sr-only">{{ t('orderDetail.refundRecordTime') }}：</span>{{ formatDate(record.created_at) }}</span>
              <span class="support-amount min-w-0 break-words font-semibold tabular-nums"><span class="sr-only">{{ t('orderDetail.refundRecordAmount') }}：</span>{{ formatMoney(record.amount, record.currency || order.currency) }}</span>
            </div>
            <p class="min-w-0 whitespace-pre-wrap [overflow-wrap:anywhere] text-muted-foreground"><span class="sr-only">{{ t('orderDetail.refundRecordReason') }}：</span>{{ refundReasonText(record.remark) }}</p>
          </div>
        </div>
        <div v-else class="text-[13px] text-muted-foreground">{{ t('orderDetail.refundRecordsEmpty') }}</div>
      </section>

      <!-- 时间信息 -->
      <section v-if="showTimeCard" class="min-w-0 rounded-xl border bg-card p-4 lg:p-5">
        <h2 class="mb-3 text-base font-semibold">{{ t('orderDetail.timeTitle') }}</h2>
        <div class="grid divide-y">
          <div class="flex items-start justify-between gap-3 py-3 first:pt-0 last:pb-0">
            <div class="flex-none text-[13px] text-muted-foreground">{{ t('orderDetail.createdAtLabel') }}</div>
            <div class="min-w-0 break-words text-right text-[13px] font-medium">{{ formatDate(order.created_at) }}</div>
          </div>
          <div v-if="order.paid_at" class="flex items-start justify-between gap-3 py-3 first:pt-0 last:pb-0">
            <div class="flex-none text-[13px] text-muted-foreground">{{ t('orderDetail.paidAtLabel') }}</div>
            <div class="min-w-0 break-words text-right text-[13px] font-medium">{{ formatDate(order.paid_at) }}</div>
          </div>
          <div v-if="order.expires_at" class="flex items-start justify-between gap-3 py-3 first:pt-0 last:pb-0">
            <div class="flex-none text-[13px] text-muted-foreground">{{ t('orderDetail.expiresAtLabel') }}</div>
            <div class="min-w-0 break-words text-right text-[13px] font-medium">{{ formatDate(order.expires_at) }}</div>
          </div>
          <div v-if="order.canceled_at" class="flex items-start justify-between gap-3 py-3 first:pt-0 last:pb-0">
            <div class="flex-none text-[13px] text-muted-foreground">{{ t('orderDetail.canceledAtLabel') }}</div>
            <div class="min-w-0 break-words text-right text-[13px] font-medium">{{ formatDate(order.canceled_at) }}</div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { toRef } from 'vue'
import { useI18n } from 'vue-i18n'
import { Badge } from '@/components/ui/badge'
import VaultOrderItem from './VaultOrderItem.vue'
import VaultOrderFulfillment from './VaultOrderFulfillment.vue'
import { useOrderDisplayHelpers } from '../../../composables/useOrderDisplayHelpers'

const props = defineProps<{
  order: any
  variant: 'user' | 'guest'
  fulfillmentDownloading: boolean
}>()

const emit = defineEmits<{
  (e: 'download', orderNo: string): void
}>()

const { t } = useI18n()

const {
  showTimeCard, showRefundRecordsCard, refundRecords, resolvedChildStatus,
  statusLabel, statusVariant, formatDate, refundReasonText,
  formatMoney, formatDiscountMoney, hasDiscountAmount, hasAmount,
} = useOrderDisplayHelpers(toRef(props, 'order'))
</script>

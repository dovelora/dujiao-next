<template>
  <div class="purchase-support-page mx-auto w-full max-w-[1120px] px-4 pb-8 sm:px-6">
    <div class="my-6">
      <h1 class="mb-1.5 text-[28px] font-semibold leading-snug">{{ t('guestOrders.title') }}</h1>
      <p class="text-sm leading-6 text-muted-foreground">{{ t('guestOrders.subtitle') }}</p>
    </div>

    <!-- 查询表单 -->
    <div class="mb-5 rounded-xl border bg-card p-4 lg:p-5">
      <div v-if="hasSavedAuth" class="mb-4 flex flex-wrap items-center justify-between gap-3 rounded-md border px-3.5 py-2.5 text-[12.5px] text-muted-foreground">
        <span class="min-w-0 break-all">{{ t('guestOrders.savedHint', { email: savedAuth.email || '-' }) }}</span>
        <button type="button" class="text-[12.5px] text-muted-foreground transition-colors hover:text-primary" @click="clearSaved">{{ t('guestOrders.clearSaved') }}</button>
      </div>
      <div class="grid grid-cols-2 items-end gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_minmax(0,1fr)_auto]">
        <label class="grid min-w-0 gap-2 text-[13px] text-muted-foreground">
          {{ t('guestOrders.emailPlaceholder') }}
          <Input v-model="email" type="email" autocomplete="email" class="h-11 min-w-0" />
        </label>
        <label class="grid min-w-0 gap-2 text-[13px] text-muted-foreground">
          {{ t('guestOrders.passwordPlaceholder') }}
          <Input v-model="orderPassword" type="password" autocomplete="current-password" class="h-11 min-w-0" />
        </label>
        <label class="col-span-2 grid min-w-0 gap-2 text-[13px] text-muted-foreground sm:col-span-1">
          {{ t('guestOrders.orderNoPlaceholder') }}
          <Input v-model="orderNo" type="text" autocomplete="off" class="h-11 min-w-0" />
        </label>
        <Button type="button" class="col-span-2 h-11 rounded-full px-4 font-semibold sm:col-span-1" :disabled="loading" @click="handleSearch">
          <Search /> {{ loading ? t('guestOrders.searching') : t('guestOrders.search') }}
        </Button>
      </div>
      <p class="mt-3 text-[13px] text-muted-foreground">{{ t('guestOrders.tip') }}</p>
      <div v-if="error" class="mt-3.5 rounded-sm bg-destructive/10 px-3 py-2.5 text-[13px] font-semibold text-destructive">{{ error }}</div>
    </div>

    <!-- 空态 -->
    <div v-if="orders.length === 0 && !loading" class="flex flex-col items-center gap-3 rounded-xl border border-dashed py-16 text-center text-muted-foreground">
      <ClipboardList class="h-10 w-10 opacity-60" />
      <p>{{ emptyMessage }}</p>
    </div>

    <!-- 列表 -->
    <div v-else class="grid gap-3.5">
      <div v-for="order in orders" :key="order.order_no" class="grid min-w-0 items-center gap-4 rounded-xl border bg-card p-4 transition hover:border-hairline-strong lg:grid-cols-[minmax(0,1fr)_160px_300px] lg:gap-5 lg:p-5">
        <div class="grid min-w-0 gap-1.5">
          <div class="break-all text-[13px] font-medium text-foreground">{{ t('orders.orderNo') }}：{{ order.order_no }}</div>
          <div v-if="hasDiscount(order)" class="flex flex-wrap gap-2.5 text-xs">
            <span v-if="hasDiscountAmount(order.discount_amount)" class="text-destructive">{{ t('orderDetail.couponDiscountLabel') }}：{{ formatDiscountMoney(order.discount_amount, order.currency) }}</span>
            <span v-if="hasDiscountAmount(order.promotion_discount_amount)" class="text-destructive">{{ t('orderDetail.promotionDiscountLabel') }}：{{ formatDiscountMoney(order.promotion_discount_amount, order.currency) }}</span>
          </div>
          <div class="text-[13px] text-muted-foreground">{{ formatDate(order.created_at) }}</div>
        </div>
        <div class="support-amount min-w-0 break-words text-[22px] font-semibold tabular-nums lg:text-right">{{ formatMoney(order.total_amount, order.currency) }}</div>
        <div class="flex flex-wrap items-center gap-2.5 lg:justify-end">
          <Badge :variant="statusVariant(order.status)" class="rounded-full">{{ statusLabel(order.status) }}</Badge>
          <RouterLink
            :to="{ name: 'guest-order-detail', params: { order_no: order.order_no } }"
            :class="buttonVariants({ variant: 'outline', size: 'sm' })"
            class="rounded-full"
          >
            <Eye /> {{ t('guestOrders.viewDetails') }}
          </RouterLink>
          <Button v-if="order.status === 'pending_payment'" as-child size="sm" class="rounded-full">
            <RouterLink :to="`/pay?guest=1&order_no=${order.order_no}`"><CreditCard /> {{ t('guestOrders.payNow') }}</RouterLink>
          </Button>
        </div>
      </div>

      <div v-if="pagination.total_page > 1" class="mt-2 flex flex-wrap items-center justify-center gap-3.5">
        <Button type="button" variant="outline" size="icon" class="rounded-full" :disabled="loading || pagination.page <= 1" @click="changePage(pagination.page - 1)"><ChevronLeft /></Button>
        <span class="text-sm font-bold">{{ pagination.page }} / {{ pagination.total_page }}</span>
        <Button type="button" variant="outline" size="icon" class="rounded-full" :disabled="loading || pagination.page >= pagination.total_page" @click="changePage(pagination.page + 1)"><ChevronRight /></Button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ChevronLeft, ChevronRight, ClipboardList, Search, Eye, CreditCard } from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import { Button, buttonVariants } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useGuestOrders } from '../../composables/useGuestOrders'

const { t } = useI18n()

const {
  savedAuth, email, orderPassword, orderNo, loading, error, orders, pagination,
  hasSavedAuth, clearSaved, handleSearch, emptyMessage, changePage,
  statusLabel, statusVariant, formatMoney, formatDiscountMoney, hasDiscountAmount, hasDiscount, formatDate,
} = useGuestOrders()
</script>

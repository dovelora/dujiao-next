<template>
  <div class="wallet-transactions rounded-xl border bg-card p-4 sm:p-6">
    <header class="mb-4 flex items-start justify-between gap-3">
      <div class="min-w-0">
        <h3 class="text-base font-semibold">{{ t('personalCenter.wallet.detailTitle') }}</h3>
        <p class="mt-1 text-xs text-muted-foreground">
          {{ t('personalCenter.wallet.transactionsLabel') }} {{ totalTransactions }}
          <span v-if="totalTransactions > 0 && totalPages > 0"> · {{ t('orders.pageInfo', { page: currentPage, total: totalPages }) }}</span>
        </p>
      </div>
      <Button type="button" variant="outline" size="sm" class="shrink-0" @click="$emit('refresh')">
        {{ t('orders.filters.refresh') }}
      </Button>
    </header>

    <div v-if="loading" class="space-y-3">
      <div v-for="idx in 3" :key="idx" class="h-16 animate-pulse rounded-xl border bg-muted"></div>
    </div>
    <div v-else-if="transactions.length === 0" class="rounded-xl border border-dashed px-4 py-6 text-sm text-muted-foreground">
      {{ t('personalCenter.wallet.empty') }}
    </div>
    <template v-else>
      <ul class="divide-y sm:hidden">
        <li v-for="item in transactions" :key="item.id" class="wallet-transaction grid grid-cols-[minmax(0,1fr)_auto] items-start gap-x-3 gap-y-2 py-3 first:pt-0 last:pb-0">
          <div class="flex min-w-0 flex-wrap items-center gap-2">
            <span class="text-sm font-medium">{{ transactionTypeLabel(item.type) }}</span>
            <Badge size="sm" :variant="directionVariant(item.direction)">{{ directionLabel(item.direction) }}</Badge>
          </div>
          <span class="max-w-40 break-all text-right text-base font-semibold tabular-nums" :class="item.direction === 'in' ? 'text-success' : 'text-destructive'">{{ signedAmount(item.direction, item.amount, item.currency) }}</span>
          <div class="min-w-0 text-xs leading-5 text-muted-foreground">
            <time :datetime="item.created_at">{{ formatDate(item.created_at) }}</time>
            <p v-if="item.remark" class="mt-1 break-words [overflow-wrap:anywhere]">{{ item.remark }}</p>
          </div>
          <div class="max-w-40 text-right text-xs leading-5">
            <p class="text-muted-foreground">{{ t('personalCenter.wallet.table.balanceAfter') }}</p>
            <p class="break-all tabular-nums text-foreground">{{ formatMoney(item.balance_after, item.currency) }}</p>
          </div>
        </li>
      </ul>
      <div class="hidden overflow-x-auto rounded-xl border sm:block">
        <Table>
          <TableHeader>
            <TableRow class="bg-muted/50">
              <TableHead class="px-4">{{ t('personalCenter.wallet.table.createdAt') }}</TableHead>
              <TableHead class="px-4">{{ t('personalCenter.wallet.table.type') }}</TableHead>
              <TableHead class="px-4">{{ t('personalCenter.wallet.table.direction') }}</TableHead>
              <TableHead class="px-4">{{ t('personalCenter.wallet.table.amount') }}</TableHead>
              <TableHead class="px-4">{{ t('personalCenter.wallet.table.balanceAfter') }}</TableHead>
              <TableHead class="px-4">{{ t('personalCenter.wallet.table.remark') }}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="item in transactions" :key="item.id">
              <TableCell class="px-4 text-xs text-muted-foreground">{{ formatDate(item.created_at) }}</TableCell>
              <TableCell class="px-4 text-xs text-muted-foreground">{{ transactionTypeLabel(item.type) }}</TableCell>
              <TableCell class="px-4">
                <Badge size="sm" :variant="directionVariant(item.direction)">
                  {{ directionLabel(item.direction) }}
                </Badge>
              </TableCell>
              <TableCell class="px-4 font-mono text-sm" :class="item.direction === 'in' ? 'text-success' : 'text-destructive'">
                {{ signedAmount(item.direction, item.amount, item.currency) }}
              </TableCell>
              <TableCell class="px-4 font-mono text-sm text-foreground">
                {{ formatMoney(item.balance_after, item.currency) }}
              </TableCell>
              <TableCell class="px-4 text-xs text-muted-foreground">{{ item.remark || '-' }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </template>

    <div v-if="totalPages > 1" class="mt-6 flex flex-wrap items-center justify-center gap-3">
      <Button variant="outline" size="sm" :disabled="currentPage <= 1" @click="$emit('changePage', currentPage - 1)">
        {{ t('orders.prevPage') }}
      </Button>
      <span class="rounded-full border text-muted-foreground px-4 py-2 text-sm">
        {{ t('orders.pageInfo', { page: currentPage, total: totalPages }) }}
      </span>
      <Button variant="outline" size="sm" :disabled="currentPage >= totalPages" @click="$emit('changePage', currentPage + 1)">
        {{ t('orders.nextPage') }}
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

defineProps<{
  loading: boolean
  transactions: Array<{
    id: number
    created_at: string
    type: string
    direction: string
    amount: string
    currency: string
    balance_after: string
    remark: string
  }>
  totalTransactions: number
  currentPage: number
  totalPages: number
}>()

defineEmits<{
  (e: 'refresh'): void
  (e: 'changePage', page: number): void
}>()

const { t } = useI18n()

const formatMoney = (amount?: string, currency?: string) => {
  if (amount === null || amount === undefined || amount === '') return '-'
  if (currency === null || currency === undefined || currency === '') {
    return String(amount)
  }
  return `${amount} ${currency}`
}

const formatDate = (raw?: string) => {
  if (!raw) return '-'
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return raw
  return date.toLocaleString()
}

const directionLabel = (direction?: string) => {
  if (direction === 'in') return t('personalCenter.wallet.directionIn')
  if (direction === 'out') return t('personalCenter.wallet.directionOut')
  return direction || '-'
}

const directionVariant = (direction?: string): 'success' | 'destructive' | 'warning' => {
  if (direction === 'in') return 'success'
  if (direction === 'out') return 'destructive'
  return 'warning'
}

const transactionTypeLabel = (type?: string) => {
  const key = `personalCenter.wallet.types.${type || ''}`
  const translated = t(key)
  if (translated === key) return type || '-'
  return translated
}

const signedAmount = (direction: string, amount?: string, currency?: string) => {
  const base = formatMoney(amount, currency)
  if (base === '-') return base
  if (direction === 'out') return `-${base}`
  return `+${base}`
}
</script>

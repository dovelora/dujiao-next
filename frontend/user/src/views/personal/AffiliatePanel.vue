<template>
  <div class="affiliate-panel space-y-4">
    <section class="space-y-3">
      <header>
        <h2 class="text-xl font-semibold">{{ t('personalCenter.affiliate.title') }}</h2>
        <p class="mt-1 text-xs text-muted-foreground">{{ t('personalCenter.affiliate.subtitle') }}</p>
      </header>

      <Alert v-if="panelAlert" class="mb-5" :variant="pageAlertVariant(panelAlert.level)" :class="pageAlertToneClass(panelAlert.level)">
        <AlertDescription>{{ panelAlert.message }}</AlertDescription>
      </Alert>

      <div v-if="loading" class="space-y-3">
        <div v-for="idx in 3" :key="idx" class="h-16 animate-pulse rounded-xl border bg-muted"></div>
      </div>

      <template v-else-if="dashboard?.opened">
        <div class="affiliate-promotion rounded-xl border bg-card p-4">
          <div class="flex items-center justify-between gap-3">
            <div class="min-w-0">
              <p class="text-xs text-muted-foreground">{{ t('personalCenter.affiliate.affiliateCode') }}</p>
              <p class="mt-1 break-all font-mono text-sm font-semibold">{{ dashboard.affiliate_code || '-' }}</p>
            </div>
            <Button type="button" class="h-10 shrink-0 px-3 text-xs font-semibold" @click="copyPromotionUrl">
              {{ t('personalCenter.affiliate.copyPromotionUrl') }}
            </Button>
          </div>
          <p class="mt-3 break-all text-xs leading-5 text-muted-foreground">{{ promotionUrl }}</p>
          <div class="mt-3 flex flex-wrap items-center justify-between gap-x-3 gap-y-1 border-t pt-3 text-xs">
            <p><span class="text-muted-foreground">{{ t('personalCenter.affiliate.conversionRate') }}</span> <span class="font-semibold tabular-nums">{{ conversionRateText }}</span></p>
            <p class="text-muted-foreground">{{ t('personalCenter.affiliate.conversionDetail', { clicks: dashboard.click_count || 0, orders: dashboard.valid_order_count || 0 }) }}</p>
          </div>
        </div>
        <dl class="affiliate-summary grid grid-cols-3 gap-x-3 gap-y-1 rounded-xl border bg-card p-4">
          <div class="row-span-2 grid min-w-0 grid-rows-subgrid">
            <dt class="text-xs leading-5 text-muted-foreground">{{ t('personalCenter.affiliate.pendingCommission') }}</dt>
            <dd class="break-all text-lg font-semibold tabular-nums">{{ dashboard.pending_commission || '0.00' }}</dd>
          </div>
          <div class="row-span-2 grid min-w-0 grid-rows-subgrid">
            <dt class="text-xs leading-5 text-muted-foreground">{{ t('personalCenter.affiliate.availableCommission') }}</dt>
            <dd class="break-all text-lg font-semibold tabular-nums text-primary">{{ dashboard.available_commission || '0.00' }}</dd>
          </div>
          <div class="row-span-2 grid min-w-0 grid-rows-subgrid">
            <dt class="text-xs leading-5 text-muted-foreground">{{ t('personalCenter.affiliate.withdrawnCommission') }}</dt>
            <dd class="break-all text-lg font-semibold tabular-nums">{{ dashboard.withdrawn_commission || '0.00' }}</dd>
          </div>
        </dl>
      </template>

      <div v-else class="rounded-xl border border-dashed p-5">
        <p class="text-sm text-muted-foreground">{{ t('personalCenter.affiliate.notOpened') }}</p>
        <Button type="button" :disabled="opening" class="mt-4 font-bold" @click="openAffiliate">
          {{ opening ? t('personalCenter.affiliate.opening') : t('personalCenter.affiliate.openButton') }}
        </Button>
      </div>
    </section>

    <div v-if="dashboard?.opened" class="rounded-xl border bg-card p-4 sm:p-6">
      <h3 class="text-base font-semibold text-foreground">{{ t('personalCenter.affiliate.withdrawTitle') }}</h3>
      <p class="mt-1 text-xs text-muted-foreground">{{ t('personalCenter.affiliate.withdrawSubtitle') }}</p>
      <form class="affiliate-withdraw mt-4 grid grid-cols-2 items-end gap-3 lg:grid-cols-4" @submit.prevent="handleApplyWithdraw">
        <div class="min-w-0">
          <Label :for="`${withdrawFormId}-amount`" class="mb-1 block text-xs">{{ t('personalCenter.affiliate.withdrawAmountLabel') }}</Label>
          <Input
            :id="`${withdrawFormId}-amount`"
            v-model="withdrawForm.amount"
            type="text"
            inputmode="decimal"
            class="h-11"
            :placeholder="t('personalCenter.affiliate.withdrawAmountPlaceholder')"
          />
        </div>
        <div class="min-w-0">
          <Label :for="`${withdrawFormId}-channel`" class="mb-1 block text-xs">{{ t('personalCenter.affiliate.withdrawChannelLabel') }}</Label>
          <Select v-if="channelOptions.length > 0" v-model="withdrawForm.channel">
            <SelectTrigger :id="`${withdrawFormId}-channel`" class="h-11 w-full"><SelectValue :placeholder="t('personalCenter.affiliate.withdrawChannelPlaceholder')" /></SelectTrigger>
            <SelectContent>
              <SelectItem v-for="channel in channelOptions" :key="channel" :value="channel">{{ channel }}</SelectItem>
            </SelectContent>
          </Select>
          <Input
            v-else
            :id="`${withdrawFormId}-channel`"
            v-model="withdrawForm.channel"
            type="text"
            class="h-11"
            :placeholder="t('personalCenter.affiliate.withdrawChannelPlaceholder')"
          />
        </div>
        <div class="col-span-2 min-w-0 lg:col-span-1">
          <Label :for="`${withdrawFormId}-account`" class="mb-1 block text-xs">{{ t('personalCenter.affiliate.withdrawAccountLabel') }}</Label>
          <Input
            :id="`${withdrawFormId}-account`"
            v-model="withdrawForm.account"
            type="text"
            class="h-11"
            :placeholder="t('personalCenter.affiliate.withdrawAccountPlaceholder')"
          />
        </div>
        <div class="col-span-2 flex items-end lg:col-span-1">
          <Button type="submit" :disabled="submittingWithdraw" class="h-11 w-full px-5 font-bold">
            {{ submittingWithdraw ? t('personalCenter.affiliate.withdrawing') : t('personalCenter.affiliate.withdrawSubmit') }}
          </Button>
        </div>
      </form>
    </div>

    <div v-if="dashboard?.opened" class="rounded-xl border bg-card p-4 sm:p-6">
      <div class="mb-4 flex items-center justify-between gap-3">
        <h3 class="text-base font-semibold text-foreground">{{ t('personalCenter.affiliate.commissionTitle') }}</h3>
        <Button type="button" variant="outline" size="sm" @click="loadCommissions(commissionsPagination.page)">
          {{ t('orders.filters.refresh') }}
        </Button>
      </div>

      <div v-if="commissionsLoading" class="space-y-3">
        <div v-for="idx in 3" :key="idx" class="h-14 animate-pulse rounded-xl border bg-muted"></div>
      </div>
      <div v-else-if="commissions.length === 0" class="rounded-xl border border-dashed px-4 py-6 text-sm text-muted-foreground">
        {{ t('personalCenter.affiliate.commissionEmpty') }}
      </div>
      <template v-else>
        <ul class="affiliate-commissions divide-y sm:hidden">
          <li v-for="item in commissions" :key="item.id" class="grid grid-cols-[minmax(0,1fr)_auto] items-start gap-x-3 gap-y-1.5 py-3 first:pt-0 last:pb-0">
            <p class="min-w-0 break-all text-base font-semibold tabular-nums">{{ item.commission_amount }}</p>
            <Badge :variant="commissionStatusVariant(item.status)" size="sm" class="justify-self-end">{{ commissionStatusLabel(item.status) }}</Badge>
            <time :datetime="item.created_at" class="col-span-2 text-xs leading-5 text-muted-foreground">{{ formatDate(item.created_at) }}</time>
          </li>
        </ul>
        <div class="hidden overflow-x-auto rounded-xl border sm:block">
          <Table>
            <TableHeader>
              <TableRow class="bg-muted/50">
                <TableHead class="px-4">{{ t('personalCenter.affiliate.table.amount') }}</TableHead>
                <TableHead class="px-4">{{ t('personalCenter.affiliate.table.status') }}</TableHead>
                <TableHead class="px-4">{{ t('personalCenter.affiliate.table.createdAt') }}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="item in commissions" :key="item.id">
                <TableCell class="px-4 font-mono text-xs text-foreground">{{ item.commission_amount }}</TableCell>
                <TableCell class="px-4">
                  <Badge :variant="commissionStatusVariant(item.status)" size="sm">
                    {{ commissionStatusLabel(item.status) }}
                  </Badge>
                </TableCell>
                <TableCell class="px-4 text-xs text-muted-foreground">{{ formatDate(item.created_at) }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </template>

      <PaginationNav
        :current-page="commissionsPagination.page"
        :total-pages="commissionsPagination.total_page"
        :loading="commissionsLoading"
        :scroll-top="false"
        @change-page="loadCommissions"
      />
    </div>

    <div v-if="dashboard?.opened" class="rounded-xl border bg-card p-4 sm:p-6">
      <div class="mb-4 flex items-center justify-between gap-3">
        <h3 class="text-base font-semibold text-foreground">{{ t('personalCenter.affiliate.withdrawRecordTitle') }}</h3>
        <Button type="button" variant="outline" size="sm" @click="loadWithdraws(withdrawsPagination.page)">
          {{ t('orders.filters.refresh') }}
        </Button>
      </div>

      <div v-if="withdrawsLoading" class="space-y-3">
        <div v-for="idx in 3" :key="idx" class="h-14 animate-pulse rounded-xl border bg-muted"></div>
      </div>
      <div v-else-if="withdraws.length === 0" class="rounded-xl border border-dashed px-4 py-6 text-sm text-muted-foreground">
        {{ t('personalCenter.affiliate.withdrawEmpty') }}
      </div>
      <template v-else>
        <ul class="affiliate-withdraws divide-y sm:hidden">
          <li v-for="item in withdraws" :key="item.id" class="grid grid-cols-[minmax(0,1fr)_auto] items-start gap-x-3 gap-y-1.5 py-3 first:pt-0 last:pb-0">
            <p class="min-w-0 break-all text-base font-semibold tabular-nums">{{ item.amount }}</p>
            <Badge :variant="withdrawStatusVariant(item.status)" size="sm" class="justify-self-end">{{ withdrawStatusLabel(item.status) }}</Badge>
            <p class="min-w-0 break-words text-xs text-muted-foreground">{{ item.channel }}</p>
            <time :datetime="item.created_at" class="text-right text-xs leading-5 text-muted-foreground">{{ formatDate(item.created_at) }}</time>
          </li>
        </ul>
        <div class="hidden overflow-x-auto rounded-xl border sm:block">
          <Table>
            <TableHeader>
              <TableRow class="bg-muted/50">
                <TableHead class="px-4">{{ t('personalCenter.affiliate.withdrawTable.amount') }}</TableHead>
                <TableHead class="px-4">{{ t('personalCenter.affiliate.withdrawTable.channel') }}</TableHead>
                <TableHead class="px-4">{{ t('personalCenter.affiliate.withdrawTable.status') }}</TableHead>
                <TableHead class="px-4">{{ t('personalCenter.affiliate.withdrawTable.createdAt') }}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="item in withdraws" :key="item.id">
                <TableCell class="px-4 font-mono text-xs text-foreground">{{ item.amount }}</TableCell>
                <TableCell class="px-4 text-xs text-muted-foreground">{{ item.channel }}</TableCell>
                <TableCell class="px-4">
                  <Badge :variant="withdrawStatusVariant(item.status)" size="sm">
                    {{ withdrawStatusLabel(item.status) }}
                  </Badge>
                </TableCell>
                <TableCell class="px-4 text-xs text-muted-foreground">{{ formatDate(item.created_at) }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </template>

      <PaginationNav
        :current-page="withdrawsPagination.page"
        :total-pages="withdrawsPagination.total_page"
        :loading="withdrawsLoading"
        :scroll-top="false"
        @change-page="loadWithdraws"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, useId } from 'vue'
import { useI18n } from 'vue-i18n'
import { affiliateAPI, type AffiliateCommissionData, type AffiliateDashboardData, type AffiliateWithdrawData } from '../../api'
import {
  AFFILIATE_COMMISSION_STATUS_AVAILABLE,
  AFFILIATE_COMMISSION_STATUS_PENDING_CONFIRM,
  AFFILIATE_COMMISSION_STATUS_REJECTED,
  AFFILIATE_COMMISSION_STATUS_WITHDRAWN,
  AFFILIATE_WITHDRAW_STATUS_PAID,
  AFFILIATE_WITHDRAW_STATUS_PENDING_REVIEW,
  AFFILIATE_WITHDRAW_STATUS_REJECTED,
} from '../../constants/affiliate'
import { useAppStore } from '../../stores/app'
import { pageAlertVariant, pageAlertToneClass, type PageAlert } from '../../utils/alerts'
import type { BadgeTone } from '../../utils/status'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import PaginationNav from '../../components/PaginationNav.vue'

const { t } = useI18n()
const appStore = useAppStore()
const withdrawFormId = useId()

const loading = ref(true)
const opening = ref(false)
const submittingWithdraw = ref(false)
const commissionsLoading = ref(false)
const withdrawsLoading = ref(false)
const dashboard = ref<AffiliateDashboardData | null>(null)
const panelAlert = ref<PageAlert | null>(null)

const commissions = ref<AffiliateCommissionData[]>([])
const withdraws = ref<AffiliateWithdrawData[]>([])

const commissionsPagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  total_page: 1,
})

const withdrawsPagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  total_page: 1,
})

const withdrawForm = reactive({
  amount: '',
  channel: '',
  account: '',
})

const channelOptions = computed(() => {
  const channels = appStore.config?.affiliate?.withdraw_channels
  if (!Array.isArray(channels)) return []
  return channels.map((item: any) => String(item || '').trim()).filter(Boolean)
})

const promotionUrl = computed(() => {
  if (!dashboard.value?.affiliate_code) return '-'
  const path = dashboard.value.promotion_path || `/?aff=${dashboard.value.affiliate_code}`
  const origin = typeof window !== 'undefined' ? window.location.origin : ''
  return `${origin}${path}`
})

const conversionRateText = computed(() => {
  const value = Number(dashboard.value?.conversion_rate || 0)
  if (!Number.isFinite(value)) return '0.00%'
  return `${value.toFixed(2)}%`
})

const formatDate = (raw?: string) => {
  if (!raw) return '-'
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return raw
  return date.toLocaleString()
}

const loadDashboard = async () => {
  try {
    const response = await affiliateAPI.dashboard()
    dashboard.value = response.data.data || null
  } catch (err: any) {
    panelAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.affiliate.errors.loadFailed'),
    }
  }
}

const loadCommissions = async (page = 1) => {
  commissionsLoading.value = true
  try {
    const response = await affiliateAPI.commissions({
      page,
      page_size: commissionsPagination.page_size,
    })
    commissions.value = response.data.data || []
    Object.assign(commissionsPagination, response.data.pagination || commissionsPagination)
  } catch {
    commissions.value = []
  } finally {
    commissionsLoading.value = false
  }
}

const loadWithdraws = async (page = 1) => {
  withdrawsLoading.value = true
  try {
    const response = await affiliateAPI.withdraws({
      page,
      page_size: withdrawsPagination.page_size,
    })
    withdraws.value = response.data.data || []
    Object.assign(withdrawsPagination, response.data.pagination || withdrawsPagination)
  } catch {
    withdraws.value = []
  } finally {
    withdrawsLoading.value = false
  }
}

const reloadOpenedData = async () => {
  if (!dashboard.value?.opened) return
  await Promise.all([loadCommissions(1), loadWithdraws(1)])
}

const initialize = async () => {
  loading.value = true
  panelAlert.value = null
  await appStore.loadConfig()
  await loadDashboard()
  await reloadOpenedData()
  loading.value = false
}

const openAffiliate = async () => {
  opening.value = true
  panelAlert.value = null
  try {
    await affiliateAPI.open()
    await loadDashboard()
    await reloadOpenedData()
    panelAlert.value = {
      level: 'success',
      message: t('personalCenter.affiliate.openSuccess'),
    }
  } catch (err: any) {
    panelAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.affiliate.errors.openFailed'),
    }
  } finally {
    opening.value = false
  }
}

const handleApplyWithdraw = async () => {
  panelAlert.value = null
  if (!withdrawForm.amount.trim()) {
    panelAlert.value = {
      level: 'warning',
      message: t('personalCenter.affiliate.errors.withdrawAmountRequired'),
    }
    return
  }
  if (!withdrawForm.channel.trim()) {
    panelAlert.value = {
      level: 'warning',
      message: t('personalCenter.affiliate.errors.withdrawChannelRequired'),
    }
    return
  }
  if (!withdrawForm.account.trim()) {
    panelAlert.value = {
      level: 'warning',
      message: t('personalCenter.affiliate.errors.withdrawAccountRequired'),
    }
    return
  }

  submittingWithdraw.value = true
  try {
    await affiliateAPI.applyWithdraw({
      amount: withdrawForm.amount.trim(),
      channel: withdrawForm.channel.trim(),
      account: withdrawForm.account.trim(),
    })
    withdrawForm.amount = ''
    withdrawForm.account = ''
    panelAlert.value = {
      level: 'success',
      message: t('personalCenter.affiliate.withdrawSuccess'),
    }
    await Promise.all([loadDashboard(), loadWithdraws(1)])
  } catch (err: any) {
    panelAlert.value = {
      level: 'error',
      message: err?.message || t('personalCenter.affiliate.errors.withdrawFailed'),
    }
  } finally {
    submittingWithdraw.value = false
  }
}

const copyPromotionUrl = async () => {
  if (!dashboard.value?.affiliate_code || !promotionUrl.value || promotionUrl.value === '-') return
  try {
    await navigator.clipboard.writeText(promotionUrl.value)
    panelAlert.value = {
      level: 'success',
      message: t('personalCenter.affiliate.copySuccess'),
    }
  } catch {
    panelAlert.value = {
      level: 'error',
      message: t('personalCenter.affiliate.errors.copyFailed'),
    }
  }
}

const commissionStatusLabel = (status?: string) => {
  if (status === AFFILIATE_COMMISSION_STATUS_PENDING_CONFIRM) return t('personalCenter.affiliate.commissionStatus.pendingConfirm')
  if (status === AFFILIATE_COMMISSION_STATUS_AVAILABLE) return t('personalCenter.affiliate.commissionStatus.available')
  if (status === AFFILIATE_COMMISSION_STATUS_REJECTED) return t('personalCenter.affiliate.commissionStatus.rejected')
  if (status === AFFILIATE_COMMISSION_STATUS_WITHDRAWN) return t('personalCenter.affiliate.commissionStatus.withdrawn')
  return status || '-'
}

const commissionStatusVariant = (status?: string): BadgeTone => {
  if (status === AFFILIATE_COMMISSION_STATUS_PENDING_CONFIRM) return 'warning'
  if (status === AFFILIATE_COMMISSION_STATUS_AVAILABLE) return 'success'
  if (status === AFFILIATE_COMMISSION_STATUS_WITHDRAWN) return 'info'
  return 'neutral'
}

const withdrawStatusLabel = (status?: string) => {
  if (status === AFFILIATE_WITHDRAW_STATUS_PENDING_REVIEW) return t('personalCenter.affiliate.withdrawStatus.pendingReview')
  if (status === AFFILIATE_WITHDRAW_STATUS_REJECTED) return t('personalCenter.affiliate.withdrawStatus.rejected')
  if (status === AFFILIATE_WITHDRAW_STATUS_PAID) return t('personalCenter.affiliate.withdrawStatus.paid')
  return status || '-'
}

const withdrawStatusVariant = (status?: string): BadgeTone => {
  if (status === AFFILIATE_WITHDRAW_STATUS_PENDING_REVIEW) return 'warning'
  if (status === AFFILIATE_WITHDRAW_STATUS_PAID) return 'success'
  return 'neutral'
}

onMounted(() => {
  initialize()
})
</script>

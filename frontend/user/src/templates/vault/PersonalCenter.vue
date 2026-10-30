<template>
  <div class="mx-auto w-full max-w-[1180px] px-4 pb-10 pt-4 sm:px-6 sm:pt-[22px]">
    <!-- 账户头部 -->
    <header class="mb-4 flex flex-wrap items-center justify-between gap-3 rounded-xl border bg-card p-3.5 sm:mb-[22px] sm:gap-[18px] sm:px-7 sm:py-[26px]">
      <div class="flex min-w-0 items-center gap-3 sm:gap-4">
        <div class="grid h-10 w-10 flex-none sm:h-[62px] sm:w-[62px] place-items-center rounded-md bg-primary/10 text-[26px] font-extrabold text-primary">{{ displayInitial }}</div>
        <div class="min-w-0">
          <p class="hidden text-[13px] font-bold uppercase tracking-[0.04em] text-primary sm:block">{{ t('personalCenter.title') }}</p>
          <h1 class="my-1 break-words text-[26px] font-extrabold">{{ userProfileStore.displayName }}</h1>
          <p class="break-all text-sm text-muted-foreground">{{ userProfileStore.profile?.email || t('personalCenter.subtitle') }}</p>
        </div>
      </div>
      <div class="flex-wrap items-center gap-2" :class="currentSection === 'overview' ? 'hidden sm:flex' : 'flex'">
        <span class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[12.5px] font-semibold" :class="emailVerifiedVariant === 'success' ? 'bg-[color:var(--teal-soft)] text-[color:var(--teal-strong)]' : 'bg-[color:var(--gold-soft)] text-[color:var(--gold-strong)]'">{{ emailVerifiedLabel }}</span>
        <span v-if="userProfileStore.currentLevel" class="inline-flex items-center gap-1.5 rounded-full bg-[color:var(--gold-soft)] px-2.5 py-1 text-[12.5px] font-semibold text-[color:var(--gold-strong)]">
          <img v-if="isImagePath(userProfileStore.currentLevel?.icon)" :src="getImageUrl(userProfileStore.currentLevel!.icon)" class="h-3.5 w-3.5 object-contain" alt="" />
          <span v-else-if="userProfileStore.currentLevel?.icon">{{ userProfileStore.currentLevel.icon }}</span>
          <Crown v-else class="h-3.5 w-3.5" />
          {{ levelName(userProfileStore.currentLevel) }}
        </span>
      </div>
    </header>

    <div class="grid items-start gap-4 sm:gap-[22px] lg:grid-cols-[248px_1fr]">
      <!-- 侧栏 -->
      <aside class="min-w-0 sticky top-[90px] max-[900px]:static">
        <nav ref="sectionNav" class="flex flex-col gap-0.5 rounded-xl border bg-card p-1 sm:p-2.5 max-[900px]:flex-row max-[900px]:overflow-x-auto">
          <button
            v-for="item in visibleSectionItems"
            :key="item.key"
            :aria-current="currentSection === item.key ? 'page' : undefined"
            type="button"
            class="flex w-full items-center gap-2 rounded-sm px-3 py-2.5 text-left text-[13px] leading-5 sm:gap-[11px] sm:px-3.5 sm:py-[11px] sm:text-[14.5px] font-semibold transition-colors max-[900px]:w-auto max-[900px]:shrink-0 max-[900px]:whitespace-nowrap"
            :class="currentSection === item.key ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-secondary hover:text-foreground'"
            @click="switchSection(item.key)"
          >
            <component :is="item.icon" class="h-4 w-4 flex-none sm:h-[18px] sm:w-[18px]" />
            <span>{{ t(item.label) }}</span>
          </button>
        </nav>
      </aside>

      <!-- 内容 -->
      <section class="grid min-w-0 gap-4 sm:gap-[18px]">
        <div v-if="globalAlert" class="rounded-sm px-3.5 py-3 text-[13px] font-semibold" :class="globalAlert.level === 'error' ? 'bg-destructive/10 text-destructive' : (globalAlert.level === 'success' ? 'bg-[color:var(--teal-soft)] text-[color:var(--teal-strong)]' : 'bg-warning/10 text-warning')">
          {{ globalAlert.message }}
        </div>

        <!-- 概览 -->
        <template v-if="currentSection === 'overview'">
          <div class="grid gap-3.5 grid-cols-2 lg:grid-cols-4">
            <div class="overview-stat flex flex-row-reverse items-start justify-between gap-2 rounded-lg border bg-card p-3 sm:flex-row sm:items-center sm:justify-start sm:gap-3 sm:p-4">
              <div class="overview-stat-icon grid h-7 w-7 flex-none sm:h-[42px] sm:w-[42px] place-items-center rounded-xl bg-primary/10 text-primary"><ShoppingBag class="h-4 w-4 sm:h-5 sm:w-5" /></div>
              <div class="overview-stat-content min-w-0 flex-1">
                <div class="overview-stat-label text-xs text-muted-foreground">{{ t('personalCenter.tabs.orders') }}</div>
                <div class="overview-stat-value mt-[3px] text-lg font-extrabold tabular-nums">{{ userProfileStore.loadingOrders ? '—' : userProfileStore.ordersTotal }}</div>
              </div>
            </div>
            <div class="overview-stat flex flex-row-reverse items-start justify-between gap-2 rounded-lg border bg-card p-3 sm:flex-row sm:items-center sm:justify-start sm:gap-3 sm:p-4" :class="{ 'overview-stat-duplicate': userProfileStore.memberLevels.length > 0 }">
              <div class="overview-stat-icon grid h-7 w-7 flex-none sm:h-[42px] sm:w-[42px] place-items-center rounded-xl bg-[color:var(--plum-soft)] text-[color:var(--plum)]"><Crown class="h-4 w-4 sm:h-5 sm:w-5" /></div>
              <div class="overview-stat-content min-w-0 flex-1">
                <div class="overview-stat-label text-xs text-muted-foreground">{{ t('personalCenter.memberLevel.currentLevel') }}</div>
                <div class="overview-stat-value mt-[3px] flex items-center gap-1.5 text-[15px] font-extrabold">
                  <img v-if="isImagePath(userProfileStore.currentLevel?.icon)" :src="getImageUrl(userProfileStore.currentLevel!.icon)" class="h-3.5 w-3.5 object-contain" alt="" />
                  <span>{{ levelName(userProfileStore.currentLevel) }}</span>
                </div>
              </div>
            </div>
            <div class="overview-stat flex flex-row-reverse items-start justify-between gap-2 rounded-lg border bg-card p-3 sm:flex-row sm:items-center sm:justify-start sm:gap-3 sm:p-4" :class="{ 'overview-stat-duplicate': userProfileStore.memberLevels.length > 0 }">
              <div class="overview-stat-icon grid h-7 w-7 flex-none sm:h-[42px] sm:w-[42px] place-items-center rounded-xl bg-[color:var(--gold-soft)] text-[color:var(--gold-strong)]"><Percent class="h-4 w-4 sm:h-5 sm:w-5" /></div>
              <div class="overview-stat-content min-w-0 flex-1">
                <div class="overview-stat-label text-xs text-muted-foreground">{{ t('personalCenter.memberLevel.discountRate') }}</div>
                <div class="overview-stat-value mt-[3px] text-lg font-extrabold">{{ discountText }}</div>
              </div>
            </div>
            <div class="overview-stat flex flex-row-reverse items-start justify-between gap-2 rounded-lg border bg-card p-3 sm:flex-row sm:items-center sm:justify-start sm:gap-3 sm:p-4">
              <div class="overview-stat-icon grid h-7 w-7 flex-none sm:h-[42px] sm:w-[42px] place-items-center rounded-xl bg-[color:var(--teal-soft)] text-[color:var(--teal-strong)]"><ShieldCheck class="h-4 w-4 sm:h-5 sm:w-5" /></div>
              <div class="overview-stat-content min-w-0 flex-1">
                <div class="overview-stat-label text-xs text-muted-foreground">{{ t('personalCenter.overview.accountLabel') }}</div>
                <div class="overview-stat-value mt-[3px]"><span class="inline-flex items-center rounded-full px-0 py-0 text-[12.5px] sm:px-2.5 sm:py-1 font-semibold" :class="emailVerifiedVariant === 'success' ? 'text-[color:var(--teal-strong)] sm:bg-[color:var(--teal-soft)]' : 'text-[color:var(--gold-strong)] sm:bg-[color:var(--gold-soft)]'">{{ emailVerifiedLabel }}</span></div>
              </div>
            </div>
          </div>

          <!-- 会员等级卡 -->
          <div v-if="userProfileStore.memberLevels.length > 0" class="rounded-xl border bg-card p-4 sm:p-[22px]">
            <div class="flex flex-wrap items-center justify-between gap-3.5">
              <div class="flex min-w-0 items-center gap-2.5 sm:gap-3.5">
                <div class="grid h-8 w-8 flex-none sm:h-[46px] sm:w-[46px] place-items-center rounded-md bg-primary/10 text-[22px]">
                  <img v-if="isImagePath(userProfileStore.currentLevel?.icon)" :src="getImageUrl(userProfileStore.currentLevel!.icon)" class="h-5 w-5 object-contain sm:h-7 sm:w-7" alt="" />
                  <span v-else-if="userProfileStore.currentLevel?.icon">{{ userProfileStore.currentLevel.icon }}</span>
                  <Crown v-else class="h-5 w-5 sm:h-6 sm:w-6" />
                </div>
                <div>
                  <p class="text-[11px] uppercase tracking-[0.06em] text-muted-foreground">{{ t('personalCenter.memberLevel.currentLevel') }}</p>
                  <p class="mt-0.5 text-lg font-bold">{{ levelName(userProfileStore.currentLevel) }}</p>
                </div>
              </div>
              <div class="ml-auto flex flex-col items-end gap-1 text-right">
                <p class="text-xs text-muted-foreground">{{ t('personalCenter.memberLevel.discountRate') }}</p>
                <p class="text-sm font-semibold text-foreground">{{ discountText }}</p>
                <span v-if="!userProfileStore.nextLevel && userProfileStore.currentLevel" class="inline-flex items-center rounded-full bg-[color:var(--teal-soft)] px-2.5 py-1 text-[12.5px] font-semibold text-[color:var(--teal-strong)]">{{ t('personalCenter.memberLevel.highestLevel') }}</span>
              </div>
            </div>

            <div v-if="userProfileStore.nextLevel" class="mt-4 border-t pt-4 sm:mt-[18px] sm:rounded-lg sm:border sm:border-primary/20 sm:bg-primary/10 sm:p-4">
              <div class="flex items-center gap-3">
                <div class="grid h-6 w-6 flex-none sm:h-[38px] sm:w-[38px] place-items-center rounded-md bg-secondary text-lg opacity-80">
                  <img v-if="isImagePath(userProfileStore.nextLevel.icon)" :src="getImageUrl(userProfileStore.nextLevel.icon)" class="h-3.5 w-3.5 object-contain" alt="" />
                  <span v-else-if="userProfileStore.nextLevel.icon">{{ userProfileStore.nextLevel.icon }}</span>
                  <Crown v-else class="h-5 w-5" />
                </div>
                <div>
                  <p class="text-[11px] uppercase tracking-[0.06em] text-muted-foreground">{{ t('personalCenter.memberLevel.nextLevel') }}</p>
                  <div class="flex items-center gap-2">
                    <span class="text-sm font-bold text-muted-foreground">{{ levelName(userProfileStore.nextLevel) }}</span>
                    <span v-if="userProfileStore.nextLevel.discount_rate < 100" class="text-xs font-semibold text-primary">{{ t('personalCenter.memberLevel.discountOff', { n: userProfileStore.nextLevel.discount_rate / 10 }) }}</span>
                  </div>
                </div>
              </div>

              <div v-if="userProfileStore.upgradeProgress" class="mt-3.5 grid gap-3">
                <div v-if="userProfileStore.upgradeProgress.rechargePercent !== null">
                  <div class="mb-1.5 flex justify-between text-xs">
                    <span class="text-muted-foreground">{{ t('personalCenter.memberLevel.rechargeProgress') }}</span>
                    <span class="text-muted-foreground tabular-nums">{{ userProfileStore.upgradeProgress.recharged.toFixed(2) }} / {{ userProfileStore.upgradeProgress.rechargeThreshold.toFixed(2) }}</span>
                  </div>
                  <div class="h-1.5 overflow-hidden rounded-full bg-secondary"><div class="h-full rounded-full bg-primary transition-[width] duration-700" :style="{ width: userProfileStore.upgradeProgress.rechargePercent + '%' }"></div></div>
                </div>
                <div v-if="userProfileStore.upgradeProgress.spendPercent !== null">
                  <div class="mb-1.5 flex justify-between text-xs">
                    <span class="text-muted-foreground">{{ t('personalCenter.memberLevel.spendProgress') }}</span>
                    <span class="text-muted-foreground tabular-nums">{{ userProfileStore.upgradeProgress.spent.toFixed(2) }} / {{ userProfileStore.upgradeProgress.spendThreshold.toFixed(2) }}</span>
                  </div>
                  <div class="h-1.5 overflow-hidden rounded-full bg-secondary"><div class="h-full rounded-full bg-primary transition-[width] duration-700" :style="{ width: userProfileStore.upgradeProgress.spendPercent + '%' }"></div></div>
                </div>
              </div>
            </div>
          </div>

          <!-- 最近订单 -->
          <div class="rounded-xl border bg-card p-4 sm:p-[22px]">
            <div class="mb-4 flex items-center justify-between gap-3.5">
              <h3 class="text-lg font-bold">{{ t('personalCenter.overview.recentOrdersTitle') }}</h3>
              <Button as-child variant="ghost" size="sm" class="rounded-full"><RouterLink to="/me/orders">{{ t('personalCenter.overview.viewAllOrders') }}</RouterLink></Button>
            </div>
            <div v-if="userProfileStore.loadingOrders" class="grid gap-2.5">
              <div v-for="idx in 3" :key="idx" class="h-16 rounded-md bg-secondary"></div>
            </div>
            <div v-else-if="userProfileStore.recentOrders.length === 0" class="rounded-md border border-dashed p-[18px] text-[13px] text-muted-foreground">{{ t('personalCenter.overview.emptyOrders') }}</div>
            <div v-else class="grid gap-2.5">
              <div v-for="order in userProfileStore.recentOrders" :key="order.order_no" class="recent-order grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-3 gap-y-2 rounded-lg border p-3 transition hover:border-hairline-strong sm:p-4">
                <div class="recent-order-reference min-w-0 break-words text-xs text-muted-foreground [overflow-wrap:anywhere]">{{ t('orders.orderNo') }}：{{ order.order_no }}</div>
                <Badge :variant="statusVariant(order.status)" class="recent-order-status justify-self-end rounded-full">{{ statusLabel(order.status) }}</Badge>
                <div class="recent-order-details min-w-0">
                  <div class="break-words text-base font-bold tabular-nums">{{ formatMoney(order.total_amount, order.currency) }}</div>
                  <div class="mt-1 break-words text-xs leading-5 text-muted-foreground">{{ formatDate(order.created_at) }}</div>
                </div>
                <div class="recent-order-actions flex flex-col items-end gap-2 min-[380px]:flex-row min-[380px]:items-center">
                  <Button as-child variant="outline" size="sm" class="rounded-lg"><RouterLink :to="`/orders/${order.order_no}`">{{ t('orders.viewDetails') }}</RouterLink></Button>
                  <Button v-if="order.status === 'pending_payment'" as-child size="sm" class="rounded-lg"><RouterLink :to="`/pay?order_no=${order.order_no}`">{{ t('orders.payNow') }}</RouterLink></Button>
                </div>
              </div>
            </div>
          </div>
        </template>

        <!-- 其余面板：复用现有组件（shadcn 风，嵌于 vault 作用域内） -->
        <div v-else class="min-w-0">
          <ProfilePanel v-if="currentSection === 'profile'" />
          <SecurityPanel v-else-if="currentSection === 'security'" />
          <OrdersPanel v-else-if="currentSection === 'orders'" />
          <WalletPanel v-else-if="currentSection === 'wallet'" />
          <AffiliatePanel v-else-if="currentSection === 'affiliate'" />
          <div v-else-if="currentSection === 'reseller' && canAccessResellerConsole" class="rounded-xl border bg-card p-[22px]">
            <h2 class="text-lg font-bold">{{ t('resellerConsole.title') }}</h2>
            <p class="mt-2 text-muted-foreground">{{ t('resellerConsole.dashboard.description') }}</p>
            <Button as-child size="sm" class="mt-4 rounded-full"><RouterLink to="/reseller">{{ t('resellerConsole.nav.dashboard') }}</RouterLink></Button>
          </div>
          <GiftCardPanel v-else-if="currentSection === 'giftCard'" />
          <ApiPanel v-else-if="currentSection === 'api'" />
          <OrdersPanel v-else />
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ref, watch } from 'vue'
import { useResizeObserver } from '@vueuse/core'
import { Crown, ShoppingBag, ShieldCheck, Percent } from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { getImageUrl } from '../../utils/image'
import ProfilePanel from '../../views/personal/ProfilePanel.vue'
import SecurityPanel from '../../views/personal/SecurityPanel.vue'
import OrdersPanel from '../../views/personal/OrdersPanel.vue'
import WalletPanel from '../../views/personal/WalletPanel.vue'
import GiftCardPanel from '../../views/personal/GiftCardPanel.vue'
import AffiliatePanel from '../../views/personal/AffiliatePanel.vue'
import ApiPanel from '../../views/personal/ApiPanel.vue'
import { usePersonalCenter, type PersonalSection } from '../../composables/usePersonalCenter'

const { t, locale } = useI18n()

const props = withDefaults(defineProps<{ section?: PersonalSection }>(), {
  section: 'overview',
})

const {
  userProfileStore, canAccessResellerConsole, visibleSectionItems, currentSection, globalAlert,
  displayInitial, switchSection, statusLabel, statusVariant, formatMoney, formatDate,
  emailVerifiedLabel, emailVerifiedVariant, discountText, isImagePath, levelName,
} = usePersonalCenter(() => props.section)

const sectionNav = ref<HTMLElement | null>(null)
const revealActiveSection = () => {
  const nav = sectionNav.value
  const active = nav?.querySelector<HTMLElement>('[aria-current="page"]')
  if (!nav || !active || nav.scrollWidth <= nav.clientWidth) return
  const navBox = nav.getBoundingClientRect()
  const activeBox = active.getBoundingClientRect()
  if (activeBox.left < navBox.left) nav.scrollLeft += activeBox.left - navBox.left - 8
  else if (activeBox.right > navBox.right) nav.scrollLeft += activeBox.right - navBox.right + 8
}
watch([currentSection, locale], revealActiveSection, { flush: 'post' })
useResizeObserver(sectionNav, revealActiveSection)
</script>

<style scoped>
@media (width < 640px) {
  .overview-stat {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 24px;
    grid-row: span 2;
    grid-template-rows: subgrid;
    align-items: start;
    gap: 4px 8px;
  }

  .overview-stat-duplicate {
    display: none;
  }

  .overview-stat-icon {
    grid-column: 2;
    grid-row: 1;
    width: 24px;
    height: 24px;
  }

  .overview-stat-content {
    display: contents;
  }

  .overview-stat-label {
    grid-column: 1;
    grid-row: 1;
    display: flex;
    align-items: center;
    min-height: 24px;
    line-height: 20px;
  }

  .overview-stat-value {
    grid-column: 1 / -1;
    grid-row: 2;
    margin: 0;
    min-height: 24px;
  }

  .overview-stat-value,
  .overview-stat-value > span {
    font-size: 16px;
    font-weight: 600;
    line-height: 24px;
  }
}
</style>

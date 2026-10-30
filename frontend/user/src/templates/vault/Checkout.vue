<template>
  <div class="checkout-page mx-auto w-full max-w-[1120px] px-4 pb-4 pt-4 sm:px-6 lg:flex lg:min-h-[calc(100dvh-71px)] lg:flex-col lg:pt-6">
    <nav :aria-label="t('nav.home')" class="mb-4 flex flex-wrap items-center justify-between gap-x-3 gap-y-2 text-[13px] font-medium text-muted-foreground">
      <div class="flex items-center gap-1.5">
        <RouterLink to="/" class="hover:text-primary">{{ t('nav.home') }}</RouterLink>
        <ChevronRight class="h-4 w-4 flex-none" />
        <span class="text-foreground" aria-current="page">{{ t('checkout.title') }}</span>
      </div>
      <p v-if="!userAuthStore.isAuthenticated" class="ml-auto hidden flex-wrap items-center justify-end gap-1.5 text-xs text-muted-foreground max-[900px]:flex">
        {{ t('checkout.accountLoginHint') }}
        <RouterLink class="inline-flex min-h-8 items-center font-medium text-primary underline-offset-4 hover:underline" :to="{ path: '/auth/login', query: { redirect: '/checkout' } }">{{ t('navbar.login') }}</RouterLink>
      </p>
    </nav>
    <h1 class="sr-only">{{ t('checkout.title') }}</h1>

    <!-- 空 -->
    <div v-if="purchaseItems.length === 0" class="my-8 flex flex-col items-center gap-3 rounded-xl border border-dashed py-16 text-center text-muted-foreground">
      <ShoppingBag class="h-10 w-10 opacity-60" />
      <p>{{ t('checkout.empty') }}</p>
      <Button as-child class="mt-2 rounded-full" size="sm"><RouterLink to="/products">{{ t('checkout.emptyAction') }}</RouterLink></Button>
    </div>

    <div v-else class="grid min-w-0 items-start gap-4 lg:flex-1 lg:gap-8 lg:grid-cols-[minmax(0,1fr)_368px]">
      <fieldset :disabled="submitting" class="contents">
      <div class="grid min-w-0 gap-4 lg:flex lg:flex-col lg:self-stretch">
        <article v-for="item in purchaseItems" :key="purchaseItemKey(item)" class="grid min-w-0 grid-cols-[48px_minmax(0,1fr)] gap-3 rounded-xl border bg-card p-3.5 sm:p-4 lg:grid-cols-[64px_minmax(0,1fr)] lg:gap-4 lg:p-5">
          <RouterLink :to="`/products/${item.slug}`" class="relative grid h-12 w-12 place-items-center lg:h-16 lg:w-16 overflow-hidden rounded-md bg-secondary">
            <img v-if="checkoutItemImage(item)" :src="checkoutItemImage(item)" :alt="getLocalizedText(item.title)" class="h-full w-full object-cover" />
            <Package v-else class="h-7 w-7 text-muted-foreground" />
          </RouterLink>
          <div class="min-w-0">
            <RouterLink :to="`/products/${item.slug}`" class="line-clamp-2 break-words text-[15px] font-semibold leading-[23px] hover:text-primary lg:text-base lg:leading-6">{{ getLocalizedText(item.title) }}</RouterLink>
            <p v-if="itemSkuDisplay(item)" class="mt-1 text-xs leading-5 text-muted-foreground lg:text-[13px]">{{ t('checkout.skuLabel') }}：{{ itemSkuDisplay(item) }}</p>
            <p v-if="itemStockHint(item)" class="mt-1 text-xs leading-5" :class="itemStockExceeded(item) ? 'text-warning' : 'text-muted-foreground'">{{ itemStockHint(item) }}</p>
          </div>
          <div class="col-span-2 flex flex-wrap items-center justify-between gap-3 border-t pt-3">
            <div class="inline-flex items-center overflow-hidden rounded-lg border border-hairline-strong">
              <button type="button" class="grid h-11 w-11 place-items-center disabled:opacity-35" :aria-label="t('checkout.decreaseQuantity')" :disabled="item.quantity <= itemPurchaseMin(item)" @click="updateQuantity(item, item.quantity - 1)"><Minus class="h-4 w-4" /></button>
              <input type="number" :aria-label="t('checkout.quantityLabel')" class="h-11 w-10 border-x bg-transparent text-center font-semibold tabular-nums outline-none [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none" :value="item.quantity" :min="itemPurchaseMin(item)" :max="itemMaxQuantity(item)" @change="handleQuantityChange(item, $event)" />
              <button type="button" class="grid h-11 w-11 place-items-center disabled:opacity-35" :aria-label="t('checkout.increaseQuantity')" :disabled="item.quantity >= itemMaxQuantity(item)" @click="updateQuantity(item, item.quantity + 1)"><Plus class="h-4 w-4" /></button>
            </div>
            <div class="min-w-0 text-right">
              <span class="block text-xs text-muted-foreground">{{ t('checkout.totalPriceLabel') }}</span>
              <strong class="checkout-total break-words text-lg font-semibold" :class="checkoutItemHasPriceDiscount(item) ? 'text-primary' : 'text-foreground'">{{ checkoutItemPriceParts(item).integer }}{{ checkoutItemPriceParts(item).decimal }} {{ checkoutItemCurrency }}</strong>
              <span v-if="checkoutItemHasPriceDiscount(item)" class="block text-xs text-muted-foreground line-through">{{ checkoutItemOriginalPriceParts(item).integer }}{{ checkoutItemOriginalPriceParts(item).decimal }} {{ checkoutItemCurrency }}</span>
            </div>
          </div>
        </article>

        <div class="min-w-0 lg:flex lg:flex-1 lg:flex-col">
        <section class="checkout-product-details min-w-0 rounded-xl border bg-card p-4 lg:basis-3/4 lg:shrink-0 lg:p-5">
          <h2 class="mb-3 border-b pb-3 text-base font-semibold">{{ t('productDetail.details') }}</h2>
          <p v-if="purchaseDescription" class="mb-3 whitespace-pre-line break-words text-sm leading-7 text-muted-foreground">{{ purchaseDescription }}</p>
          <div v-if="purchaseContent" class="theme-prose prose prose-sm max-w-none overflow-x-auto break-words dark:prose-invert prose-a:text-primary prose-img:rounded-md" v-html="purchaseContent"></div>
          <p v-if="!purchaseDescription && !purchaseContent" class="text-sm leading-7 text-muted-foreground">{{ t('checkout.detailsEmpty') }}</p>
        </section>
        </div>

        <!-- 自定义表单（复用） -->
        <CheckoutManualForm
          v-if="manualFormProducts.length"
          :manual-form-products="manualFormProducts"
          v-model="manualFormData"
          :submit-attempted="submitAttempted"
          :get-manual-field-label="getManualFieldLabel"
          :get-manual-field-placeholder="getManualFieldPlaceholder"
          :manual-field-error="manualFieldError"
        />

      </div>

      <!-- 支付 -->
      <aside class="min-w-0 border-t pt-4 lg:sticky lg:top-[90px] lg:border-t-0 lg:border-l lg:pl-8 lg:pt-0">

        <!-- 优惠码 -->
        <details v-if="!isResellerTenant" class="mb-4 border-b pb-3">
          <summary class="cursor-pointer text-sm font-medium">{{ t('checkout.couponTitle') }}</summary>
          <Input v-model="couponCode" :disabled="submitting" type="text" class="mt-3 h-11" :aria-label="t('checkout.couponTitle')" :placeholder="t('checkout.couponPlaceholder')" />
        </details>

        <!-- 下单方式 -->
        <section ref="guestInfo" v-if="!userAuthStore.isAuthenticated" class="mb-4 border-b pb-4">
          <div class="grid grid-cols-2 gap-2">
            <label class="grid min-w-0 grid-rows-[1fr_auto] gap-1.5 text-xs text-muted-foreground lg:text-[13px]">
              {{ t('checkout.guestEmailLabel') }}
              <Input v-model="guestEmail" :aria-invalid="submitAttempted && !guestEmailValid" type="email" autocomplete="email" class="h-11 min-w-0 text-sm" />
            </label>
            <label class="grid min-w-0 grid-rows-[1fr_auto] gap-1.5 text-xs text-muted-foreground lg:text-[13px]">
              {{ t('checkout.guestPasswordLabel', { count: MIN_GUEST_ORDER_PASSWORD_LENGTH }) }}
              <Input v-model="guestPassword" :aria-invalid="submitAttempted && [...guestPassword.trim()].length < MIN_GUEST_ORDER_PASSWORD_LENGTH" type="password" :minlength="MIN_GUEST_ORDER_PASSWORD_LENGTH" autocomplete="new-password" class="h-11 min-w-0 text-sm" />
            </label>
          </div>

          <div v-if="guestCaptchaEnabled" class="mt-3.5">
            <p class="mb-2 text-[11px] font-bold uppercase tracking-[0.1em] text-muted-foreground">{{ t('auth.common.captchaLabel') }}</p>
            <ImageCaptcha
              v-if="captchaProvider === 'image'"
              ref="guestImageCaptchaRef"
              v-model="guestCaptchaPayload"
              :disabled="submitting"
              @config-stale="handleGuestCaptchaConfigStale"
            />
            <TurnstileCaptcha
              v-else-if="captchaProvider === 'turnstile'"
              ref="guestTurnstileRef"
              v-model="guestTurnstileToken"
              :site-key="guestTurnstileSiteKey"
            />
          </div>

          <p class="mt-2 text-xs leading-5 text-muted-foreground lg:text-[13px]">{{ t('checkout.guestDetailsHint') }}</p>
          <p v-if="guestEmail && !guestEmailValid" class="mt-2 text-[13px] text-destructive">{{ t('error.email_invalid') }}</p>
        </section>

        <!-- 支付方式 -->
        <div ref="paymentMethods" class="mb-4">
          <h3 class="mb-3 text-sm font-bold">{{ t('checkout.paymentMethod') }}</h3>

          <div v-if="showBalanceOption" class="mb-3 rounded-sm border bg-secondary p-3.5">
            <div class="flex items-start justify-between gap-2.5">
              <div>
                <div class="text-xs text-muted-foreground">{{ t('payment.walletBalanceLabel') }}</div>
                <div class="mt-0.5 font-bold text-foreground">{{ walletLoading ? t('common.loading') : formatPrice(walletBalance, previewCurrency) }}</div>
              </div>
              <label class="inline-flex items-center gap-1.5 text-xs text-muted-foreground lg:text-[13px]">
                <input v-model="useBalance" type="checkbox" class="h-4 w-4 accent-[var(--ui-accent)]" :disabled="walletOnlyPayment" />
                <span>{{ t('payment.useBalance') }}</span>
              </label>
            </div>
            <div v-if="walletOnlyPayment" class="mt-2 text-xs text-warning">{{ t('payment.walletOnlyHint') }}</div>
            <div v-if="useBalance" class="mt-2.5 grid gap-0.5 text-xs text-muted-foreground">
              <div>{{ t('payment.walletDeductLabel') }}：{{ expectedWalletPaidDisplay }}</div>
              <div v-if="!walletOnlyPayment">{{ t('payment.onlinePayLabel') }}：{{ expectedOnlinePayDisplay }}</div>
              <div v-if="walletOnlyPayment && expectedOnlinePayCents > 0" class="text-warning">{{ t('payment.walletInsufficientHint') }}</div>
            </div>
          </div>

          <template v-if="!walletOnlyPayment">
            <div v-if="requiresOnlineChannel && paymentChannels.length > 0" role="radiogroup" :aria-label="t('checkout.paymentMethod')" class="grid grid-cols-2 gap-2.5">
              <label
                v-for="channel in paymentChannels"
                :key="channel.id"
                class="relative flex min-h-14 min-w-0 cursor-pointer flex-col justify-center rounded-xl border p-3 transition-colors focus-within:ring-2 focus-within:ring-primary/30"
                :class="[
                  selectedChannelId === channel.id && !isChannelDisabledForAmount(channel) ? 'border-primary bg-primary/5' : 'border-border bg-card hover:border-primary/50',
                  isChannelDisabledForAmount(channel) ? 'cursor-not-allowed opacity-50' : '',
                ]"
                :title="isChannelDisabledForAmount(channel) ? channelAmountLimitHint(channel) : ''"
              >
                <input
                  type="radio"
                  name="checkout-payment-channel"
                  class="sr-only"
                  :value="channel.id"
                  :checked="selectedChannelId === channel.id"
                  :disabled="isChannelDisabledForAmount(channel)"
                  @change="handleSelectChannel(channel)"
                />
                <div class="flex items-center gap-2">
                  <span v-if="channel.icon" class="grid h-7 w-7 flex-none place-items-center rounded-lg bg-secondary" aria-hidden="true">
                    <img :src="getImageUrl(channel.icon)" alt="" loading="lazy" class="h-5 w-5 object-contain" />
                  </span>
                  <span class="min-w-0 flex-1 break-words text-[13px] font-semibold leading-5 text-foreground lg:text-sm">{{ channel.name }}</span>
                  <span class="grid h-4 w-4 flex-none place-items-center rounded-full border" :class="selectedChannelId === channel.id ? 'border-primary bg-primary text-primary-foreground' : 'border-muted-foreground/40'" aria-hidden="true">
                    <Check v-if="selectedChannelId === channel.id" class="h-3 w-3" />
                  </span>
                </div>
                <div v-if="channel.fee_policy === 'customer_surcharge'" class="mt-2 flex flex-wrap gap-x-2 text-[11px] leading-4 text-muted-foreground">
                  <span>{{ t('payment.feeLabel') }} {{ formatChannelFeeRate(channel) }}</span>
                  <span>{{ t('payment.fixedFeeLabel') }} {{ formatChannelFixedFee(channel) }}</span>
                </div>
                <div v-if="isChannelDisabledForAmount(channel)" class="mt-1 text-[11px] text-warning">{{ channelAmountLimitHint(channel) }}</div>
              </label>
            </div>
            <div v-else-if="requiresOnlineChannel && paymentChannels.length === 0" class="text-[13px] text-muted-foreground">{{ t('checkout.noPaymentChannels') }}</div>
          </template>
          <div v-if="!requiresOnlineChannel" class="text-[13px] text-[color:var(--teal-strong)]">{{ t('checkout.walletCoversAll') }}</div>
        </div>

        <div class="grid gap-2.5">
          <template v-if="!isResellerTenant">
            <div v-if="hasPositiveAmount(previewCoupon)" class="flex items-center justify-between text-[13.5px]"><span class="text-muted-foreground">{{ t('checkout.previewCoupon') }}</span><span class="font-semibold" :class="hasPositiveAmount(previewCoupon) ? 'text-primary' : 'text-foreground'">{{ formatDiscountPrice(previewCoupon, previewCurrency) }}</span></div>
            <div v-if="hasPositiveAmount(previewPromotion)" class="flex items-center justify-between text-[13.5px]"><span class="text-muted-foreground">{{ t('checkout.previewPromotion') }}</span><span class="font-semibold" :class="hasPositiveAmount(previewPromotion) ? 'text-primary' : 'text-foreground'">{{ formatDiscountPrice(previewPromotion, previewCurrency) }}</span></div>
            <div v-if="hasPositiveAmount(previewWholesale)" class="flex items-center justify-between text-[13.5px]"><span class="text-muted-foreground">{{ t('checkout.previewWholesale') }}</span><span class="font-semibold" :class="hasPositiveAmount(previewWholesale) ? 'text-[color:var(--teal-strong)]' : 'text-foreground'">{{ formatDiscountPrice(previewWholesale, previewCurrency) }}</span></div>
          </template>
          <div v-if="Number(previewMemberDiscount) > 0" class="flex items-center justify-between text-[13.5px]"><span class="text-muted-foreground">{{ t('checkout.previewMemberDiscount') }}</span><span class="font-semibold text-[color:var(--gold-strong)]">-{{ formatPrice(previewMemberDiscount, previewCurrency) }}</span></div>
        </div>

        <div v-if="previewLoading || couponRefreshing" class="mt-3 text-xs text-muted-foreground">{{ previewStatusText }}</div>
        <div ref="paymentError" v-if="checkoutAlert?.level === 'error'" role="alert" class="mt-3.5 rounded-sm bg-destructive/10 px-3 py-2.5 text-[13px] font-semibold text-destructive">{{ checkoutAlert.message }}</div>

        <Button v-if="previewError" type="button" variant="outline" class="mt-2" :disabled="previewLoading || submitting" @click="retryPreview">{{ t('common.retry') }}</Button>

        <Teleport defer to=".vault-scope" :disabled="isDesktop">
          <div aria-hidden="true" class="flex-none lg:hidden" :style="{ height: `${payBarHeight}px` }"></div>
          <div ref="payBar" class="checkout-paybar fixed inset-x-0 bottom-0 z-40 grid grid-cols-[minmax(0,1fr)_132px] items-center gap-5 border-t bg-card px-4 pt-3 pb-[calc(12px+env(safe-area-inset-bottom,0px))] shadow-[0_-4px_20px_rgb(0_0_0/0.05)] lg:static lg:mt-4 lg:bg-transparent lg:px-0 lg:pb-0 lg:pt-4 lg:shadow-none">
            <div class="min-w-0">
              <p class="text-xs text-muted-foreground">{{ t('checkout.previewTotal') }}</p>
              <p class="checkout-total break-all text-[22px] font-semibold leading-8 text-primary">{{ formatPrice(previewTotal, previewCurrency) }}</p>
            </div>
            <Button class="h-11 w-full rounded-full font-semibold" :disabled="submitting" :class="{ 'opacity-50': !canSubmit }" @click="handlePayClick">
              {{ submitting ? t('checkout.submitting') : t('checkout.payNow') }}
            </Button>
          </div>
        </Teleport>
      </aside>
      </fieldset>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { useElementSize, useMediaQuery } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { Check, ChevronRight, Minus, Package, Plus, ShoppingBag } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import ImageCaptcha from '../../components/captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '../../components/captcha/TurnstileCaptcha.vue'
import CheckoutManualForm from '../../components/checkout/CheckoutManualForm.vue'
import { useCheckout } from '../../composables/useCheckout'
import { getProductDescriptionText } from '../../utils/productDescription'
import { processHtmlForDisplay } from '../../utils/content'
import { toast } from '../../composables/useToast'
import { MIN_GUEST_ORDER_PASSWORD_LENGTH } from '../../utils/guestOrderAuth'

const { t } = useI18n()
const isDesktop = useMediaQuery('(min-width: 1024px)')
const payBar = ref<HTMLElement | null>(null)
const guestInfo = ref<HTMLElement | null>(null)
const paymentMethods = ref<HTMLElement | null>(null)
const paymentError = ref<HTMLElement | null>(null)
const { height: payBarHeight } = useElementSize(payBar, { width: 0, height: 84 }, { box: 'border-box' })

const {
  userAuthStore, getLocalizedText, formatPrice, getImageUrl,
  purchaseItems, purchaseDetails, purchaseItemKey, checkoutItemImage, itemSkuDisplay,
  itemStockExceeded, itemStockHint, itemPurchaseMin, itemMaxQuantity, updateQuantity, handleQuantityChange,
  checkoutItemCurrency, checkoutItemPriceParts, checkoutItemOriginalPriceParts, checkoutItemHasPriceDiscount,
  manualFormProducts, manualFormData, submitAttempted, getManualFieldLabel, getManualFieldPlaceholder, manualFieldError,
  couponCode, isResellerTenant,
  guestEmail, guestPassword, guestEmailValid,
  guestCaptchaEnabled, captchaProvider, guestCaptchaPayload, guestTurnstileToken, guestTurnstileSiteKey,
  guestImageCaptchaRef, guestTurnstileRef, handleGuestCaptchaConfigStale,
  previewCurrency, previewCoupon, previewPromotion, previewWholesale, previewMemberDiscount, previewTotal,
  previewLoading, previewError, retryPreview, couponRefreshing, previewStatusText, hasPositiveAmount, formatDiscountPrice, checkoutAlert,
  showBalanceOption, walletLoading, walletBalance, useBalance, walletOnlyPayment,
  expectedWalletPaidDisplay, expectedOnlinePayDisplay, expectedOnlinePayCents,
  requiresOnlineChannel, paymentChannels, selectedChannelId, isChannelDisabledForAmount, channelAmountLimitHint,
  handleSelectChannel, formatChannelFeeRate, formatChannelFixedFee,
  submitting, canSubmit, handleSubmit,
} = useCheckout()

const purchaseDescription = computed(() => getProductDescriptionText(getLocalizedText(purchaseDetails.value.description)))
const purchaseContent = computed(() => {
  const html = processHtmlForDisplay(getLocalizedText(purchaseDetails.value.content))
  return getProductDescriptionText(html) || /<(?:img|video|audio|svg)\b/i.test(html) ? html : ''
})

const handlePayClick = async () => {
  if (submitting.value) return
  if (!canSubmit.value) {
    submitAttempted.value = true
    const message = checkoutAlert.value?.message || t('checkout.previewLoading')
    toast.error(message)
    await nextTick()
    const target = message === t('checkout.errors.selectPayment')
      ? paymentMethods.value
      : guestInfo.value?.querySelector<HTMLElement>('input[aria-invalid="true"]') || paymentError.value
    target?.scrollIntoView({ behavior: 'smooth', block: 'center' })
    target?.focus({ preventScroll: true })
    return
  }
  await handleSubmit()
  if (checkoutAlert.value?.level === 'error') {
    toast.error(checkoutAlert.value.message)
    await nextTick()
    paymentError.value?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }
}

// 字符串模板 ref，逻辑在 composable 内，显式标记避免 noUnusedLocals 误报。
void guestImageCaptchaRef
void guestTurnstileRef
</script>

<style scoped>
.checkout-total {
  font-family: "Rubik", var(--ui-font-family);
}
</style>

<template>
  <div class="product-detail-page min-h-screen bg-background text-foreground">
    <div class="product-detail-shell">
      <!-- Loading Skeleton -->
      <div v-if="loading" class="space-y-5">
        <div class="h-4 w-56 rounded theme-skeleton"></div>
        <div class="product-hero product-hero--loading">
          <div class="min-h-[520px] rounded-[18px] theme-skeleton"></div>
          <div class="flex flex-col gap-5 px-6 py-8">
            <div class="h-10 w-3/5 rounded theme-skeleton"></div>
            <div class="flex gap-3">
              <div v-for="i in 3" :key="i" class="h-12 w-28 rounded-xl theme-skeleton"></div>
            </div>
            <div class="h-14 w-48 rounded theme-skeleton"></div>
            <div class="mt-3 flex gap-3">
              <div v-for="i in 3" :key="i" class="h-16 w-32 rounded-xl theme-skeleton"></div>
            </div>
            <div class="mt-auto flex gap-3">
              <div class="h-13 w-56 rounded-xl theme-skeleton"></div>
              <div class="h-13 w-56 rounded-xl theme-skeleton"></div>
            </div>
          </div>
        </div>
        <div class="h-[420px] rounded-[24px] theme-skeleton"></div>
      </div>

      <div v-else-if="product">
        <BreadcrumbNav
          class="product-breadcrumb"
          :items="[
            { label: t('nav.home'), to: '/' },
            { label: t('nav.products'), to: '/products' },
            { label: getLocalizedText(product.title) },
          ]"
        />

        <!-- Main purchase card -->
        <section class="product-hero">
          <ProductImageGallery
            :images="images"
            :current-image="currentImage"
            :product-title="getLocalizedText(product.title)"
            @update:current-image="currentImage = $event"
          />

          <div class="product-buy-panel">
            <h1 class="product-title">{{ getLocalizedText(product.title) }}</h1>

            <!-- Purchase / fulfillment / stock states keep their own semantic colors. -->
            <div class="product-status-row">
              <div
                class="product-status"
                :class="product.purchase_type === 'guest' ? 'product-status--guest' : 'product-status--member'"
              >
                <UserPlus v-if="product.purchase_type === 'guest'" />
                <Lock v-else />
                <span>{{ getPurchaseTypeLabel(product.purchase_type) }}</span>
              </div>

              <div
                class="product-status"
                :class="product.fulfillment_type === 'auto' ? 'product-status--auto' : 'product-status--manual'"
              >
                <Zap v-if="product.fulfillment_type === 'auto'" />
                <Pencil v-else />
                <span>{{ getFulfillmentTypeLabel(product.fulfillment_type) }}</span>
              </div>

              <div
                class="product-status"
                :class="`product-status--stock-${getStockBadgeVariant(product.stock_status)}`"
              >
                <Package />
                <span>{{ getStockStatusLabel(product) }}</span>
              </div>
            </div>

            <!-- Price -->
            <div ref="priceSection" class="product-price-block">
              <div class="product-price-label-row">
                <span class="product-section-label">{{ t('products.price') }}</span>
                <span
                  v-if="(selectedSku && hasSkuPromotionPrice(selectedSku)) || (!selectedSku && hasPromotionPrice(product))"
                  class="price-flag price-flag--promotion"
                >
                  {{ t('products.promotionTag') }}
                </span>
                <span v-if="showSelectedSkuMemberBadge" class="price-flag price-flag--member">
                  {{ t('products.memberPriceTag') }}
                </span>
                <span v-if="hasSelectedSkuWholesalePrice" class="price-flag price-flag--wholesale">
                  {{ t('products.wholesaleTag') }}
                </span>
              </div>

              <template v-if="selectedSku && hasSelectedSkuWholesalePrice">
                <div class="product-price-row">
                  <span
                    class="product-price"
                    :class="selectedSkuWholesaleFinalIsMember ? 'product-price--member' : 'product-price--wholesale'"
                  >
                    {{ formatPrice(selectedSkuWholesaleFinalPrice!, siteCurrency) }}
                  </span>
                  <span class="product-price-original">{{ formatPrice(selectedSku.price_amount, siteCurrency) }}</span>
                </div>
                <p
                  class="product-price-note"
                  :class="selectedSkuWholesaleFinalIsMember ? 'product-price--member' : 'product-price--wholesale'"
                >
                  {{ selectedSkuWholesaleFinalIsMember ? t('products.memberPriceTag') : t('products.wholesaleTag') }} ·
                  {{ t('products.saveAmount') }}
                  {{ formatPrice(Number(selectedSku.price_amount) - Number(selectedSkuWholesaleFinalPrice), siteCurrency) }}
                </p>
              </template>

              <template v-else-if="selectedSku && hasSkuPromotionPrice(selectedSku)">
                <div class="product-price-row">
                  <span
                    class="product-price"
                    :class="selectedSkuPromotionFinalIsMember ? 'product-price--member' : 'product-price--promotion'"
                  >
                    {{ formatPrice(selectedSkuPromotionFinalIsMember ? selectedSkuPromotionFinalPrice! : selectedSkuPromotionPrice!, siteCurrency) }}
                  </span>
                  <span class="product-price-original">{{ formatPrice(selectedSku.price_amount, siteCurrency) }}</span>
                </div>
                <p
                  class="product-price-note"
                  :class="selectedSkuPromotionFinalIsMember ? 'product-price--member' : 'product-price--promotion'"
                >
                  <template v-if="selectedSkuPromotionFinalIsMember">
                    {{ t('products.memberPriceTag') }} · {{ t('products.saveAmount') }}
                    {{ formatPrice(Number(selectedSku.price_amount) - Number(selectedSkuPromotionFinalPrice), siteCurrency) }}
                  </template>
                  <template v-else>
                    {{ t('products.saveAmount') }} {{ formatPrice(getSkuPromotionSaveAmount(selectedSku), siteCurrency) }}
                  </template>
                </p>
              </template>

              <template v-else-if="selectedSku && hasMemberPrice">
                <div class="product-price-row">
                  <span class="product-price product-price--member">
                    {{ formatPrice(selectedSkuMemberPrice!, siteCurrency) }}
                  </span>
                  <span class="product-price-original">{{ formatPrice(selectedSku.price_amount, siteCurrency) }}</span>
                </div>
                <p class="product-price-note product-price--member">
                  {{ t('products.memberPriceTag') }} · {{ t('products.saveAmount') }}
                  {{ formatPrice(Number(selectedSku.price_amount) - selectedSkuMemberPrice!, siteCurrency) }}
                </p>
              </template>

              <div v-else-if="selectedSku" class="product-price-row">
                <span class="product-price">{{ formatPrice(selectedSku.price_amount, siteCurrency) }}</span>
              </div>

              <template v-else-if="hasPromotionPrice(product)">
                <div class="product-price-row">
                  <span class="product-price product-price--promotion">
                    {{ formatPrice(getPromotionPriceAmount(product), siteCurrency) }}
                  </span>
                  <span class="product-price-original">{{ formatPrice(product.price_amount, siteCurrency) }}</span>
                </div>
                <p class="product-price-note product-price--promotion">
                  {{ t('products.saveAmount') }} {{ formatPrice(getPromotionSaveAmount(product), siteCurrency) }}
                </p>
              </template>

              <div v-else class="product-price-row">
                <span class="product-price">{{ formatPrice(product.price_amount, siteCurrency) }}</span>
              </div>
            </div>

            <!-- Optional pricing rules are retained, but kept visually compact. -->
            <div v-if="selectedSkuWholesaleRules.length" class="product-rule product-rule--wholesale">
              <h2>{{ t('products.wholesaleRulesTitle') }}</h2>
              <div class="flex flex-wrap gap-1.5">
                <span
                  v-for="tier in selectedSkuWholesaleRules"
                  :key="`${tier.sku_id || tier.sku_code || 'all'}-${tier.min_quantity}`"
                  class="product-rule-pill"
                >
                  {{ formatWholesaleTier(tier) }}
                </span>
              </div>
            </div>

            <div v-if="hasPromotionRules(product)" class="product-rule product-rule--promotion">
              <h2><Tag /> {{ t('products.promotionRulesTitle') }}</h2>
              <ul>
                <li v-for="rule in getPromotionRules(product)" :key="rule.id">
                  {{ formatPromotionRule(rule) }}
                </li>
              </ul>
            </div>

            <!-- SKUs use their content width and wrap only when the viewport requires it. -->
            <div v-if="activeSkus.length" class="product-sku-block">
              <h2 class="product-section-label">{{ t('productDetail.skuTitle') }}</h2>
              <div class="product-sku-list">
                <button
                  v-for="sku in activeSkus"
                  :key="sku.id"
                  type="button"
                  class="product-sku"
                  :class="[
                    normalizeSkuId(sku.id) === selectedSkuId ? 'product-sku--selected' : '',
                    !isSkuPurchasable(sku) ? 'product-sku--disabled' : '',
                  ]"
                  :disabled="!isSkuPurchasable(sku)"
                  @click="selectedSkuId = normalizeSkuId(sku.id)"
                >
                  <span class="product-sku-name">{{ skuDisplayText(sku) }}</span>
                  <span class="product-sku-stock" :class="skuStockBadgeClass(sku)">
                    {{ skuStockText(sku) }}
                  </span>
                </button>
              </div>
              <p v-if="requiresSKUSelection" class="mt-2 text-xs font-medium text-warning">
                {{ t('productDetail.skuRequired') }}
              </p>
            </div>

            <div class="product-quantity-row">
              <h2 class="product-section-label">{{ t('productDetail.quantity') }}</h2>
              <div class="product-quantity-control">
                <button
                  type="button"
                  :aria-label="t('productDetail.quantity')"
                  :disabled="quantity <= quantityEffectiveMin"
                  @click="quantity = Math.max(quantityEffectiveMin, quantity - 1)"
                >
                  <Minus />
                </button>
                <input
                  type="text"
                  inputmode="numeric"
                  :aria-label="t('productDetail.quantity')"
                  :value="quantity"
                  @change="handleQuantityInput($event)"
                  @keydown.enter.prevent="($event.target as HTMLInputElement)?.blur()"
                />
                <button
                  type="button"
                  :aria-label="t('productDetail.quantity')"
                  :disabled="quantityEffectiveLimit !== null && quantity >= quantityEffectiveLimit"
                  @click="quantity = quantity + 1"
                >
                  <Plus />
                </button>
              </div>
            </div>

            <div class="product-purchase-messages">
              <Alert v-if="cannotPurchaseReason" variant="destructive">
                <AlertDescription class="font-semibold">{{ cannotPurchaseReason }}</AlertDescription>
              </Alert>
              <Alert v-if="purchaseWarning" class="border-warning/30 bg-warning/10 text-warning">
                <AlertDescription class="font-semibold text-warning">{{ purchaseWarning }}</AlertDescription>
              </Alert>
            </div>

            <!-- Desktop actions. Mobile actions live in the fixed purchase bar. -->
            <div class="product-actions">
              <Button v-if="requiresLogin" class="product-action product-action--login" @click="goLogin">
                {{ t('productDetail.loginToBuy') }}
              </Button>
              <template v-else>
                <Button
                  variant="secondary"
                  class="product-action product-action--cart"
                  :disabled="!canPurchase"
                  @click="addToCart"
                >
                  <ShoppingCart />
                  {{ t('productDetail.addToCart') }}
                </Button>
                <Button class="product-action product-action--buy" :disabled="!canPurchase" @click="buyNow">
                  {{ t('productDetail.buyNow') }}
                </Button>
              </template>
            </div>
          </div>
        </section>

        <!-- Long-form product content deliberately keeps generous space. -->
        <section class="product-details-card">
          <h2 class="product-details-title">
            <span><FileText /></span>
            {{ t('productDetail.details') }}
          </h2>
          <div
            v-if="product.content"
            v-html="processHtmlForDisplay(getLocalizedText(product.content))"
            class="product-rich-text prose prose-gray dark:prose-invert max-w-none theme-prose"
          ></div>
        </section>

        <section v-if="relatedPosts.length" class="related-posts-card">
          <h2 class="product-details-title">
            <span><Newspaper /></span>
            {{ t('productDetail.relatedPosts') }}
          </h2>
          <div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
            <router-link
              v-for="rp in relatedPosts"
              :key="rp.id"
              :to="`/blog/${rp.slug}`"
              class="related-post"
            >
              <img
                v-if="rp.thumbnail"
                :src="getImageUrl(rp.thumbnail)"
                :alt="getLocalizedText(rp.title)"
                loading="lazy"
              />
              <div>
                <h3>{{ getLocalizedText(rp.title) }}</h3>
                <p v-if="rp.summary">{{ getLocalizedText(rp.summary) }}</p>
                <time>{{ formatRelatedPostDate(rp.published_at) }}</time>
              </div>
            </router-link>
          </div>
        </section>

        <ProductMobileBar
          :visible="!!product && !loading"
          :requires-login="requiresLogin"
          :can-purchase="canPurchase"
          @add-to-cart="addToCart"
          @buy-now="buyNow"
          @go-login="goLogin"
        />
      </div>

      <EmptyState v-else size="lg" icon="alert" :title="t('productDetail.notFound')">
        <template #action>
          <Button class="h-10 rounded-full" @click="loadProduct">
            <RotateCw />
            {{ t('errorBoundary.retry') }}
          </Button>
          <Button variant="secondary" as-child class="h-10 rounded-full">
            <router-link to="/products">{{ t('productDetail.backToProducts') }}</router-link>
          </Button>
        </template>
      </EmptyState>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import {
  FileText,
  Lock,
  Minus,
  Newspaper,
  Package,
  Pencil,
  Plus,
  RotateCw,
  ShoppingCart,
  Tag,
  UserPlus,
  Zap,
} from 'lucide-vue-next'
import { getImageUrl } from '../utils/image'
import { processHtmlForDisplay } from '../utils/content'
import { useProductDetail } from '../composables/useProductDetail'
import ProductImageGallery from '../components/product/ProductImageGallery.vue'
import ProductMobileBar from '../components/product/ProductMobileBar.vue'
import BreadcrumbNav from '../components/BreadcrumbNav.vue'
import EmptyState from '../components/EmptyState.vue'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'

const { t } = useI18n()

const {
  getLocalizedText, siteCurrency, formatPrice,
  getPurchaseTypeLabel, getFulfillmentTypeLabel, getStockBadgeVariant, getStockStatusLabel,
  hasPromotionPrice, getPromotionPriceAmount, getPromotionSaveAmount,
  hasSkuPromotionPrice, getSkuPromotionSaveAmount,
  hasPromotionRules, getPromotionRules,
  formatPromotionRule, formatWholesaleTier, formatRelatedPostDate, normalizeSkuId,
  loading, product, relatedPosts, currentImage, selectedSkuId, quantity, purchaseWarning,
  activeSkus, selectedSku,
  selectedSkuMemberPrice, hasMemberPrice,
  hasSelectedSkuWholesalePrice, selectedSkuWholesaleFinalIsMember, selectedSkuWholesaleFinalPrice,
  selectedSkuWholesaleRules,
  selectedSkuPromotionPrice, selectedSkuPromotionFinalIsMember, selectedSkuPromotionFinalPrice,
  showSelectedSkuMemberBadge,
  isSkuPurchasable, skuDisplayText, skuStockText, skuStockBadgeClass,
  quantityEffectiveLimit, quantityEffectiveMin, handleQuantityInput,
  requiresLogin, requiresSKUSelection, canPurchase, cannotPurchaseReason,
  images,
  addToCart, buyNow, goLogin, loadProduct,
} = useProductDetail()
</script>

<style scoped>
.product-detail-page {
  padding: 88px 0 54px;
  background:
    radial-gradient(circle at 78% 12%, color-mix(in srgb, var(--ui-accent) 7%, transparent), transparent 28rem),
    var(--ui-bg-page);
}

.product-detail-shell {
  width: min(1320px, calc(100% - 32px));
  margin: 0 auto;
}

.product-breadcrumb {
  margin-bottom: 20px;
}

.product-hero {
  display: grid;
  grid-template-columns: minmax(0, 650px) minmax(0, 1fr);
  gap: 22px;
  min-height: 560px;
  padding: 18px;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--ui-border) 72%, transparent);
  border-radius: 28px;
  background:
    linear-gradient(125deg, var(--ui-bg-elevated) 0%, var(--ui-bg-elevated) 58%, color-mix(in srgb, var(--ui-accent) 4%, var(--ui-bg-elevated)) 100%);
  box-shadow: 0 24px 58px -34px rgba(22, 42, 73, 0.34), 0 8px 20px -16px rgba(22, 42, 73, 0.2);
}

.product-hero--loading {
  grid-template-columns: minmax(0, 650px) minmax(0, 1fr);
}

.product-buy-panel {
  display: flex;
  min-width: 0;
  flex-direction: column;
  padding: 18px 28px 8px 4px;
}

.product-title {
  margin: 0 0 18px;
  font-size: clamp(2rem, 2.35vw, 2.25rem);
  font-weight: 800;
  letter-spacing: -0.035em;
  color: var(--ui-text-primary);
}

.product-status-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 22px;
}

.product-status {
  display: inline-flex;
  width: max-content;
  height: 48px;
  align-items: center;
  gap: 9px;
  padding: 0 16px;
  border-radius: 13px;
  font-size: 14px;
  font-weight: 700;
  white-space: nowrap;
}

.product-status svg {
  width: 20px;
  height: 20px;
  stroke-width: 1.9;
}

.product-status--guest,
.product-status--auto {
  color: var(--ui-info);
  background: var(--ui-info-soft);
}

.product-status--member,
.product-status--stock-info,
.product-status--stock-success {
  color: var(--ui-success);
  background: var(--ui-success-soft);
}

.product-status--manual,
.product-status--stock-warning {
  color: var(--ui-warning);
  background: var(--ui-warning-soft);
}

.product-status--stock-destructive,
.product-status--stock-danger {
  color: var(--ui-danger);
  background: var(--ui-danger-soft);
}

.product-status--stock-neutral,
.product-status--stock-secondary {
  color: var(--ui-text-muted);
  background: var(--ui-bg-soft);
}

.product-price-block {
  margin-bottom: 20px;
}

.product-price-label-row,
.product-price-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}

.product-price-label-row {
  min-height: 20px;
  gap: 7px;
  margin-bottom: 4px;
}

.product-section-label {
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.01em;
  color: var(--ui-text-secondary);
}

.price-flag {
  display: inline-flex;
  align-items: center;
  min-height: 21px;
  padding: 1px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
}

.price-flag--promotion {
  color: var(--ui-danger);
  background: var(--ui-danger-soft);
}

.price-flag--member {
  color: var(--ui-warning);
  background: var(--ui-warning-soft);
}

.price-flag--wholesale {
  color: var(--ui-success);
  background: var(--ui-success-soft);
}

.product-price-row {
  gap: 12px;
  align-items: baseline;
}

.product-price {
  font-size: clamp(2.15rem, 2.7vw, 2.5rem);
  line-height: 1.14;
  font-weight: 820;
  letter-spacing: -0.025em;
  color: var(--ui-accent);
  font-variant-numeric: tabular-nums;
}

.product-price--promotion {
  color: var(--ui-danger);
}

.product-price--member {
  color: var(--ui-warning);
}

.product-price--wholesale {
  color: var(--ui-success);
}

.product-price-original {
  font-size: 14px;
  font-weight: 600;
  color: var(--ui-text-muted);
  text-decoration: line-through;
}

.product-price-note {
  margin-top: 3px;
  font-size: 12px;
  font-weight: 650;
}

.product-rule {
  width: fit-content;
  max-width: 100%;
  margin-bottom: 12px;
  padding: 10px 12px;
  border-radius: 12px;
  font-size: 12px;
}

.product-rule h2 {
  display: flex;
  align-items: center;
  gap: 5px;
  margin-bottom: 5px;
  font-size: 12px;
  font-weight: 750;
}

.product-rule h2 svg {
  width: 14px;
  height: 14px;
}

.product-rule ul {
  display: grid;
  gap: 2px;
}

.product-rule--wholesale {
  color: var(--ui-success);
  background: var(--ui-success-soft);
}

.product-rule--promotion {
  color: var(--ui-warning);
  background: var(--ui-warning-soft);
}

.product-rule-pill {
  display: inline-flex;
  padding: 2px 7px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--ui-bg-elevated) 72%, transparent);
  font-weight: 650;
}

.product-sku-block {
  margin-bottom: 18px;
}

.product-sku-block > .product-section-label {
  display: block;
  margin-bottom: 9px;
}

.product-sku-list {
  display: flex;
  flex-wrap: wrap;
  align-items: stretch;
  gap: 10px;
}

.product-sku {
  display: inline-flex;
  width: max-content;
  min-width: 112px;
  min-height: 62px;
  max-width: 100%;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  gap: 4px;
  padding: 9px 18px;
  border: 1px solid transparent;
  border-radius: 12px;
  background: var(--ui-bg-soft);
  color: var(--ui-text-primary);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--ui-border) 52%, transparent), 0 5px 14px -12px rgba(22, 42, 73, 0.35);
  transition: transform var(--ui-duration-fast) var(--ui-ease-out), box-shadow var(--ui-duration-fast) var(--ui-ease-out), background var(--ui-duration-fast) var(--ui-ease-out);
}

.product-sku:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: inset 0 0 0 1px var(--ui-border-strong), 0 10px 20px -16px rgba(22, 42, 73, 0.4);
}

.product-sku--selected {
  border-color: color-mix(in srgb, var(--ui-accent) 60%, transparent);
  background: var(--ui-accent-soft);
  color: var(--ui-accent);
  box-shadow: 0 10px 22px -16px color-mix(in srgb, var(--ui-accent) 50%, transparent);
}

.product-sku--disabled {
  cursor: not-allowed;
  opacity: 0.48;
}

.product-sku-name {
  max-width: 220px;
  overflow: hidden;
  font-size: 14px;
  font-weight: 750;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-sku-stock {
  margin: 0;
  padding: 0;
  border: 0;
  background: transparent;
  font-size: 12px;
  font-weight: 650;
  line-height: 1.2;
}

.product-quantity-row {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 9px;
  margin-bottom: 18px;
}

.product-quantity-control {
  display: inline-flex;
  width: 128px;
  height: 44px;
  overflow: hidden;
  border-radius: 11px;
  background: var(--ui-bg-soft);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--ui-border) 62%, transparent);
}

.product-quantity-control button {
  display: grid;
  width: 40px;
  flex: none;
  place-items: center;
  color: var(--ui-text-muted);
  transition: color var(--ui-duration-fast), background var(--ui-duration-fast);
}

.product-quantity-control button:hover:not(:disabled) {
  color: var(--ui-text-primary);
  background: var(--ui-bg-muted);
}

.product-quantity-control button:disabled {
  opacity: 0.35;
}

.product-quantity-control button svg {
  width: 16px;
  height: 16px;
}

.product-quantity-control input {
  width: 48px;
  min-width: 0;
  border: 0;
  border-inline: 1px solid var(--ui-border);
  outline: 0;
  background: transparent;
  text-align: center;
  font-size: 14px;
  font-weight: 750;
  color: var(--ui-text-primary);
  font-variant-numeric: tabular-nums;
}

.product-purchase-messages {
  display: grid;
  gap: 8px;
  margin-bottom: 12px;
}

.product-purchase-messages:empty {
  display: none;
}

.product-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: auto;
}

.product-action {
  width: 228px;
  height: 52px;
  border-radius: 13px;
  font-size: 15px;
  font-weight: 750;
  box-shadow: none;
}

.product-action--cart {
  color: var(--ui-accent);
  background: var(--ui-accent-soft);
}

.product-action--cart:hover {
  background: color-mix(in srgb, var(--ui-accent) 20%, var(--ui-bg-elevated));
}

.product-action--buy,
.product-action--login {
  background: linear-gradient(135deg, var(--ui-accent), color-mix(in srgb, var(--ui-accent) 78%, #5a73ff));
  box-shadow: 0 14px 26px -18px color-mix(in srgb, var(--ui-accent) 72%, transparent);
}

.product-details-card,
.related-posts-card {
  margin-top: 30px;
  padding: 28px 32px;
  border: 1px solid color-mix(in srgb, var(--ui-border) 72%, transparent);
  border-radius: 24px;
  background: var(--ui-bg-elevated);
  box-shadow: 0 18px 40px -32px rgba(22, 42, 73, 0.34);
}

.product-details-card {
  min-height: 420px;
}

.product-details-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0 0 24px;
  font-size: 20px;
  font-weight: 800;
  letter-spacing: -0.02em;
}

.product-details-title > span {
  display: grid;
  width: 28px;
  height: 28px;
  flex: none;
  place-items: center;
  border-radius: 8px;
  background: var(--ui-accent-soft);
  color: var(--ui-accent);
}

.product-details-title svg {
  width: 16px;
  height: 16px;
}

.product-rich-text {
  padding-top: 20px;
  border-top: 1px solid color-mix(in srgb, var(--ui-border) 62%, transparent);
  color: var(--ui-text-secondary);
  line-height: 1.75;
}

.related-post {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 14px;
  min-height: 118px;
  padding: 14px;
  border-radius: 16px;
  background: var(--ui-bg-soft);
  color: inherit;
  transition: transform var(--ui-duration-normal) var(--ui-ease-out), box-shadow var(--ui-duration-normal) var(--ui-ease-out);
}

.related-post:hover {
  transform: translateY(-2px);
  box-shadow: var(--ui-shadow-soft);
}

.related-post img {
  width: 96px;
  height: 90px;
  border-radius: 11px;
  object-fit: cover;
}

.related-post h3 {
  display: -webkit-box;
  overflow: hidden;
  font-size: 15px;
  font-weight: 750;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.related-post p {
  display: -webkit-box;
  overflow: hidden;
  margin-top: 5px;
  font-size: 12px;
  color: var(--ui-text-muted);
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.related-post time {
  display: block;
  margin-top: 7px;
  font-size: 11px;
  color: var(--ui-text-muted);
}

@media (max-width: 1100px) {
  .product-hero {
    grid-template-columns: minmax(0, 1.03fr) minmax(0, 0.97fr);
    gap: 16px;
  }

  .product-buy-panel {
    padding-right: 12px;
  }

  .product-status-row {
    gap: 8px;
  }

  .product-status {
    padding-inline: 12px;
  }
}

@media (max-width: 1023px) {
  .product-detail-page {
    padding-top: 78px;
    padding-bottom: 92px;
  }

  .product-detail-shell {
    width: min(760px, calc(100% - 32px));
  }

  .product-breadcrumb {
    margin-bottom: 14px;
  }

  .product-hero,
  .product-hero--loading {
    display: block;
    min-height: 0;
    padding: 0;
    overflow: visible;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
  }

  .product-buy-panel {
    margin-top: 12px;
    padding: 22px;
    border: 1px solid color-mix(in srgb, var(--ui-border) 68%, transparent);
    border-radius: 20px;
    background: var(--ui-bg-elevated);
    box-shadow: 0 18px 38px -30px rgba(22, 42, 73, 0.38);
  }

  .product-actions {
    display: none;
  }

  .product-details-card,
  .related-posts-card {
    margin-top: 18px;
  }
}

@media (max-width: 639px) {
  .product-detail-page {
    padding-top: 74px;
  }

  .product-detail-shell {
    width: calc(100% - 32px);
  }

  .product-buy-panel {
    min-height: 374px;
    padding: 18px;
    border-radius: 18px;
  }

  .product-title {
    margin-bottom: 14px;
    font-size: 26px;
  }

  .product-status-row {
    flex-wrap: nowrap;
    gap: 7px;
    margin-bottom: 17px;
  }

  .product-status {
    min-width: 0;
    height: 34px;
    gap: 6px;
    padding: 0 9px;
    border-radius: 10px;
    font-size: 11px;
  }

  .product-status svg {
    width: 14px;
    height: 14px;
  }

  .product-price-block {
    margin-bottom: 16px;
  }

  .product-section-label {
    font-size: 12px;
  }

  .product-price {
    font-size: 30px;
  }

  .product-sku-block {
    margin-bottom: 16px;
  }

  .product-sku-list {
    flex-wrap: wrap;
    gap: 7px;
  }

  .product-sku {
    min-width: 96px;
    min-height: 58px;
    padding: 8px 11px;
    border-radius: 10px;
  }

  .product-sku-name {
    max-width: 128px;
    font-size: 12px;
  }

  .product-sku-stock {
    font-size: 10px;
  }

  .product-quantity-row {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 0;
  }

  .product-quantity-control {
    width: 112px;
    height: 40px;
  }

  .product-quantity-control button {
    width: 36px;
  }

  .product-quantity-control input {
    width: 40px;
  }

  .product-purchase-messages {
    margin-top: 12px;
    margin-bottom: 0;
  }

  .product-details-card,
  .related-posts-card {
    padding: 22px 20px;
    border-radius: 18px;
  }

  .product-details-card {
    min-height: 370px;
  }

  .product-details-title {
    margin-bottom: 18px;
    font-size: 18px;
  }

  .product-rich-text {
    padding-top: 16px;
    font-size: 14px;
  }

  .related-post {
    grid-template-columns: 76px minmax(0, 1fr);
  }

  .related-post img {
    width: 76px;
    height: 76px;
  }
}
</style>

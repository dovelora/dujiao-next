<template>
  <Transition
    enter-active-class="transition duration-300 ease-out"
    enter-from-class="translate-y-full opacity-0"
    enter-to-class="translate-y-0 opacity-100"
    leave-active-class="transition duration-200 ease-in"
    leave-from-class="translate-y-0 opacity-100"
    leave-to-class="translate-y-full opacity-0">
    <div v-if="visible"
      class="product-mobile-bar lg:hidden fixed bottom-0 left-0 right-0 z-40 bg-card/95 backdrop-blur-xl border-t shadow-2xl theme-safe-bottom">
      <div class="grid items-center gap-2 px-4 py-3 grid-cols-[minmax(0,1fr)_auto]">
        <!-- Price -->
        <div class="flex-1 min-w-0">
          <span v-if="showMemberPrice" class="theme-price-sm text-amber-600 dark:text-amber-300 break-all block">
            {{ memberPriceDisplay }}
          </span>
          <span v-else-if="showSkuPromotionPrice" class="theme-price-sm text-rose-600 dark:text-rose-300 break-all block">
            {{ skuPromotionPriceDisplay }}
          </span>
          <span v-else-if="showSkuPrice" class="theme-price-sm text-primary break-all block">
            {{ skuPriceDisplay }}
          </span>
          <span v-else-if="showProductPromotionPrice" class="theme-price-sm text-rose-600 dark:text-rose-300 break-all block">
            {{ productPromotionPriceDisplay }}
          </span>
          <span v-else class="theme-price-sm text-primary break-all block">
            {{ productPriceDisplay }}
          </span>
        </div>
        <!-- Actions -->
        <Button size="lg" class="h-11 px-4 font-semibold" :disabled="!canPurchase" @click="$emit('buyNow')">
          {{ t('productDetail.buyNow') }}
        </Button>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'

const { t } = useI18n()

defineProps<{
  visible: boolean
  canPurchase: boolean
  showMemberPrice: boolean
  memberPriceDisplay: string
  showSkuPromotionPrice: boolean
  skuPromotionPriceDisplay: string
  showSkuPrice: boolean
  skuPriceDisplay: string
  showProductPromotionPrice: boolean
  productPromotionPriceDisplay: string
  productPriceDisplay: string
}>()

defineEmits<{
  buyNow: []
}>()
</script>

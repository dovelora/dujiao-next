<template>
  <Transition
    enter-active-class="transition duration-300 ease-out"
    enter-from-class="translate-y-full opacity-0"
    enter-to-class="translate-y-0 opacity-100"
    leave-active-class="transition duration-200 ease-in"
    leave-from-class="translate-y-0 opacity-100"
    leave-to-class="translate-y-full opacity-0"
  >
    <div
      v-if="visible"
      class="product-mobile-bar theme-safe-bottom fixed bottom-0 left-0 right-0 z-40 border-t bg-card/95 shadow-[var(--shadow)] backdrop-blur-xl lg:hidden"
    >
      <div class="grid items-center gap-2 px-4 py-3 grid-cols-[minmax(0,1fr)_auto]">
        <!-- 价格 -->
        <div class="min-w-0 flex-1">
          <span v-if="showMemberPrice" class="block break-all text-lg font-semibold tabular-nums text-[color:var(--gold-strong)]">{{ memberPriceDisplay }}</span>
          <span v-else-if="showSkuPromotionPrice" class="block break-all text-lg font-semibold tabular-nums text-primary">{{ skuPromotionPriceDisplay }}</span>
          <span v-else-if="showSkuPrice" class="block break-all text-lg font-semibold tabular-nums text-primary">{{ skuPriceDisplay }}</span>
          <span v-else-if="showProductPromotionPrice" class="block break-all text-lg font-semibold tabular-nums text-primary">{{ productPromotionPriceDisplay }}</span>
          <span v-else class="block break-all text-lg font-semibold tabular-nums text-primary">{{ productPriceDisplay }}</span>
        </div>
        <!-- 操作 -->
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

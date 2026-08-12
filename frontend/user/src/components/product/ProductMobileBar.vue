<template>
  <Transition
    enter-active-class="transition duration-300 ease-out"
    enter-from-class="translate-y-full opacity-0"
    enter-to-class="translate-y-0 opacity-100"
    leave-active-class="transition duration-200 ease-in"
    leave-from-class="translate-y-0 opacity-100"
    leave-to-class="translate-y-full opacity-0"
  >
    <div v-if="visible" class="product-mobile-bar theme-safe-bottom">
      <div class="product-mobile-actions">
        <Button v-if="requiresLogin" class="product-mobile-button product-mobile-button--buy" @click="$emit('goLogin')">
          {{ t('productDetail.loginToBuy') }}
        </Button>
        <template v-else>
          <Button
            variant="secondary"
            class="product-mobile-button product-mobile-button--cart"
            :disabled="!canPurchase"
            @click="$emit('addToCart')"
          >
            <ShoppingCart />
            {{ t('productDetail.addToCart') }}
          </Button>
          <Button
            class="product-mobile-button product-mobile-button--buy"
            :disabled="!canPurchase"
            @click="$emit('buyNow')"
          >
            {{ t('productDetail.buyNow') }}
          </Button>
        </template>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ShoppingCart } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const { t } = useI18n()

defineProps<{
  visible: boolean
  requiresLogin: boolean
  canPurchase: boolean
}>()

defineEmits<{
  addToCart: []
  buyNow: []
  goLogin: []
}>()
</script>

<style scoped>
.product-mobile-bar {
  position: fixed;
  z-index: 50;
  right: 0;
  bottom: 0;
  left: 0;
  display: none;
  background: color-mix(in srgb, var(--ui-bg-elevated) 92%, transparent);
  box-shadow: 0 -14px 34px -24px rgba(22, 42, 73, 0.46);
  backdrop-filter: blur(18px);
}

.product-mobile-actions {
  display: flex;
  gap: 8px;
  width: min(390px, 100%);
  min-height: 72px;
  margin: 0 auto;
  padding: 10px 16px 12px;
}

.product-mobile-button {
  height: 50px;
  flex: 1;
  border-radius: 14px;
  font-size: 14px;
  font-weight: 750;
  box-shadow: none;
}

.product-mobile-button--cart {
  color: var(--ui-accent);
  background: var(--ui-accent-soft);
}

.product-mobile-button--buy {
  background: linear-gradient(135deg, var(--ui-accent), color-mix(in srgb, var(--ui-accent) 78%, #5a73ff));
  box-shadow: 0 14px 24px -18px color-mix(in srgb, var(--ui-accent) 72%, transparent);
}

@media (max-width: 1023px) {
  .product-mobile-bar {
    display: block;
  }
}
</style>

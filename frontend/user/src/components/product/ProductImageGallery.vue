<template>
  <div class="product-media">
    <div
      class="product-media-frame"
      @touchstart="onImageTouchStart"
      @touchend="onImageTouchEnd"
    >
      <img
        v-if="currentImage"
        :src="currentImage"
        :alt="productTitle"
        class="product-media-image"
      />
      <div v-else class="product-media-empty">
        <ImageIcon />
      </div>

      <div v-if="images.length > 1" class="product-media-thumbnails" aria-label="Product images">
        <button
          v-for="(image, index) in images"
          :key="index"
          type="button"
          :aria-label="`Image ${index + 1}`"
          :aria-current="currentImage === image ? 'true' : undefined"
          :class="currentImage === image ? 'is-active' : ''"
          @click="$emit('update:currentImage', image)"
        >
          <img :src="image" :alt="`Image ${index + 1}`" loading="lazy" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Image as ImageIcon } from 'lucide-vue-next'

const props = defineProps<{
  images: string[]
  currentImage: string
  productTitle: string
}>()

const emit = defineEmits<{
  'update:currentImage': [value: string]
}>()

let touchStartX = 0

const onImageTouchStart = (event: TouchEvent) => {
  touchStartX = event.touches[0]?.clientX ?? 0
}

const onImageTouchEnd = (event: TouchEvent) => {
  const touchEndX = event.changedTouches[0]?.clientX ?? 0
  const difference = touchStartX - touchEndX
  if (Math.abs(difference) < 50 || props.images.length <= 1) return

  const currentIndex = props.images.indexOf(props.currentImage)
  if (currentIndex === -1) return

  const nextIndex = difference > 0
    ? (currentIndex + 1) % props.images.length
    : (currentIndex - 1 + props.images.length) % props.images.length

  emit('update:currentImage', props.images[nextIndex] ?? '')
}
</script>

<style scoped>
.product-media {
  min-width: 0;
  height: 520px;
}

.product-media-frame {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
  border-radius: 18px;
  background: var(--ui-bg-soft);
  box-shadow: 0 22px 42px -28px rgba(6, 18, 45, 0.55);
}

.product-media-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.product-media-empty {
  display: grid;
  width: 100%;
  height: 100%;
  place-items: center;
  color: var(--ui-text-muted);
  background:
    radial-gradient(circle at 50% 40%, var(--ui-accent-soft), transparent 46%),
    var(--ui-bg-soft);
}

.product-media-empty svg {
  width: 88px;
  height: 88px;
  stroke-width: 1.25;
}

.product-media-thumbnails {
  position: absolute;
  right: 14px;
  bottom: 14px;
  left: 14px;
  display: flex;
  justify-content: center;
  gap: 8px;
  overflow-x: auto;
  padding: 8px;
  border-radius: 14px;
  background: color-mix(in srgb, #05070f 66%, transparent);
  backdrop-filter: blur(14px);
  scrollbar-width: none;
}

.product-media-thumbnails::-webkit-scrollbar {
  display: none;
}

.product-media-thumbnails button {
  width: 48px;
  height: 48px;
  flex: none;
  overflow: hidden;
  border: 2px solid transparent;
  border-radius: 9px;
  opacity: 0.62;
  transition: opacity var(--ui-duration-fast), border-color var(--ui-duration-fast), transform var(--ui-duration-fast);
}

.product-media-thumbnails button:hover,
.product-media-thumbnails button.is-active {
  border-color: #ffffff;
  opacity: 1;
  transform: translateY(-1px);
}

.product-media-thumbnails img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

@media (max-width: 1023px) {
  .product-media {
    height: auto;
    aspect-ratio: 4 / 3;
  }

  .product-media-frame {
    border-radius: 18px;
  }
}

@media (max-width: 639px) {
  .product-media-frame {
    border-radius: 16px;
    box-shadow: 0 16px 30px -24px rgba(6, 18, 45, 0.52);
  }

  .product-media-thumbnails {
    right: 10px;
    bottom: 10px;
    left: 10px;
    justify-content: flex-start;
    padding: 6px;
  }

  .product-media-thumbnails button {
    width: 40px;
    height: 40px;
    border-radius: 7px;
  }
}
</style>

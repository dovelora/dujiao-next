<template>
  <Card class="group relative overflow-hidden transition-shadow hover:shadow-md">
    <div class="flex flex-col items-start justify-between gap-2 p-3 sm:flex-row sm:gap-3 sm:p-5">
      <div class="min-w-0 w-full sm:w-auto sm:flex-1">
        <p class="break-words text-xs font-semibold leading-5 sm:truncate sm:text-[11px] sm:uppercase sm:tracking-[0.16em] text-muted-foreground">{{ label }}</p>
        <div class="mt-2 break-all text-xl font-bold sm:truncate sm:font-black leading-tight text-foreground sm:text-2xl" :class="{ 'font-mono': mono }">
          <slot name="value">{{ value }}</slot>
        </div>
        <p v-if="hint" class="mt-1 break-words text-xs sm:truncate text-muted-foreground">{{ hint }}</p>
      </div>
      <span
        v-if="icon"
        class="order-first flex h-8 w-8 shrink-0 sm:order-last sm:h-10 sm:w-10 items-center justify-center rounded-xl transition-transform group-hover:scale-105"
        :class="toneClass"
      >
        <component :is="icon" class="h-5 w-5" />
      </span>
    </div>
  </Card>
</template>

<script setup lang="ts">
import { computed, type Component } from 'vue'
import { Card } from '@/components/ui/card'

const props = withDefaults(
  defineProps<{
    label: string
    value?: string | number
    hint?: string
    icon?: Component
    tone?: 'accent' | 'success' | 'warning' | 'info' | 'neutral'
    mono?: boolean
  }>(),
  { tone: 'accent', mono: false },
)

const toneClass = computed(
  () =>
    ({
      success: 'bg-success/10 text-success',
      warning: 'bg-warning/10 text-warning',
      info: 'bg-info/10 text-info',
      neutral: 'bg-muted text-muted-foreground',
      accent: 'bg-primary/10 text-primary',
    })[props.tone],
)
</script>

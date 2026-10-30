<template>
  <aside class="catalog-categories min-w-0">
    <div class="relative lg:hidden">
      <label :for="mobileSelectId" class="sr-only">{{ t('products.categories') }}</label>
      <select
        :id="mobileSelectId"
        :value="selectedCategory ?? ''"
        class="h-11 w-full min-w-0 appearance-none rounded-xl border bg-card pl-3.5 pr-10 text-base text-foreground"
        @change="selectMobileCategory"
      >
        <option value="">{{ t('products.allCategories') }}</option>
        <template v-for="grp in categoryGroups" :key="grp.id">
          <option :value="grp.id">{{ catName(grp) }}</option>
          <optgroup v-if="grp.children.length" :label="catName(grp)">
            <option v-for="child in grp.children" :key="child.id" :value="child.id">{{ catName(child) }}</option>
          </optgroup>
        </template>
      </select>
      <ChevronDown class="pointer-events-none absolute right-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
    </div>
    <div class="hidden rounded-xl border bg-card p-4 lg:block">
      <div class="mb-3 flex items-center gap-2 px-1">
        <span class="h-4 w-1 flex-none rounded-full bg-primary"></span>
        <h4 class="text-sm font-bold">{{ t('products.categories') }}</h4>
      </div>
      <!-- 桌面分类列表 -->
      <div class="grid gap-1">
        <button
          type="button"
          class="w-full rounded-lg px-3 py-2.5 text-left text-sm font-semibold transition-colors"
          :class="selectedCategory === null
            ? 'bg-primary/10 text-primary'
            : 'text-muted-foreground hover:bg-secondary hover:text-foreground'"
          @click="$emit('select', null)"
        >
          {{ t('products.allCategories') }}
        </button>
        <template v-for="grp in categoryGroups" :key="grp.id">
          <div class="flex w-full min-w-0 items-center gap-0.5">
            <button
              type="button"
              class="flex min-w-0 flex-1 items-center gap-2.5 rounded-lg px-3 py-2.5 text-left text-sm font-semibold transition-colors"
              :class="selectedCategory === grp.id
                ? 'bg-primary/10 text-primary'
                : 'text-muted-foreground hover:bg-secondary hover:text-foreground'"
              @click="$emit('select', grp.id)"
            >
              <img v-if="grp.icon" :src="getImageUrl(grp.icon)" :alt="catName(grp)" loading="lazy" class="h-5 w-5 flex-none rounded-md object-cover" />
              <span class="min-w-0 truncate">{{ catName(grp) }}</span>
            </button>
            <button
              v-if="grp.children.length"
              type="button"
              class="grid h-9 w-9 flex-none place-items-center rounded-lg transition-colors hover:bg-secondary hover:text-foreground"
              :class="expandedParentIds.includes(grp.id) ? 'text-primary' : 'text-muted-foreground'"
              :aria-expanded="expandedParentIds.includes(grp.id)"
              :aria-label="catName(grp)"
              @click="$emit('toggle', grp.id)"
            >
              <ChevronDown class="h-4 w-4 transition-transform" :class="{ 'rotate-180': expandedParentIds.includes(grp.id) }" />
            </button>
          </div>
          <template v-if="grp.children.length && expandedParentIds.includes(grp.id)">
            <button
              v-for="child in grp.children"
              :key="child.id"
              type="button"
              class="block w-full min-w-0 truncate rounded-lg py-2 pl-9 pr-3 text-left text-[13.5px] font-semibold transition-colors"
              :class="selectedCategory === child.id ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-secondary hover:text-foreground'"
              @click="$emit('select', child.id)"
            >
              {{ catName(child) }}
            </button>
          </template>
        </template>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useId } from 'vue'
import { ChevronDown } from 'lucide-vue-next'
import { getImageUrl } from '../../../utils/image'
import { useLocalized } from '../../../composables/useProduct'
import type { PublicCategory } from '../../../utils/category'

defineProps<{
  categoryGroups: (PublicCategory & { children: PublicCategory[] })[]
  selectedCategory: number | null
  expandedParentIds: number[]
}>()

const emit = defineEmits<{ select: [id: number | null]; toggle: [id: number] }>()

const { t } = useI18n()
const mobileSelectId = useId()
const selectMobileCategory = (event: Event) => {
  const value = (event.target as HTMLSelectElement).value
  emit('select', value === '' ? null : Number(value))
}
const { getLocalizedText } = useLocalized()
const catName = (cat: PublicCategory) => getLocalizedText(cat.name) || cat.slug || ''
</script>

<template>
  <div class="wallet-recharge rounded-xl border bg-card p-4 sm:p-6">
    <header class="mb-4">
      <h3 class="text-base font-semibold">{{ t('personalCenter.wallet.rechargeTitle') }}</h3>
      <p class="mt-1 text-xs text-muted-foreground">{{ t('personalCenter.wallet.rechargeSubtitle') }}</p>
    </header>
    <form @submit.prevent="$emit('submit')">
      <div class="grid grid-cols-2 items-end gap-3 lg:grid-cols-[1fr_1fr_2fr]">
        <div class="min-w-0">
          <Label :for="`${formId}-amount`" class="mb-1 block text-xs">{{ t('personalCenter.wallet.amountLabel') }}</Label>
          <Input
            :id="`${formId}-amount`"
            :model-value="amount"
            @update:model-value="(v) => $emit('update:amount', String(v).trim())"
            type="text"
            inputmode="decimal"
            :placeholder="t('personalCenter.wallet.amountPlaceholder')"
            class="h-11"
          />
        </div>
        <div class="min-w-0">
          <Label :for="`${formId}-channel`" class="mb-1 block text-xs">{{ t('personalCenter.wallet.channelLabel') }}</Label>
          <Select
            :model-value="String(channelId)"
            @update:model-value="(v) => $emit('update:channelId', Number(v))"
            :disabled="!hasChannels || channelLoading || recharging"
          >
            <SelectTrigger :id="`${formId}-channel`" class="h-11 w-full">
              <SelectValue :placeholder="t('personalCenter.wallet.channelPlaceholder')" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="0">{{ t('personalCenter.wallet.channelPlaceholder') }}</SelectItem>
              <SelectItem v-for="channel in channels" :key="channel.id" :value="String(channel.id)">
                {{ channel.name }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div class="col-span-2 min-w-0 lg:col-span-1">
          <Label :for="`${formId}-remark`" class="mb-1 block text-xs">{{ t('personalCenter.wallet.remarkLabel') }}</Label>
          <Input
            :id="`${formId}-remark`"
            :model-value="remark"
            @update:model-value="(v) => $emit('update:remark', String(v).trim())"
            type="text"
            :placeholder="t('personalCenter.wallet.remarkPlaceholder')"
            class="h-11"
          />
        </div>
      </div>
      <div v-if="selectedChannel?.fee_policy === 'customer_surcharge'" class="wallet-fees mt-4 grid grid-cols-3 gap-3 rounded-lg border border-warning/40 p-3 text-sm">
        <div class="row-span-2 grid min-w-0 grid-rows-subgrid gap-1">
          <div class="text-xs text-warning">{{ t('payment.feeRateLabel') }}</div>
          <div class="break-words text-sm font-semibold text-foreground">{{ feeRateDisplay }}</div>
        </div>
        <div class="row-span-2 grid min-w-0 grid-rows-subgrid gap-1">
          <div class="text-xs text-warning">{{ t('payment.fixedFeeLabel') }}</div>
          <div class="break-words text-sm font-semibold text-foreground">{{ fixedFeeDisplay }}</div>
        </div>
        <div class="row-span-2 grid min-w-0 grid-rows-subgrid gap-1">
          <div class="text-xs text-warning">{{ t('payment.feeAmountLabel') }}</div>
          <div class="break-words text-sm font-semibold text-foreground">{{ feeAmountDisplay }}</div>
        </div>
      </div>
      <p v-if="channelLoading" class="mt-3 text-xs text-muted-foreground">
        {{ t('common.loading') }}
      </p>
      <p v-else-if="!hasChannels" class="mt-3 text-xs text-amber-600">
        {{ t('payment.channelEmpty') }}
      </p>
      <Button
        type="submit"
        :disabled="recharging || channelLoading || !hasChannels"
        class="mt-4 h-11 w-full font-semibold"
      >
        {{ recharging ? t('personalCenter.wallet.recharging') : t('personalCenter.wallet.rechargeSubmit') }}
      </Button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useId } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

defineProps<{
  amount: string
  channelId: number
  remark: string
  channels: Array<{ id: number; name: string }>
  hasChannels: boolean
  channelLoading: boolean
  recharging: boolean
  selectedChannel: { id: number; name: string; fee_policy?: string } | null
  feeRateDisplay: string
  fixedFeeDisplay: string
  feeAmountDisplay: string
}>()

defineEmits<{
  (e: 'submit'): void
  (e: 'update:amount', value: string): void
  (e: 'update:channelId', value: number): void
  (e: 'update:remark', value: string): void
}>()

const { t } = useI18n()
const formId = useId()
</script>

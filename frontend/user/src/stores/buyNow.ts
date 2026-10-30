import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'

export interface PurchaseItem {
    productId: number
    skuId?: number
    skuCode?: string
    skuSpecValues?: Record<string, any>
    skuManualStockTotal?: number
    skuManualStockLocked?: number
    skuManualStockSold?: number
    skuAutoStockAvailable?: number
    skuUpstreamStock?: number
    skuStockStatus?: string
    skuStockDisplayMode?: string
    skuStockDisplay?: string
    skuStockRangeMin?: number
    skuStockRangeMax?: number
    skuStockQuantityHidden?: boolean
    skuStockEnforced?: boolean
    skuStockSnapshotAt?: string
    slug: string
    title: any
    priceAmount: string
    wholesalePrices?: Array<{ sku_id?: number; sku_code?: string; min_quantity: number; unit_price: string | number }>
    image?: string
    quantity: number
    minPurchaseQuantity?: number
    maxPurchaseQuantity?: number
    fulfillmentType?: string
    manualFormSchema?: any
    paymentChannelIds?: number[]
}

const storageKey = 'checkout_item'

const readSelection = (): PurchaseItem | null => {
    try {
        const raw = sessionStorage.getItem(storageKey)
        if (!raw) return null
        const value = JSON.parse(raw)
        if (!value || Array.isArray(value) || typeof value !== 'object') return null
        if (!Number.isSafeInteger(value.productId) || value.productId <= 0) return null
        if (!Number.isSafeInteger(value.quantity) || value.quantity <= 0) return null
        if (typeof value.slug !== 'string' || !value.slug.trim()) return null
        if (typeof value.priceAmount !== 'string') return null
        return value as PurchaseItem
    } catch {
        return null
    }
}

export const useBuyNowStore = defineStore('buyNow', () => {
    const item = ref<PurchaseItem | null>(readSelection())

    watch(item, value => {
        try {
            if (value) sessionStorage.setItem(storageKey, JSON.stringify(value))
            else sessionStorage.removeItem(storageKey)
        } catch {
            // 浏览器禁用存储时，本次页面内仍可完成购买。
        }
    }, { deep: true, flush: 'sync' })

    const hasItem = computed(() => item.value !== null)

    const setItem = (newItem: PurchaseItem) => {
        item.value = { ...newItem }
    }

    const clear = () => {
        item.value = null
    }

    return {
        item,
        hasItem,
        setItem,
        clear,
    }
})

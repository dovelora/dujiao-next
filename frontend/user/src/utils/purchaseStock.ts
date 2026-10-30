import { productAPI } from '../api'
import type { PurchaseItem } from '../stores/buyNow'
import { resolvePurchasePricingSnapshot } from './purchasePricingSnapshot'
import { normalizeSkuId } from './sku'

const normalizeStockNumber = (value: unknown) => {
  const numberValue = Number(value)
  if (!Number.isFinite(numberValue)) return 0
  return Math.max(Math.floor(numberValue), 0)
}

const normalizeManualStockTotal = (value: unknown) => {
  const numberValue = Number(value)
  if (!Number.isFinite(numberValue)) return 0
  const integerValue = Math.floor(numberValue)
  if (integerValue === -1) return -1
  return Math.max(integerValue, 0)
}

const normalizeOptionalLimitNumber = (value: unknown) => {
  const numberValue = Number(value)
  if (!Number.isFinite(numberValue)) return undefined
  const integerValue = Math.floor(numberValue)
  if (integerValue <= 0) return undefined
  return integerValue
}

// purchaseLimit 返回所选商品的单次最大购买上限（无限制时返回 null）。
export const purchaseLimit = (item: Pick<PurchaseItem, 'maxPurchaseQuantity'>) =>
  normalizeOptionalLimitNumber(item.maxPurchaseQuantity) ?? null

// purchaseMinimum 返回所选商品的单次最小购买下限（默认 1）。
export const purchaseMinimum = (item: Pick<PurchaseItem, 'minPurchaseQuantity'>) => {
  const minimum = normalizeOptionalLimitNumber(item.minPurchaseQuantity)
  return minimum && minimum > 0 ? minimum : 1
}

export const resolveActiveSkus = (product: any) => {
  const rows = Array.isArray(product?.skus) ? product.skus : []
  return rows.filter((sku: any) => Boolean(sku?.is_active))
}

const resolveMatchedSku = (item: PurchaseItem, activeSkus: any[]) => {
  const skuId = normalizeSkuId(item.skuId)
  if (skuId > 0) {
    const matchedById = activeSkus.find((sku: any) => normalizeSkuId(sku?.id) === skuId)
    return matchedById || null
  }

  if (skuId === 0 && activeSkus.length === 1) {
    return activeSkus[0]
  }

  return null
}

const shouldEnforceSkuStock = (product: any, sku: any) => {
  if (!product || !sku) return false
  if (sku?.stock_quantity_hidden === true || product?.stock_quantity_hidden === true) return false
  const type = String(product?.fulfillment_type || '').trim()
  if (type === 'auto') return true
  if (type === 'upstream') return true
  if (type !== 'manual') return false
  const total = normalizeManualStockTotal(sku?.manual_stock_total)
  if (total === -1) return false
  return true
}

export const refreshPurchaseSnapshot = async (
  item: PurchaseItem,
  onProduct?: (product: any) => void,
): Promise<Partial<PurchaseItem>> => {
  try {
    const response = await productAPI.detail(item.slug)
    const product = response.data?.data
    if (!product || Number(product.id) !== item.productId) return {}
    onProduct?.(product)
    const activeSkus = resolveActiveSkus(product)
    if (activeSkus.length === 0) return {}

    const matchedSku = resolveMatchedSku(item, activeSkus)
    if (!matchedSku) return {}

    return buildPurchaseSkuSnapshot(item, product, matchedSku)

  } catch {
    return {}
  }
}

export const buildPurchaseSkuSnapshot = (item: PurchaseItem, product: any, matchedSku: any): Partial<PurchaseItem> => {
    const manualStockTotal = normalizeManualStockTotal(matchedSku?.manual_stock_total)
    const manualStockLocked = normalizeStockNumber(matchedSku?.manual_stock_locked)
    const manualStockSold = normalizeStockNumber(matchedSku?.manual_stock_sold)
    const autoStockAvailable = normalizeStockNumber(matchedSku?.auto_stock_available)
    const upstreamStock = normalizeManualStockTotal(matchedSku?.upstream_stock)
    const skuStockEnforced = shouldEnforceSkuStock(product, matchedSku)
    const stockRangeMin = normalizeStockNumber(matchedSku?.stock_range_min)
    const stockRangeMax = normalizeStockNumber(matchedSku?.stock_range_max)

    return {
      skuId: normalizeSkuId(matchedSku.id),
      ...resolvePurchasePricingSnapshot(product, matchedSku),
      skuCode: String(matchedSku?.sku_code || ''),
      priceAmount: String(matchedSku?.price_amount ?? ''),
      skuSpecValues: (matchedSku?.spec_values && typeof matchedSku.spec_values === 'object')
        ? matchedSku.spec_values
        : {},
      skuManualStockTotal: manualStockTotal,
      skuManualStockLocked: manualStockLocked,
      skuManualStockSold: manualStockSold,
      skuAutoStockAvailable: autoStockAvailable,
      skuUpstreamStock: upstreamStock,
      skuStockStatus: String(matchedSku?.stock_status || ''),
      skuStockDisplayMode: String(matchedSku?.stock_display_mode || product?.stock_display_mode || ''),
      skuStockDisplay: String(matchedSku?.stock_display || ''),
      skuStockRangeMin: stockRangeMin > 0 ? stockRangeMin : undefined,
      skuStockRangeMax: stockRangeMax > 0 ? stockRangeMax : undefined,
      skuStockQuantityHidden: Boolean(matchedSku?.stock_quantity_hidden || product?.stock_quantity_hidden),
      skuStockEnforced,
      skuStockSnapshotAt: new Date().toISOString(),
      minPurchaseQuantity: normalizeOptionalLimitNumber(product?.min_purchase_quantity),
      maxPurchaseQuantity: normalizeOptionalLimitNumber(product?.max_purchase_quantity),
      fulfillmentType: String(product?.fulfillment_type || item.fulfillmentType || ''),
      manualFormSchema: product?.manual_form_schema || {},
      paymentChannelIds: Array.isArray(product?.payment_channel_ids) && product.payment_channel_ids.length > 0
        ? product.payment_channel_ids.map(Number)
        : undefined,
    }
}

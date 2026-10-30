import type { PurchaseItem } from '../stores/buyNow'
import { amountToCents, centsToAmount } from './money.ts'

// 购买信息只同步目录基础价和批发阶梯；活动、会员与优惠券金额仍以服务端结算预览为准。
export const resolvePurchasePricingSnapshot = (product: any, sku: any): Partial<PurchaseItem> => {
  const patch: Partial<PurchaseItem> = {
    wholesalePrices: Array.isArray(product?.wholesale_prices)
      ? product.wholesale_prices
      : undefined,
  }

  const priceCents = amountToCents(sku?.price_amount)
  if (priceCents !== null && priceCents > 0) {
    patch.priceAmount = centsToAmount(priceCents)
  }

  return patch
}

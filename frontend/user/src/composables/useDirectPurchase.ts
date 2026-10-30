import { useRouter } from 'vue-router'
import { useBuyNowStore } from '../stores/buyNow'
import { getFirstImageUrl } from '../utils/image'

export function useDirectPurchase() {
  const router = useRouter()
  const store = useBuyNowStore()

  const startPurchase = (product: any) => {
    if (!product?.id || !product?.slug) return
    const minimum = Number(product.min_purchase_quantity)
    store.setItem({
      productId: Number(product.id),
      skuId: 0,
      slug: product.slug,
      title: product.title,
      priceAmount: String(product.price_amount || '0.00'),
      image: getFirstImageUrl(product.images),
      quantity: Number.isSafeInteger(minimum) && minimum > 0 ? minimum : 1,
      fulfillmentType: product.fulfillment_type,
      manualFormSchema: product.manual_form_schema || {},
      paymentChannelIds: Array.isArray(product.payment_channel_ids) && product.payment_channel_ids.length ? product.payment_channel_ids : undefined,
    })
    void router.push('/checkout')
  }

  return { startPurchase }
}

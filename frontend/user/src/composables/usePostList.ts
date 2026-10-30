import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '../stores/app'
import { postAPI } from '../api'
import { debounceAsync } from '../utils/debounce'
import { usePageSeo } from './usePageSeo'

/**
 * 文章/公告列表共享逻辑（Blog/Notice，classic + vault 双模板共用）。
 * 完整保留原 views/Blog.vue 与 Notice.vue 的行为，仅抽离为 composable。
 * Blog 启用搜索；Notice 无搜索框（searchKeyword 恒为空，watch 永不触发，行为一致）。
 */
export function usePostList(
  type: 'blog' | 'notice',
  seo: { title: () => string; canonicalPath: string },
) {
  const router = useRouter()
  const route = useRoute()
  const appStore = useAppStore()

  usePageSeo({
    title: seo.title,
    canonicalPath: () => seo.canonicalPath,
  })

  const loading = ref(true)
  const posts = ref<any[]>([])
  const readPage = () => {
    const value = typeof route.query.page === 'string' ? Number(route.query.page) : 1
    return Number.isSafeInteger(value) && value > 0 && value <= 100000 ? value : 1
  }
  const readSearch = () => typeof route.query.q === 'string' ? route.query.q.trim() : ''
  const currentPage = ref(readPage())
  const pageSize = ref(12)
  const total = ref(0)
  const totalPages = ref(0)
  const searchKeyword = ref(readSearch())
  let loadVersion = 0

  const getLocalizedText = (jsonData: any) => {
    if (!jsonData) return ''
    const locale = appStore.locale
    return jsonData[locale] || jsonData['zh-CN'] || jsonData['en-US'] || ''
  }

  const formatDate = (dateString: string) => {
    if (!dateString) return ''
    const date = new Date(dateString)
    return date.toLocaleDateString(appStore.locale, {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
  }

  const loadPosts = async () => {
    const version = ++loadVersion
    loading.value = true
    try {
      const params: Record<string, any> = {
        type,
        page: currentPage.value,
        page_size: pageSize.value,
      }
      const keyword = searchKeyword.value.trim()
      if (keyword) {
        params.search = keyword
      }
      const response = await postAPI.list(params)
      if (version !== loadVersion) return
      posts.value = response.data.data || []
      if (response.data.pagination) {
        total.value = response.data.pagination.total || 0
        totalPages.value = response.data.pagination.total_page || 0
      }
    } catch (error) {
      console.error('Failed to load posts:', error)
    } finally {
      if (version === loadVersion) loading.value = false
    }
  }

  const updateQuery = async (page: number, keyword: string) => {
    const query = { ...route.query }
    if (page > 1) query.page = String(page)
    else delete query.page
    if (keyword) query.q = keyword
    else delete query.q
    await router.replace({ query })
  }
  const debouncedSearch = debounceAsync(() => updateQuery(1, searchKeyword.value.trim()), 300)

  const goToPost = (slug: string) => {
    router.push(`/blog/${slug}`)
  }

  const changePage = (page: number) => {
    if (page < 1 || page > totalPages.value) return
    void updateQuery(page, searchKeyword.value.trim())
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  watch(searchKeyword, () => {
    if (searchKeyword.value.trim() !== readSearch()) void debouncedSearch()
  })

  watch(() => [route.query.page, route.query.q], () => {
    currentPage.value = readPage()
    searchKeyword.value = readSearch()
    void loadPosts()
  })

  onMounted(() => {
    loadPosts()
  })

  onUnmounted(() => {
    loadVersion++
    debouncedSearch.cancel()
  })

  return {
    loading,
    posts,
    currentPage,
    totalPages,
    total,
    searchKeyword,
    getLocalizedText,
    formatDate,
    goToPost,
    changePage,
  }
}

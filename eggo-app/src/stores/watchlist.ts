import { defineStore } from 'pinia'
import { ref } from 'vue'
import { get, post, del } from '@/utils/request'

export interface WatchlistItem {
  fund_code: string
  fund_name: string
  fund_type?: string
  remark?: string
  // 实时估值（FundDistribution 组件拉取后注入）
  estReturn?: number
  estNav?: number
  lastNav?: number
  updateTime?: string
}

export const useWatchlistStore = defineStore('watchlist', () => {
  const items = ref<WatchlistItem[]>([])
  const loading = ref(false)
  const loaded = ref(false) // 是否已加载过（避免重复请求）

  async function fetchList() {
    loading.value = true
    try {
      const res = await get<WatchlistItem[]>('/watchlist')
      if (res.code === 0 && res.data) {
        items.value = res.data
        loaded.value = true
      }
    } catch (e) {
      console.error('[watchlist] fetchList error:', e)
    } finally {
      loading.value = false
    }
  }

  async function add(fundCode: string): Promise<boolean> {
    try {
      const res = await post<{ added: boolean; fundCode: string; fundName: string; already: boolean }>(
        '/watchlist',
        { fund_code: fundCode }
      )
      if (res.code === 0) {
        // 重新拉列表以获取最新顺序和名称
        await fetchList()
        return true
      }
      return false
    } catch (e) {
      console.error('[watchlist] add error:', e)
      return false
    }
  }

  async function remove(fundCode: string): Promise<boolean> {
    try {
      const res = await del<{ removed: boolean }>(`/watchlist/${fundCode}`)
      if (res.code === 0) {
        items.value = items.value.filter(i => i.fund_code !== fundCode)
        return true
      }
      return false
    } catch (e) {
      console.error('[watchlist] remove error:', e)
      return false
    }
  }

  function isWatched(fundCode: string): boolean {
    return items.value.some(i => i.fund_code === fundCode)
  }

  return { items, loading, loaded, fetchList, add, remove, isWatched }
})

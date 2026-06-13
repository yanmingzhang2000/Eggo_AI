import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  EggStatusResponse,
  ChickenStatus,
  FundMetric,
  NewsClue,
  Decision,
} from '@/types/egg'
import { get } from '@/utils/request'

export const useEggStore = defineStore('egg', () => {
  // ── 状态 ──────────────────────────────────────────────────────
  const loading = ref(false)
  const error = ref<string | null>(null)
  const eggData = ref<EggStatusResponse | null>(null)
  const lastFetchTime = ref<string>('')

  // ── 计算属性 ──────────────────────────────────────────────────
  const chickenStatus = computed<ChickenStatus | null>(() => eggData.value?.chickenStatus ?? null)
  const todayMetrics = computed<FundMetric[]>(() => eggData.value?.todayMetrics ?? [])
  const newsClues = computed<NewsClue[]>(() => eggData.value?.newsClues ?? [])
  const decision = computed<Decision | null>(() => eggData.value?.decision ?? null)
  const hasData = computed(() => !!eggData.value)

  // ── 操作 ──────────────────────────────────────────────────────
  async function fetchEggStatus() {
    loading.value = true
    error.value = null

    try {
      const res = await get<EggStatusResponse>('/egg/status')
      if (res.code === 0 && res.data) {
        eggData.value = res.data
        lastFetchTime.value = res.data.generatedAt
      } else {
        error.value = res.message || '获取数据失败'
      }
    } catch (err: unknown) {
      error.value = (err as Error).message || '网络异常'
      console.error('[useEggStore] fetchEggStatus error:', err)
    } finally {
      loading.value = false
    }
  }

  async function refresh() {
    await fetchEggStatus()
  }

  return {
    loading,
    error,
    eggData,
    lastFetchTime,
    chickenStatus,
    todayMetrics,
    newsClues,
    decision,
    hasData,
    fetchEggStatus,
    refresh,
  }
})

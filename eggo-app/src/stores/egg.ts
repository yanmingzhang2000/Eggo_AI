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
  /** 母鸡状态 */
  const chickenStatus = computed<ChickenStatus | null>(() => eggData.value?.chickenStatus ?? null)

  /** 今日基金指标 */
  const todayMetrics = computed<FundMetric[]>(() => eggData.value?.todayMetrics ?? [])

  /** AI 新闻线索 */
  const newsClues = computed<NewsClue[]>(() => eggData.value?.newsClues ?? [])

  /** 最终决策 */
  const decision = computed<Decision | null>(() => eggData.value?.decision ?? null)

  /** 是否有数据 */
  const hasData = computed(() => !!eggData.value)

  /** 母鸡状态文案 */
  const stateText = computed(() => {
    if (!chickenStatus.value) return '加载中...'
    return chickenStatus.value.stateDesc
  })

  /** 决策建议文案 */
  const suggestionText = computed(() => {
    if (!decision.value) return ''
    return decision.value.suggestion
  })

  /** 是否触发铁律 */
  const hasTriggeredRule = computed(() => {
    return decision.value?.triggeredRule !== 'none'
  })

  // ── 操作 ──────────────────────────────────────────────────────
  /** 拉取母鸡状态 */
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

  /** 下拉刷新 */
  async function refresh() {
    await fetchEggStatus()
  }

  /** 清空数据 */
  function $reset() {
    eggData.value = null
    error.value = null
    lastFetchTime.value = ''
  }

  return {
    // 状态
    loading,
    error,
    eggData,
    lastFetchTime,
    // 计算属性
    chickenStatus,
    todayMetrics,
    newsClues,
    decision,
    hasData,
    stateText,
    suggestionText,
    hasTriggeredRule,
    // 操作
    fetchEggStatus,
    refresh,
    $reset,
  }
})

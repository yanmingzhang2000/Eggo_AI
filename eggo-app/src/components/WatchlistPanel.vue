<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useWatchlistStore } from '@/stores/watchlist'
import { get } from '@/utils/request'

const emit = defineEmits<{
  (e: 'viewDetail', fundCode: string, fundName: string): void
}>()

const store = useWatchlistStore()
const quotes = ref<Record<string, { estReturn: number; estNav: number; lastNav: number; updateTime: string }>>({})
const loadingQuotes = ref(false)
const removing = ref<string | null>(null)

async function fetchQuotes() {
  if (store.items.length === 0) return
  loadingQuotes.value = true
  try {
    const codes = store.items.map(i => i.fund_code).join(',')
    const res = await get<Array<{
      code: string; estReturn: number; estNav: number; lastNav: number; updateTime: string
    }>>('/market/fund-quotes', { codes })
    if (res.code === 0 && res.data) {
      const map: typeof quotes.value = {}
      for (const q of res.data) {
        map[q.code] = { estReturn: q.estReturn, estNav: q.estNav, lastNav: q.lastNav, updateTime: q.updateTime }
      }
      quotes.value = map
    }
  } catch (e) {
    console.error('[WatchlistPanel] fetchQuotes error:', e)
  } finally {
    loadingQuotes.value = false
  }
}

async function handleRemove(fundCode: string) {
  removing.value = fundCode
  await store.remove(fundCode)
  removing.value = null
  // 重新拉估值（列表变了）
  await fetchQuotes()
}

function returnColor(val: number): string {
  if (val > 0.01) return '#ff4d4f'
  if (val < -0.01) return '#00d68f'
  return '#787878'
}

function formatReturn(val: number): string {
  const sign = val > 0 ? '+' : ''
  return `${sign}${val.toFixed(2)}%`
}

onMounted(async () => {
  if (!store.loaded) {
    await store.fetchList()
  }
  await fetchQuotes()
})
</script>

<template>
  <section class="watchlist-panel">
    <div class="wp__header">
      <h2 class="wp__title">⭐ 我的收藏</h2>
      <span class="wp__count">{{ store.items.length }} 只</span>
    </div>

    <!-- 空状态 -->
    <div v-if="!store.loading && store.items.length === 0" class="wp__empty">
      <p class="wp__empty-icon">🐣</p>
      <p class="wp__empty-text">还没有收藏的基金</p>
      <p class="wp__empty-sub">在基金详情页点击「收藏」即可添加</p>
    </div>

    <!-- 加载中 -->
    <div v-else-if="store.loading" class="wp__loading">加载中...</div>

    <!-- 列表 -->
    <div v-else class="wp__list">
      <div
        v-for="item in store.items"
        :key="item.fund_code"
        class="wp__row"
        @click="emit('viewDetail', item.fund_code, item.fund_name)"
      >
        <!-- 左侧：名称+代码 -->
        <div class="wp__row-left">
          <span class="wp__name">{{ item.fund_name }}</span>
          <span class="wp__code">{{ item.fund_code }}</span>
        </div>

        <!-- 中间：实时估值 -->
        <div class="wp__row-mid">
          <span
            class="wp__return"
            :style="{ color: quotes[item.fund_code] ? returnColor(quotes[item.fund_code].estReturn) : '#787878' }"
          >
            {{ quotes[item.fund_code] ? formatReturn(quotes[item.fund_code].estReturn) : (loadingQuotes ? '···' : '--') }}
          </span>
          <span class="wp__nav" v-if="quotes[item.fund_code]">
            估值 {{ quotes[item.fund_code].estNav > 0 ? quotes[item.fund_code].estNav.toFixed(4) : '--' }}
          </span>
        </div>

        <!-- 右侧：删除按钮 -->
        <button
          class="wp__remove"
          :disabled="removing === item.fund_code"
          @click.stop="handleRemove(item.fund_code)"
          title="取消收藏"
        >
          {{ removing === item.fund_code ? '···' : '✕' }}
        </button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.watchlist-panel {
  margin-bottom: 32px;
}

.wp__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.wp__title {
  font-size: 16px;
  color: var(--text-secondary);
  font-weight: 600;
  margin: 0;
}

.wp__count {
  font-size: 12px;
  color: var(--text-tertiary);
}

/* 空状态 */
.wp__empty {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  padding: 32px 16px;
  text-align: center;
}

.wp__empty-icon {
  font-size: 36px;
  margin: 0 0 8px;
}

.wp__empty-text {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0 0 4px;
}

.wp__empty-sub {
  font-size: 12px;
  color: var(--text-tertiary);
  margin: 0;
}

.wp__loading {
  text-align: center;
  padding: 24px;
  font-size: 13px;
  color: var(--text-tertiary);
}

/* 列表 */
.wp__list {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  overflow: hidden;
}

.wp__row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  cursor: pointer;
  transition: background 0.15s;
}

.wp__row + .wp__row {
  border-top: 1px solid var(--border-color);
}

.wp__row:hover {
  background: rgba(255, 215, 0, 0.04);
}

.wp__row-left {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.wp__name {
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.wp__code {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

.wp__row-mid {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 3px;
  min-width: 80px;
}

.wp__return {
  font-size: 16px;
  font-weight: 700;
  font-family: var(--font-mono);
}

.wp__nav {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

.wp__remove {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
  background: transparent;
  border: 1px solid var(--border-color);
  border-radius: 50%;
  color: var(--text-tertiary);
  font-size: 11px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.wp__remove:hover:not(:disabled) {
  border-color: rgba(255, 77, 79, 0.5);
  color: #ff4d4f;
  background: rgba(255, 77, 79, 0.08);
}

.wp__remove:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
</style>

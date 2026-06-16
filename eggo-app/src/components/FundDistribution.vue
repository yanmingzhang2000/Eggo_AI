<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { get, getCache, setCache } from '@/utils/request'
import { useEggStore } from '@/stores/egg'

interface FundQuote {
  code: string
  name: string
  estReturn: number
  estNav: number
  lastNav: number
  updateTime: string
}

interface FundDistribution {
  riseCount: number
  fallCount: number
  flatCount: number
  total: number
  avgReturn: number
  topFunds: FundQuote[]
  flopFunds: FundQuote[]
  updateAt: string
}

const eggStore = useEggStore()

const emit = defineEmits<{
  (e: 'viewDetail', fundCode: string, fundName: string): void
}>()

const dist = ref<FundDistribution | null>(null)
const watchlistQuotes = ref<FundQuote[]>([])
const loading = ref(true)
const refreshing = ref(false)

let timer: ReturnType<typeof setInterval> | null = null

// 从 egg store 拿自选基金代码
const watchlistCodes = computed(() =>
  eggStore.todayMetrics?.map((m: any) => m.fundCode).filter(Boolean) ?? []
)

async function fetchDist(silent = false) {
  if (!silent) loading.value = true

  // 先读缓存
  const cached = getCache<FundDistribution>('/market/fund-distribution')
  if (cached) dist.value = cached

  try {
    const res = await get<FundDistribution>('/market/fund-distribution')
    if (res.code === 0 && res.data) {
      dist.value = res.data
      setCache('/market/fund-distribution', res.data)
    }
  } catch (e) {
    console.error('Failed to fetch fund distribution:', e)
  } finally {
    loading.value = false
  }
}

async function fetchWatchlist(silent = false) {
  if (watchlistCodes.value.length === 0) return

  const cacheUrl = '/market/fund-quotes'
  const cached = getCache<FundQuote[]>(cacheUrl)
  if (cached) watchlistQuotes.value = cached

  try {
    const codes = watchlistCodes.value.join(',')
    const res = await get<FundQuote[]>('/market/fund-quotes', { codes })
    if (res.code === 0 && res.data) {
      watchlistQuotes.value = res.data
      setCache(cacheUrl, res.data)
    }
  } catch (e) {
    console.error('Failed to fetch watchlist quotes:', e)
  }
}

async function refresh() {
  refreshing.value = true
  await Promise.all([fetchDist(true), fetchWatchlist(true)])
  refreshing.value = false
}

onMounted(async () => {
  await Promise.all([fetchDist(), fetchWatchlist()])
  timer = setInterval(refresh, 3 * 60 * 1000) // 每3分钟静默刷新
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

function returnColor(val: number): string {
  if (val > 0.01) return '#ff4d4f'
  if (val < -0.01) return '#00d68f'
  return '#787878'
}

function formatReturn(val: number): string {
  const sign = val > 0 ? '+' : ''
  return `${sign}${val.toFixed(2)}%`
}

// 涨跌分布色条百分比
const riseRatio = computed(() => {
  if (!dist.value || dist.value.total === 0) return 0
  return (dist.value.riseCount / dist.value.total) * 100
})
const fallRatio = computed(() => {
  if (!dist.value || dist.value.total === 0) return 0
  return (dist.value.fallCount / dist.value.total) * 100
})
const flatRatio = computed(() => {
  if (!dist.value || dist.value.total === 0) return 0
  return (dist.value.flatCount / dist.value.total) * 100
})

// 市场情绪
const sentiment = computed(() => {
  if (!dist.value) return { text: '--', color: '#787878' }
  const { riseCount, fallCount, total } = dist.value
  if (total === 0) return { text: '--', color: '#787878' }
  const ratio = riseCount / total
  if (ratio > 0.65) return { text: '全面上攻', color: '#ff4d4f' }
  if (ratio > 0.55) return { text: '多头占优', color: '#ff7875' }
  if (ratio > 0.45) return { text: '震荡分化', color: '#f7ba1e' }
  if (ratio > 0.35) return { text: '空头占优', color: '#52c41a' }
  return { text: '全面回调', color: '#00d68f' }
})
</script>

<template>
  <section class="fund-dist">
    <div class="fund-dist__header">
      <h2 class="fund-dist__title">📊 基金涨跌分布</h2>
      <div class="fund-dist__right">
        <span v-if="refreshing" class="fund-dist__hint">刷新中...</span>
        <button class="fund-dist__refresh" @click="refresh" :disabled="refreshing">
          <span :class="{ spin: refreshing }">↻</span>
        </button>
      </div>
    </div>

    <!-- 自选基金实时估值 -->
    <div v-if="watchlistCodes.length > 0" class="watchlist-section">
      <div class="section-label">我的自选</div>
      <div class="watchlist-grid">
          <div
              v-for="q in watchlistQuotes"
              :key="q.code"
              class="wl-card wl-card--clickable"
              @click="emit('viewDetail', q.code, q.name)"
          >
          <div class="wl-card__top">
            <span class="wl-card__name">{{ q.name }}</span>
            <span class="wl-card__code">{{ q.code }}</span>
          </div>
          <div class="wl-card__return" :style="{ color: returnColor(q.estReturn) }">
            {{ formatReturn(q.estReturn) }}
          </div>
          <div class="wl-card__nav">
            估值 <span>{{ q.estNav > 0 ? q.estNav.toFixed(4) : '--' }}</span>
            · 昨净 <span>{{ q.lastNav > 0 ? q.lastNav.toFixed(4) : '--' }}</span>
          </div>
          <div class="wl-card__time">{{ q.updateTime || '--' }}</div>
        </div>
        <div v-if="watchlistQuotes.length === 0 && !loading" class="wl-empty">
          自选基金估值加载中…
        </div>
      </div>
    </div>

    <!-- 全市场涨跌分布 -->
    <div class="market-section">
      <div class="section-label">
        全市场基金
        <span v-if="dist" class="section-label__sub">· 样本 {{ dist.total }} 只</span>
      </div>

      <div v-if="loading && !dist" class="dist-loading">加载中...</div>

      <template v-else-if="dist">
        <!-- 情绪 + 均涨跌 -->
        <div class="dist-summary">
          <div class="dist-summary__item">
            <span class="dist-summary__val" :style="{ color: sentiment.color }">
              {{ sentiment.text }}
            </span>
            <span class="dist-summary__label">市场情绪</span>
          </div>
          <div class="dist-summary__item">
            <span class="dist-summary__val" :style="{ color: returnColor(dist.avgReturn) }">
              {{ formatReturn(dist.avgReturn) }}
            </span>
            <span class="dist-summary__label">平均涨跌</span>
          </div>
          <div class="dist-summary__item">
            <span class="dist-summary__val" style="color: #ff4d4f">{{ dist.riseCount }}</span>
            <span class="dist-summary__label">上涨只数</span>
          </div>
          <div class="dist-summary__item">
            <span class="dist-summary__val" style="color: #00d68f">{{ dist.fallCount }}</span>
            <span class="dist-summary__label">下跌只数</span>
          </div>
        </div>

        <!-- 涨跌色条 -->
        <div class="dist-bar">
          <div class="dist-bar__track">
            <div class="dist-bar__seg dist-bar__seg--rise" :style="{ width: riseRatio + '%' }"></div>
            <div class="dist-bar__seg dist-bar__seg--flat" :style="{ width: flatRatio + '%' }"></div>
            <div class="dist-bar__seg dist-bar__seg--fall" :style="{ width: fallRatio + '%' }"></div>
          </div>
          <div class="dist-bar__legend">
            <span><i class="dot dot--rise"></i>涨 {{ dist.riseCount }}（{{ riseRatio.toFixed(0) }}%）</span>
            <span><i class="dot dot--flat"></i>平 {{ dist.flatCount }}</span>
            <span><i class="dot dot--fall"></i>跌 {{ dist.fallCount }}（{{ fallRatio.toFixed(0) }}%）</span>
          </div>
        </div>

        <!-- 涨跌榜 -->
        <div class="rank-row">
          <div class="rank-col">
            <div class="rank-col__title" style="color:#ff4d4f">涨幅 Top5</div>
            <div
              v-for="f in dist.topFunds"
              :key="f.code"
              class="rank-item rank-item--clickable"
              @click="emit('viewDetail', f.code, f.name)"
            >
              <span class="rank-item__name">{{ f.name }}</span>
              <span class="rank-item__return" style="color:#ff4d4f">{{ formatReturn(f.estReturn) }}</span>
            </div>
          </div>
          <div class="rank-col">
            <div class="rank-col__title" style="color:#00d68f">跌幅 Top5</div>
            <div
              v-for="f in dist.flopFunds"
              :key="f.code"
              class="rank-item rank-item--clickable"
              @click="emit('viewDetail', f.code, f.name)"
            >
              <span class="rank-item__name">{{ f.name }}</span>
              <span class="rank-item__return" style="color:#00d68f">{{ formatReturn(f.estReturn) }}</span>
            </div>
          </div>
        </div>

        <div class="dist-footer">更新于 {{ dist.updateAt }}</div>
      </template>
    </div>
  </section>
</template>

<style scoped>
.fund-dist {
  margin-bottom: 32px;
}

.fund-dist__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.fund-dist__title {
  font-size: 16px;
  color: var(--text-secondary);
  font-weight: 600;
  margin: 0;
}

.fund-dist__right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.fund-dist__hint {
  font-size: 11px;
  color: var(--text-tertiary);
}

.fund-dist__refresh {
  width: 28px;
  height: 28px;
  background: rgba(255, 215, 0, 0.08);
  border: 1px solid rgba(255, 215, 0, 0.4);
  color: var(--accent);
  border-radius: 50%;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.fund-dist__refresh:hover {
  background: rgba(255, 215, 0, 0.18);
  border-color: #f7ba1e;
}

.fund-dist__refresh:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.spin {
  display: inline-block;
  animation: spin 1s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

/* 分区标签 */
.section-label {
  font-size: 12px;
  color: var(--text-tertiary);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 1px;
  margin-bottom: 10px;
}

.section-label__sub {
  font-weight: 400;
  text-transform: none;
  letter-spacing: 0;
}

/* ── 自选基金 ── */
.watchlist-section {
  margin-bottom: 20px;
}

.watchlist-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.wl-card {
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 14px;
  padding: 14px;
  transition: border-color 0.2s, transform 0.15s;
}

.wl-card--clickable {
  cursor: pointer;
}

.wl-card--clickable:hover {
  border-color: #f7ba1e;
  transform: translateY(-1px);
}

.wl-card__top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.wl-card__name {
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 70%;
}

.wl-card__code {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

.wl-card__return {
  font-size: 20px;
  font-weight: 700;
  font-family: var(--font-mono);
  margin-bottom: 6px;
}

.wl-card__nav {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

.wl-card__nav span {
  color: var(--text-secondary);
}

.wl-card__time {
  font-size: 10px;
  color: var(--text-tertiary);
  margin-top: 4px;
}

.wl-empty {
  font-size: 13px;
  color: var(--text-tertiary);
  padding: 12px 0;
}

/* ── 全市场分布 ── */
.market-section {
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 16px;
  padding: 16px;
}

.dist-loading {
  font-size: 13px;
  color: var(--text-tertiary);
  padding: 24px 0;
  text-align: center;
}

/* 摘要数字 */
.dist-summary {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  margin-bottom: 16px;
}

.dist-summary__item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.dist-summary__val {
  font-size: 18px;
  font-weight: 700;
  font-family: var(--font-mono);
  line-height: 1;
}

.dist-summary__label {
  font-size: 11px;
  color: var(--text-tertiary);
}

/* 色条 */
.dist-bar__track {
  display: flex;
  height: 8px;
  border-radius: 4px;
  overflow: hidden;
  background: #2a2a2a;
  margin-bottom: 10px;
}

.dist-bar__seg {
  transition: width 0.4s ease;
}

.dist-bar__seg--rise { background: #ff4d4f; }
.dist-bar__seg--flat { background: #4a4a4a; }
.dist-bar__seg--fall { background: #00d68f; }

.dist-bar__legend {
  display: flex;
  gap: 14px;
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 16px;
}

.dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  margin-right: 4px;
  vertical-align: middle;
}

.dot--rise { background: #ff4d4f; }
.dot--flat { background: #4a4a4a; }
.dot--fall { background: #00d68f; }

/* 涨跌榜 */
.rank-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 12px;
}

.rank-col__title {
  font-size: 12px;
  font-weight: 600;
  margin-bottom: 8px;
}

.rank-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 5px 0;
  border-bottom: 1px solid #1e1e1e;
  transition: opacity 0.15s;
}

.rank-item--clickable {
  cursor: pointer;
}

.rank-item--clickable:hover {
  opacity: 0.8;
}

.rank-item__name {
  font-size: 12px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 65%;
}

.rank-item__return {
  font-size: 12px;
  font-weight: 700;
  font-family: var(--font-mono);
  white-space: nowrap;
}

.dist-footer {
  font-size: 11px;
  color: var(--text-tertiary);
  text-align: right;
}

@media (max-width: 480px) {
  .watchlist-grid {
    grid-template-columns: 1fr;
  }

  .dist-summary {
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }

  .rank-row {
    grid-template-columns: 1fr;
  }
}
</style>

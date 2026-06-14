<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { get } from '@/utils/request'

interface IndexData {
  name: string
  code: string
  price: number
  change: number
  changePercent: number
  volume: string
  high: number
  low: number
}

interface SectorData {
  name: string
  code: string
  changePercent: number
  price: number
}

interface MarketStats {
  upCount: number
  downCount: number
  flatCount: number
  limitUp: number
  limitDown: number
  totalVolume: string
  northFlow: string
  sentiment: string
  updateAt: string
}

interface KLineData {
  date: string
  open: number
  close: number
  high: number
  low: number
}

interface IndexOption {
  name: string
  code: string
}

const indices = ref<IndexData[]>([])
const sectors = ref<SectorData[]>([])
const concepts = ref<SectorData[]>([])
const marketStats = ref<MarketStats>({
  upCount: 0,
  downCount: 0,
  flatCount: 0,
  limitUp: 0,
  limitDown: 0,
  totalVolume: '--',
  northFlow: '--',
  sentiment: '--',
  updateAt: '--',
})

const loading = ref(true)
const activeTab = ref<'index' | 'sector' | 'concept' | 'stats'>('index')
let refreshTimer: ReturnType<typeof setInterval> | null = null

// 历史趋势相关
const showHistory = ref(false)
const historyLoading = ref(false)
const indexOptions = ref<IndexOption[]>([])
const selectedCode = ref('1.000300')
const klineData = ref<KLineData[]>([])

async function fetchIndices() {
  try {
    const res = await get<IndexData[]>('/market/indices')
    if (res.code === 0 && res.data) {
      indices.value = res.data
    }
  } catch (err) {
    console.error('Failed to fetch indices:', err)
  }
}

async function fetchSectors() {
  try {
    const res = await get<SectorData[]>('/market/sectors')
    if (res.code === 0 && res.data) {
      sectors.value = res.data
    }
  } catch (err) {
    console.error('Failed to fetch sectors:', err)
  }
}

async function fetchConcepts() {
  try {
    const res = await get<SectorData[]>('/market/concepts')
    if (res.code === 0 && res.data) {
      concepts.value = res.data
    }
  } catch (err) {
    console.error('Failed to fetch concepts:', err)
  }
}

async function fetchOverview() {
  try {
    const res = await get<MarketStats>('/market/overview')
    if (res.code === 0 && res.data) {
      marketStats.value = res.data
    }
  } catch (err) {
    console.error('Failed to fetch overview:', err)
  }
}

async function fetchAll() {
  loading.value = true
  await Promise.all([fetchIndices(), fetchSectors(), fetchConcepts(), fetchOverview()])
  loading.value = false
}

async function fetchIndexOptions() {
  try {
    const res = await get<IndexOption[]>('/market/index-options')
    if (res.code === 0 && res.data) {
      indexOptions.value = res.data
    }
  } catch (err) {
    console.error('Failed to fetch index options:', err)
  }
}

async function fetchHistory() {
  historyLoading.value = true
  try {
    const res = await get<KLineData[]>('/market/history', { code: selectedCode.value, days: 120 })
    if (res.code === 0 && res.data) {
      klineData.value = res.data
    }
  } catch (err) {
    console.error('Failed to fetch history:', err)
  } finally {
    historyLoading.value = false
  }
}

function openHistory() {
  showHistory.value = true
  if (indexOptions.value.length === 0) {
    fetchIndexOptions()
  }
  fetchHistory()
}

function closeHistory() {
  showHistory.value = false
}

function selectIndex(code: string) {
  selectedCode.value = code
  fetchHistory()
}

function returnColor(val: number): string {
  if (val > 0) return '#ff4d4f'
  if (val < 0) return '#00b96b'
  return '#666'
}

function returnArrow(val: number): string {
  if (val > 0) return '↑'
  if (val < 0) return '↓'
  return '→'
}

// SVG 折线图生成（逐段变色）
function generateChartPaths(data: KLineData[]): { up: string; down: string } {
  if (data.length < 2) return { up: '', down: '' }
  const closes = data.map(d => d.close)
  const min = Math.min(...closes)
  const max = Math.max(...closes)
  const range = max - min || 1
  const width = 460
  const height = 200
  const padding = { top: 20, bottom: 30, left: 50, right: 20 }
  const chartW = width - padding.left - padding.right
  const chartH = height - padding.top - padding.bottom

  const points = closes.map((val, i) => {
    const x = padding.left + (i / (closes.length - 1)) * chartW
    const y = padding.top + chartH - ((val - min) / range) * chartH
    return { x, y }
  })

  let upPath = ''
  let downPath = ''

  for (let i = 1; i < points.length; i++) {
    const prev = points[i - 1]
    const curr = points[i]
    const d = `M${prev.x},${prev.y} L${curr.x},${curr.y}`
    if (closes[i] >= closes[i - 1]) {
      upPath += (upPath ? ' ' : '') + d
    } else {
      downPath += (downPath ? ' ' : '') + d
    }
  }

  return { up: upPath, down: downPath }
}

function generateChartFill(data: KLineData[]): string {
  if (data.length < 2) return ''
  const closes = data.map(d => d.close)
  const min = Math.min(...closes)
  const max = Math.max(...closes)
  const range = max - min || 1
  const width = 460
  const height = 200
  const padding = { top: 20, bottom: 30, left: 50, right: 20 }
  const chartW = width - padding.left - padding.right
  const chartH = height - padding.top - padding.bottom

  const points = closes.map((val, i) => {
    const x = padding.left + (i / (closes.length - 1)) * chartW
    const y = padding.top + chartH - ((val - min) / range) * chartH
    return `${x},${y}`
  })

  return `M${points.join(' L')} L${points[points.length - 1].split(',')[0]},${padding.top + chartH} L${points[0].split(',')[0]},${padding.top + chartH} Z`
}

function getChartYLabels(data: KLineData[]): { value: string; y: number }[] {
  if (data.length < 2) return []
  const closes = data.map(d => d.close)
  const min = Math.min(...closes)
  const max = Math.max(...closes)
  const range = max - min || 1
  const height = 200
  const padding = { top: 20, bottom: 30 }
  const chartH = height - padding.top - padding.bottom

  const labels = []
  for (let i = 0; i <= 4; i++) {
    const val = min + (range * i) / 4
    const y = padding.top + chartH - (i / 4) * chartH
    labels.push({ value: val.toFixed(2), y })
  }
  return labels
}

function getChartXLabels(data: KLineData[]): { date: string; x: number }[] {
  if (data.length < 2) return []
  const width = 460
  const padding = { left: 50, right: 20 }
  const chartW = width - padding.left - padding.right
  const step = Math.floor(data.length / 4)

  const labels = []
  for (let i = 0; i <= 4; i++) {
    const idx = Math.min(i * step, data.length - 1)
    const x = padding.left + (idx / (data.length - 1)) * chartW
    const date = data[idx].date.slice(5) // MM-DD
    labels.push({ date, x })
  }
  return labels
}

onMounted(() => {
  fetchAll()
  // 每 60 秒自动刷新
  refreshTimer = setInterval(fetchAll, 60000)
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<template>
  <section class="dashboard">
    <div class="dashboard__header">
      <h2 class="dashboard__title">📈 大盘看板</h2>
      <div class="dashboard__right">
        <button class="dashboard__history" @click="openHistory">看历史趋势</button>
        <span v-if="loading" class="dashboard__loading">加载中...</span>
        <button class="dashboard__refresh" @click="fetchAll" :disabled="loading">
          <span :class="{ 'spin': loading }">↻</span>
        </button>
      </div>
    </div>

    <!-- Tab 切换 -->
    <div class="tabs">
      <button
        :class="['tab', { 'tab--active': activeTab === 'index' }]"
        @click="activeTab = 'index'"
      >指数行情</button>
      <button
        :class="['tab', { 'tab--active': activeTab === 'sector' }]"
        @click="activeTab = 'sector'"
      >行业板块</button>
      <button
        :class="['tab', { 'tab--active': activeTab === 'concept' }]"
        @click="activeTab = 'concept'"
      >概念板块</button>
      <button
        :class="['tab', { 'tab--active': activeTab === 'stats' }]"
        @click="activeTab = 'stats'"
      >市场情绪</button>
    </div>

    <!-- 指数行情 -->
    <div v-if="activeTab === 'index'" class="indices-grid">
      <div v-for="idx in indices" :key="idx.code" class="index-card">
        <div class="index-card__top">
          <span class="index-card__name">{{ idx.name }}</span>
          <span class="index-card__code">{{ idx.code }}</span>
        </div>
        <div class="index-card__price" :style="{ color: returnColor(idx.change) }">
          {{ idx.price.toFixed(2) }}
        </div>
        <div class="index-card__change">
          <span :style="{ color: returnColor(idx.change) }">
            {{ returnArrow(idx.change) }} {{ Math.abs(idx.change).toFixed(2) }}
          </span>
          <span :style="{ color: returnColor(idx.changePercent) }">
            {{ idx.changePercent > 0 ? '+' : '' }}{{ idx.changePercent.toFixed(2) }}%
          </span>
        </div>
        <div class="index-card__meta">
          <div class="meta-item">
            <span class="meta-label">成交</span>
            <span class="meta-value">{{ idx.volume }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label">最高</span>
            <span class="meta-value">{{ idx.high.toFixed(2) }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label">最低</span>
            <span class="meta-value">{{ idx.low.toFixed(2) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 行业板块 -->
    <div v-if="activeTab === 'sector'" class="sectors-list">
      <div v-for="(s, i) in sectors" :key="s.code" class="sector-row">
        <div class="sector-rank" :style="{ color: i < 3 ? '#ffd700' : '#666' }">{{ i + 1 }}</div>
        <div class="sector-info">
          <div class="sector-name">{{ s.name }}</div>
          <div class="sector-reason" v-if="s.price">最新价: {{ s.price.toFixed(2) }}</div>
        </div>
        <div class="sector-change" :style="{ color: returnColor(s.changePercent) }">
          {{ s.changePercent > 0 ? '+' : '' }}{{ s.changePercent.toFixed(2) }}%
        </div>
      </div>
      <div v-if="sectors.length === 0 && !loading" class="empty-hint">暂无数据</div>
    </div>

    <!-- 概念板块 -->
    <div v-if="activeTab === 'concept'" class="sectors-list">
      <div v-for="(s, i) in concepts" :key="s.code" class="sector-row">
        <div class="sector-rank" :style="{ color: i < 3 ? '#ffd700' : '#666' }">{{ i + 1 }}</div>
        <div class="sector-info">
          <div class="sector-name">{{ s.name }}</div>
          <div class="sector-reason" v-if="s.price">最新价: {{ s.price.toFixed(2) }}</div>
        </div>
        <div class="sector-change" :style="{ color: returnColor(s.changePercent) }">
          {{ s.changePercent > 0 ? '+' : '' }}{{ s.changePercent.toFixed(2) }}%
        </div>
      </div>
      <div v-if="concepts.length === 0 && !loading" class="empty-hint">暂无数据</div>
    </div>

    <!-- 市场情绪 -->
    <div v-if="activeTab === 'stats'" class="stats-panel">
      <div class="stats-grid">
        <div class="stat-card stat-card--up">
          <div class="stat-card__value">{{ marketStats.upCount }}</div>
          <div class="stat-card__label">上涨家数</div>
        </div>
        <div class="stat-card stat-card--down">
          <div class="stat-card__value">{{ marketStats.downCount }}</div>
          <div class="stat-card__label">下跌家数</div>
        </div>
        <div class="stat-card stat-card--flat">
          <div class="stat-card__value">{{ marketStats.flatCount }}</div>
          <div class="stat-card__label">平盘家数</div>
        </div>
      </div>

      <div class="stats-grid stats-grid--2">
        <div class="stat-card">
          <div class="stat-card__value" style="color: #ff4d4f">{{ marketStats.limitUp }}</div>
          <div class="stat-card__label">涨停</div>
        </div>
        <div class="stat-card">
          <div class="stat-card__value" style="color: #00b96b">{{ marketStats.limitDown }}</div>
          <div class="stat-card__label">跌停</div>
        </div>
      </div>

      <div class="stats-detail">
        <div class="detail-row">
          <span class="detail-label">北向资金</span>
          <span class="detail-value" :style="{ color: marketStats.northFlow?.startsWith('+') ? '#ff4d4f' : '#00b96b' }">{{ marketStats.northFlow || '--' }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">市场情绪</span>
          <span class="detail-value" style="color: #ffd700">{{ marketStats.sentiment }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">更新时间</span>
          <span class="detail-value">{{ marketStats.updateAt }}</span>
        </div>
      </div>

      <!-- 涨跌分布条 -->
      <div class="bar-chart" v-if="marketStats.upCount + marketStats.downCount + marketStats.flatCount > 0">
        <div class="bar-label">涨跌分布</div>
        <div class="bar-container">
          <div class="bar-segment bar-segment--up" :style="{ width: (marketStats.upCount / (marketStats.upCount + marketStats.downCount + marketStats.flatCount) * 100) + '%' }"></div>
          <div class="bar-segment bar-segment--flat" :style="{ width: (marketStats.flatCount / (marketStats.upCount + marketStats.downCount + marketStats.flatCount) * 100) + '%' }"></div>
          <div class="bar-segment bar-segment--down" :style="{ width: (marketStats.downCount / (marketStats.upCount + marketStats.downCount + marketStats.flatCount) * 100) + '%' }"></div>
        </div>
        <div class="bar-legend">
          <span><i class="dot dot--up"></i>涨 {{ marketStats.upCount }}</span>
          <span><i class="dot dot--flat"></i>平 {{ marketStats.flatCount }}</span>
          <span><i class="dot dot--down"></i>跌 {{ marketStats.downCount }}</span>
        </div>
      </div>
    </div>

    <!-- 历史趋势弹窗 -->
    <Teleport to="body">
      <div v-if="showHistory" class="modal-overlay" @click.self="closeHistory">
        <div class="modal">
          <div class="modal__header">
            <h3 class="modal__title">历史趋势（近120个交易日）</h3>
            <button class="modal__close" @click="closeHistory">✕</button>
          </div>

          <div class="modal__tabs">
            <button
              v-for="opt in indexOptions"
              :key="opt.code"
              :class="['modal__tab', { 'modal__tab--active': selectedCode === opt.code }]"
              @click="selectIndex(opt.code)"
            >{{ opt.name }}</button>
          </div>

          <div class="modal__chart">
            <div v-if="historyLoading" class="chart-loading">加载中...</div>
            <svg v-else viewBox="0 0 460 200" class="chart-svg">
              <!-- Y 轴标签 -->
              <text
                v-for="(label, i) in getChartYLabels(klineData)"
                :key="'y'+i"
                :x="42"
                :y="label.y + 4"
                class="chart-label"
                text-anchor="end"
              >{{ label.value }}</text>

              <!-- 网格线 -->
              <line
                v-for="(label, i) in getChartYLabels(klineData)"
                :key="'g'+i"
                :x1="50"
                :x2="440"
                :y1="label.y"
                :y2="label.y"
                class="chart-grid"
              />

              <!-- 渐变填充 -->
              <defs>
                <linearGradient id="chartGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#999" stop-opacity="0.2" />
                  <stop offset="100%" stop-color="#999" stop-opacity="0" />
                </linearGradient>
              </defs>
              <path
                :d="generateChartFill(klineData)"
                fill="url(#chartGrad)"
              />

              <!-- 涨（红） -->
              <path
                v-if="generateChartPaths(klineData).up"
                :d="generateChartPaths(klineData).up"
                fill="none"
                stroke="#ff4d4f"
                stroke-width="2"
                stroke-linecap="round"
              />

              <!-- 跌（绿） -->
              <path
                v-if="generateChartPaths(klineData).down"
                :d="generateChartPaths(klineData).down"
                fill="none"
                stroke="#00b96b"
                stroke-width="2"
                stroke-linecap="round"
              />

              <!-- X 轴标签 -->
              <text
                v-for="(label, i) in getChartXLabels(klineData)"
                :key="'x'+i"
                :x="label.x"
                :y="192"
                class="chart-label"
                text-anchor="middle"
              >{{ label.date }}</text>
            </svg>
          </div>

          <div class="modal__footer">
            <span v-if="klineData.length > 0" class="chart-info">
              {{ klineData[0].date }} ~ {{ klineData[klineData.length - 1].date }}
              &nbsp;|&nbsp;起点 {{ klineData[0].close.toFixed(2) }}
              &nbsp;→&nbsp;终点 {{ klineData[klineData.length - 1].close.toFixed(2) }}
              &nbsp;
              <span :style="{ color: klineData[klineData.length - 1].close >= klineData[0].close ? '#ff4d4f' : '#00b96b' }">
                ({{ klineData[klineData.length - 1].close >= klineData[0].close ? '+' : '' }}{{ ((klineData[klineData.length - 1].close - klineData[0].close) / klineData[0].close * 100).toFixed(2) }}%)
              </span>
            </span>
          </div>
        </div>
      </div>
    </Teleport>
  </section>
</template>

<style scoped>
.dashboard {
  margin-bottom: 32px;
}

.dashboard__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.dashboard__title {
  font-size: 16px;
  color: var(--text-secondary);
  font-weight: 600;
  margin: 0;
}

.dashboard__right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dashboard__loading {
  font-size: 11px;
  color: var(--text-tertiary);
}

.dashboard__refresh {
  width: 28px;
  height: 28px;
  background: rgba(255, 215, 0, 0.1);
  border: 1px solid rgba(255, 215, 0, 0.2);
  color: var(--accent);
  border-radius: 50%;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dashboard__refresh:hover {
  background: rgba(255, 215, 0, 0.2);
}

.dashboard__refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.dashboard__history {
  padding: 6px 12px;
  border-radius: 8px;
  border: 1px solid rgba(255, 215, 0, 0.3);
  background: rgba(255, 215, 0, 0.1);
  color: var(--accent);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.dashboard__history:hover {
  background: rgba(255, 215, 0, 0.2);
}

.spin {
  display: inline-block;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Tabs */
.tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.tab {
  padding: 8px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  background: transparent;
  color: var(--text-tertiary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.tab:hover {
  color: var(--text-secondary);
  border-color: rgba(255, 215, 0, 0.2);
}

.tab--active {
  background: rgba(255, 215, 0, 0.1);
  border-color: rgba(255, 215, 0, 0.3);
  color: var(--accent);
}

/* Index Cards */
.indices-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.index-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  padding: 16px;
}

.index-card__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.index-card__name {
  font-size: 13px;
  color: var(--text-secondary);
  font-weight: 600;
}

.index-card__code {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

.index-card__price {
  font-size: 22px;
  font-weight: 700;
  font-family: var(--font-mono);
  margin-bottom: 4px;
}

.index-card__change {
  display: flex;
  gap: 12px;
  font-size: 13px;
  font-family: var(--font-mono);
  margin-bottom: 12px;
}

.index-card__meta {
  display: flex;
  gap: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.meta-label {
  font-size: 11px;
  color: var(--text-tertiary);
}

.meta-value {
  font-size: 12px;
  color: var(--text-secondary);
  font-family: var(--font-mono);
}

/* Sectors */
.sectors-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sector-row {
  display: flex;
  align-items: center;
  gap: 12px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 14px 16px;
}

.sector-rank {
  font-size: 16px;
  font-weight: 700;
  font-family: var(--font-mono);
  min-width: 24px;
  text-align: center;
}

.sector-info {
  flex: 1;
  min-width: 0;
}

.sector-name {
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 600;
}

.sector-reason {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sector-change {
  font-size: 15px;
  font-weight: 700;
  font-family: var(--font-mono);
  min-width: 72px;
  text-align: right;
}

.empty-hint {
  text-align: center;
  padding: 32px;
  color: var(--text-tertiary);
  font-size: 14px;
}

/* Stats Panel */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 12px;
}

.stats-grid--2 {
  grid-template-columns: repeat(2, 1fr);
}

.stat-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 16px;
  text-align: center;
}

.stat-card__value {
  font-size: 24px;
  font-weight: 700;
  font-family: var(--font-mono);
  margin-bottom: 4px;
}

.stat-card--up .stat-card__value { color: #ff4d4f; }
.stat-card--down .stat-card__value { color: #00b96b; }
.stat-card--flat .stat-card__value { color: #666; }

.stat-card__label {
  font-size: 12px;
  color: var(--text-tertiary);
}

.stats-detail {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
}

.detail-row + .detail-row {
  border-top: 1px solid var(--border-color);
}

.detail-label {
  font-size: 13px;
  color: var(--text-tertiary);
}

.detail-value {
  font-size: 14px;
  font-weight: 600;
  font-family: var(--font-mono);
}

/* Bar Chart */
.bar-chart {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 16px;
}

.bar-label {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-bottom: 12px;
}

.bar-container {
  display: flex;
  height: 8px;
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 12px;
}

.bar-segment {
  transition: width 0.3s;
}

.bar-segment--up { background: #ff4d4f; }
.bar-segment--flat { background: #666; }
.bar-segment--down { background: #00b96b; }

.bar-legend {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: var(--text-tertiary);
}

.dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 4px;
  vertical-align: middle;
}

.dot--up { background: #ff4d4f; }
.dot--flat { background: #666; }
.dot--down { background: #00b96b; }

@media (max-width: 480px) {
  .indices-grid {
    grid-template-columns: 1fr;
  }

  .index-card__price {
    font-size: 18px;
  }

  .stats-grid {
    grid-template-columns: repeat(3, 1fr);
  }

  .modal {
    width: 95%;
    max-height: 85vh;
  }
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.modal {
  background: #1a1a2e;
  border: 1px solid var(--border-color);
  border-radius: 20px;
  width: 520px;
  max-width: 95vw;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

.modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px 0;
}

.modal__title {
  font-size: 16px;
  color: var(--text-primary);
  font-weight: 600;
  margin: 0;
}

.modal__close {
  width: 28px;
  height: 28px;
  background: rgba(255, 255, 255, 0.05);
  border: none;
  color: var(--text-tertiary);
  border-radius: 50%;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.modal__close:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}

.modal__tabs {
  display: flex;
  gap: 6px;
  padding: 16px 24px;
  flex-wrap: wrap;
}

.modal__tab {
  padding: 6px 12px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  background: transparent;
  color: var(--text-tertiary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.modal__tab:hover {
  color: var(--text-secondary);
  border-color: rgba(255, 215, 0, 0.2);
}

.modal__tab--active {
  background: rgba(255, 215, 0, 0.1);
  border-color: rgba(255, 215, 0, 0.3);
  color: var(--accent);
}

.modal__chart {
  padding: 0 24px;
  min-height: 220px;
}

.chart-loading {
  text-align: center;
  padding: 60px 0;
  color: var(--text-tertiary);
  font-size: 14px;
}

.chart-svg {
  width: 100%;
  height: auto;
}

.chart-label {
  font-size: 9px;
  fill: var(--text-tertiary);
  font-family: var(--font-mono);
}

.chart-grid {
  stroke: var(--border-color);
  stroke-width: 0.5;
  stroke-dasharray: 2,2;
}

.modal__footer {
  padding: 16px 24px 20px;
}

.chart-info {
  font-size: 12px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}
</style>

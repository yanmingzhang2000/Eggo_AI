<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { get } from '@/utils/request'
import HistoryTrend from './HistoryTrend.vue'

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

const showHistory = ref(false)

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

onMounted(() => {
  fetchAll()
  refreshTimer = setInterval(fetchAll, 60000)
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<template>
  <!-- 历史趋势全屏页 -->
  <HistoryTrend v-if="showHistory" @back="showHistory = false" />

  <!-- 大盘看板 -->
  <section v-else class="dashboard">
    <div class="dashboard__header">
      <h2 class="dashboard__title">📈 大盘看板</h2>
      <div class="dashboard__right">
        <button class="dashboard__history" @click="showHistory = true">看历史趋势</button>
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
}
</style>

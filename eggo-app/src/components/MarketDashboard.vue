<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { get, getCache, setCache } from '@/utils/request'
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

// 先读缓存立即展示，再后台静默拉新数据
async function fetchIndices() {
  const cached = getCache<IndexData[]>('/market/indices')
  if (cached) indices.value = cached

  try {
    const res = await get<IndexData[]>('/market/indices')
    if (res.code === 0 && res.data) {
      indices.value = res.data
      setCache('/market/indices', res.data)
    }
  } catch (err) {
    console.error('Failed to fetch indices:', err)
  }
}

async function fetchSectors() {
  const cached = getCache<SectorData[]>('/market/sectors')
  if (cached) sectors.value = cached

  try {
    const res = await get<SectorData[]>('/market/sectors')
    if (res.code === 0 && res.data) {
      sectors.value = res.data
      setCache('/market/sectors', res.data)
    }
  } catch (err) {
    console.error('Failed to fetch sectors:', err)
  }
}

async function fetchConcepts() {
  const cached = getCache<SectorData[]>('/market/concepts')
  if (cached) concepts.value = cached

  try {
    const res = await get<SectorData[]>('/market/concepts')
    if (res.code === 0 && res.data) {
      concepts.value = res.data
      setCache('/market/concepts', res.data)
    }
  } catch (err) {
    console.error('Failed to fetch concepts:', err)
  }
}

async function fetchOverview() {
  const cached = getCache<MarketStats>('/market/overview')
  if (cached) marketStats.value = cached

  try {
    const res = await get<MarketStats>('/market/overview')
    if (res.code === 0 && res.data) {
      marketStats.value = res.data
      setCache('/market/overview', res.data)
    }
  } catch (err) {
    console.error('Failed to fetch overview:', err)
  }
}

async function fetchAll() {
  // 有缓存时不显示全局 loading，只在后台刷新
  const hasCache = !!getCache('/market/indices')
  if (!hasCache) loading.value = true
  await Promise.all([fetchIndices(), fetchSectors(), fetchConcepts(), fetchOverview()])
  loading.value = false
}

function returnColor(val: number): string {
  if (val > 0) return '#ff5c5e'
  if (val < 0) return '#00d68f'
  return '#787878'
}

function returnArrow(val: number): string {
  if (val > 0) return '↑'
  if (val < 0) return '↓'
  return '→'
}

const BENCHMARK_MAP: Record<string, string> = {
  '1.000001': '沪深主动基金常见参考基准',
  '0.399001': '深市主动基金常见参考基准',
  '0.399006': '创业板主题基金业绩基准',
  '1.000300': '沪深300指数基金 / 大部分主动基金业绩基准',
  '1.000905': '中证500指数基金 / 中小盘基金基准',
  '1.000852': '中证1000指数基金 / 小盘基金基准',
}

function indexBenchmark(code: string): string {
  return BENCHMARK_MAP[code] ?? ''
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
        <div class="index-card__benchmark">{{ indexBenchmark(idx.code) }}</div>
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
      <div v-if="sectors.length === 0 && !loading" class="empty-hint">
        <span class="empty-hint__icon">📡</span>
        <span class="empty-hint__text">暂无板块数据</span>
        <span class="empty-hint__sub">境外服务器暂不支持 A 股板块行情</span>
      </div>
    </div>

    <!-- 概念板块 -->
    <div v-if="activeTab === 'concept'" class="sectors-list">
      <div v-for="(s, i) in concepts" :key="s.code" class="sector-row">
        <div class="sector-rank" :style="{ color: i < 3 ? '#ffd700' : '#787878' }">{{ i + 1 }}</div>
        <div class="sector-info">
          <div class="sector-name">{{ s.name }}</div>
          <div class="sector-reason" v-if="s.price">最新价: {{ s.price.toFixed(2) }}</div>
        </div>
        <div class="sector-change" :style="{ color: returnColor(s.changePercent) }">
          {{ s.changePercent > 0 ? '+' : '' }}{{ s.changePercent.toFixed(2) }}%
        </div>
      </div>
      <div v-if="concepts.length === 0 && !loading" class="empty-hint">
        <span class="empty-hint__icon">📡</span>
        <span class="empty-hint__text">暂无概念数据</span>
        <span class="empty-hint__sub">境外服务器暂不支持 A 股概念行情</span>
      </div>
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

/* 按钮统一金色主色 */
.dashboard__history {
  padding: 6px 12px;
  border-radius: 8px;
  border: 1px solid rgba(255, 215, 0, 0.4);
  background: rgba(255, 215, 0, 0.08);
  color: var(--accent);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.dashboard__history:hover {
  background: rgba(255, 215, 0, 0.18);
  border-color: rgba(255, 215, 0, 0.7);
  box-shadow: 0 0 8px rgba(255, 215, 0, 0.2);
}

.dashboard__refresh {
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

.dashboard__refresh:hover {
  background: rgba(255, 215, 0, 0.18);
  border-color: rgba(255, 215, 0, 0.7);
  box-shadow: 0 0 8px rgba(255, 215, 0, 0.2);
}

.dashboard__refresh:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.spin {
  display: inline-block;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Tabs — 金色主色 */
.tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.tab {
  padding: 7px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  background: transparent;
  color: var(--text-tertiary);   /* 辅助文字级，AA 合格 */
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.tab:hover {
  color: var(--text-secondary);  /* 次要数据级 */
  border-color: rgba(255, 215, 0, 0.3);
}

.tab--active {
  background: rgba(255, 215, 0, 0.1);
  border-color: rgba(255, 215, 0, 0.5);
  color: var(--accent);          /* 核心数据级，金色 */
  font-weight: 600;
}

/* ── 指数卡片：2×3 网格，深灰渐变背景 + 描边，hover 金色高亮 ── */
.indices-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.index-card {
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;       /* 深灰描边 */
  border-radius: 16px;
  padding: 16px;
  transition: border-color 0.2s, box-shadow 0.2s;
  cursor: default;
}

.index-card:hover {
  border-color: #f7ba1e;
  box-shadow: 0 0 12px rgba(247, 186, 30, 0.15);
}

.index-card__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.index-card__name {
  font-size: 13px;
  color: var(--text-secondary);   /* 次要数据级 */
  font-weight: 600;
}

.index-card__code {
  font-size: 11px;
  color: var(--text-tertiary);    /* 辅助文字级 */
  font-family: var(--font-mono);
}

/* 价格：核心数据，由 returnColor 动态设红/绿，字号最大 */
.index-card__price {
  font-size: 22px;
  font-weight: 700;
  font-family: var(--font-mono);
  margin-bottom: 4px;
  line-height: 1.2;
}

.index-card__change {
  display: flex;
  gap: 12px;
  font-size: 13px;
  font-family: var(--font-mono);
  font-weight: 600;
  margin-bottom: 12px;
}

.index-card__meta {
  display: flex;
  gap: 12px;
  padding-top: 10px;
  border-top: 1px solid #2a2a2a;
}

.index-card__benchmark {
  margin-top: 8px;
  font-size: 11px;
  color: #f7ba1e;
  opacity: 0.8;
  line-height: 1.3;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.meta-label {
  font-size: 11px;
  color: var(--text-tertiary);    /* 辅助文字级 */
}

.meta-value {
  font-size: 12px;
  color: var(--text-secondary);   /* 次要数据级 */
  font-family: var(--font-mono);
}

/* ── 板块列表 —— 与指数卡片统一背景和描边 ── */
.sectors-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sector-row {
  display: flex;
  align-items: center;
  gap: 12px;
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 12px;
  padding: 14px 16px;
  transition: border-color 0.2s;
}

.sector-row:hover {
  border-color: rgba(255, 215, 0, 0.35);
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
  color: var(--text-primary);     /* 核心数据级 */
  font-weight: 600;
}

.sector-reason {
  font-size: 12px;
  color: var(--text-tertiary);    /* 辅助文字级 */
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

/* ── 空状态 —— 与卡片统一背景描边，文字中灰，提示语金色 ── */
.empty-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 36px 16px;
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 16px;
  text-align: center;
}

.empty-hint__icon {
  font-size: 28px;
  line-height: 1;
  opacity: 0.6;
}

.empty-hint__text {
  font-size: 14px;
  color: var(--text-secondary);   /* 中灰 */
  font-weight: 600;
}

.empty-hint__sub {
  font-size: 12px;
  color: #f7ba1e; /* 金色提示，与 hover 色一致 */
}

/* ── 市场情绪面板 ── */
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
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 12px;
  padding: 16px;
  text-align: center;
  transition: border-color 0.2s;
}

.stat-card:hover {
  border-color: rgba(255, 215, 0, 0.35);
}

.stat-card__value {
  font-size: 24px;
  font-weight: 700;
  font-family: var(--font-mono);
  margin-bottom: 4px;
  color: var(--text-primary);
}

.stat-card--up   .stat-card__value { color: #ff5c5e; }
.stat-card--down .stat-card__value { color: #00d68f; }
.stat-card--flat .stat-card__value { color: var(--text-tertiary); }

.stat-card__label {
  font-size: 12px;
  color: var(--text-tertiary);    /* 辅助文字级 */
}

.stats-detail {
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
}

.detail-row + .detail-row {
  border-top: 1px solid #2a2a2a;
}

.detail-label {
  font-size: 13px;
  color: var(--text-tertiary);    /* 辅助文字级 */
}

.detail-value {
  font-size: 14px;
  font-weight: 600;
  font-family: var(--font-mono);
  color: var(--text-secondary);   /* 次要数据级，特殊值由内联 style 覆盖 */
}

/* Bar Chart */
.bar-chart {
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
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
  background: #2a2a2a;
}

.bar-segment {
  transition: width 0.3s;
}

.bar-segment--up   { background: #ff5c5e; }
.bar-segment--flat { background: #787878; }
.bar-segment--down { background: #00d68f; }

.bar-legend {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: var(--text-secondary);   /* 次要数据级，可读性更强 */
}

.dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 4px;
  vertical-align: middle;
}

.dot--up   { background: #ff5c5e; }
.dot--flat { background: #787878; }
.dot--down { background: #00d68f; }

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

<script setup lang="ts">
import { ref, computed } from 'vue'

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
  changePercent: number
  leader: string
  reason: string
}

const indices = ref<IndexData[]>([
  { name: '上证指数', code: '000001', price: 3387.52, change: 28.36, changePercent: 0.84, volume: '4823亿', high: 3395.18, low: 3362.41 },
  { name: '深证成指', code: '399001', price: 10245.68, change: -45.23, changePercent: -0.44, volume: '5612亿', high: 10312.55, low: 10198.32 },
  { name: '创业板指', code: '399006', price: 2078.33, change: 15.67, changePercent: 0.76, volume: '2341亿', high: 2089.44, low: 2058.12 },
  { name: '沪深300', code: '000300', price: 3956.78, change: 12.45, changePercent: 0.32, volume: '3215亿', high: 3968.22, low: 3938.56 },
  { name: '中证500', code: '000905', price: 6123.45, change: -38.92, changePercent: -0.63, volume: '1876亿', high: 6178.33, low: 6098.21 },
  { name: '中证1000', code: '000852', price: 6534.21, change: 52.18, changePercent: 0.81, volume: '1543亿', high: 6558.44, low: 6478.33 },
])

const sectors = ref<SectorData[]>([
  { name: 'AI算力', changePercent: 3.28, leader: '中际旭创', reason: '海外大厂资本开支超预期' },
  { name: '半导体', changePercent: 2.15, leader: '北方华创', reason: '国产替代加速' },
  { name: '新能源车', changePercent: 1.87, leader: '比亚迪', reason: '5月销量数据亮眼' },
  { name: '白酒', changePercent: -1.23, leader: '贵州茅台', reason: '消费复苏低于预期' },
  { name: '医药', changePercent: -0.89, leader: '恒瑞医药', reason: '集采压力持续' },
  { name: '银行', changePercent: 0.45, leader: '招商银行', reason: '高股息防御属性' },
])

const marketStats = ref({
  upCount: 2856,
  downCount: 1823,
  flatCount: 345,
  limitUp: 68,
  limitDown: 12,
  totalVolume: '1.23万亿',
  northFlow: '+45.6亿',
  sentiment: '偏多',
})

const activeTab = ref<'index' | 'sector' | 'stats'>('index')

function returnColor(val: number): string {
  if (val > 0) return '#ff4d4f'
  if (val < 0) return '#00c3ff'
  return '#666'
}

function returnArrow(val: number): string {
  if (val > 0) return '↑'
  if (val < 0) return '↓'
  return '→'
}

const updateTime = ref('2026-06-14 15:00:00')
</script>

<template>
  <section class="dashboard">
    <div class="dashboard__header">
      <h2 class="dashboard__title">📈 大盘看板</h2>
      <span class="dashboard__time">{{ updateTime }}</span>
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
      >板块热点</button>
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

    <!-- 板块热点 -->
    <div v-if="activeTab === 'sector'" class="sectors-list">
      <div v-for="(s, i) in sectors" :key="s.name" class="sector-row">
        <div class="sector-rank" :style="{ color: i < 3 ? '#ffd700' : '#666' }">{{ i + 1 }}</div>
        <div class="sector-info">
          <div class="sector-name">{{ s.name }}</div>
          <div class="sector-reason">{{ s.reason }}</div>
        </div>
        <div class="sector-leader">
          <span class="sector-leader__label">龙头</span>
          <span class="sector-leader__name">{{ s.leader }}</span>
        </div>
        <div class="sector-change" :style="{ color: returnColor(s.changePercent) }">
          {{ s.changePercent > 0 ? '+' : '' }}{{ s.changePercent.toFixed(2) }}%
        </div>
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
          <div class="stat-card__value" style="color: #00c3ff">{{ marketStats.limitDown }}</div>
          <div class="stat-card__label">跌停</div>
        </div>
      </div>

      <div class="stats-detail">
        <div class="detail-row">
          <span class="detail-label">两市成交</span>
          <span class="detail-value">{{ marketStats.totalVolume }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">北向资金</span>
          <span class="detail-value" style="color: #ff4d4f">{{ marketStats.northFlow }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">市场情绪</span>
          <span class="detail-value" style="color: #ffd700">{{ marketStats.sentiment }}</span>
        </div>
      </div>

      <!-- 涨跌分布条 -->
      <div class="bar-chart">
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

.dashboard__time {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

/* Tabs */
.tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
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

.sector-leader {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.sector-leader__label {
  font-size: 10px;
  color: var(--text-tertiary);
}

.sector-leader__name {
  font-size: 13px;
  color: var(--text-secondary);
}

.sector-change {
  font-size: 15px;
  font-weight: 700;
  font-family: var(--font-mono);
  min-width: 72px;
  text-align: right;
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
.stat-card--down .stat-card__value { color: #00c3ff; }
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
.bar-segment--down { background: #00c3ff; }

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
.dot--down { background: #00c3ff; }

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

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { get } from '@/utils/request'

interface FundDetailData {
  fundCode: string
  fundName: string
  unitNav: number
  accNav?: number
  dailyReturn: number
  navDate: string
  weekReturn: number
  monthReturn?: number
  quarterReturn?: number
  yearReturn?: number
  consecutiveUp: number
  consecutiveDown: number
  hasNegNews: boolean
  hasPosPolicy: boolean
  sentimentCool: boolean
  maxDrawdown?: number
  fundType?: string
  manager?: string
  custodian?: string
  inceptionDate?: string
  benchmark?: string
}

interface NavPoint {
  date: string
  unitNav: number
  accNav: number
  dailyReturn: number
}

interface NavHistoryData {
  fundCode: string
  fundName: string
  points: NavPoint[]
}

interface FundAnalysis {
  fundCode: string
  fundName: string
  signalType: string
  signalText: string
  signalEmoji: string
  confidence: number
  reason: string
  trendLabel: string
  trendColor: string
  suggestions: string[]
}

interface AnalysisResponse {
  fundCode: string
  fundName: string
  analysis: FundAnalysis
  generatedAt: string
}

const props = defineProps<{
  fundCode: string
  fundName: string
}>()

const emit = defineEmits<{
  (e: 'back'): void
}>()

const detail = ref<FundDetailData | null>(null)
const navHistory = ref<NavHistoryData | null>(null)
const analysis = ref<AnalysisResponse | null>(null)
const loadingDetail = ref(true)
const loadingNav = ref(true)
const loadingAnalysis = ref(true)
const activeRange = ref('3m')

const ranges = [
  { key: '1w', label: '近1周', days: 5 },
  { key: '1m', label: '近1月', days: 22 },
  { key: '3m', label: '近3月', days: 66 },
  { key: '6m', label: '近6月', days: 132 },
  { key: '1y', label: '近1年', days: 252 },
]

async function fetchDetail() {
  loadingDetail.value = true
  try {
    const res = await get<FundDetailData>(`/funds/${props.fundCode}/detail`)
    if (res.code === 0 && res.data) {
      detail.value = res.data
    }
  } catch (e) {
    console.error('Failed to fetch fund detail:', e)
  } finally {
    loadingDetail.value = false
  }
}

async function fetchNavHistory() {
  loadingNav.value = true
  const days = ranges.find(r => r.key === activeRange.value)?.days || 66
  try {
    const res = await get<NavHistoryData>(`/funds/${props.fundCode}/nav-history`, { days })
    if (res.code === 0 && res.data) {
      navHistory.value = res.data
    }
  } catch (e) {
    console.error('Failed to fetch nav history:', e)
  } finally {
    loadingNav.value = false
  }
}

async function fetchAnalysis() {
  loadingAnalysis.value = true
  try {
    const res = await get<AnalysisResponse>(`/funds/${props.fundCode}/analysis`)
    if (res.code === 0 && res.data) {
      analysis.value = res.data
    }
  } catch (e) {
    console.error('Failed to fetch analysis:', e)
  } finally {
    loadingAnalysis.value = false
  }
}

function returnColor(val: number): string {
  if (val > 0) return '#ff4d4f'
  if (val < 0) return '#00d68f'
  return '#787878'
}

function returnArrow(val: number): string {
  if (val > 0) return '↑'
  if (val < 0) return '↓'
  return '→'
}

function formatReturn(val: number): string {
  const sign = val > 0 ? '+' : ''
  return `${sign}${val.toFixed(2)}%`
}

function formatNav(val: number): string {
  return val.toFixed(4)
}

// ── NAV Chart (SVG) ──
const W = 360
const H = 180
const PAD = { top: 16, right: 8, bottom: 24, left: 8 }
const chartW = W - PAD.left - PAD.right
const chartH = H - PAD.top - PAD.bottom

const chartPoints = computed(() => {
  const data = navHistory.value
  if (!data || data.points.length < 2) return []
  const pts = data.points.map(p => p.unitNav)
  const min = Math.min(...pts)
  const max = Math.max(...pts)
  const range = max - min || 1
  const padding = range * 0.05
  return data.points.map((p, i) => ({
    x: PAD.left + (i / (data.points.length - 1)) * chartW,
    y: PAD.top + (1 - (p.unitNav - min + padding) / (range + padding * 2)) * chartH,
  }))
})

const svgPath = computed(() => {
  if (chartPoints.value.length < 2) return ''
  return 'M' + chartPoints.value.map(p => `${p.x.toFixed(1)},${p.y.toFixed(1)}`).join(' L')
})

const areaPath = computed(() => {
  if (chartPoints.value.length < 2) return ''
  const first = chartPoints.value[0]
  const last = chartPoints.value[chartPoints.value.length - 1]
  const bottomY = PAD.top + chartH
  return svgPath.value + ` L${last.x.toFixed(1)},${bottomY} L${first.x.toFixed(1)},${bottomY} Z`
})

const chartColor = computed(() => {
  if (!navHistory.value || navHistory.value.points.length < 2) return '#787878'
  const first = navHistory.value.points[0].unitNav
  const last = navHistory.value.points[navHistory.value.points.length - 1].unitNav
  if (last > first) return '#ff4d4f'
  if (last < first) return '#00d68f'
  return '#787878'
})

const periodReturn = computed(() => {
  if (!navHistory.value || navHistory.value.points.length < 2) return { val: '--', color: '#787878' }
  const first = navHistory.value.points[0].unitNav
  const last = navHistory.value.points[navHistory.value.points.length - 1].unitNav
  const pct = first > 0 ? ((last - first) / first * 100) : 0
  const sign = pct > 0 ? '+' : ''
  return {
    val: `${sign}${pct.toFixed(2)}%`,
    color: pct > 0 ? '#ff4d4f' : pct < 0 ? '#00d68f' : '#787878',
  }
})

const startNav = computed(() => {
  if (!navHistory.value || navHistory.value.points.length < 2) return '--'
  return navHistory.value.points[0].unitNav.toFixed(4)
})

const endNav = computed(() => {
  if (!navHistory.value || navHistory.value.points.length < 2) return '--'
  return navHistory.value.points[navHistory.value.points.length - 1].unitNav.toFixed(4)
})

const navStartDate = computed(() => {
  if (!navHistory.value || navHistory.value.points.length < 2) return '--'
  return navHistory.value.points[0].date
})

const navEndDate = computed(() => {
  if (!navHistory.value || navHistory.value.points.length < 2) return '--'
  return navHistory.value.points[navHistory.value.points.length - 1].date
})

// 分析建议卡片的样式
const signalStyle = computed(() => {
  if (!analysis.value) return { bg: 'rgba(255,255,255,0.05)', border: 'var(--border-color)', icon: '⚪' }
  const type = analysis.value.analysis.signalType
  switch (type) {
    case 'buy':
      return { bg: 'rgba(0, 214, 143, 0.1)', border: 'rgba(0, 214, 143, 0.3)', icon: '📈' }
    case 'sell':
      return { bg: 'rgba(255, 77, 79, 0.1)', border: 'rgba(255, 77, 79, 0.3)', icon: '📉' }
    case 'hold':
      return { bg: 'rgba(255, 215, 0, 0.1)', border: 'rgba(255, 215, 0, 0.3)', icon: '🤝' }
    default:
      return { bg: 'rgba(255,255,255,0.03)', border: 'var(--border-color)', icon: '👁️' }
  }
})

watch(activeRange, () => {
  fetchNavHistory()
})

onMounted(() => {
  fetchDetail()
  fetchNavHistory()
  fetchAnalysis()
})
</script>

<template>
  <div class="fund-detail-page">
    <!-- 顶部导航 -->
    <div class="detail-nav">
      <button class="back-btn" @click="emit('back')">← 返回</button>
      <div class="detail-nav__info">
        <span class="detail-nav__name">{{ fundName }}</span>
        <span class="detail-nav__code">{{ fundCode }}</span>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loadingDetail" class="loading-state">加载中...</div>

    <template v-else-if="detail">
      <!-- 价格摘要 -->
      <div class="price-section">
        <div class="price-section__main">
          <div class="price-section__nav">
            <span class="price-section__nav-value">{{ formatNav(detail.unitNav) }}</span>
            <span class="price-section__nav-label">单位净值</span>
          </div>
          <div class="price-section__change" :style="{ color: returnColor(detail.dailyReturn) }">
            <span class="price-section__change-arrow">{{ returnArrow(detail.dailyReturn) }}</span>
            <span class="price-section__change-val">{{ Math.abs(detail.dailyReturn).toFixed(2) }}%</span>
          </div>
        </div>
        <div class="price-section__meta">
          <span>净值日期: {{ detail.navDate }}</span>
          <span v-if="detail.accNav">累计净值: {{ formatNav(detail.accNav) }}</span>
        </div>
      </div>

      <!-- NAV 趋势图 -->
      <div class="chart-section">
        <div class="chart-section__header">
          <h3 class="chart-section__title">净值走势</h3>
          <div class="chart-section__range">
            <button
              v-for="r in ranges"
              :key="r.key"
              :class="['range-btn', { 'range-btn--active': activeRange === r.key }]"
              @click="activeRange = r.key"
            >{{ r.label }}</button>
          </div>
        </div>

        <div class="chart-container">
          <div v-if="loadingNav" class="chart-loading">加载中...</div>
          <div v-else-if="!navHistory || navHistory.points.length < 2" class="chart-empty">暂无净值数据</div>
          <svg v-else :viewBox="`0 0 ${W} ${H}`" class="chart-svg">
            <defs>
              <linearGradient id="navGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" :stop-color="chartColor" stop-opacity="0.2" />
                <stop offset="100%" :stop-color="chartColor" stop-opacity="0.02" />
              </linearGradient>
            </defs>
            <path :d="areaPath" fill="url(#navGrad)" />
            <polyline :points="svgPath" fill="none" :stroke="chartColor" stroke-width="2" stroke-linejoin="round" stroke-linecap="round" />
          </svg>
        </div>

        <div class="chart-footer" v-if="navHistory && navHistory.points.length >= 2">
          <span>起始: {{ navStartDate }} {{ startNav }}</span>
          <span :style="{ color: periodReturn.color }">区间: {{ periodReturn.val }}</span>
          <span>截止: {{ navEndDate }} {{ endNav }}</span>
        </div>
      </div>

      <!-- 关键指标 -->
      <div class="metrics-section">
        <h3 class="metrics-section__title">关键指标</h3>
        <div class="metrics-grid">
          <div class="metric-item">
            <span class="metric-item__label">近1周</span>
            <span class="metric-item__val" :style="{ color: returnColor(detail.weekReturn) }">{{ formatReturn(detail.weekReturn) }}</span>
          </div>
          <div class="metric-item" v-if="detail.monthReturn !== undefined">
            <span class="metric-item__label">近1月</span>
            <span class="metric-item__val" :style="{ color: returnColor(detail.monthReturn) }">{{ formatReturn(detail.monthReturn) }}</span>
          </div>
          <div class="metric-item" v-if="detail.quarterReturn !== undefined">
            <span class="metric-item__label">近3月</span>
            <span class="metric-item__val" :style="{ color: returnColor(detail.quarterReturn) }">{{ formatReturn(detail.quarterReturn) }}</span>
          </div>
          <div class="metric-item" v-if="detail.yearReturn !== undefined">
            <span class="metric-item__label">近1年</span>
            <span class="metric-item__val" :style="{ color: returnColor(detail.yearReturn) }">{{ formatReturn(detail.yearReturn) }}</span>
          </div>
          <div class="metric-item">
            <span class="metric-item__label">连涨</span>
            <span class="metric-item__val" style="color: #ff4d4f">{{ detail.consecutiveUp }}天</span>
          </div>
          <div class="metric-item">
            <span class="metric-item__label">连跌</span>
            <span class="metric-item__val" style="color: #00d68f">{{ detail.consecutiveDown }}天</span>
          </div>
          <div class="metric-item" v-if="detail.maxDrawdown !== undefined">
            <span class="metric-item__label">最大回撤</span>
            <span class="metric-item__val" style="color: #f7ba1e">{{ detail.maxDrawdown.toFixed(2) }}%</span>
          </div>
        </div>
      </div>

      <!-- AI 分析建议 -->
      <div class="signal-section" v-if="analysis">
        <h3 class="signal-section__title">🧠 AI 智能分析</h3>
        <div
          class="signal-card"
          :style="{
            background: signalStyle.bg,
            borderColor: signalStyle.border,
          }"
        >
          <div class="signal-card__header">
            <span class="signal-card__icon">{{ analysis.analysis.signalEmoji }}</span>
            <div class="signal-card__info">
              <span class="signal-card__signal" :style="{ color: analysis.analysis.trendColor }">
                {{ analysis.analysis.signalText }}
              </span>
              <span class="signal-card__trend">
                趋势判断: <strong :style="{ color: analysis.analysis.trendColor }">{{ analysis.analysis.trendLabel }}</strong>
              </span>
            </div>
            <span class="signal-card__confidence">
              可信度 {{ (analysis.analysis.confidence * 100).toFixed(0) }}%
            </span>
          </div>

          <div class="signal-card__reason">
            <span class="signal-card__reason-label">分析依据</span>
            <p class="signal-card__reason-text">{{ analysis.analysis.reason }}</p>
          </div>

          <div class="signal-card__suggestions">
            <span class="signal-card__suggestions-label">操作建议</span>
            <ul class="signal-card__suggestions-list">
              <li v-for="(item, i) in analysis.analysis.suggestions" :key="i">{{ item }}</li>
            </ul>
          </div>
        </div>
      </div>

      <!-- 基金基本信息 -->
      <div class="info-section" v-if="detail.fundType || detail.manager || detail.custodian">
        <h3 class="info-section__title">基金档案</h3>
        <div class="info-grid">
          <div class="info-row" v-if="detail.fundType">
            <span class="info-row__label">基金类型</span>
            <span class="info-row__val">{{ detail.fundType }}</span>
          </div>
          <div class="info-row" v-if="detail.manager">
            <span class="info-row__label">基金经理</span>
            <span class="info-row__val">{{ detail.manager }}</span>
          </div>
          <div class="info-row" v-if="detail.custodian">
            <span class="info-row__label">托管人</span>
            <span class="info-row__val">{{ detail.custodian }}</span>
          </div>
          <div class="info-row" v-if="detail.inceptionDate">
            <span class="info-row__label">成立日期</span>
            <span class="info-row__val">{{ detail.inceptionDate }}</span>
          </div>
          <div class="info-row" v-if="detail.benchmark">
            <span class="info-row__label">业绩基准</span>
            <span class="info-row__val">{{ detail.benchmark }}</span>
          </div>
        </div>
      </div>
    </template>

    <div v-else class="error-state">获取基金详情失败</div>
  </div>
</template>

<style scoped>
.fund-detail-page {
  max-width: 800px;
  margin: 0 auto;
  padding: 24px 16px;
}

/* ── 导航 ── */
.detail-nav {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.back-btn {
  padding: 8px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  background: transparent;
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.back-btn:hover {
  border-color: rgba(255, 215, 0, 0.3);
  color: var(--accent);
}

.detail-nav__info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.detail-nav__name {
  font-size: 16px;
  color: var(--text-primary);
  font-weight: 700;
}

.detail-nav__code {
  font-size: 12px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

/* ── 价格摘要 ── */
.price-section {
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 16px;
  padding: 20px;
  margin-bottom: 16px;
}

.price-section__main {
  display: flex;
  align-items: baseline;
  gap: 16px;
  margin-bottom: 8px;
}

.price-section__nav-value {
  font-size: 32px;
  font-weight: 700;
  color: var(--text-primary);
  font-family: var(--font-mono);
  line-height: 1;
}

.price-section__nav-label {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-left: 4px;
}

.price-section__change {
  font-size: 18px;
  font-weight: 700;
  font-family: var(--font-mono);
}

.price-section__change-arrow {
  margin-right: 2px;
}

.price-section__meta {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

/* ── 图表 ── */
.chart-section {
  margin-bottom: 16px;
}

.chart-section__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.chart-section__title {
  font-size: 14px;
  color: var(--text-secondary);
  font-weight: 600;
  margin: 0;
}

.chart-section__range {
  display: flex;
  gap: 4px;
}

.range-btn {
  padding: 4px 10px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  background: transparent;
  color: var(--text-tertiary);
  font-size: 11px;
  cursor: pointer;
  transition: all 0.2s;
}

.range-btn:hover {
  color: var(--text-secondary);
  border-color: rgba(255, 215, 0, 0.3);
}

.range-btn--active {
  background: rgba(255, 215, 0, 0.1);
  border-color: rgba(255, 215, 0, 0.5);
  color: var(--accent);
}

.chart-container {
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 12px;
  padding: 8px;
}

.chart-loading,
.chart-empty {
  text-align: center;
  padding: 40px;
  font-size: 13px;
  color: var(--text-tertiary);
}

.chart-svg {
  width: 100%;
  height: auto;
  display: block;
}

.chart-footer {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
  margin-top: 8px;
  padding: 0 4px;
}

/* ── 关键指标 ── */
.metrics-section {
  margin-bottom: 16px;
}

.metrics-section__title {
  font-size: 14px;
  color: var(--text-secondary);
  font-weight: 600;
  margin: 0 0 12px;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}

.metric-item {
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 12px;
  padding: 14px 12px;
  text-align: center;
}

.metric-item__label {
  display: block;
  font-size: 11px;
  color: var(--text-tertiary);
  margin-bottom: 4px;
}

.metric-item__val {
  font-size: 16px;
  font-weight: 700;
  font-family: var(--font-mono);
}

/* ── AI 分析建议 ── */
.signal-section {
  margin-bottom: 16px;
}

.signal-section__title {
  font-size: 14px;
  color: var(--text-secondary);
  font-weight: 600;
  margin: 0 0 12px;
}

.signal-card {
  border: 1px solid;
  border-radius: 16px;
  padding: 20px;
}

.signal-card__header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}

.signal-card__icon {
  font-size: 28px;
  line-height: 1;
}

.signal-card__info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.signal-card__signal {
  font-size: 18px;
  font-weight: 700;
}

.signal-card__trend {
  font-size: 13px;
  color: var(--text-tertiary);
}

.signal-card__confidence {
  font-size: 11px;
  color: var(--text-tertiary);
  white-space: nowrap;
  padding: 3px 8px;
  background: rgba(255,255,255,0.05);
  border-radius: 6px;
}

.signal-card__reason {
  background: rgba(255,215,0,0.04);
  border: 1px solid rgba(255,215,0,0.1);
  border-radius: 10px;
  padding: 12px 16px;
  margin-bottom: 16px;
}

.signal-card__reason-label {
  display: block;
  font-size: 11px;
  color: var(--text-tertiary);
  margin-bottom: 4px;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.signal-card__reason-text {
  font-size: 13px;
  color: var(--accent);
  line-height: 1.6;
  margin: 0;
}

.signal-card__suggestions-label {
  display: block;
  font-size: 11px;
  color: var(--text-tertiary);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.signal-card__suggestions-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.signal-card__suggestions-list li {
  position: relative;
  padding: 6px 0 6px 16px;
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
}

.signal-card__suggestions-list li::before {
  content: '›';
  position: absolute;
  left: 0;
  color: var(--accent);
  font-weight: 700;
}

/* ── 基金档案 ── */
.info-section {
  margin-bottom: 16px;
}

.info-section__title {
  font-size: 14px;
  color: var(--text-secondary);
  font-weight: 600;
  margin: 0 0 12px;
}

.info-grid {
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 12px;
  overflow: hidden;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
}

.info-row + .info-row {
  border-top: 1px solid #2a2a2a;
}

.info-row__label {
  font-size: 13px;
  color: var(--text-tertiary);
}

.info-row__val {
  font-size: 13px;
  color: var(--text-secondary);
  text-align: right;
  max-width: 60%;
  word-break: break-all;
}

/* ── 状态 ── */
.loading-state,
.error-state {
  text-align: center;
  padding: 60px 16px;
  font-size: 14px;
  color: var(--text-tertiary);
}

@media (max-width: 480px) {
  .metrics-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .price-section__nav-value {
    font-size: 26px;
  }

  .chart-footer {
    flex-direction: column;
    gap: 4px;
  }
}
</style>

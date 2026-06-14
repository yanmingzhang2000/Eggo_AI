<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { get } from '@/utils/request'

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

const emit = defineEmits<{
  (e: 'back'): void
}>()

const indexOptions = ref<IndexOption[]>([])
const selectedCodes = ref<string[]>(['1.000300'])
const activeRange = ref('3m')
const historyData = ref<Record<string, KLineData[]>>({})
const loading = ref(false)

const ranges = [
  { key: '1w', label: '近1周', days: 5 },
  { key: '1m', label: '近1月', days: 22 },
  { key: '3m', label: '近3月', days: 66 },
  { key: '6m', label: '近6月', days: 132 },
  { key: '1y', label: '近1年', days: 252 },
]

const currentDays = computed(() => {
  return ranges.find(r => r.key === activeRange.value)?.days || 66
})

const colors = ['#ff4d4f', '#00b96b', '#ffd700', '#9b59b6', '#3498db', '#e67e22']

const allDates = computed(() => {
  const dateSet = new Set<string>()
  for (const code of selectedCodes.value) {
    const data = historyData.value[code]
    if (data) {
      data.forEach(d => dateSet.add(d.date))
    }
  }
  return Array.from(dateSet).sort()
})

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
  loading.value = true
  try {
    const promises = selectedCodes.value.map(async (code) => {
      const res = await get<KLineData[]>('/market/history', { code, days: currentDays.value })
      if (res.code === 0 && res.data) {
        historyData.value[code] = res.data
      }
    })
    await Promise.all(promises)
  } catch (err) {
    console.error('Failed to fetch history:', err)
  } finally {
    loading.value = false
  }
}

function toggleIndex(code: string) {
  const idx = selectedCodes.value.indexOf(code)
  if (idx >= 0) {
    selectedCodes.value.splice(idx, 1)
  } else {
    selectedCodes.value.push(code)
  }
}

function isSelected(code: string): boolean {
  return selectedCodes.value.includes(code)
}

function selectRange(key: string) {
  activeRange.value = key
}

// 归一化数据：将不同指数的价格映射到 0-100 的相对值，便于对比
function normalizeData(code: string): { x: number; y: number }[] {
  const data = historyData.value[code]
  if (!data || data.length < 2) return []

  const closes = data.map(d => d.close)
  const min = Math.min(...closes)
  const max = Math.max(...closes)
  const range = max - min || 1

  return closes.map((val, i) => ({
    x: i / (closes.length - 1),
    y: 1 - (val - min) / range,
  }))
}

function generatePath(code: string): string {
  const points = normalizeData(code)
  if (points.length < 2) return ''

  const width = 700
  const height = 350
  const padding = { top: 20, bottom: 40, left: 60, right: 30 }
  const chartW = width - padding.left - padding.right
  const chartH = height - padding.top - padding.bottom

  const pathPoints = points.map(p => {
    const x = padding.left + p.x * chartW
    const y = padding.top + p.y * chartH
    return `${x},${y}`
  })

  return `M${pathPoints.join(' L')}`
}

function generateFillPath(code: string): string {
  const points = normalizeData(code)
  if (points.length < 2) return ''

  const width = 700
  const height = 350
  const padding = { top: 20, bottom: 40, left: 60, right: 30 }
  const chartW = width - padding.left - padding.right
  const chartH = height - padding.top - padding.bottom

  const pathPoints = points.map(p => {
    const x = padding.left + p.x * chartW
    const y = padding.top + p.y * chartH
    return `${x},${y}`
  })

  const firstX = padding.left
  const lastX = padding.left + chartW
  const bottomY = padding.top + chartH

  return `M${pathPoints.join(' L')} L${lastX},${bottomY} L${firstX},${bottomY} Z`
}

function getLineColor(code: string): string {
  const idx = selectedCodes.value.indexOf(code)
  return colors[idx % colors.length]
}

function getIndexName(code: string): string {
  return indexOptions.value.find(o => o.code === code)?.name || code
}

function getChartYLabels(): { value: string; y: number }[] {
  const height = 350
  const padding = { top: 20, bottom: 40 }
  const chartH = height - padding.top - padding.bottom
  const labels = []
  for (let i = 0; i <= 4; i++) {
    const pct = (4 - i) * 25
    const y = padding.top + (i / 4) * chartH
    labels.push({ value: `${pct}%`, y })
  }
  return labels
}

function getChartXLabels(): { date: string; x: number }[] {
  if (allDates.value.length < 2) return []
  const width = 700
  const padding = { left: 60, right: 30 }
  const chartW = width - padding.left - padding.right
  const step = Math.max(1, Math.floor(allDates.value.length / 6))

  const labels = []
  for (let i = 0; i < allDates.value.length; i += step) {
    const x = padding.left + (i / (allDates.value.length - 1)) * chartW
    const date = allDates.value[i].slice(5)
    labels.push({ date, x })
  }
  return labels
}

watch(selectedCodes, () => {
  fetchHistory()
}, { deep: true })

watch(activeRange, () => {
  fetchHistory()
})

onMounted(() => {
  fetchIndexOptions()
  fetchHistory()
})
</script>

<template>
  <div class="history-page">
    <header class="history-header">
      <button class="back-btn" @click="emit('back')">← 返回</button>
      <h2 class="history-title">📈 历史趋势</h2>
    </header>

    <!-- 指数选择 -->
    <div class="section">
      <div class="section-label">选择指数（可多选）</div>
      <div class="index-chips">
        <button
          v-for="opt in indexOptions"
          :key="opt.code"
          :class="['chip', { 'chip--active': isSelected(opt.code) }]"
          :style="isSelected(opt.code) ? { borderColor: getLineColor(opt.code), color: getLineColor(opt.code) } : {}"
          @click="toggleIndex(opt.code)"
        >{{ opt.name }}</button>
      </div>
    </div>

    <!-- 时间范围选择 -->
    <div class="section">
      <div class="section-label">时间范围</div>
      <div class="range-chips">
        <button
          v-for="r in ranges"
          :key="r.key"
          :class="['chip', { 'chip--active': activeRange === r.key }]"
          @click="selectRange(r.key)"
        >{{ r.label }}</button>
      </div>
    </div>

    <!-- 图表 -->
    <div class="chart-container">
      <div v-if="loading" class="chart-loading">加载中...</div>
      <div v-else-if="selectedCodes.length === 0" class="chart-empty">请选择至少一个指数</div>
      <svg v-else viewBox="0 0 700 350" class="chart-svg">
        <!-- Y 轴标签（百分比） -->
        <text
          v-for="(label, i) in getChartYLabels()"
          :key="'y'+i"
          :x="52"
          :y="label.y + 4"
          class="chart-label"
          text-anchor="end"
        >{{ label.value }}</text>

        <!-- 网格线 -->
        <line
          v-for="(label, i) in getChartYLabels()"
          :key="'g'+i"
          :x1="60"
          :x2="670"
          :y1="label.y"
          :y2="label.y"
          class="chart-grid"
        />

        <!-- 各指数折线 -->
        <g v-for="code in selectedCodes" :key="code">
          <path
            :d="generateFillPath(code)"
            :fill="getLineColor(code)"
            fill-opacity="0.05"
          />
          <path
            :d="generatePath(code)"
            fill="none"
            :stroke="getLineColor(code)"
            stroke-width="2"
            stroke-linecap="round"
          />
        </g>

        <!-- X 轴标签 -->
        <text
          v-for="(label, i) in getChartXLabels()"
          :key="'x'+i"
          :x="label.x"
          :y="330"
          class="chart-label"
          text-anchor="middle"
        >{{ label.date }}</text>
      </svg>
    </div>

    <!-- 图例 -->
    <div class="legend" v-if="selectedCodes.length > 0 && !loading">
      <div
        v-for="code in selectedCodes"
        :key="code"
        class="legend-item"
      >
        <span class="legend-dot" :style="{ background: getLineColor(code) }"></span>
        <span class="legend-name">{{ getIndexName(code) }}</span>
        <span class="legend-price" v-if="historyData[code]?.length > 0">
          {{ historyData[code][0].close.toFixed(2) }} → {{ historyData[code][historyData[code].length - 1].close.toFixed(2) }}
          <span :style="{ color: historyData[code][historyData[code].length - 1].close >= historyData[code][0].close ? '#ff4d4f' : '#00b96b' }">
            ({{ historyData[code][historyData[code].length - 1].close >= historyData[code][0].close ? '+' : '' }}{{ ((historyData[code][historyData[code].length - 1].close - historyData[code][0].close) / historyData[code][0].close * 100).toFixed(2) }}%)
          </span>
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.history-page {
  max-width: 800px;
  margin: 0 auto;
  padding: 24px 16px;
}

.history-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}

.back-btn {
  padding: 8px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  background: transparent;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.back-btn:hover {
  border-color: rgba(255, 215, 0, 0.3);
  color: var(--accent);
}

.history-title {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
  color: var(--text-primary);
}

.section {
  margin-bottom: 20px;
}

.section-label {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-bottom: 10px;
}

.index-chips,
.range-chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.chip {
  padding: 8px 16px;
  border-radius: 20px;
  border: 1px solid var(--border-color);
  background: transparent;
  color: var(--text-tertiary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.chip:hover {
  border-color: rgba(255, 215, 0, 0.3);
  color: var(--text-secondary);
}

.chip--active {
  background: rgba(255, 215, 0, 0.1);
  border-color: rgba(255, 215, 0, 0.3);
  color: var(--accent);
}

.chart-container {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  padding: 24px;
  min-height: 400px;
  margin-bottom: 20px;
}

.chart-loading,
.chart-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 350px;
  color: var(--text-tertiary);
  font-size: 14px;
}

.chart-svg {
  width: 100%;
  height: auto;
}

.chart-label {
  font-size: 10px;
  fill: var(--text-tertiary);
  font-family: var(--font-mono);
}

.chart-grid {
  stroke: var(--border-color);
  stroke-width: 0.5;
  stroke-dasharray: 2,2;
}

.legend {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 12px;
}

.legend-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.legend-name {
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 600;
  min-width: 70px;
}

.legend-price {
  font-size: 13px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}
</style>

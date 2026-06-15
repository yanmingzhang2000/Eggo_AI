<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { get, getCache, setCache } from '@/utils/request'

interface IntradayPoint {
  time: string
  price: number
  open: number
  high: number
  low: number
  vol: number
}

interface IntradayData {
  code: string
  name: string
  preClose: number
  points: IntradayPoint[]
  updateAt: string
}

const props = defineProps<{
  code: string
}>()

const emit = defineEmits<{
  (e: 'back'): void
}>()

const data = ref<IntradayData | null>(null)
const loading = ref(true)
let timer: ReturnType<typeof setInterval> | null = null

// SVG 尺寸
const W = 360
const H = 200
const PAD = { top: 20, right: 50, bottom: 28, left: 8 }
const chartW = W - PAD.left - PAD.right
const chartH = H - PAD.top - PAD.bottom

const INDEX_NAMES: Record<string, string> = {
  '1.000001': '上证指数',
  '0.399001': '深证成指',
  '0.399006': '创业板指',
  '1.000300': '沪深300',
  '1.000905': '中证500',
  '1.000852': '中证1000',
}

async function fetchData() {
  const cacheUrl = '/market/intraday'
  const cached = getCache<IntradayData>(cacheUrl, { code: props.code })
  if (cached) data.value = cached

  try {
    const res = await get<IntradayData>('/market/intraday', { code: props.code })
    if (res.code === 0 && res.data) {
      data.value = res.data
      setCache(cacheUrl, res.data, { code: props.code })
    }
  } catch (e) {
    console.error('Failed to fetch intraday:', e)
  } finally {
    loading.value = false
  }
}

// 计算 Y 轴范围
const yRange = computed(() => {
  if (!data.value || data.value.points.length === 0) {
    return { min: 0, max: 1, diff: 1 }
  }
  const prices = data.value.points.map(p => p.price)
  const preClose = data.value.preClose
  // 范围包含昨收价
  prices.push(preClose)
  const min = Math.min(...prices)
  const max = Math.max(...prices)
  // 上下留 10% 余量
  const padding = (max - min) * 0.1 || max * 0.005
  return {
    min: min - padding,
    max: max + padding,
    diff: (max + padding) - (min - padding),
  }
})

// 数据点 → SVG 坐标
const svgPoints = computed(() => {
  if (!data.value || data.value.points.length === 0) return ''
  const pts = data.value.points
  const total = pts.length
  if (total === 0) return ''

  return pts.map((p, i) => {
    const x = PAD.left + (i / (total - 1 || 1)) * chartW
    const y = PAD.top + (1 - (p.price - yRange.value.min) / yRange.value.diff) * chartH
    return `${x},${y}`
  }).join(' ')
})

// 面积填充路径（折线 + 底部闭合）
const areaPath = computed(() => {
  if (!data.value || data.value.points.length === 0) return ''
  const pts = data.value.points
  const total = pts.length

  const linePoints = pts.map((p, i) => {
    const x = PAD.left + (i / (total - 1 || 1)) * chartW
    const y = PAD.top + (1 - (p.price - yRange.value.min) / yRange.value.diff) * chartH
    return `${x},${y}`
  })

  const first = linePoints[0]
  const last = linePoints[linePoints.length - 1]
  const bottom = PAD.top + chartH

  return `M${first} ` + linePoints.slice(1).map(p => `L${p}`).join(' ') + ` L${last.split(',')[0]},${bottom} L${first.split(',')[0]},${bottom} Z`
})

// 昨收线 Y 坐标
const preCloseY = computed(() => {
  if (!data.value) return 0
  return PAD.top + (1 - (data.value.preClose - yRange.value.min) / yRange.value.diff) * chartH
})

// 整体涨跌颜色
const mainColor = computed(() => {
  if (!data.value || data.value.points.length === 0) return '#787878'
  const last = data.value.points[data.value.points.length - 1].price
  if (last > data.value.preClose) return '#ff4d4f'
  if (last < data.value.preClose) return '#00d68f'
  return '#787878'
})

// 最新价
const currentPrice = computed(() => {
  if (!data.value || data.value.points.length === 0) return '--'
  return data.value.points[data.value.points.length - 1].price.toFixed(2)
})

// 涨跌额/幅
const changeInfo = computed(() => {
  if (!data.value || data.value.points.length === 0) {
    return { val: '--', pct: '--', color: '#787878' }
  }
  const last = data.value.points[data.value.points.length - 1].price
  const diff = last - data.value.preClose
  const pct = data.value.preClose > 0 ? (diff / data.value.preClose) * 100 : 0
  const color = diff > 0 ? '#ff4d4f' : diff < 0 ? '#00d68f' : '#787878'
  const sign = diff > 0 ? '+' : ''
  return {
    val: `${sign}${diff.toFixed(2)}`,
    pct: `${sign}${pct.toFixed(2)}%`,
    color,
  }
})

// 时间轴标签（取首、中、尾）
const timeLabels = computed(() => {
  if (!data.value || data.value.points.length === 0) return []
  const pts = data.value.points
  const mid = Math.floor(pts.length / 2)
  return [pts[0]?.time, pts[mid]?.time, pts[pts.length - 1]?.time].filter(Boolean)
})

// Y 轴标签（3个）
const yLabels = computed(() => {
  const { min, max } = yRange.value
  const step = (max - min) / 4
  return [
    max.toFixed(2),
    ((max + min) / 2).toFixed(2),
    min.toFixed(2),
  ]
})

onMounted(() => {
  fetchData()
  // 交易时段每 30 秒刷新
  timer = setInterval(fetchData, 30000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

watch(() => props.code, () => {
  loading.value = true
  fetchData()
})
</script>

<template>
  <div class="intraday">
    <!-- 顶部导航 -->
    <div class="intraday__nav">
      <button class="intraday__back" @click="emit('back')">← 返回</button>
      <span class="intraday__title">{{ INDEX_NAMES[props.code] || props.code }}</span>
      <span class="intraday__code">{{ props.code }}</span>
    </div>

    <!-- 价格摘要 -->
    <div class="intraday__summary">
      <span class="intraday__price" :style="{ color: mainColor }">{{ currentPrice }}</span>
      <span class="intraday__change" :style="{ color: changeInfo.color }">
        {{ changeInfo.val }} ({{ changeInfo.pct }})
      </span>
      <span class="intraday__pre">昨收 {{ data?.preClose?.toFixed(2) || '--' }}</span>
    </div>

    <!-- 图表 -->
    <div class="intraday__chart" v-if="!loading && data && data.points.length > 0">
      <svg :viewBox="`0 0 ${W} ${H}`" class="intraday__svg">
        <!-- 昨收虚线 -->
        <line
          :x1="PAD.left"
          :y1="preCloseY"
          :x2="W - PAD.right"
          :y2="preCloseY"
          stroke="#f7ba1e"
          stroke-width="0.8"
          stroke-dasharray="4,3"
          opacity="0.6"
        />

        <!-- 面积填充 -->
        <defs>
          <linearGradient id="areaGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" :stop-color="mainColor" stop-opacity="0.25" />
            <stop offset="100%" :stop-color="mainColor" stop-opacity="0.02" />
          </linearGradient>
        </defs>
        <path :d="areaPath" fill="url(#areaGrad)" />

        <!-- 折线 -->
        <polyline
          :points="svgPoints"
          fill="none"
          :stroke="mainColor"
          stroke-width="1.5"
          stroke-linejoin="round"
          stroke-linecap="round"
        />

        <!-- Y 轴标签 -->
        <text
          v-for="(label, i) in yLabels"
          :key="'y' + i"
          :x="W - PAD.right + 4"
          :y="PAD.top + (i / 2) * chartH + 4"
          class="axis-label"
          fill="#787878"
          font-size="9"
        >{{ label }}</text>

        <!-- 时间轴标签 -->
        <text
          v-for="(label, i) in timeLabels"
          :key="'t' + i"
          :x="PAD.left + (i / 2) * chartW"
          :y="H - 4"
          class="axis-label"
          fill="#787878"
          font-size="9"
          text-anchor="middle"
        >{{ label }}</text>

        <!-- 昨收标签 -->
        <text
          :x="W - PAD.right + 4"
          :y="preCloseY + 3"
          fill="#f7ba1e"
          font-size="8"
          opacity="0.7"
        >昨收</text>
      </svg>
    </div>

    <!-- 加载/空状态 -->
    <div v-else class="intraday__empty">
      <template v-if="loading">加载中...</template>
      <template v-else>暂无分时数据</template>
    </div>

    <!-- 底部信息 -->
    <div class="intraday__footer" v-if="data">
      <span>更新于 {{ data.updateAt }}</span>
      <span>数据源：东方财富</span>
    </div>
  </div>
</template>

<style scoped>
.intraday {
  padding: 0;
}

.intraday__nav {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.intraday__back {
  padding: 6px 14px;
  border-radius: 8px;
  border: 1px solid rgba(255, 215, 0, 0.4);
  background: rgba(255, 215, 0, 0.08);
  color: var(--accent);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.intraday__back:hover {
  background: rgba(255, 215, 0, 0.18);
  border-color: #f7ba1e;
}

.intraday__title {
  font-size: 18px;
  color: var(--text-primary);
  font-weight: 700;
}

.intraday__code {
  font-size: 12px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

.intraday__summary {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.intraday__price {
  font-size: 28px;
  font-weight: 700;
  font-family: var(--font-mono);
  line-height: 1;
}

.intraday__change {
  font-size: 14px;
  font-weight: 600;
  font-family: var(--font-mono);
}

.intraday__pre {
  font-size: 12px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

.intraday__chart {
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 16px;
  padding: 8px;
  margin-bottom: 12px;
}

.intraday__svg {
  width: 100%;
  height: auto;
  display: block;
}

.axis-label {
  font-family: var(--font-mono, monospace);
}

.intraday__empty {
  text-align: center;
  padding: 60px 16px;
  font-size: 14px;
  color: var(--text-tertiary);
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 16px;
}

.intraday__footer {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--text-tertiary);
}
</style>

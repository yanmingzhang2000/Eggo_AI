<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { get } from '@/utils/request'

interface PositionView {
  fundCode: string
  fundName: string
  shares: number
  avgCost: number
  totalCost: number
  currentNav: number
  marketValue: number
  unrealizedPnL: number
  returnPct: number
  dividendMethod: string
}

const props = defineProps<{ accountId: number }>()

const emit = defineEmits<{
  (e: 'refresh'): void
}>()

const positions = ref<PositionView[]>([])
const loading = ref(true)

function fmt(val: number): string {
  return val.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function fmtShares(val: number): string {
  return val.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 4 })
}

function returnColor(val: number): string {
  if (val > 0) return '#ff4d4f'
  if (val < 0) return '#00d68f'
  return '#787878'
}

async function fetchPositions() {
  try {
    const res = await get<PositionView[]>(`/portfolio/accounts/${props.accountId}/positions`)
    if (res.code === 0 && res.data) {
      positions.value = res.data
    }
  } catch {} finally {
    loading.value = false
  }
}

onMounted(fetchPositions)
watch(() => props.accountId, fetchPositions)
</script>

<template>
  <div class="pos">
    <div class="pos__header">
      <span class="pos__title">持仓明细</span>
      <span class="pos__count" v-if="positions.length > 0">{{ positions.length }} 只基金</span>
    </div>

    <div v-if="loading" class="pos__loading">加载中...</div>

    <div v-else-if="positions.length === 0" class="pos__empty">
      <div class="pos__empty-icon">🥚</div>
      <div class="pos__empty-text">还没有持仓</div>
      <div class="pos__empty-sub">买入第一只基金，开始养鸡之旅</div>
    </div>

    <div v-else class="pos__list">
      <div v-for="pos in positions" :key="pos.fundCode" class="pos-card">
        <div class="pos-card__top">
          <div class="pos-card__name">{{ pos.fundName || pos.fundCode }}</div>
          <div class="pos-card__code">{{ pos.fundCode }}</div>
        </div>

        <div class="pos-card__body">
          <div class="pos-card__row">
            <span class="pos-card__label">持有份额</span>
            <span class="pos-card__val">{{ fmtShares(pos.shares) }}</span>
          </div>
          <div class="pos-card__row">
            <span class="pos-card__label">成本价</span>
            <span class="pos-card__val">{{ pos.avgCost.toFixed(4) }}</span>
          </div>
          <div class="pos-card__row">
            <span class="pos-card__label">现价</span>
            <span class="pos-card__val" :style="{ color: returnColor(pos.currentNav - pos.avgCost) }">
              {{ pos.currentNav > 0 ? pos.currentNav.toFixed(4) : '--' }}
            </span>
          </div>
          <div class="pos-card__row">
            <span class="pos-card__label">市值</span>
            <span class="pos-card__val">¥ {{ fmt(pos.marketValue) }}</span>
          </div>
        </div>

        <div class="pos-card__footer">
          <span class="pos-card__pnl" :style="{ color: returnColor(pos.unrealizedPnL) }">
            {{ pos.unrealizedPnL > 0 ? '+' : '' }}{{ fmt(pos.unrealizedPnL) }}
          </span>
          <span class="pos-card__pct" :style="{ color: returnColor(pos.returnPct) }">
            {{ pos.returnPct > 0 ? '+' : '' }}{{ pos.returnPct.toFixed(2) }}%
          </span>
          <span class="pos-card__dividend">
            {{ pos.dividendMethod === 'reinvest' ? '红利再投' : '现金分红' }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.pos {
  margin-bottom: 20px;
}

.pos__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.pos__title {
  font-size: 14px;
  color: var(--text-secondary);
  font-weight: 600;
}

.pos__count {
  font-size: 12px;
  color: var(--text-tertiary);
}

.pos__loading,
.pos__empty {
  text-align: center;
  padding: 32px 16px;
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 14px;
}

.pos__empty-icon {
  font-size: 32px;
  margin-bottom: 8px;
}

.pos__empty-text {
  font-size: 14px;
  color: var(--text-secondary);
  font-weight: 600;
}

.pos__empty-sub {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 4px;
}

.pos__list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.pos-card {
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 14px;
  padding: 14px;
  transition: border-color 0.2s;
}

.pos-card:hover {
  border-color: rgba(255, 215, 0, 0.3);
}

.pos-card__top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.pos-card__name {
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 600;
}

.pos-card__code {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

.pos-card__body {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
  margin-bottom: 10px;
}

.pos-card__row {
  display: flex;
  justify-content: space-between;
}

.pos-card__label {
  font-size: 12px;
  color: var(--text-tertiary);
}

.pos-card__val {
  font-size: 12px;
  color: var(--text-secondary);
  font-family: var(--font-mono);
  font-weight: 600;
}

.pos-card__footer {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-top: 10px;
  border-top: 1px solid #2a2a2a;
}

.pos-card__pnl {
  font-size: 14px;
  font-weight: 700;
  font-family: var(--font-mono);
}

.pos-card__pct {
  font-size: 13px;
  font-weight: 600;
  font-family: var(--font-mono);
}

.pos-card__dividend {
  margin-left: auto;
  font-size: 11px;
  color: var(--text-tertiary);
  padding: 2px 8px;
  background: rgba(255, 215, 0, 0.06);
  border-radius: 4px;
}
</style>

<script setup lang="ts">
import { computed } from 'vue'
import { useEggStore } from '@/stores/egg'

const store = useEggStore()
const metrics = computed(() => store.todayMetrics)

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
</script>

<template>
  <section v-if="metrics.length > 0" class="section">
    <div class="section__header">
      <h2 class="section__title">📊 持仓指标</h2>
      <span class="section__count">{{ metrics.length }} 只</span>
    </div>

    <div class="metrics-grid">
      <div v-for="m in metrics" :key="m.fundCode" class="metric-card">
        <div class="metric-card__top">
          <span class="metric-card__name">{{ m.fundName }}</span>
          <span class="metric-card__code">{{ m.fundCode }}</span>
        </div>

        <div class="metric-card__body">
          <div class="metric-card__item">
            <span class="metric-card__label">净值</span>
            <span class="metric-card__value">{{ m.unitNav.toFixed(4) }}</span>
          </div>
          <div class="metric-card__item">
            <span class="metric-card__label">日涨跌</span>
            <span class="metric-card__value" :style="{ color: returnColor(m.dailyReturn) }">
              {{ returnArrow(m.dailyReturn) }} {{ Math.abs(m.dailyReturn).toFixed(2) }}%
            </span>
          </div>
        </div>

        <div class="metric-card__tags">
          <span v-if="m.consecutiveUp >= 5" class="tag tag--gold">连涨{{ m.consecutiveUp }}天</span>
          <span v-if="m.hasNegNews" class="tag tag--red">舆情风险</span>
          <span v-if="m.hasPosPolicy" class="tag tag--green">政策利好</span>
          <span v-if="m.sentimentCool" class="tag tag--blue">舆情降温</span>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.section {
  margin-bottom: 32px;
}

.section__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.section__title {
  font-size: 16px;
  color: var(--text-secondary);
  font-weight: 600;
  margin: 0;
}

.section__count {
  font-size: 12px;
  color: var(--text-tertiary);
}

.metrics-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.metric-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  padding: 20px;
}

.metric-card__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.metric-card__name {
  font-size: 15px;
  color: var(--text-primary);
  font-weight: 600;
}

.metric-card__code {
  font-size: 12px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

.metric-card__body {
  display: flex;
  gap: 32px;
  margin-bottom: 16px;
}

.metric-card__item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.metric-card__label {
  font-size: 12px;
  color: var(--text-tertiary);
}

.metric-card__value {
  font-size: 20px;
  color: var(--text-primary);
  font-weight: 700;
  font-family: var(--font-mono);
}

.metric-card__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag {
  padding: 4px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

.tag--gold {
  background: rgba(255, 215, 0, 0.1);
  border: 1px solid rgba(255, 215, 0, 0.2);
  color: #ffd700;
}

.tag--red {
  background: rgba(255, 77, 79, 0.1);
  border: 1px solid rgba(255, 77, 79, 0.2);
  color: #ff4d4f;
}

.tag--green {
  background: rgba(82, 196, 26, 0.1);
  border: 1px solid rgba(82, 196, 26, 0.2);
  color: #52c41a;
}

.tag--blue {
  background: rgba(0, 195, 255, 0.1);
  border: 1px solid rgba(0, 195, 255, 0.2);
  color: #00c3ff;
}
</style>

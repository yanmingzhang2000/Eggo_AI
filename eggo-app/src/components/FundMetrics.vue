<script setup lang="ts">
import { computed } from 'vue'
import { useEggStore } from '@/stores/egg'
import type { FundMetric } from '@/types/egg'

const store = useEggStore()
const metrics = computed(() => store.todayMetrics)

/** 涨跌幅颜色 */
function returnColor(val: number): string {
  if (val > 0) return '#ff4d4f'
  if (val < 0) return '#00c3ff'
  return '#666'
}

/** 涨跌幅箭头 */
function returnArrow(val: number): string {
  if (val > 0) return '↑'
  if (val < 0) return '↓'
  return '→'
}
</script>

<template>
  <view v-if="metrics.length > 0" class="section">
    <view class="section__header">
      <text class="section__title">📊 持仓指标</text>
      <text class="section__count">{{ metrics.length }} 只</text>
    </view>

    <view class="metrics-grid">
      <view
        v-for="m in metrics"
        :key="m.fundCode"
        class="metric-card"
      >
        <!-- 基金名称 -->
        <view class="metric-card__top">
          <text class="metric-card__name">{{ m.fundName }}</text>
          <text class="metric-card__code">{{ m.fundCode }}</text>
        </view>

        <!-- 核心数据 -->
        <view class="metric-card__body">
          <view class="metric-card__nav">
            <text class="metric-card__nav-label">净值</text>
            <text class="metric-card__nav-value">{{ m.unitNav.toFixed(4) }}</text>
          </view>

          <view class="metric-card__return">
            <text class="metric-card__return-label">日涨跌</text>
            <text
              class="metric-card__return-value"
              :style="{ color: returnColor(m.dailyReturn) }"
            >
              {{ returnArrow(m.dailyReturn) }} {{ Math.abs(m.dailyReturn).toFixed(2) }}%
            </text>
          </view>
        </view>

        <!-- 标签 -->
        <view class="metric-card__tags">
          <view v-if="m.consecutiveUp >= 5" class="tag tag--gold">
            <text class="tag__text">连涨{{ m.consecutiveUp }}天</text>
          </view>
          <view v-if="m.hasNegNews" class="tag tag--red">
            <text class="tag__text">舆情风险</text>
          </view>
          <view v-if="m.hasPosPolicy" class="tag tag--green">
            <text class="tag__text">政策利好</text>
          </view>
          <view v-if="m.sentimentCool" class="tag tag--blue">
            <text class="tag__text">舆情降温</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
.section {
  margin: 32rpx 32rpx 0;
}

.section__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20rpx;
}

.section__title {
  font-size: 28rpx;
  color: var(--text-secondary);
  font-weight: 600;
  letter-spacing: 1rpx;
}

.section__count {
  font-size: 22rpx;
  color: var(--text-tertiary);
}

.metrics-grid {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.metric-card {
  background: var(--bg-card);
  border: 1rpx solid var(--border-color);
  border-radius: 20rpx;
  padding: 28rpx;
}

.metric-card__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20rpx;
}

.metric-card__name {
  font-size: 28rpx;
  color: var(--text-primary);
  font-weight: 600;
}

.metric-card__code {
  font-size: 22rpx;
  color: var(--text-tertiary);
  font-family: 'SF Mono', 'Fira Code', monospace;
}

.metric-card__body {
  display: flex;
  gap: 40rpx;
  margin-bottom: 20rpx;
}

.metric-card__nav,
.metric-card__return {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}

.metric-card__nav-label,
.metric-card__return-label {
  font-size: 22rpx;
  color: var(--text-tertiary);
}

.metric-card__nav-value {
  font-size: 32rpx;
  color: var(--text-primary);
  font-weight: 700;
  font-family: 'SF Mono', 'Fira Code', monospace;
}

.metric-card__return-value {
  font-size: 32rpx;
  font-weight: 700;
  font-family: 'SF Mono', 'Fira Code', monospace;
}

.metric-card__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}

.tag {
  padding: 6rpx 16rpx;
  border-radius: 8rpx;
}

.tag--gold {
  background: rgba(255, 215, 0, 0.1);
  border: 1rpx solid rgba(255, 215, 0, 0.2);
}

.tag--red {
  background: rgba(255, 77, 79, 0.1);
  border: 1rpx solid rgba(255, 77, 79, 0.2);
}

.tag--green {
  background: rgba(82, 196, 26, 0.1);
  border: 1rpx solid rgba(82, 196, 26, 0.2);
}

.tag--blue {
  background: rgba(0, 195, 255, 0.1);
  border: 1rpx solid rgba(0, 195, 255, 0.2);
}

.tag__text {
  font-size: 20rpx;
  font-weight: 500;
}

.tag--gold .tag__text { color: #ffd700; }
.tag--red .tag__text { color: #ff4d4f; }
.tag--green .tag__text { color: #52c41a; }
.tag--blue .tag__text { color: #00c3ff; }
</style>

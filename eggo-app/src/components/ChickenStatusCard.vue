<script setup lang="ts">
import { computed } from 'vue'
import { useEggStore } from '@/stores/egg'
import type { ChickenState } from '@/types/egg'

const store = useEggStore()

const status = computed(() => store.chickenStatus)
const decision = computed(() => store.decision)

/** 状态对应的视觉配置 */
const stateConfig = computed(() => {
  const state = status.value?.overallState as ChickenState
  const configs: Record<ChickenState, { bg: string; glow: string; border: string }> = {
    laying: {
      bg: 'linear-gradient(135deg, #1a1a0a 0%, #0a0a0a 100%)',
      glow: '0 0 60rpx rgba(255, 215, 0, 0.15)',
      border: '1rpx solid rgba(255, 215, 0, 0.3)',
    },
    resting: {
      bg: 'linear-gradient(135deg, #1a1a1a 0%, #0a0a0a 100%)',
      glow: '0 0 40rpx rgba(255, 255, 255, 0.05)',
      border: '1rpx solid rgba(255, 255, 255, 0.1)',
    },
    molting: {
      bg: 'linear-gradient(135deg, #1a0a0a 0%, #0a0a0a 100%)',
      glow: '0 0 60rpx rgba(255, 80, 80, 0.15)',
      border: '1rpx solid rgba(255, 80, 80, 0.3)',
    },
    soaking: {
      bg: 'linear-gradient(135deg, #0a1a1a 0%, #0a0a0a 100%)',
      glow: '0 0 60rpx rgba(0, 195, 255, 0.15)',
      border: '1rpx solid rgba(0, 195, 255, 0.3)',
    },
  }
  return configs[state] || configs.resting
})

/** 告警等级对应的颜色 */
const alertColor = computed(() => {
  const level = status.value?.alertLevel
  const colors: Record<string, string> = {
    none: '#666',
    info: '#00c3ff',
    warning: '#ffd700',
    critical: '#ff4d4f',
  }
  return colors[level || 'none'] || '#666'
})

/** 操作 Emoji */
const actionEmoji = computed(() => decision.value?.actionEmoji || '🐔')

/** 规则名称 */
const ruleName = computed(() => decision.value?.ruleName || '观察中')

/** 建议文案 */
const suggestion = computed(() => decision.value?.suggestion || '暂无建议')
</script>

<template>
  <view
    v-if="status"
    class="card"
    :style="{
      background: stateConfig.bg,
      boxShadow: stateConfig.glow,
      border: stateConfig.border,
    }"
  >
    <!-- 顶部：状态 Emoji + 规则 -->
    <view class="card__header">
      <text class="card__emoji">{{ status.stateEmoji }}</text>
      <view class="card__rule-tag">
        <text class="card__rule-text">{{ ruleName }}</text>
      </view>
    </view>

    <!-- 中部：状态描述 -->
    <view class="card__body">
      <text class="card__state-desc">{{ status.stateDesc }}</text>
    </view>

    <!-- 底部：决策建议 -->
    <view v-if="decision" class="card__footer">
      <view class="card__suggestion-box">
        <text class="card__suggestion-icon">💡</text>
        <text class="card__suggestion-text">{{ suggestion }}</text>
      </view>

      <!-- 置信度条 -->
      <view class="card__confidence">
        <view class="card__confidence-bar">
          <view
            class="card__confidence-fill"
            :style="{ width: `${decision.confidence * 100}%` }"
          />
        </view>
        <text class="card__confidence-text">
          置信度 {{ (decision.confidence * 100).toFixed(0) }}%
        </text>
      </view>
    </view>

    <!-- 告警指示器 -->
    <view
      v-if="status.alertLevel !== 'none'"
      class="card__alert-dot"
      :style="{ background: alertColor }"
    />
  </view>
</template>

<style scoped>
.card {
  margin: 24rpx 32rpx;
  padding: 36rpx;
  border-radius: 24rpx;
  position: relative;
  overflow: hidden;
  transition: all 0.3s ease;
}

.card::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle at 30% 30%, rgba(255, 215, 0, 0.03), transparent 60%);
  pointer-events: none;
}

.card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20rpx;
}

.card__emoji {
  font-size: 64rpx;
}

.card__rule-tag {
  padding: 8rpx 24rpx;
  background: rgba(255, 215, 0, 0.1);
  border: 1rpx solid rgba(255, 215, 0, 0.2);
  border-radius: 32rpx;
}

.card__rule-text {
  font-size: 24rpx;
  color: var(--accent);
  font-weight: 600;
  letter-spacing: 1rpx;
}

.card__body {
  margin-bottom: 24rpx;
}

.card__state-desc {
  font-size: 30rpx;
  color: var(--text-primary);
  font-weight: 500;
  line-height: 1.6;
}

.card__footer {
  border-top: 1rpx solid rgba(255, 255, 255, 0.06);
  padding-top: 24rpx;
}

.card__suggestion-box {
  display: flex;
  align-items: flex-start;
  gap: 12rpx;
  background: rgba(255, 215, 0, 0.05);
  padding: 20rpx 24rpx;
  border-radius: 16rpx;
  margin-bottom: 20rpx;
}

.card__suggestion-icon {
  font-size: 28rpx;
  flex-shrink: 0;
}

.card__suggestion-text {
  font-size: 26rpx;
  color: var(--accent);
  font-weight: 500;
  line-height: 1.5;
}

.card__confidence {
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.card__confidence-bar {
  flex: 1;
  height: 6rpx;
  background: rgba(255, 255, 255, 0.08);
  border-radius: 3rpx;
  overflow: hidden;
}

.card__confidence-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--accent), #ffed4a);
  border-radius: 3rpx;
  transition: width 0.6s ease;
}

.card__confidence-text {
  font-size: 22rpx;
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.card__alert-dot {
  position: absolute;
  top: 24rpx;
  right: 24rpx;
  width: 16rpx;
  height: 16rpx;
  border-radius: 50%;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(1.3); }
}
</style>

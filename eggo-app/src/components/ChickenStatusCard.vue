<script setup lang="ts">
import { computed } from 'vue'
import { useEggStore } from '@/stores/egg'
import type { ChickenState } from '@/types/egg'

const store = useEggStore()
const status = computed(() => store.chickenStatus)
const decision = computed(() => store.decision)

const stateConfig: Record<ChickenState, { bg: string; glow: string; border: string }> = {
  laying: {
    bg: 'linear-gradient(135deg, #1a1a0a 0%, #0a0a0a 100%)',
    glow: '0 0 60px rgba(255, 215, 0, 0.15)',
    border: '1px solid rgba(255, 215, 0, 0.3)',
  },
  resting: {
    bg: 'linear-gradient(135deg, #1a1a1a 0%, #0a0a0a 100%)',
    glow: '0 0 40px rgba(255, 255, 255, 0.05)',
    border: '1px solid rgba(255, 255, 255, 0.1)',
  },
  molting: {
    bg: 'linear-gradient(135deg, #1a0a0a 0%, #0a0a0a 100%)',
    glow: '0 0 60px rgba(255, 80, 80, 0.15)',
    border: '1px solid rgba(255, 80, 80, 0.3)',
  },
  soaking: {
    bg: 'linear-gradient(135deg, #0a1a1a 0%, #0a0a0a 100%)',
    glow: '0 0 60px rgba(0, 195, 255, 0.15)',
    border: '1px solid rgba(0, 195, 255, 0.3)',
  },
}

const currentConfig = computed(() => {
  return stateConfig[status.value?.overallState || 'resting']
})
</script>

<template>
  <div
    v-if="status"
    class="card"
    :style="{
      background: currentConfig.bg,
      boxShadow: currentConfig.glow,
      border: currentConfig.border,
    }"
  >
    <!-- 顶部 -->
    <div class="card__header">
      <span class="card__emoji">{{ status.stateEmoji }}</span>
      <span class="card__rule-tag">{{ decision?.ruleName || '观察中' }}</span>
    </div>

    <!-- 状态描述 -->
    <p class="card__state-desc">{{ status.stateDesc }}</p>

    <!-- 决策建议 -->
    <div v-if="decision" class="card__footer">
      <div class="card__suggestion-box">
        <span class="card__suggestion-icon">💡</span>
        <span class="card__suggestion-text">{{ decision.suggestion }}</span>
      </div>

      <!-- 置信度条 -->
      <div class="card__confidence">
        <div class="card__confidence-bar">
          <div
            class="card__confidence-fill"
            :style="{ width: `${decision.confidence * 100}%` }"
          ></div>
        </div>
        <span class="card__confidence-text">
          置信度 {{ (decision.confidence * 100).toFixed(0) }}%
        </span>
      </div>
    </div>

    <!-- 告警指示器 -->
    <div
      v-if="status.alertLevel !== 'none'"
      class="card__alert-dot"
    ></div>
  </div>
</template>

<style scoped>
.card {
  padding: 28px;
  border-radius: 20px;
  position: relative;
  overflow: hidden;
  transition: all 0.3s ease;
  margin-bottom: 24px;
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
  margin-bottom: 16px;
}

.card__emoji {
  font-size: 48px;
}

.card__rule-tag {
  padding: 6px 16px;
  background: rgba(255, 215, 0, 0.1);
  border: 1px solid rgba(255, 215, 0, 0.2);
  border-radius: 20px;
  font-size: 13px;
  color: var(--accent);
  font-weight: 600;
  letter-spacing: 1px;
}

.card__state-desc {
  font-size: 18px;
  color: var(--text-primary);
  font-weight: 500;
  line-height: 1.6;
  margin: 0 0 20px;
}

.card__footer {
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  padding-top: 20px;
}

.card__suggestion-box {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  background: rgba(255, 215, 0, 0.05);
  padding: 16px 20px;
  border-radius: 12px;
  margin-bottom: 16px;
}

.card__suggestion-icon {
  font-size: 18px;
  flex-shrink: 0;
}

.card__suggestion-text {
  font-size: 15px;
  color: var(--accent);
  font-weight: 500;
  line-height: 1.5;
}

.card__confidence {
  display: flex;
  align-items: center;
  gap: 12px;
}

.card__confidence-bar {
  flex: 1;
  height: 4px;
  background: rgba(255, 255, 255, 0.08);
  border-radius: 2px;
  overflow: hidden;
}

.card__confidence-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--accent), #ffed4a);
  border-radius: 2px;
  transition: width 0.6s ease;
}

.card__confidence-text {
  font-size: 12px;
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.card__alert-dot {
  position: absolute;
  top: 20px;
  right: 20px;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #ff4d4f;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(1.3); }
}
</style>

<script setup lang="ts">
import { onMounted } from '@dcloudio/uni-app'
import { useEggStore } from '@/stores/egg'
import ChickenStatusCard from '@/components/ChickenStatusCard.vue'
import NewsFeed from '@/components/NewsFeed.vue'
import FundMetrics from '@/components/FundMetrics.vue'

const store = useEggStore()

onMounted(() => {
  store.fetchEggStatus()
})

function onPullDownRefresh() {
  store.refresh().finally(() => {
    uni.stopPullDownRefresh()
  })
}
</script>

<template>
  <view class="page">
    <!-- 顶部状态栏占位 -->
    <view :style="{ height: `${statusBarHeight}px` }" />

    <!-- Header -->
    <view class="header">
      <text class="header__title">🐔 鸡生蛋</text>
      <text class="header__sub">Eggo · 智能基金决策</text>
    </view>

    <!-- 加载态 -->
    <view v-if="store.loading && !store.hasData" class="loading-wrap">
      <view class="loading-spinner" />
      <text class="loading-text">正在计算母鸡状态...</text>
    </view>

    <!-- 错误态 -->
    <view v-else-if="store.error && !store.hasData" class="error-wrap">
      <text class="error-emoji">⚠️</text>
      <text class="error-text">{{ store.error }}</text>
      <view class="error-btn" @tap="store.fetchEggStatus">
        <text class="error-btn__text">重试</text>
      </view>
    </view>

    <!-- 主内容 -->
    <scroll-view
      v-else
      scroll-y
      class="scroll-area"
      :refresher-enabled="true"
      :refresher-triggered="store.loading"
      @refresherrefresh="onPullDownRefresh"
    >
      <!-- 母鸡状态卡片 -->
      <ChickenStatusCard />

      <!-- 基金指标 -->
      <FundMetrics />

      <!-- AI 情报线 -->
      <NewsFeed />

      <!-- 底部安全区 -->
      <view :style="{ height: '120rpx' }" />
    </scroll-view>
  </view>
</template>

<script lang="ts">
export default {
  data() {
    return {
      statusBarHeight: 0,
    }
  },
  onLoad() {
    const sysInfo = uni.getSystemInfoSync()
    this.statusBarHeight = sysInfo.statusBarHeight || 0
  },
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: var(--bg-primary);
}

.header {
  padding: 24rpx 32rpx 16rpx;
}

.header__title {
  font-size: 40rpx;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: 2rpx;
}

.header__sub {
  display: block;
  font-size: 22rpx;
  color: var(--text-tertiary);
  margin-top: 4rpx;
  letter-spacing: 4rpx;
  text-transform: uppercase;
}

.loading-wrap,
.error-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding-top: 300rpx;
}

.loading-spinner {
  width: 64rpx;
  height: 64rpx;
  border: 4rpx solid var(--border-color);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-text {
  margin-top: 24rpx;
  font-size: 26rpx;
  color: var(--text-tertiary);
}

.error-emoji {
  font-size: 80rpx;
}

.error-text {
  margin-top: 16rpx;
  font-size: 28rpx;
  color: var(--text-secondary);
}

.error-btn {
  margin-top: 32rpx;
  padding: 16rpx 48rpx;
  background: var(--accent);
  border-radius: 48rpx;
}

.error-btn__text {
  font-size: 28rpx;
  color: #000;
  font-weight: 600;
}

.scroll-area {
  height: calc(100vh - 200rpx);
}
</style>

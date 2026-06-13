<script setup lang="ts">
import { onMounted } from 'vue'
import { useEggStore } from '@/stores/egg'
import ChickenStatusCard from '@/components/ChickenStatusCard.vue'
import FundMetrics from '@/components/FundMetrics.vue'
import NewsFeed from '@/components/NewsFeed.vue'

const store = useEggStore()

onMounted(() => {
  store.fetchEggStatus()
})

function handleRefresh() {
  store.refresh()
}
</script>

<template>
  <div class="page">
    <!-- Header -->
    <header class="header">
      <div class="header__left">
        <h1 class="header__title">🐔 鸡生蛋</h1>
        <span class="header__sub">Eggo · 智能基金决策</span>
      </div>
      <button class="header__refresh" @click="handleRefresh" :disabled="store.loading">
        <span :class="{ 'spin': store.loading }">↻</span>
      </button>
    </header>

    <!-- 加载态 -->
    <div v-if="store.loading && !store.hasData" class="loading-wrap">
      <div class="loading-spinner"></div>
      <p class="loading-text">正在计算母鸡状态...</p>
    </div>

    <!-- 错误态 -->
    <div v-else-if="store.error && !store.hasData" class="error-wrap">
      <p class="error-emoji">⚠️</p>
      <p class="error-text">{{ store.error }}</p>
      <button class="error-btn" @click="store.fetchEggStatus">重试</button>
    </div>

    <!-- 主内容 -->
    <main v-else class="main">
      <ChickenStatusCard />
      <FundMetrics />
      <NewsFeed />
    </main>

    <!-- 底部 -->
    <footer class="footer">
      <p>鸡生蛋 Eggo © 2025 · 仅供参考，不构成投资建议</p>
    </footer>
  </div>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: var(--bg-primary);
  color: var(--text-primary);
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 32px;
  border-bottom: 1px solid var(--border-color);
  position: sticky;
  top: 0;
  background: rgba(10, 10, 10, 0.9);
  backdrop-filter: blur(20px);
  z-index: 100;
}

.header__title {
  font-size: 24px;
  font-weight: 700;
  margin: 0;
  letter-spacing: 2px;
}

.header__sub {
  display: block;
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 4px;
  letter-spacing: 4px;
  text-transform: uppercase;
}

.header__refresh {
  background: rgba(255, 215, 0, 0.1);
  border: 1px solid rgba(255, 215, 0, 0.2);
  color: var(--accent);
  width: 40px;
  height: 40px;
  border-radius: 50%;
  font-size: 20px;
  cursor: pointer;
  transition: all 0.2s;
}

.header__refresh:hover {
  background: rgba(255, 215, 0, 0.2);
}

.header__refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spin {
  display: inline-block;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-wrap,
.error-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding-top: 200px;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border-color);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.loading-text {
  margin-top: 16px;
  font-size: 14px;
  color: var(--text-tertiary);
}

.error-emoji {
  font-size: 48px;
  margin: 0;
}

.error-text {
  margin-top: 8px;
  font-size: 16px;
  color: var(--text-secondary);
}

.error-btn {
  margin-top: 20px;
  padding: 10px 32px;
  background: var(--accent);
  color: #000;
  border: none;
  border-radius: 24px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s;
}

.error-btn:hover {
  transform: scale(1.05);
}

.main {
  max-width: 800px;
  margin: 0 auto;
  padding: 24px 16px;
}

.footer {
  text-align: center;
  padding: 32px 16px;
  color: var(--text-tertiary);
  font-size: 12px;
  border-top: 1px solid var(--border-color);
  margin-top: 48px;
}
</style>

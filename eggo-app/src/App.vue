<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useEggStore } from '@/stores/egg'
import AuthPage from '@/components/AuthPage.vue'
import ChickenStatusCard from '@/components/ChickenStatusCard.vue'
import FundMetrics from '@/components/FundMetrics.vue'
import NewsFeed from '@/components/NewsFeed.vue'
import MarketDashboard from '@/components/MarketDashboard.vue'

const authStore = useAuthStore()
const eggStore = useEggStore()

const isLoggedIn = computed(() => authStore.isLoggedIn)

onMounted(() => {
  if (isLoggedIn.value) {
    eggStore.fetchEggStatus()
  }
})

function handleRefresh() {
  eggStore.refresh()
}

function handleLogout() {
  authStore.logout()
}
</script>

<template>
  <!-- 未登录：显示登录/注册页 -->
  <AuthPage v-if="!isLoggedIn" />

  <!-- 已登录：显示主页面 -->
  <div v-else class="page">
    <!-- Header -->
    <header class="header">
      <div class="header__left">
        <h1 class="header__title">🐔 鸡生蛋</h1>
        <span class="header__sub">Eggo · 智能基金决策</span>
      </div>
      <div class="header__right">
        <span class="header__user">
          {{ authStore.isGuest ? '👤 游客' : '👋 ' + authStore.username }}
        </span>
        <button class="header__btn" @click="handleRefresh" :disabled="eggStore.loading">
          <span :class="{ 'spin': eggStore.loading }">↻</span>
        </button>
        <button class="header__btn header__btn--logout" @click="handleLogout">
          退出
        </button>
      </div>
    </header>

    <!-- 加载态 -->
    <div v-if="eggStore.loading && !eggStore.hasData" class="loading-wrap">
      <div class="loading-spinner"></div>
      <p class="loading-text">正在计算母鸡状态...</p>
    </div>

    <!-- 错误态 -->
    <div v-else-if="eggStore.error && !eggStore.hasData" class="error-wrap">
      <p class="error-emoji">⚠️</p>
      <p class="error-text">{{ eggStore.error }}</p>
      <button class="error-btn" @click="eggStore.fetchEggStatus">重试</button>
    </div>

    <!-- 主内容 -->
    <main v-else class="main">
      <MarketDashboard />
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
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
  position: sticky;
  top: 0;
  background: rgba(10, 10, 10, 0.95);
  backdrop-filter: blur(20px);
  z-index: 100;
}

.header__title {
  font-size: 20px;
  font-weight: 700;
  margin: 0;
  letter-spacing: 2px;
}

.header__sub {
  display: block;
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 2px;
  letter-spacing: 3px;
  text-transform: uppercase;
}

.header__right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header__user {
  font-size: 13px;
  color: var(--text-secondary);
}

.header__btn {
  width: 36px;
  height: 36px;
  background: rgba(255, 215, 0, 0.1);
  border: 1px solid rgba(255, 215, 0, 0.2);
  color: var(--accent);
  border-radius: 50%;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.header__btn:hover {
  background: rgba(255, 215, 0, 0.2);
}

.header__btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.header__btn--logout {
  width: auto;
  border-radius: 18px;
  padding: 0 14px;
  font-size: 12px;
  background: transparent;
  color: var(--text-tertiary);
  border-color: var(--border-color);
}

.header__btn--logout:hover {
  color: #ff4d4f;
  border-color: rgba(255, 77, 79, 0.3);
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

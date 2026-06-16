<script setup lang="ts">
import { onMounted, computed, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useEggStore } from '@/stores/egg'
import AuthPage from '@/components/AuthPage.vue'
import ChickenStatusCard from '@/components/ChickenStatusCard.vue'
import FundMetrics from '@/components/FundMetrics.vue'
import FundDistribution from '@/components/FundDistribution.vue'
import NewsFeed from '@/components/NewsFeed.vue'
import MarketDashboard from '@/components/MarketDashboard.vue'
import Portfolio from '@/components/portfolio/Portfolio.vue'
import FundDetail from '@/components/FundDetail.vue'
import WatchlistPanel from '@/components/WatchlistPanel.vue'
import { useWatchlistStore } from '@/stores/watchlist'

const activeView = ref<'home' | 'portfolio' | 'fundDetail'>('home')
const selectedFund = ref<{ code: string; name: string } | null>(null)
const searchCode = ref('')

const authStore = useAuthStore()
const eggStore = useEggStore()
const watchlistStore = useWatchlistStore()

const isLoggedIn = computed(() => authStore.isLoggedIn)

onMounted(() => {
  if (isLoggedIn.value) {
    eggStore.fetchEggStatus()
    // 已登录且非游客，预加载 watchlist
    if (!authStore.isGuest) {
      watchlistStore.fetchList()
    }
  }
})

function handleRefresh() {
  eggStore.refresh()
}

function handleLogout() {
  authStore.logout()
}

function handleViewDetail(fundCode: string, fundName: string) {
  selectedFund.value = { code: fundCode, name: fundName }
  activeView.value = 'fundDetail'
}

function handleBackFromDetail() {
  selectedFund.value = null
  activeView.value = 'home'
}

function handleSearchFund() {
  const code = searchCode.value.trim()
  if (!code) return
  selectedFund.value = { code, name: code }
  activeView.value = 'fundDetail'
  searchCode.value = ''
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
        <h1 class="header__title"><img src="/logo.svg" class="header__logo" alt="鸡生蛋" /> 鸡生蛋</h1>
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

    <!-- 主内容 -->
    <main class="main">
      <!-- 首次加载错误 -->
      <div v-if="eggStore.error && !eggStore.hasData" class="error-wrap">
        <p class="error-emoji">⚠️</p>
        <p class="error-text">{{ eggStore.error }}</p>
        <button class="error-btn" @click="eggStore.fetchEggStatus">重试</button>
      </div>
      <!-- 顶部导航 Tab -->
      <div class="view-tabs">
        <button
          :class="['view-tab', { 'view-tab--active': activeView === 'home' }]"
          @click="activeView = 'home'"
        >📊 行情看板</button>
        <button
          :class="['view-tab', { 'view-tab--active': activeView === 'portfolio' }]"
          @click="activeView = 'portfolio'"
        >🐔 我的鸡笼</button>
      </div>

      <!-- 基金详情 -->
      <FundDetail
        v-if="activeView === 'fundDetail' && selectedFund"
        :fund-code="selectedFund.code"
        :fund-name="selectedFund.name"
        @back="handleBackFromDetail"
      />

      <!-- 行情看板 -->
      <template v-else-if="activeView === 'home'">
        <!-- 基金搜索入口 -->
        <div class="fund-search">
          <input
            v-model="searchCode"
            class="fund-search__input"
            type="text"
            placeholder="输入基金代码直接查看详情，如 000001"
            maxlength="6"
            @keyup.enter="handleSearchFund"
          />
          <button class="fund-search__btn" @click="handleSearchFund" :disabled="!searchCode.trim()">
            查看
          </button>
        </div>
        <WatchlistPanel v-if="!authStore.isGuest" @view-detail="handleViewDetail" />
        <MarketDashboard />
        <FundDistribution @view-detail="handleViewDetail" />
        <ChickenStatusCard />
        <FundMetrics @view-detail="handleViewDetail" />
        <NewsFeed />
      </template>

      <!-- 我的鸡笼 -->
      <Portfolio v-else-if="activeView === 'portfolio'" />
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

.header__logo {
  width: 24px;
  height: 24px;
  vertical-align: middle;
  object-fit: contain;
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

.error-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding-top: 200px;
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

/* 顶部导航 Tab */
.view-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 24px;
}

.view-tab {
  flex: 1;
  padding: 12px 16px;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  background: transparent;
  color: var(--text-tertiary);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  text-align: center;
}

.view-tab:hover {
  color: var(--text-secondary);
  border-color: rgba(255, 215, 0, 0.3);
}

.view-tab--active {
  background: rgba(255, 215, 0, 0.1);
  border-color: rgba(255, 215, 0, 0.5);
  color: var(--accent);
}

.footer {
  text-align: center;
  padding: 32px 16px;
  color: var(--text-tertiary);
  font-size: 12px;
  border-top: 1px solid var(--border-color);
  margin-top: 48px;
}

/* 基金搜索入口 */
.fund-search {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
}

.fund-search__input {
  flex: 1;
  padding: 10px 14px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  color: var(--text-primary);
  font-size: 14px;
  font-family: var(--font-mono);
  outline: none;
  transition: border-color 0.2s;
}

.fund-search__input::placeholder {
  color: var(--text-tertiary);
  font-family: inherit;
}

.fund-search__input:focus {
  border-color: rgba(255, 215, 0, 0.5);
}

.fund-search__btn {
  padding: 10px 18px;
  background: rgba(255, 215, 0, 0.12);
  border: 1px solid rgba(255, 215, 0, 0.35);
  border-radius: 10px;
  color: var(--accent);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.fund-search__btn:hover:not(:disabled) {
  background: rgba(255, 215, 0, 0.22);
  border-color: rgba(255, 215, 0, 0.6);
}

.fund-search__btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
</style>

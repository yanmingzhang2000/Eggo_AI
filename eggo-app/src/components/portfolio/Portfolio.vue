<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { get, post } from '@/utils/request'
import AccountCard from './AccountCard.vue'
import PositionList from './PositionList.vue'
import BuyModal from './BuyModal.vue'
import PendingOrders from './PendingOrders.vue'

interface AccountSummary {
  userId: number
  initialBalance: number
  cashBalance: number
  frozenCash: number
  totalAssets: number
  totalProfit: number
  totalReturn: number
  pendingCount: number
}

const account = ref<AccountSummary | null>(null)
const hasAccount = ref(false)
const loading = ref(true)
const showBuy = ref(false)
const refreshKey = ref(0)

// 创建账户表单
const showCreate = ref(false)
const createBalance = ref(1000000)

async function fetchAccount() {
  try {
    const res = await get<AccountSummary>('/portfolio/account')
    if (res.code === 0 && res.data) {
      account.value = res.data
      hasAccount.value = true
    }
  } catch {
    hasAccount.value = false
  } finally {
    loading.value = false
  }
}

async function createAccount() {
  try {
    const res = await post('/portfolio/account', { initialBalance: createBalance.value })
    if (res.code === 0) {
      showCreate.value = false
      await fetchAccount()
    }
  } catch (e: any) {
    alert(e?.response?.data?.message || '创建失败')
  }
}

function onBuyDone() {
  showBuy.value = false
  refreshKey.value++
  fetchAccount()
}

onMounted(fetchAccount)
</script>

<template>
  <section class="portfolio">
    <!-- 未创建账户 -->
    <div v-if="!loading && !hasAccount" class="portfolio__create">
      <div class="create-card">
        <div class="create-card__icon">🐔</div>
        <h2 class="create-card__title">开启你的养鸡之旅</h2>
        <p class="create-card__sub">设置初始虚拟资金，开始模拟基金投资</p>

        <div class="create-balance">
          <input
            v-model.number="createBalance"
            type="number"
            class="create-input"
            min="1000"
            step="10000"
          />
          <span class="create-unit">元</span>
        </div>

        <div class="create-templates">
          <button class="tmpl-btn" @click="createBalance = 100000">10万·新手实战</button>
          <button class="tmpl-btn" @click="createBalance = 500000">50万·进阶玩家</button>
          <button class="tmpl-btn" @click="createBalance = 1000000">100万·土豪梭哈</button>
        </div>

        <button class="create-submit" @click="createAccount" :disabled="createBalance < 1000">
          🐣 破壳启动
        </button>
      </div>
    </div>

    <!-- 已有账户 -->
    <template v-else-if="hasAccount">
      <AccountCard :key="refreshKey" @buy="showBuy = true" />

      <PendingOrders :key="'po' + refreshKey" />

      <PositionList :key="'pos' + refreshKey" @refresh="refreshKey++" />
    </template>

    <!-- 加载中 -->
    <div v-else class="portfolio__loading">加载中...</div>

    <!-- 买入弹窗 -->
    <BuyModal v-if="showBuy" @close="showBuy = false" @done="onBuyDone" />
  </section>
</template>

<style scoped>
.portfolio {
  padding: 0;
}

/* ── 创建账户 ── */
.portfolio__create {
  display: flex;
  justify-content: center;
  padding: 40px 0;
}

.create-card {
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 20px;
  padding: 40px 32px;
  text-align: center;
  max-width: 400px;
  width: 100%;
}

.create-card__icon {
  font-size: 48px;
  margin-bottom: 12px;
}

.create-card__title {
  font-size: 20px;
  color: var(--text-primary);
  font-weight: 700;
  margin: 0 0 8px;
}

.create-card__sub {
  font-size: 13px;
  color: var(--text-tertiary);
  margin: 0 0 24px;
}

.create-balance {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-bottom: 16px;
}

.create-input {
  width: 180px;
  padding: 12px 16px;
  background: #0a0a0a;
  border: 1px solid #333;
  border-radius: 12px;
  color: var(--accent);
  font-size: 24px;
  font-weight: 700;
  font-family: var(--font-mono);
  text-align: center;
  outline: none;
  transition: border-color 0.2s;
}

.create-input:focus {
  border-color: #f7ba1e;
}

.create-unit {
  font-size: 16px;
  color: var(--text-tertiary);
}

.create-templates {
  display: flex;
  gap: 8px;
  justify-content: center;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.tmpl-btn {
  padding: 8px 14px;
  border-radius: 8px;
  border: 1px solid rgba(255, 215, 0, 0.3);
  background: rgba(255, 215, 0, 0.06);
  color: var(--accent);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.tmpl-btn:hover {
  background: rgba(255, 215, 0, 0.15);
  border-color: rgba(255, 215, 0, 0.6);
}

.create-submit {
  padding: 14px 40px;
  background: var(--accent);
  color: #000;
  border: none;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
}

.create-submit:hover {
  background: #ffe066;
  transform: translateY(-1px);
}

.create-submit:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  transform: none;
}

.portfolio__loading {
  text-align: center;
  padding: 60px;
  color: var(--text-tertiary);
  font-size: 14px;
}
</style>

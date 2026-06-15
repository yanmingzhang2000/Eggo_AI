<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { get, post, del } from '@/utils/request'
import AccountCard from './AccountCard.vue'
import PositionList from './PositionList.vue'
import BuyModal from './BuyModal.vue'
import PendingOrders from './PendingOrders.vue'
import CreateCageModal from './CreateCageModal.vue'

interface AccountSummary {
  id: number
  name: string
  userId: number
  initialBalance: number
  cashBalance: number
  frozenCash: number
  totalAssets: number
  totalProfit: number
  totalReturn: number
  pendingCount: number
  createdAt: string
}

const cages = ref<AccountSummary[]>([])
const selectedCage = ref<AccountSummary | null>(null)
const loading = ref(true)
const showCreate = ref(false)
const showBuy = ref(false)
const refreshKey = ref(0)

async function fetchCages() {
  try {
    const res = await get<AccountSummary[]>('/portfolio/accounts')
    if (res.code === 0 && res.data) {
      cages.value = res.data
    }
  } catch {
    cages.value = []
  } finally {
    loading.value = false
  }
}

function selectCage(cage: AccountSummary) {
  selectedCage.value = cage
  refreshKey.value++
}

function goBack() {
  selectedCage.value = null
  fetchCages()
}

function onCageCreated() {
  showCreate.value = false
  fetchCages()
}

async function deleteCage(cage: AccountSummary) {
  if (!confirm(`确定删除「${cage.name}」？此操作不可恢复。`)) return
  try {
    await del(`/portfolio/accounts/${cage.id}`)
    if (selectedCage.value?.id === cage.id) {
      selectedCage.value = null
    }
    await fetchCages()
  } catch (e: any) {
    alert(e?.response?.data?.message || '删除失败')
  }
}

function onBuyDone() {
  showBuy.value = false
  refreshKey.value++
}

function fmt(val: number): string {
  return val.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

onMounted(fetchCages)
</script>

<template>
  <section class="portfolio">
    <!-- 加载中 -->
    <div v-if="loading" class="portfolio__loading">加载中...</div>

    <!-- 无鸡笼：引导创建 -->
    <div v-else-if="!selectedCage && cages.length === 0" class="portfolio__create">
      <div class="create-card">
        <div class="create-card__icon">🐔</div>
        <h2 class="create-card__title">开启你的养鸡之旅</h2>
        <p class="create-card__sub">创建你的第一个鸡笼，开始模拟基金投资</p>
        <button class="create-submit" @click="showCreate = true">
          🐣 创建鸡笼
        </button>
      </div>
    </div>

    <!-- 鸡笼列表 -->
    <template v-else-if="!selectedCage">
      <div class="cage-header">
        <h2 class="cage-header__title">🐔 我的鸡笼</h2>
        <button class="cage-add-btn" @click="showCreate = true">+ 新建鸡笼</button>
      </div>

      <div class="cage-grid">
        <div
          v-for="cage in cages"
          :key="cage.id"
          class="cage-card"
          @click="selectCage(cage)"
        >
          <div class="cage-card__top">
            <span class="cage-card__name">{{ cage.name }}</span>
            <button class="cage-card__del" @click.stop="deleteCage(cage)">×</button>
          </div>

          <div class="cage-card__assets">
            <span class="cage-card__total">¥ {{ fmt(cage.totalAssets) }}</span>
          </div>

          <div class="cage-card__row">
            <span class="cage-card__label">初始资金</span>
            <span class="cage-card__value">¥ {{ fmt(cage.initialBalance) }}</span>
          </div>
          <div class="cage-card__row">
            <span class="cage-card__label">可用现金</span>
            <span class="cage-card__value">¥ {{ fmt(cage.cashBalance) }}</span>
          </div>

          <div class="cage-card__footer">
            <span
              :class="['cage-card__return', cage.totalReturn >= 0 ? 'return--up' : 'return--down']"
            >
              {{ cage.totalReturn >= 0 ? '+' : '' }}{{ cage.totalReturn }}%
            </span>
            <span class="cage-card__date">{{ cage.createdAt }}</span>
          </div>
        </div>
      </div>
    </template>

    <!-- 鸡笼详情 -->
    <template v-else>
      <button class="back-btn" @click="goBack">← 返回鸡笼列表</button>

      <div class="cage-detail-header">
        <h2 class="cage-detail-name">{{ selectedCage.name }}</h2>
        <button class="cage-detail-del" @click="deleteCage(selectedCage)">删除鸡笼</button>
      </div>

      <AccountCard
        :key="'ac' + refreshKey"
        :account-id="selectedCage.id"
        @buy="showBuy = true"
      />

      <PendingOrders :key="'po' + refreshKey" :account-id="selectedCage.id" />

      <PositionList
        :key="'pos' + refreshKey"
        :account-id="selectedCage.id"
        @refresh="refreshKey++"
      />
    </template>

    <!-- 创建鸡笼弹窗 -->
    <CreateCageModal
      v-if="showCreate"
      @close="showCreate = false"
      @done="onCageCreated"
    />

    <!-- 买入弹窗 -->
    <BuyModal
      v-if="showBuy && selectedCage"
      :account-id="selectedCage.id"
      @close="showBuy = false"
      @done="onBuyDone"
    />
  </section>
</template>

<style scoped>
.portfolio { padding: 0; }

/* ── 无鸡笼 ── */
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

.create-card__icon { font-size: 48px; margin-bottom: 12px; }

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

/* ── 鸡笼列表头 ── */
.cage-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.cage-header__title {
  font-size: 18px;
  color: var(--text-primary);
  font-weight: 700;
  margin: 0;
}

.cage-add-btn {
  padding: 8px 16px;
  border-radius: 10px;
  border: 1px solid rgba(255, 215, 0, 0.4);
  background: rgba(255, 215, 0, 0.08);
  color: var(--accent);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.cage-add-btn:hover {
  background: rgba(255, 215, 0, 0.18);
}

/* ── 鸡笼卡片网格 ── */
.cage-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 14px;
}

.cage-card {
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 16px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s;
}

.cage-card:hover {
  border-color: rgba(255, 215, 0, 0.4);
  transform: translateY(-2px);
}

.cage-card__top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.cage-card__name {
  font-size: 15px;
  color: var(--text-primary);
  font-weight: 700;
}

.cage-card__del {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--text-tertiary);
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.cage-card__del:hover {
  background: rgba(255, 77, 79, 0.15);
  color: #ff4d4f;
}

.cage-card__assets {
  margin-bottom: 12px;
}

.cage-card__total {
  font-size: 22px;
  color: var(--accent);
  font-weight: 700;
  font-family: var(--font-mono);
}

.cage-card__row {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
}

.cage-card__label {
  font-size: 12px;
  color: var(--text-tertiary);
}

.cage-card__value {
  font-size: 12px;
  color: var(--text-secondary);
  font-family: var(--font-mono);
}

.cage-card__footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  padding-top: 10px;
  border-top: 1px solid #2a2a2a;
}

.cage-card__return {
  font-size: 14px;
  font-weight: 700;
  font-family: var(--font-mono);
}

.return--up { color: #00d68f; }
.return--down { color: #ff4d4f; }

.cage-card__date {
  font-size: 11px;
  color: var(--text-tertiary);
}

/* ── 鸡笼详情头 ── */
.cage-detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.cage-detail-name {
  font-size: 18px;
  color: var(--text-primary);
  font-weight: 700;
  margin: 0;
}

.cage-detail-del {
  padding: 6px 12px;
  border-radius: 8px;
  border: 1px solid rgba(255, 77, 79, 0.3);
  background: transparent;
  color: var(--text-tertiary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}

.cage-detail-del:hover {
  background: rgba(255, 77, 79, 0.15);
  color: #ff4d4f;
}

.back-btn {
  display: inline-block;
  margin-bottom: 16px;
  padding: 8px 16px;
  border-radius: 10px;
  border: 1px solid var(--border-color);
  background: transparent;
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.back-btn:hover {
  color: var(--accent);
  border-color: rgba(255, 215, 0, 0.3);
}

.portfolio__loading {
  text-align: center;
  padding: 60px;
  color: var(--text-tertiary);
  font-size: 14px;
}
</style>

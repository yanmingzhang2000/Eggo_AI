<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { get } from '@/utils/request'

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

const emit = defineEmits<{
  (e: 'buy'): void
}>()

const account = ref<AccountSummary | null>(null)

function fmt(val: number): string {
  return val.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function returnColor(val: number): string {
  if (val > 0) return '#ff4d4f'
  if (val < 0) return '#00d68f'
  return '#787878'
}

async function fetchAccount() {
  try {
    const res = await get<AccountSummary>('/portfolio/account')
    if (res.code === 0 && res.data) {
      account.value = res.data
    }
  } catch {}
}

onMounted(fetchAccount)
</script>

<template>
  <div class="acct" v-if="account">
    <!-- 总资产 -->
    <div class="acct__hero">
      <div class="acct__label">总资产（元）</div>
      <div class="acct__total">¥ {{ fmt(account.totalAssets) }}</div>
      <div class="acct__return" :style="{ color: returnColor(account.totalReturn) }">
        {{ account.totalReturn > 0 ? '+' : '' }}{{ account.totalReturn.toFixed(2) }}%
        <span class="acct__profit">
          （{{ account.totalProfit > 0 ? '+' : '' }}{{ fmt(account.totalProfit) }}）
        </span>
      </div>
    </div>

    <!-- 明细 -->
    <div class="acct__detail">
      <div class="acct__item">
        <span class="acct__item-label">可用现金</span>
        <span class="acct__item-val">¥ {{ fmt(account.cashBalance) }}</span>
      </div>
      <div class="acct__item" v-if="account.frozenCash > 0">
        <span class="acct__item-label">冻结资金</span>
        <span class="acct__item-val" style="color: #f7ba1e">¥ {{ fmt(account.frozenCash) }}</span>
      </div>
      <div class="acct__item">
        <span class="acct__item-label">初始资金</span>
        <span class="acct__item-val">¥ {{ fmt(account.initialBalance) }}</span>
      </div>
    </div>

    <!-- 操作 -->
    <div class="acct__actions">
      <button class="act-btn act-btn--buy" @click="emit('buy')">买入基金</button>
    </div>

    <!-- 待结算提示 -->
    <div v-if="account.pendingCount > 0" class="acct__pending">
      ⏳ 有 {{ account.pendingCount }} 笔订单等待今晚净值结算
    </div>
  </div>
</template>

<style scoped>
.acct {
  margin-bottom: 20px;
}

.acct__hero {
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 16px;
  padding: 20px;
  text-align: center;
  margin-bottom: 12px;
}

.acct__label {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-bottom: 4px;
}

.acct__total {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  font-family: var(--font-mono);
  margin-bottom: 4px;
}

.acct__return {
  font-size: 14px;
  font-weight: 600;
  font-family: var(--font-mono);
}

.acct__profit {
  font-size: 12px;
  font-weight: 400;
  opacity: 0.8;
}

.acct__detail {
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 12px;
  padding: 12px 16px;
  margin-bottom: 12px;
}

.acct__item {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
}

.acct__item + .acct__item {
  border-top: 1px solid #2a2a2a;
}

.acct__item-label {
  font-size: 13px;
  color: var(--text-tertiary);
}

.acct__item-val {
  font-size: 13px;
  color: var(--text-secondary);
  font-family: var(--font-mono);
  font-weight: 600;
}

.acct__actions {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
}

.act-btn {
  flex: 1;
  padding: 12px;
  border-radius: 12px;
  border: none;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
}

.act-btn--buy {
  background: var(--accent);
  color: #000;
}

.act-btn--buy:hover {
  background: #ffe066;
}

.acct__pending {
  text-align: center;
  padding: 10px;
  background: rgba(247, 186, 30, 0.08);
  border: 1px solid rgba(247, 186, 30, 0.3);
  border-radius: 10px;
  font-size: 12px;
  color: #f7ba1e;
}
</style>

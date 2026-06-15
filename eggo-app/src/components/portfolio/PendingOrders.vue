<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { get } from '@/utils/request'

interface PendingOrder {
  id: number
  fundCode: string
  fundName: string
  orderType: string
  amount: number
  status: string
  orderDate: string
}

const props = defineProps<{ accountId: number }>()

const orders = ref<PendingOrder[]>([])

function fmt(val: number): string {
  return val.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

async function fetchOrders() {
  try {
    const res = await get<PendingOrder[]>(`/portfolio/accounts/${props.accountId}/orders/pending`)
    if (res.code === 0 && res.data) {
      orders.value = res.data
    }
  } catch {}
}

onMounted(fetchOrders)
watch(() => props.accountId, fetchOrders)
</script>

<template>
  <div class="pending" v-if="orders.length > 0">
    <div class="pending__header">
      <span class="pending__title">待结算订单</span>
      <span class="pending__count">{{ orders.length }} 笔</span>
    </div>

    <div class="pending__list">
      <div v-for="order in orders" :key="order.id" class="pending-card">
        <div class="pending-card__left">
          <span class="pending-card__type" :class="order.orderType === 'buy' ? 'type--buy' : 'type--sell'">
            {{ order.orderType === 'buy' ? '买入' : '卖出' }}
          </span>
          <div class="pending-card__info">
            <span class="pending-card__name">{{ order.fundName || order.fundCode }}</span>
            <span class="pending-card__code">{{ order.fundCode }}</span>
          </div>
        </div>
        <div class="pending-card__right">
          <span class="pending-card__amount">¥ {{ fmt(order.amount) }}</span>
          <span class="pending-card__status">⏳ 等待净值结算</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.pending {
  margin-bottom: 20px;
}

.pending__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.pending__title {
  font-size: 14px;
  color: var(--text-secondary);
  font-weight: 600;
}

.pending__count {
  font-size: 12px;
  color: var(--text-tertiary);
}

.pending__list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.pending-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(145deg, #1c1c1c, #141414);
  border: 1px solid #2a2a2a;
  border-radius: 12px;
  padding: 12px 14px;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { border-color: #2a2a2a; }
  50% { border-color: rgba(247, 186, 30, 0.3); }
}

.pending-card__left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.pending-card__type {
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 4px;
}

.type--buy {
  background: rgba(255, 77, 79, 0.15);
  color: #ff4d4f;
}

.type--sell {
  background: rgba(0, 214, 143, 0.15);
  color: #00d68f;
}

.pending-card__name {
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 600;
  display: block;
}

.pending-card__code {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

.pending-card__right {
  text-align: right;
}

.pending-card__amount {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary);
  font-family: var(--font-mono);
  display: block;
}

.pending-card__status {
  font-size: 11px;
  color: #f7ba1e;
}
</style>

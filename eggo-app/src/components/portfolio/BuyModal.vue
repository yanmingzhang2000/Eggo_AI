<script setup lang="ts">
import { ref } from 'vue'
import { post, get } from '@/utils/request'

const props = defineProps<{ accountId: number }>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'done'): void
}>()

const fundCode = ref('')
const fundName = ref('')
const amount = ref<number>(10000)
const submitting = ref(false)
const errorMsg = ref('')

const quickAmounts = [10000, 50000, 100000, 500000]

async function submit() {
  if (!fundCode.value || amount.value <= 0) return

  submitting.value = true
  errorMsg.value = ''

  try {
    const res = await post(`/portfolio/accounts/${props.accountId}/buy`, {
      fundCode: fundCode.value,
      fundName: fundName.value || fundCode.value,
      amount: amount.value,
    })
    if (res.code === 0) {
      emit('done')
    } else {
      errorMsg.value = res.message || '买入失败'
    }
  } catch (e: any) {
    errorMsg.value = e?.response?.data?.message || '网络异常'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="modal-mask" @click.self="emit('close')">
    <div class="modal">
      <div class="modal__header">
        <span class="modal__title">买入基金</span>
        <button class="modal__close" @click="emit('close')">✕</button>
      </div>

      <div class="modal__body">
        <!-- 基金代码 -->
        <div class="field">
          <label class="field__label">基金代码</label>
          <input
            v-model="fundCode"
            class="field__input"
            placeholder="如 110011"
            maxlength="6"
          />
        </div>

        <!-- 基金名称（可选） -->
        <div class="field">
          <label class="field__label">基金名称（选填）</label>
          <input
            v-model="fundName"
            class="field__input"
            placeholder="方便识别"
          />
        </div>

        <!-- 买入金额 -->
        <div class="field">
          <label class="field__label">买入金额（元）</label>
          <input
            v-model.number="amount"
            type="number"
            class="field__input field__input--amount"
            min="1"
          />
          <div class="field__quick">
            <button
              v-for="q in quickAmounts"
              :key="q"
              class="quick-btn"
              :class="{ 'quick-btn--active': amount === q }"
              @click="amount = q"
            >{{ (q / 10000).toFixed(0) }}万</button>
          </div>
        </div>

        <!-- 提示 -->
        <div class="modal__tip">
          <div class="tip-row">📌 T 日买入，当晚 22:00 后按真实净值结算份额</div>
          <div class="tip-row">⏳ 提交后资金冻结，T+1 日确认份额并开始计算收益</div>
        </div>

        <!-- 错误 -->
        <div v-if="errorMsg" class="modal__error">{{ errorMsg }}</div>
      </div>

      <div class="modal__footer">
        <button class="modal__cancel" @click="emit('close')">取消</button>
        <button
          class="modal__submit"
          @click="submit"
          :disabled="!fundCode || amount <= 0 || submitting"
        >
          {{ submitting ? '提交中...' : '确认买入' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 16px;
}

.modal {
  background: #1a1a1a;
  border: 1px solid #333;
  border-radius: 16px;
  width: 100%;
  max-width: 400px;
  overflow: hidden;
}

.modal__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #2a2a2a;
}

.modal__title {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary);
}

.modal__close {
  width: 28px;
  height: 28px;
  background: transparent;
  border: 1px solid #333;
  color: var(--text-tertiary);
  border-radius: 50%;
  font-size: 12px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.modal__close:hover {
  border-color: #ff4d4f;
  color: #ff4d4f;
}

.modal__body {
  padding: 20px;
}

.field {
  margin-bottom: 16px;
}

.field__label {
  display: block;
  font-size: 12px;
  color: var(--text-tertiary);
  margin-bottom: 6px;
}

.field__input {
  width: 100%;
  padding: 10px 14px;
  background: #0a0a0a;
  border: 1px solid #333;
  border-radius: 10px;
  color: var(--text-primary);
  font-size: 14px;
  font-family: var(--font-mono);
  outline: none;
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.field__input:focus {
  border-color: #f7ba1e;
}

.field__input--amount {
  font-size: 20px;
  font-weight: 700;
  color: var(--accent);
}

.field__quick {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}

.quick-btn {
  flex: 1;
  padding: 6px;
  background: rgba(255, 215, 0, 0.06);
  border: 1px solid rgba(255, 215, 0, 0.2);
  border-radius: 6px;
  color: var(--text-tertiary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.quick-btn:hover,
.quick-btn--active {
  background: rgba(255, 215, 0, 0.15);
  border-color: rgba(255, 215, 0, 0.5);
  color: var(--accent);
}

.modal__tip {
  background: rgba(247, 186, 30, 0.06);
  border: 1px solid rgba(247, 186, 30, 0.15);
  border-radius: 10px;
  padding: 12px;
  margin-bottom: 0;
}

.tip-row {
  font-size: 12px;
  color: var(--text-tertiary);
  line-height: 1.6;
}

.modal__error {
  margin-top: 12px;
  padding: 10px;
  background: rgba(255, 77, 79, 0.1);
  border: 1px solid rgba(255, 77, 79, 0.3);
  border-radius: 8px;
  font-size: 13px;
  color: #ff4d4f;
}

.modal__footer {
  display: flex;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #2a2a2a;
}

.modal__cancel {
  flex: 1;
  padding: 12px;
  background: transparent;
  border: 1px solid #333;
  border-radius: 10px;
  color: var(--text-tertiary);
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.modal__cancel:hover {
  border-color: #555;
  color: var(--text-secondary);
}

.modal__submit {
  flex: 2;
  padding: 12px;
  background: var(--accent);
  border: none;
  border-radius: 10px;
  color: #000;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
}

.modal__submit:hover {
  background: #ffe066;
}

.modal__submit:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
</style>

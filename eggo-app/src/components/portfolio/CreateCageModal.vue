<script setup lang="ts">
import { ref } from 'vue'
import { post } from '@/utils/request'

const emit = defineEmits<{ close: []; done: [] }>()

const name = ref('')
const balance = ref(1000000)
const submitting = ref(false)

async function submit() {
  if (balance.value < 1000 || submitting.value) return
  submitting.value = true
  try {
    const res = await post('/portfolio/accounts', {
      name: name.value || '我的鸡笼',
      initialBalance: balance.value,
    })
    if (res.code === 0) {
      emit('done')
    }
  } catch (e: any) {
    alert(e?.response?.data?.message || '创建失败')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="modal-mask" @click.self="emit('close')">
    <div class="modal">
      <div class="modal__header">
        <h3 class="modal__title">🐣 新建鸡笼</h3>
        <button class="modal__close" @click="emit('close')">×</button>
      </div>

      <div class="modal__body">
        <label class="field-label">鸡笼名称</label>
        <input
          v-model="name"
          type="text"
          class="field-input"
          placeholder="给鸡笼起个名字..."
          maxlength="30"
        />

        <label class="field-label">初始资金</label>
        <input
          v-model.number="balance"
          type="number"
          class="field-input field-input--amount"
          min="1000"
          step="10000"
        />
        <span class="field-hint">至少 1,000 元</span>

        <div class="templates">
          <button class="tmpl-btn" @click="balance = 100000">10万</button>
          <button class="tmpl-btn" @click="balance = 500000">50万</button>
          <button class="tmpl-btn" @click="balance = 1000000">100万</button>
          <button class="tmpl-btn" @click="balance = 3000000">300万</button>
        </div>
      </div>

      <div class="modal__footer">
        <button class="btn btn--cancel" @click="emit('close')">取消</button>
        <button
          class="btn btn--confirm"
          @click="submit"
          :disabled="balance < 1000 || submitting"
        >
          {{ submitting ? '创建中...' : '确定创建' }}
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
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.modal {
  background: #1a1a1a;
  border: 1px solid #2a2a2a;
  border-radius: 20px;
  width: 100%;
  max-width: 420px;
  overflow: hidden;
}

.modal__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px 0;
}

.modal__title {
  font-size: 18px;
  color: var(--text-primary);
  font-weight: 700;
  margin: 0;
}

.modal__close {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: var(--text-tertiary);
  font-size: 20px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal__close:hover { color: var(--text-primary); }

.modal__body { padding: 20px 24px; }

.field-label {
  display: block;
  font-size: 13px;
  color: var(--text-secondary);
  font-weight: 600;
  margin-bottom: 6px;
}

.field-input {
  width: 100%;
  padding: 12px 14px;
  background: #0a0a0a;
  border: 1px solid #333;
  border-radius: 10px;
  color: var(--text-primary);
  font-size: 14px;
  outline: none;
  box-sizing: border-box;
  transition: border-color 0.2s;
}

.field-input:focus { border-color: var(--accent); }

.field-input--amount {
  font-size: 22px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--accent);
}

.field-hint {
  display: block;
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 4px;
}

.templates {
  display: flex;
  gap: 8px;
  margin-top: 14px;
}

.tmpl-btn {
  flex: 1;
  padding: 8px 0;
  border-radius: 8px;
  border: 1px solid rgba(255, 215, 0, 0.3);
  background: rgba(255, 215, 0, 0.06);
  color: var(--accent);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.tmpl-btn:hover {
  background: rgba(255, 215, 0, 0.15);
  border-color: rgba(255, 215, 0, 0.6);
}

.modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 0 24px 20px;
}

.btn {
  padding: 10px 24px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.btn--cancel {
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-tertiary);
}

.btn--cancel:hover { color: var(--text-secondary); }

.btn--confirm {
  background: var(--accent);
  color: #000;
}

.btn--confirm:hover { background: #ffe066; }

.btn--confirm:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
</style>

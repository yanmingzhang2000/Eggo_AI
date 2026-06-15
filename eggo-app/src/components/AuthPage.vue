<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

const mode = ref<'login' | 'register'>('login')
const username = ref('')
const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function handleSubmit() {
  error.value = ''
  loading.value = true

  try {
    if (mode.value === 'login') {
      const result = await authStore.login(username.value, password.value)
      if (!result.success) {
        error.value = result.message || '登录失败'
      }
    } else {
      if (!email.value) {
        error.value = '请输入邮箱'
        return
      }
      const result = await authStore.register(username.value, email.value, password.value)
      if (!result.success) {
        error.value = result.message || '注册失败'
      }
    }
  } finally {
    loading.value = false
  }
}

async function handleGuest() {
  loading.value = true
  await authStore.guestLogin()
  loading.value = false
}

function toggleMode() {
  mode.value = mode.value === 'login' ? 'register' : 'login'
  error.value = ''
}
</script>

<template>
  <div class="auth-page">
    <div class="auth-card">
      <!-- Logo -->
      <div class="auth-card__header">
        <img src="/logo.svg" class="auth-card__logo" alt="鸡生蛋" />
        <h1 class="auth-card__title">鸡生蛋</h1>
        <p class="auth-card__subtitle">Eggo · 智能基金决策</p>
      </div>

      <!-- 表单 -->
      <form class="auth-card__form" @submit.prevent="handleSubmit">
        <h2 class="auth-card__form-title">
          {{ mode === 'login' ? '登录' : '注册' }}
        </h2>

        <!-- 错误提示 -->
        <div v-if="error" class="auth-card__error">
          {{ error }}
        </div>

        <!-- 用户名 -->
        <div class="auth-card__field">
          <label class="auth-card__label">用户名</label>
          <input
            v-model="username"
            type="text"
            class="auth-card__input"
            placeholder="请输入用户名"
            :disabled="loading"
          />
        </div>

        <!-- 邮箱（注册时显示） -->
        <div v-if="mode === 'register'" class="auth-card__field">
          <label class="auth-card__label">邮箱</label>
          <input
            v-model="email"
            type="email"
            class="auth-card__input"
            placeholder="请输入邮箱"
            :disabled="loading"
          />
        </div>

        <!-- 密码 -->
        <div class="auth-card__field">
          <label class="auth-card__label">密码</label>
          <input
            v-model="password"
            type="password"
            class="auth-card__input"
            placeholder="请输入密码"
            :disabled="loading"
          />
        </div>

        <!-- 提交按钮 -->
        <button type="submit" class="auth-card__submit" :disabled="loading">
          {{ loading ? '处理中...' : (mode === 'login' ? '登录' : '注册') }}
        </button>

        <!-- 切换登录/注册 -->
        <p class="auth-card__switch" @click="toggleMode">
          {{ mode === 'login' ? '没有账号？立即注册' : '已有账号？立即登录' }}
        </p>
      </form>

      <!-- 分割线 -->
      <div class="auth-card__divider">
        <span class="auth-card__divider-text">或</span>
      </div>

      <!-- 游客模式 -->
      <button class="auth-card__guest" @click="handleGuest" :disabled="loading">
        👤 游客模式体验
      </button>

      <p class="auth-card__guest-hint">
        游客模式可查看演示数据，注册后可管理自选基金
      </p>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-primary);
  padding: 20px;
}

.auth-card {
  width: 100%;
  max-width: 400px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 24px;
  padding: 40px 32px;
}

.auth-card__header {
  text-align: center;
  margin-bottom: 32px;
}

.auth-card__logo {
  width: 64px;
  height: 64px;
  display: block;
  margin: 0 auto 12px;
  object-fit: contain;
}

.auth-card__title {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
  letter-spacing: 2px;
}

.auth-card__subtitle {
  font-size: 12px;
  color: var(--text-tertiary);
  margin: 4px 0 0;
  letter-spacing: 4px;
  text-transform: uppercase;
}

.auth-card__form {
  margin-bottom: 24px;
}

.auth-card__form-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 20px;
}

.auth-card__error {
  background: rgba(255, 77, 79, 0.1);
  border: 1px solid rgba(255, 77, 79, 0.2);
  border-radius: 8px;
  padding: 10px 14px;
  margin-bottom: 16px;
  font-size: 13px;
  color: #ff4d4f;
}

.auth-card__field {
  margin-bottom: 16px;
}

.auth-card__label {
  display: block;
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.auth-card__input {
  width: 100%;
  padding: 12px 14px;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  color: var(--text-primary);
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.auth-card__input:focus {
  border-color: var(--accent);
}

.auth-card__input::placeholder {
  color: var(--text-tertiary);
}

.auth-card__submit {
  width: 100%;
  padding: 12px;
  background: var(--accent);
  color: #000;
  border: none;
  border-radius: 10px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
  margin-top: 8px;
}

.auth-card__submit:hover {
  opacity: 0.9;
}

.auth-card__submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.auth-card__switch {
  text-align: center;
  font-size: 13px;
  color: var(--accent);
  cursor: pointer;
  margin: 16px 0 0;
}

.auth-card__switch:hover {
  text-decoration: underline;
}

.auth-card__divider {
  display: flex;
  align-items: center;
  margin: 24px 0;
}

.auth-card__divider::before,
.auth-card__divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--border-color);
}

.auth-card__divider-text {
  padding: 0 16px;
  font-size: 12px;
  color: var(--text-tertiary);
}

.auth-card__guest {
  width: 100%;
  padding: 12px;
  background: transparent;
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.auth-card__guest:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.auth-card__guest:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.auth-card__guest-hint {
  text-align: center;
  font-size: 11px;
  color: var(--text-tertiary);
  margin: 12px 0 0;
}
</style>

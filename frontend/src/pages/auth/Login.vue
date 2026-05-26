<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '@/api'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const phone = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function handleLogin() {
  if (!phone.value || !password.value) {
    error.value = '请输入手机号和密码'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const res: any = await authApi.login({
      phone: phone.value,
      password: password.value,
    })
    if (res.code === 0) {
      auth.setAccessToken(res.data.access_token)
      await auth.fetchUserInfo()
      router.push('/')
    } else {
      error.value = res.message || '登录失败'
    }
  } catch (err: any) {
    error.value = err.message || '网络错误'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page-container">
    <div class="blue-accent"></div>
    <div class="content-wrapper">
      <!-- 品牌头部 -->
      <div class="header">
        <h1 class="title">易账</h1>
        <p class="subtitle">简单记账，清晰生活</p>
      </div>

      <!-- 登录表单 -->
      <div class="form-section">
        <form @submit.prevent="handleLogin" class="form">
          <div class="input-line">
            <label>手机号码</label>
            <input
              type="tel"
              v-model="phone"
              placeholder="请输入11位手机号"
              autocomplete="tel"
            />
          </div>

          <div class="input-line">
            <label>登录密码</label>
            <input
              type="password"
              v-model="password"
              placeholder="请输入登录密码"
              autocomplete="current-password"
            />
          </div>

          <div class="error-wrap" v-if="error">
            <svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="2" fill="none"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
            {{ error }}
          </div>

          <button type="submit" class="submit-btn" :class="{ 'is-loading': loading }" :disabled="loading">
            <span class="btn-text">立即登录</span>
            <div class="spinner"></div>
          </button>
        </form>

        <div class="card-footer">
          <span class="text-mute">没有账号？</span>
          <router-link to="/register" replace class="text-link">去注册</router-link>
        </div>
      </div>

      <div class="bottom-links">
        <span class="safe-text">已启用安全加密传输</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-container {
  min-height: 100vh;
  background-color: var(--bg-base);
  display: flex;
  flex-direction: column;
}
.blue-accent {
  height: 80px;
  background: var(--primary);
  flex-shrink: 0;
}
.content-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 0 32px;
  padding-bottom: calc(24px + env(safe-area-inset-bottom));
  max-width: 460px;
  margin: 0 auto;
  width: 100%;
}
.header {
  padding-top: 32px;
  margin-bottom: 48px;
}
.title {
  font-size: 28px;
  font-weight: 800;
  color: var(--primary);
  margin: 0 0 8px;
}
.subtitle {
  font-size: 15px;
  color: var(--text-secondary);
  margin: 0;
}

.form-section {
  flex: 1;
}
.form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.input-line {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.input-line label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}
.input-line input {
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--bg-card);
  padding: 14px;
  font-size: 14px;
  color: var(--text-primary);
  outline: none;
  transition: border-color 0.2s;
  -webkit-appearance: none;
  appearance: none;
}
.input-line input:focus {
  border-color: var(--primary);
}
.input-line input::placeholder {
  color: var(--text-tertiary);
  font-size: 14px;
}

.error-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--danger);
  font-size: 13px;
  background: var(--danger-soft);
  padding: 10px 14px;
  border-radius: 6px;
}

.submit-btn {
  background: var(--primary);
  color: white;
  border: none;
  border-radius: 8px;
  height: 48px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  justify-content: center;
  align-items: center;
  position: relative;
  overflow: hidden;
}
.submit-btn:active { background: var(--primary-dark); }
.submit-btn:disabled { opacity: 0.7; cursor: not-allowed; }
.submit-btn.is-loading .btn-text { opacity: 0; }
.spinner {
  position: absolute;
  width: 20px;
  height: 20px;
  border: 3px solid rgba(255,255,255,0.3);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 0.8s linear infinite;
  opacity: 0;
  transition: opacity 0.2s;
}
.submit-btn.is-loading .spinner { opacity: 1; }
@keyframes spin { to { transform: rotate(360deg); } }

.card-footer {
  margin-top: 24px;
  text-align: center;
  font-size: 13px;
}
.text-mute { color: var(--text-secondary); }
.text-link {
  color: var(--primary);
  font-weight: 600;
  text-decoration: none;
  margin-left: 4px;
}

.bottom-links {
  margin-top: auto;
  padding-top: 24px;
  text-align: center;
}
.safe-text {
  font-size: 12px;
  color: var(--text-tertiary);
}
</style>

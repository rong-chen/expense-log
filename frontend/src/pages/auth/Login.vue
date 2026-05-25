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
    <div class="content-wrapper">
      <!-- 品牌头部 (左对齐，更精干) -->
      <div class="header">
        <div class="logo-box">
          <img src="/icon-192.png" alt="易账 Logo" />
        </div>
        <div>
          <h1 class="title">易账</h1>
          <p class="subtitle">让记账成为一种享受</p>
        </div>
      </div>

      <!-- 登录卡片 -->
      <div class="auth-card">
        <h2 class="card-title">登录</h2>
        
        <form @submit.prevent="handleLogin" class="form">
          <div class="input-line">
            <label>手机号码</label>
            <input 
              type="tel" 
              v-model="phone" 
              placeholder="请输入 11 位手机号" 
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
        <span class="safe-text">已启用安全加密传输，守护您的隐私</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-container {
  min-height: 100vh;
  min-height: 100%;
  background-color: var(--bg-base);
  display: flex;
  flex-direction: column;
}

.content-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  align-items: center;
  padding: calc(4vh + env(safe-area-inset-top)) 24px calc(24px + env(safe-area-inset-bottom));
  max-width: 460px;
  margin: 0 auto;
  width: 100%;
}

.header {
  display: flex;
  align-items: center;
  gap: 16px;
  width: 100%;
  margin-bottom: 32px;
  padding-left: 8px;
}
.logo-box {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  overflow: hidden;
  flex-shrink: 0;
}
.logo-box img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.title {
  font-size: 26px;
  font-weight: 800;
  color: var(--primary);
  margin: 0 0 2px;
}
.subtitle {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0;
}

.auth-card {
  background: var(--bg-card);
  width: 100%;
  border-radius: 8px;
  padding: 32px 28px;
  border: 1px solid var(--border);
}
.card-title {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 28px;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.input-line {
  display: flex;
  flex-direction: column;
  gap: 6px;
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
  padding: 12px 14px;
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
  margin-top: -4px;
}

.submit-btn {
  background: var(--primary);
  color: white;
  border: none;
  border-radius: 8px;
  padding: 14px;
  font-size: 15px;
  font-weight: 600;
  margin-top: 8px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  justify-content: center;
  align-items: center;
  position: relative;
  overflow: hidden;
}
.submit-btn:active {
  background: var(--primary-dark);
}
.submit-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
.submit-btn.is-loading .btn-text {
  opacity: 0;
}
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
.submit-btn.is-loading .spinner {
  opacity: 1;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}

.card-footer {
  margin-top: 24px;
  text-align: center;
  font-size: 14px;
}
.text-mute {
  color: var(--text-secondary);
}
.text-link {
  color: var(--primary);
  font-weight: 600;
  text-decoration: none;
  margin-left: 2px;
}

.bottom-links {
  margin-top: 40px;
  text-align: center;
}
.safe-text {
  font-size: 12px;
  color: var(--text-tertiary);
}
</style>

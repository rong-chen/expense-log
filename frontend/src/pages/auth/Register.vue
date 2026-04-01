<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '@/api'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const phone = ref('')
const password = ref('')
const confirmPassword = ref('')
const invitationCode = ref('')
const loading = ref(false)
const error = ref('')

async function handleRegister() {
  if (!phone.value || !password.value || !invitationCode.value) {
    error.value = '请填写所有必填项'
    return
  }
  if (password.value.length < 6) {
    error.value = '密码至少 6 位'
    return
  }
  if (password.value !== confirmPassword.value) {
    error.value = '两次密码不一致'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const res: any = await authApi.register({
      phone: phone.value,
      password: password.value,
      invitation_code: invitationCode.value,
    })
    if (res.code === 0) {
      auth.setAccessToken(res.data.access_token)
      await auth.fetchUserInfo()
      router.push('/')
    } else {
      error.value = res.message || '注册失败'
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
    <!-- 背景几何装饰 -->
    <div class="bg-shape shape-1"></div>
    <div class="bg-shape shape-2"></div>
    
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

      <!-- 注册卡片 -->
      <div class="auth-card">
        <h2 class="card-title">快捷注册</h2>
        
        <form @submit.prevent="handleRegister" class="form">
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
            <label>设置密码</label>
            <input 
              type="password" 
              v-model="password" 
              placeholder="设置密码（至少6位）" 
              autocomplete="new-password" 
            />
          </div>

          <div class="input-line">
            <label>邀请码</label>
            <input 
              type="text" 
              v-model="invitationCode" 
              placeholder="请输入 8 位邀请码" 
              autocomplete="off" 
            />
          </div>

          <div class="error-wrap" v-if="error">
            <svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="2" fill="none"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
            {{ error }}
          </div>

          <button type="submit" class="submit-btn" :class="{ 'is-loading': loading }" :disabled="loading">
            <span class="btn-text">立即创建</span>
            <div class="spinner"></div>
          </button>
        </form>

        <div class="card-footer">
          <span class="text-mute">已有账号？</span>
          <router-link to="/login" replace class="text-link">去登录</router-link>
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
  position: relative;
  overflow: hidden;
}

/* 几何背景图 */
.bg-shape {
  position: absolute;
  border-radius: 50%;
  pointer-events: none;
  z-index: 0;
}
.shape-1 {
  width: 600px;
  height: 600px;
  background: radial-gradient(circle, rgba(230,126,34,0.06) 0%, rgba(230,126,34,0) 70%);
  top: -200px;
  right: -250px;
}
.shape-2 {
  width: 450px;
  height: 450px;
  background: radial-gradient(circle, rgba(230,126,34,0.04) 0%, rgba(230,126,34,0) 70%);
  bottom: -100px;
  left: -200px;
}

.content-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  align-items: center;
  padding: calc(4vh + env(safe-area-inset-top)) 24px calc(24px + env(safe-area-inset-bottom));
  position: relative;
  z-index: 1;
  max-width: 460px;
  margin: 0 auto;
  width: 100%;
}

/* 头部信息 */
.header {
  display: flex;
  align-items: center;
  gap: 16px;
  width: 100%;
  margin-bottom: 32px;
  padding-left: 8px; /* 轻微缩进对齐卡片内容 */
}
.logo-box {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 8px 16px rgba(230, 126, 34, 0.2);
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
  color: #1a1e23;
  margin: 0 0 2px;
  letter-spacing: 0.5px;
}
.subtitle {
  font-size: 14px;
  color: #64748b;
  margin: 0;
}

/* 核心卡片 */
.auth-card {
  background: white;
  width: 100%;
  border-radius: 24px;
  padding: 32px 28px;
  box-shadow: 0 16px 40px rgba(0,0,0,0.04), 0 4px 12px rgba(0,0,0,0.02);
  border: 1px solid rgba(255,255,255,0.8);
}
.card-title {
  font-size: 20px;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 28px;
}

/* 表单输入线 */
.form {
  display: flex;
  flex-direction: column;
  gap: 22px;
}
.input-line {
  display: flex;
  align-items: center;
  border-bottom: 1.5px solid #e2e8f0;
  padding-bottom: 12px;
  transition: border-color 0.2s;
}
.input-line:focus-within {
  border-bottom-color: var(--primary);
}
.input-line label {
  width: 66px;
  font-size: 15px;
  font-weight: 600;
  color: #334155;
  flex-shrink: 0;
  display: inline-block;
  text-align: justify;
  text-align-last: justify; /* Aligns 4-char text cleanly */
}
.input-line input {
  flex: 1;
  border: none;
  background: transparent;
  padding: 0 8px;
  font-size: 16px;
  color: #0f172a;
  outline: none;
  -webkit-appearance: none;
  appearance: none;
}
.input-line input::placeholder {
  color: #94a3b8;
  font-size: 15px;
}

/* 错误提示 */
.error-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #ef4444;
  font-size: 13px;
  background: #fef2f2;
  padding: 10px 14px;
  border-radius: 8px;
  margin-top: -8px;
}

/* 按钮 */
.submit-btn {
  background: var(--primary);
  color: white;
  border: none;
  border-radius: 12px;
  padding: 16px;
  font-size: 16px;
  font-weight: 600;
  margin-top: 12px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  justify-content: center;
  align-items: center;
  position: relative;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(230, 126, 34, 0.2);
}
.submit-btn:active {
  transform: scale(0.98);
  background: #d35400; /* Darker orange */
  box-shadow: 0 2px 6px rgba(230, 126, 34, 0.15);
}
.submit-btn:disabled {
  opacity: 0.8;
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

/* 卡片底部 */
.card-footer {
  margin-top: 24px;
  text-align: center;
  font-size: 14px;
}
.text-mute {
  color: #94a3b8;
}
.text-link {
  color: var(--primary);
  font-weight: 600;
  text-decoration: none;
  margin-left: 2px;
  transition: opacity 0.2s;
}
.text-link:active {
  opacity: 0.7;
}

/* 全页底部 */
.bottom-links {
  margin-top: 40px;
  text-align: center;
}
.safe-text {
  font-size: 12px;
  color: #cbd5e1;
  letter-spacing: 0.5px;
}
</style>

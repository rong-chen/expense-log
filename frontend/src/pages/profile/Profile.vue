<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { Lock, LogOut, ChevronRight, Settings, CalendarClock } from 'lucide-vue-next'

const auth = useAuthStore()
const router = useRouter()
const envInfo = ref('')

function logout() {
  auth.logout()
  router.replace('/login')
}

onMounted(() => {
  if (!auth.user) auth.fetchUserInfo()

  // 环境类型检测
  const ua = navigator.userAgent
  let device = 'PC端浏览器'
  if (/android/i.test(ua)) device = 'Android'
  else if (/iPad|iPhone|iPod/.test(ua)) device = 'iOS'
  else if (/Mac/i.test(ua)) device = 'macOS'
  else if (/Win/i.test(ua)) device = 'Windows'

  const isStandalone = window.matchMedia('(display-mode: standalone)').matches || 
                       (navigator as any).standalone || 
                       document.referrer.includes('android-app://')
  
  const mode = isStandalone ? 'PWA 独立渲染组件' : 'Web 浏览器模式'
  envInfo.value = `${device} · ${mode}`
})
</script>

<template>
  <div class="profile-page">
    <div class="user-card" v-if="auth.user">
      <div class="avatar">{{ auth.user.email.charAt(0).toUpperCase() }}</div>
      <div class="user-info">
        <h2>{{ auth.user.nickname || '未设置昵称' }}</h2>
        <p>{{ auth.user.email }}</p>
      </div>
    </div>

    <div class="menu-list">

      <div class="menu-item" @click="router.push('/ukey')">
        <div class="menu-icon" style="background: rgba(41, 128, 185, 0.12); color: var(--primary)">
          <Settings :size="20" />
        </div>
        <div class="menu-content">
          <span>系统设置</span>
          <ChevronRight :size="18" class="chevron" />
        </div>
      </div>

      <div class="menu-item" @click="router.push('/recurring')">
        <div class="menu-icon" style="background: rgba(155, 89, 182, 0.12); color: #9b59b6">
          <CalendarClock :size="20" />
        </div>
        <div class="menu-content">
          <span>周期账单</span>
          <ChevronRight :size="18" class="chevron" />
        </div>
      </div>
      
      <div class="menu-item" @click="router.push('/password')">
        <div class="menu-icon" style="background: rgba(26, 188, 156, 0.12); color: var(--info)">
          <Lock :size="20" />
        </div>
        <div class="menu-content">
          <span>修改密码</span>
          <ChevronRight :size="18" class="chevron" />
        </div>
      </div>
    </div>

    <button class="logout-btn" @click="logout">
      <LogOut :size="18" />
      退出登录
    </button>

    <!-- 环境检测底部标识 -->
    <div class="env-indicator">
      Current Env: {{ envInfo }}
    </div>
  </div>
</template>

<style scoped>
.profile-page {
  padding: 16px;
  padding-top: calc(16px + env(safe-area-inset-top));
}
.user-card {
  display: flex;
  align-items: center;
  background: white;
  padding: 24px;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);
  margin-bottom: 24px;
}
.avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: var(--primary);
  color: white;
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 2.2rem;
  font-weight: 600;
  margin-right: 16px;
  box-shadow: 0 4px 12px rgba(230, 126, 34, 0.2);
}
.user-info h2 {
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0 0 6px 0;
  color: var(--text-primary);
}
.user-info p {
  color: var(--text-secondary);
  margin: 0;
  font-size: 0.9rem;
}

.menu-list {
  background: white;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);
  margin-bottom: 24px;
}
.menu-item {
  display: flex;
  align-items: center;
  padding: 16px;
  cursor: pointer;
}
.menu-item:active {
  background: rgba(0, 0, 0, 0.02);
}
.menu-item:not(:last-child) .menu-content {
  border-bottom: 1px solid rgba(0,0,0,0.05);
}
.menu-icon {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  justify-content: center;
  align-items: center;
  margin-right: 16px;
}
.menu-content {
  flex: 1;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  margin: -16px 0;
  font-weight: 500;
  color: var(--text-primary);
  font-size: 1.05rem;
}
.chevron {
  color: var(--text-muted);
}

.logout-btn {
  width: 100%;
  background: white;
  color: var(--danger);
  border: none;
  padding: 18px;
  border-radius: 16px;
  font-size: 1.05rem;
  font-weight: 600;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);
  cursor: pointer;
  transition: transform 0.1s;
}
.logout-btn:active {
  transform: scale(0.98);
}

.env-indicator {
  text-align: center;
  font-size: 0.75rem;
  color: var(--text-tertiary, #95a5a6);
  margin-top: 32px;
  padding-bottom: 12px;
  opacity: 0.8;
}
</style>

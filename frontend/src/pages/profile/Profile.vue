<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { Lock, LogOut, ChevronRight, Settings, CalendarClock, ShieldCheck, Tag } from 'lucide-vue-next'

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

      <div class="menu-item" v-if="auth.user?.role === 'admin'" @click="router.push('/admin')">
        <div class="menu-icon" style="background: rgba(217, 119, 6, 0.08); color: #D97706">
          <ShieldCheck :size="20" />
        </div>
        <div class="menu-content">
          <span>管理后台</span>
          <ChevronRight :size="18" class="chevron" />
        </div>
      </div>

      <div class="menu-item" @click="router.push('/ukey')">
        <div class="menu-icon" style="background: rgba(37, 99, 235, 0.08); color: var(--primary)">
          <Settings :size="20" />
        </div>
        <div class="menu-content">
          <span>系统设置</span>
          <ChevronRight :size="18" class="chevron" />
        </div>
      </div>

      <div class="menu-item" @click="router.push('/recurring')">
        <div class="menu-icon" style="background: rgba(124, 58, 237, 0.08); color: #7C3AED">
          <CalendarClock :size="20" />
        </div>
        <div class="menu-content">
          <span>周期账单</span>
          <ChevronRight :size="18" class="chevron" />
        </div>
      </div>

      <div class="menu-item" @click="router.push('/tags')">
        <div class="menu-icon" style="background: rgba(8, 145, 178, 0.08); color: #0891B2">
          <Tag :size="20" />
        </div>
        <div class="menu-content">
          <span>标签管理</span>
          <ChevronRight :size="18" class="chevron" />
        </div>
      </div>
      
      <div class="menu-item" @click="router.push('/password')">
        <div class="menu-icon" style="background: rgba(100, 116, 139, 0.08); color: #64748B">
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
  background: var(--bg-card);
  padding: 20px;
  border-radius: 8px;
  border: 1px solid var(--border);
  margin-bottom: 16px;
}
.avatar {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: var(--primary);
  color: white;
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 1.8rem;
  font-weight: 600;
  margin-right: 14px;
}
.user-info h2 {
  font-size: 1.1rem;
  font-weight: 700;
  margin: 0 0 4px 0;
  color: var(--text-primary);
}
.user-info p {
  color: var(--text-secondary);
  margin: 0;
  font-size: 0.85rem;
}

.menu-list {
  background: var(--bg-card);
  border-radius: 8px;
  border: 1px solid var(--border);
  overflow: hidden;
  margin-bottom: 16px;
}
.menu-item {
  display: flex;
  align-items: center;
  padding: 14px 16px;
  cursor: pointer;
}
.menu-item:active {
  background: var(--bg-base);
}
.menu-item:not(:last-child) .menu-content {
  border-bottom: 1px solid var(--border-light);
}
.menu-icon {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  justify-content: center;
  align-items: center;
  margin-right: 14px;
}
.menu-content {
  flex: 1;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 0;
  margin: -14px 0;
  font-weight: 500;
  color: var(--text-primary);
  font-size: 0.95rem;
}
.chevron {
  color: var(--text-tertiary);
}

.logout-btn {
  width: 100%;
  background: var(--bg-card);
  color: var(--danger);
  border: 1px solid var(--border);
  padding: 14px;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 600;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}
.logout-btn:active {
  background: var(--danger-soft);
}

.env-indicator {
  text-align: center;
  font-size: 0.75rem;
  color: var(--text-tertiary);
  margin-top: 32px;
  padding-bottom: 12px;
}
</style>

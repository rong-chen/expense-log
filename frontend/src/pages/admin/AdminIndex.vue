<script setup lang="ts">
import { useRouter } from 'vue-router'
import { UserPlus, Users, ChevronRight } from 'lucide-vue-next'
import TopNavBar from '@/components/layout/TopNavBar.vue'

const router = useRouter()

const adminMenus = [
  {
    title: '邀请码管理',
    desc: '生成并管理新用户注册所需的邀请码',
    icon: UserPlus,
    path: '/admin/invitation',
    color: '#2563EB',
    bg: 'rgba(37, 99, 235, 0.08)'
  },
  {
    title: '用户管理',
    desc: '查看注册用户列表，管理用户角色权限',
    icon: Users,
    path: '/admin/users',
    color: '#0891B2',
    bg: 'rgba(8, 145, 178, 0.08)'
  }
]
</script>

<template>
  <div class="admin-index-page">
    <TopNavBar title="管理中心" />

    <div class="content">
      <div class="welcome-section">
        <h3>欢迎使用管理后端</h3>
        <p>请选择您需要执行的管理任务</p>
      </div>

      <div class="menu-grid">
        <div 
          v-for="menu in adminMenus" 
          :key="menu.path" 
          class="menu-card"
          @click="router.push(menu.path)"
        >
          <div class="menu-icon" :style="{ backgroundColor: menu.bg, color: menu.color }">
            <component :is="menu.icon" :size="24" />
          </div>
          <div class="menu-info">
            <h4>{{ menu.title }}</h4>
            <p>{{ menu.desc }}</p>
          </div>
          <ChevronRight :size="20" class="arrow" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.admin-index-page {
  min-height: 100vh;
  background: var(--bg-base);
  padding: 16px;
  padding-top: calc(72px + env(safe-area-inset-top));
}

.content {
  max-width: 600px;
  margin: 0 auto;
}

.welcome-section {
  margin-bottom: 24px;
}
.welcome-section h3 {
  font-size: 1.25rem;
  font-weight: 800;
  color: var(--text-primary);
  margin: 0 0 6px 0;
}
.welcome-section p {
  color: var(--text-secondary);
  font-size: 0.9rem;
  margin: 0;
}

.menu-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.menu-card {
  background: var(--bg-card);
  border-radius: 8px;
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 14px;
  border: 1px solid var(--border);
  cursor: pointer;
  transition: all 0.2s;
}

.menu-card:active {
  background: var(--bg-card-hover);
}

.menu-icon {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.menu-info {
  flex: 1;
}

.menu-info h4 {
  font-size: 1.05rem;
  font-weight: 700;
  margin: 0 0 4px 0;
  color: var(--text-primary);
}

.menu-info p {
  font-size: 0.8rem;
  color: var(--text-secondary);
  margin: 0;
  line-height: 1.4;
}

.arrow {
  color: var(--text-tertiary);
}
</style>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { BarChart3, Settings, PlusCircle } from 'lucide-vue-next'

const route = useRoute()

const navItems = [
  { name: 'dashboard', label: '概览', icon: BarChart3, path: '/' },
  { name: 'settings', label: '设置', icon: Settings, path: '/settings' },
]
</script>

<template>
  <nav class="bottom-nav">
    <!-- 左侧导航 -->
    <router-link
      :to="navItems[0].path"
      class="nav-item"
      :class="{ active: route.name === navItems[0].name }"
    >
      <component :is="navItems[0].icon" :size="22" stroke-width="2.5" />
      <span>{{ navItems[0].label }}</span>
    </router-link>

    <!-- 中间悬浮按钮 (发票上传/记账) -->
    <div class="fab-wrapper">
      <button class="fab-btn">
        <PlusCircle :size="28" color="white" stroke-width="2.5" />
      </button>
    </div>

    <!-- 右侧导航 -->
    <router-link
      :to="navItems[1].path"
      class="nav-item"
      :class="{ active: route.name === navItems[1].name }"
    >
      <component :is="navItems[1].icon" :size="22" stroke-width="2.5" />
      <span>{{ navItems[1].label }}</span>
    </router-link>
  </nav>
</template>

<style scoped>
.bottom-nav {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 64px;
  background: var(--bg-card);
  border-top: 1px solid var(--border-light);
  display: flex;
  justify-content: space-around;
  align-items: center;
  z-index: 1000;
  padding-bottom: env(safe-area-inset-bottom);
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.04);
}

.dark .bottom-nav {
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.2);
}

.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  color: var(--text-tertiary);
  font-size: 0.75rem;
  font-weight: 500;
  text-decoration: none;
  width: 60px;
  transition: color 0.2s ease;
}

.nav-item.active {
  color: var(--primary);
}

.fab-wrapper {
  position: relative;
  width: 60px;
  height: 60px;
  display: flex;
  justify-content: center;
}

.fab-btn {
  position: absolute;
  top: -24px;
  width: 56px;
  height: 56px;
  border-radius: 28px;
  background: linear-gradient(135deg, var(--primary), var(--primary-light));
  border: 4px solid var(--bg-card);
  box-shadow: 0 8px 20px rgba(230, 126, 34, 0.35);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s;
  padding: 0;
}

.fab-btn:active {
  transform: scale(0.92);
}
</style>

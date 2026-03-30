<script setup lang="ts">
import { Home, User, Receipt } from 'lucide-vue-next'
</script>

<template>
  <nav class="bottom-nav">
    <router-link to="/home" class="nav-item" active-class="active">
      <Home :size="22" strokeWidth="2" />
      <span>总览</span>
    </router-link>
    <router-link to="/bills" class="nav-item" active-class="active">
      <Receipt :size="22" strokeWidth="2" />
      <span>明细</span>
    </router-link>
    <router-link to="/profile" class="nav-item" active-class="active">
      <User :size="22" strokeWidth="2" />
      <span>我的</span>
    </router-link>
  </nav>
</template>

<style scoped>
.bottom-nav {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: rgba(255, 255, 255, 0.98);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-top: none;
  box-shadow: 0 -4px 24px rgba(44, 36, 23, 0.04);
  display: flex;
  justify-content: space-around;
  align-items: stretch;
  z-index: 1000;
  padding: 8px 0;
  padding-bottom: calc(8px + env(safe-area-inset-bottom));
}
.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #64748b;
  text-decoration: none;
  font-size: 0.75rem;
  font-weight: 500;
  gap: 3px;
  flex: 1;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}
.nav-item.active {
  color: var(--primary); /* 品牌橙色 */
  font-weight: 700;
  transform: translateY(-2px); /* 选中时稍微上浮 */
}
.nav-item:active {
  transform: scale(0.92);
}

/* 如果 App 切换到暗黑模式，既然用户要求不要黑色，我们给它一个非常优雅的浅银灰色而不是死黑 */
:global(.dark) .bottom-nav {
  background: rgba(240, 240, 243, 0.95);
  box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.1);
}
</style>

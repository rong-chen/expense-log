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
  background: var(--bg-card);
  border-top: 1px solid var(--border);
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
  color: var(--text-tertiary);
  text-decoration: none;
  font-size: 0.7rem;
  font-weight: 500;
  gap: 2px;
  flex: 1;
  transition: color 0.2s;
}
.nav-item.active {
  color: var(--primary);
  font-weight: 700;
}
.nav-item:active {
  opacity: 0.7;
}
</style>

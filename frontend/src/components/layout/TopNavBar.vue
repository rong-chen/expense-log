<script setup lang="ts">
import { ChevronLeft } from 'lucide-vue-next'
import { useRouter } from 'vue-router'

defineProps<{ title: string }>()
const router = useRouter()

function goBack() {
  router.back()
}
</script>

<template>
  <header class="top-nav-bar">
    <button class="back-btn" @click="goBack" aria-label="返回">
      <ChevronLeft :size="28" />
    </button>
    <h1 class="title">{{ title }}</h1>
    <div class="right-slot">
      <slot name="right"></slot>
    </div>
  </header>
</template>

<style scoped>
.top-nav-bar {
  position: fixed;
  top: 0;
  left: -1px;
  right: -1px;
  height: 56px;
  background: var(--bg-card);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  padding-top: env(safe-area-inset-top);
  box-sizing: content-box;
  z-index: 1000;
  border-bottom: 1px solid var(--border);
}
.back-btn {
  background: none;
  border: none;
  padding: 8px 12px 8px 0;
  margin-left: 0;
  cursor: pointer;
  color: var(--text-primary);
  display: flex;
  align-items: center;
}
.title {
  font-size: 1.05rem;
  font-weight: 700;
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  margin: 0;
  color: var(--text-primary);
}
.right-slot {
  min-width: 44px;
  display: flex;
  justify-content: flex-end;
}
</style>

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
  background: rgba(250, 248, 245, 0.85); /* 半透明暖白高斯模糊 */
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  padding-top: env(safe-area-inset-top);
  box-sizing: content-box;
  z-index: 1000;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
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
  font-size: 1.15rem;
  font-weight: 600;
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  margin: 0;
  color: var(--text-primary);
}
.right-slot {
  min-width: 44px; /* 占位，保持标题居中 */
  display: flex;
  justify-content: flex-end;
}
/* 同样拒绝死黑，跟随全站浅色/银灰高级质感 */
:global(.dark) .top-nav-bar {
  background: rgba(240, 240, 243, 0.95);
  border-bottom-color: rgba(0, 0, 0, 0.05); /* 顶部不需要特别显眼的浅色白线 */
}
</style>

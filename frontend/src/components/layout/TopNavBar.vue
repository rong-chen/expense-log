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
  background: var(--primary); /* 🌟 改为定义的蓝色方案 */
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  padding-top: env(safe-area-inset-top);
  box-sizing: content-box;
  z-index: 1000;
  border-bottom: none; /* 🌟 消除边框，融为一体 */
  box-shadow: 0 2px 8px rgba(37, 99, 235, 0.15); /* 🌟 添加轻微的蓝色投影，增加高级感 */
}
.back-btn {
  background: none;
  border: none;
  padding: 8px 12px 8px 0;
  margin-left: 0;
  cursor: pointer;
  color: #ffffff; /* 🌟 返回按钮改为白色 */
  display: flex;
  align-items: center;
  transition: opacity 0.2s;
}
.back-btn:active {
  opacity: 0.7;
}
.title {
  font-size: 1.05rem;
  font-weight: 700;
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  margin: 0;
  color: #ffffff; /* 🌟 标题文字改为白色 */
}
.right-slot {
  min-width: 44px;
  display: flex;
  justify-content: flex-end;
  color: #ffffff; /* 🌟 右侧文字或图标默认为白色 */
}
</style>

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
  left: 0;
  right: 0;
  height: calc(56px + env(safe-area-inset-top)) !important;
  background: var(--primary); /* 🌟 改为定义的蓝色方案 */
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  padding-top: env(safe-area-inset-top) !important;
  box-sizing: border-box !important;
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
  z-index: 1001; /* Ensure button stays clickable and above absolute title */
}
.back-btn:active {
  opacity: 0.7;
}
.title {
  font-size: 1.02rem;
  font-weight: 700;
  position: absolute;
  left: 50%;
  bottom: 0;
  transform: translateX(-50%);
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0;
  color: #ffffff; /* 🌟 标题文字改为白色 */
  max-width: 60%; /* 🌟 限制宽度，防止溢出或重叠按钮 */
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  text-align: center;
  z-index: 1000;
}
.right-slot {
  min-width: 44px;
  display: flex;
  justify-content: flex-end;
  color: #ffffff; /* 🌟 右侧文字或图标默认为白色 */
  z-index: 1001;
}
</style>


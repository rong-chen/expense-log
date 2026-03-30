<script setup lang="ts">
import { ref } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'
import BottomNav from '@/components/layout/BottomNav.vue'
import ReloadPrompt from '@/components/ReloadPrompt.vue'
import { Toaster } from 'vue-sonner'

const route = useRoute()
const router = useRouter()
const isRouterReady = ref(false)

// 移动端检测：禁止 PC 端使用
const isMobile = ref(true)
function checkMobile() {
  const ua = navigator.userAgent
  const mobileKeywords = /Android|iPhone|iPad|iPod|webOS|BlackBerry|IEMobile|Opera Mini/i
  const hasTouchScreen = navigator.maxTouchPoints > 0
  isMobile.value = mobileKeywords.test(ua) || hasTouchScreen
}
checkMobile()

router.isReady().then(() => {
  isRouterReady.value = true
})
</script>

<template>
  <div id="app-root">
    <!-- PC 端拦截屏 -->
    <div v-if="!isMobile" class="pc-block">
      <div class="pc-block-card">
        <div class="pc-block-icon">
          <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="2" width="14" height="20" rx="2" ry="2"/><line x1="12" y1="18" x2="12.01" y2="18"/></svg>
        </div>
        <h2>请使用手机访问</h2>
        <p>本应用仅支持手机端使用，请用手机浏览器扫码或直接访问。</p>
      </div>
    </div>

    <!-- 正常应用内容 -->
    <template v-else>
      <main class="main-content" :class="{ 'has-bottom-nav': isRouterReady && !route.meta.hideBottomNav && route.name !== 'Login' && route.name !== 'Register' }">
        <router-view />
      </main>
      
      <BottomNav v-if="isRouterReady && !route.meta.hideBottomNav && route.name !== 'Login' && route.name !== 'Register'" />

      <Toaster position="bottom-center" />
      
      <ReloadPrompt />
    </template>
  </div>
</template>

<style>
/* CSS 重置与移动端适配基础样式由 index.css 提供 */
#app-root {
  height: 100vh !important;
  height: 100% !important;
  width: 100% !important;
  background-color: var(--bg-base);
  display: flex;
  flex-direction: column;
  /* 强力截断：如果在 100vw 之外的任何元素全部斩底裁掉！ */
  overflow: hidden !important;
  overflow-x: hidden !important;
  overflow-y: hidden !important;
  position: relative;
  /* 防御级遮盖：补偿所有缩放舍入导致的 1px 细线缝隙 */
  margin-left: -1px;
  margin-right: -1px;
  padding-left: 1px;
  padding-right: 1px;
}

.main-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  -webkit-overflow-scrolling: touch; /* iOS 平滑滚动支持 */
}

/* 如果有 BottomNav，则留出足够的底部空间，确保最下方卡片能滚动到可视区中部 */
.main-content.has-bottom-nav {
  padding-bottom: calc(96px + env(safe-area-inset-bottom));
}

/* 全屏页面切入/返回过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* PWA 引导横幅样式 */
.pwa-prompt {
  background: var(--primary);
  color: white;
  padding: 12px 16px;
  padding-top: calc(12px + env(safe-area-inset-top)); /* 适配刘海屏 */
  display: flex;
  align-items: center;
  gap: 12px;
  z-index: 1000;
  box-shadow: 0 4px 12px rgba(26, 188, 156, 0.2);
  flex-shrink: 0;
}
.pwa-prompt-icon {
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.2);
  width: 36px; height: 36px;
  display: flex; justify-content: center; align-items: center;
  border-radius: 50%;
}
.pwa-prompt-content {
  flex: 1;
  font-size: 0.75rem;
  line-height: 1.4;
}
.pwa-prompt-content strong {
  font-size: 0.9rem;
  font-weight: 700;
}
.pwa-prompt-close {
  flex-shrink: 0;
  padding: 6px;
  cursor: pointer;
  background: rgba(255, 255, 255, 0.15);
  border-radius: 50%;
  display: flex;
  align-items: center; justify-content: center;
  transition: all 0.2s;
}
.pwa-prompt-close:active {
  background: rgba(255, 255, 255, 0.3);
  transform: scale(0.9);
}

/* PC 端拦截屏 */
.pc-block {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4e9f0 100%);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 9999;
}
.pc-block-card {
  text-align: center;
  max-width: 400px;
  padding: 48px 40px;
  background: white;
  border-radius: 24px;
  box-shadow: 0 12px 40px rgba(0,0,0,0.08);
}
.pc-block-icon {
  color: var(--primary);
  margin-bottom: 24px;
}
.pc-block-card h2 {
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 12px 0;
}
.pc-block-card p {
  font-size: 0.95rem;
  color: var(--text-secondary);
  margin: 0;
  line-height: 1.6;
}
</style>

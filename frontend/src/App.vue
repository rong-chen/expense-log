<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'
import BottomNav from '@/components/layout/BottomNav.vue'
import ReloadPrompt from '@/components/ReloadPrompt.vue'
import { Toaster } from 'vue-sonner'
import { X, Smartphone, Monitor } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const isRouterReady = ref(false)

router.isReady().then(() => {
  isRouterReady.value = true
})

const showPwaPrompt = ref(false)
const pwaPromptMode = ref<'mobile' | 'pc' | null>(null)

onMounted(() => {
  if (sessionStorage.getItem('pwa_prompt_dismissed')) return

  const isStandalone = window.matchMedia('(display-mode: standalone)').matches || 
                       (navigator as any).standalone || 
                       document.referrer.includes('android-app://')

  if (isStandalone) return // 是原生 PWA，无需提示

  const ua = navigator.userAgent
  const isMobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(ua)
  
  if (isMobile) {
    pwaPromptMode.value = 'mobile'
    showPwaPrompt.value = true
  } else {
    pwaPromptMode.value = 'pc'
    showPwaPrompt.value = true
  }
})

function dismissPrompt() {
  showPwaPrompt.value = false
  sessionStorage.setItem('pwa_prompt_dismissed', '1')
}
</script>

<template>
  <div id="app-root">
    


    <main class="main-content" :class="{ 'has-bottom-nav': isRouterReady && !route.meta.hideBottomNav && route.name !== 'Login' && route.name !== 'Register' }">
      <router-view />
    </main>
    
    <!-- 移动端底部导航 (路由就绪后 + 非隐藏页面才显示) -->
    <BottomNav v-if="isRouterReady && !route.meta.hideBottomNav && route.name !== 'Login' && route.name !== 'Register'" />

    <!-- 全局轻提示组件 -->
    <Toaster position="top-center" richColors />
    
    <!-- PWA 更新提示弹窗 -->
    <ReloadPrompt />
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
</style>

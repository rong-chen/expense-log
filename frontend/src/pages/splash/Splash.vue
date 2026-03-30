<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { billApi } from '@/api'
import { Loader2 } from 'lucide-vue-next'

const router = useRouter()
const auth = useAuthStore()

onMounted(async () => {
  // Simulate min loading time for smooth UX
  const minWait = new Promise(resolve => setTimeout(resolve, 1500))
  
  try {
    if (!auth.isLoggedIn) {
      await minWait
      router.replace('/login')
      return
    }

    // Parallel fetch user info & preheat cache (seamless display UX per user request)
    await Promise.allSettled([
      auth.fetchUserInfo(),
      billApi.getDashboardStats(),
      minWait
    ])
    
    // Once data is successfully warmed up, transition to Home
    router.replace('/home')
  } catch (err) {
    router.replace('/login')
  }
})
</script>

<template>
  <div class="splash-screen">
    <!-- 炫酷的环境光背景 (与 Auth 保持连贯) -->
    <div class="ambient-bg">
      <div class="ambient-orb orb-primary"></div>
      <div class="ambient-orb orb-secondary"></div>
    </div>

    <div class="logo-box animate-pulse-logo">
      <img src="/icon-192.png" alt="易账 Logo" class="app-icon-img shadow-xl" />
      <h2 class="brand-title">易账</h2>
    </div>
    
    <div class="loading-indicator">
      <Loader2 class="spinner" :size="28" stroke-width="2.5" />
    </div>
  </div>
</template>

<style scoped>
.splash-screen {
  position: relative;
  height: 100vh;
  height: 100%;
  width: 100vw;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background-color: var(--bg-base);
  overflow: hidden;
  /* 视觉重心上移，让 Logo 看起来在"黄金分割点"而不是数学正中 */
  padding-bottom: 16vh;
}

/* 沉浸式环境光效 */
.ambient-bg {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
  z-index: 0;
}
.ambient-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(90px);
  opacity: 0.35;
}
.orb-primary {
  width: 120vw; height: 120vw;
  background: var(--primary);
  bottom: -40vw; left: -20vw;
  animation: pulse-slow 10s infinite alternate;
}
.orb-secondary {
  width: 100vw; height: 100vw;
  background: var(--success);
  bottom: -30vw; right: -40vw;
  opacity: 0.25;
  animation: pulse-slow 12s infinite alternate-reverse;
}
@keyframes pulse-slow {
  0% { transform: scale(1) translate(0, 0); }
  100% { transform: scale(1.1) translate(5%, 5%); }
}

.logo-box {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 40px;
}
.animate-pulse-logo {
  animation: logo-entrance 1.2s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
@keyframes logo-entrance {
  0% { opacity: 0; transform: scale(0.85) translateY(20px); }
  100% { opacity: 1; transform: scale(1) translateY(0); }
}

.app-icon-img {
  width: 96px;
  height: 96px;
  border-radius: 24px;
  box-shadow: 0 16px 32px rgba(230, 126, 34, 0.3);
  margin-bottom: 24px;
  object-fit: cover;
}
.brand-title {
  font-size: 2rem;
  font-weight: 800;
  color: var(--text-primary);
  margin: 0;
  letter-spacing: -0.5px;
}
.loading-indicator {
  position: absolute;
  top: calc(env(safe-area-inset-top) + 24px);
  right: 24px;
  z-index: 1;
  display: flex;
  justify-content: center;
  opacity: 0;
  animation: fade-in 1s ease 0.6s forwards;
}
@keyframes fade-in {
  to { opacity: 1; }
}
.spinner {
  animation: spin 1.2s cubic-bezier(0.4, 0, 0.2, 1) infinite;
  color: var(--primary);
}
@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>

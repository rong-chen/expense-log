<script setup lang="ts">
import { useRegisterSW } from 'virtual:pwa-register/vue'
import { RefreshCw, X } from 'lucide-vue-next'

const {
  offlineReady,
  needRefresh,
  updateServiceWorker,
} = useRegisterSW({
  onRegistered(r) {
    if (r) {
      console.log('SW Registered')
    }
  },
  onRegisterError(error) {
    console.error('SW registration error', error)
  }
})

function close() {
  offlineReady.value = false
  needRefresh.value = false
}
</script>

<template>
  <div v-if="offlineReady || needRefresh" class="pwa-toast-wrapper">
    <div class="pwa-toast">
      <div class="icon-box">
        <RefreshCw :size="18" :class="{ 'spin-icon': needRefresh }" />
      </div>
      <div class="message">
        <span v-if="offlineReady">
          应用已缓存就绪，支持离线访问
        </span>
        <span v-else>
          ✨ 发现新版本！点击立即更新
        </span>
      </div>
      <div class="actions">
        <button v-if="needRefresh" @click="updateServiceWorker()" class="btn-refresh">
          更新
        </button>
        <button @click="close" class="btn-close">
          <X :size="16" />
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.pwa-toast-wrapper {
  position: fixed;
  bottom: calc(100px + max(env(safe-area-inset-bottom), 16px));
  left: 50%;
  transform: translateX(-50%);
  z-index: 9999;
  display: flex;
  justify-content: center;
  pointer-events: none;
  animation: slideUp 0.4s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  width: 90%;
  max-width: 360px;
}

@keyframes slideUp {
  0% { transform: translate(-50%, 20px); opacity: 0; }
  100% { transform: translate(-50%, 0); opacity: 1; }
}

.pwa-toast {
  background: white;
  border-radius: 16px;
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.12), 0 2px 8px rgba(0, 0, 0, 0.04);
  pointer-events: auto;
  border: 1px solid rgba(0,0,0,0.05);
  width: 100%;
}

.icon-box {
  width: 32px;
  height: 32px;
  background: rgba(230, 126, 34, 0.1);
  color: var(--primary);
  border-radius: 50%;
  display: flex;
  justify-content: center;
  align-items: center;
  flex-shrink: 0;
}

.spin-icon {
  animation: spin 3s linear infinite;
}
@keyframes spin {
  100% { transform: rotate(360deg); }
}

.message {
  flex: 1;
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 600;
  line-height: 1.4;
}

.actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-refresh {
  background: var(--primary);
  color: white;
  border: none;
  padding: 6px 14px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
}
.btn-refresh:active {
  opacity: 0.8;
}

.btn-close {
  background: #f1f5f9;
  color: var(--text-secondary);
  border: none;
  padding: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  cursor: pointer;
  transition: background-color 0.2s;
}
.btn-close:active {
  background-color: #e2e8f0;
}
</style>

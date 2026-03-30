<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Key, Plus, Trash2 } from 'lucide-vue-next'
import TopNavBar from '@/components/layout/TopNavBar.vue'
import request from '@/api'
import { toast } from 'vue-sonner'

const ukeys = ref<any[]>([])
const loading = ref(true)
const creating = ref(false)
const deletingId = ref('')
const showInstallModal = ref(false)

// iOS 快捷指令安装链接
const SHORTCUT_LINK = 'https://www.icloud.com/shortcuts/65d3aec377784390a392b14f042b8ae9'


async function fetchUkeys() {
  loading.value = true
  try {
    const res: any = await request.get('/user/ukey')
    if (res.code === 0) {
      ukeys.value = res.data || []
    }
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

async function createUkey() {
  try {
    creating.value = true
    // 默认名称
    const res: any = await request.post('/user/ukey', { name: 'iOS 快捷指令' })
    if (res.code === 0) {
      fetchUkeys()
    } else {
      toast.error(res.msg || '创建失败')
    }
  } catch (error: any) {
    toast.error(error.msg || '网络或未知错误')
  } finally {
    creating.value = false
  }
}

async function deleteUkey(id: string) {
  if (!confirm('确定要删除此凭证吗？关联的自动化指令将立即失效！')) return
  try {
    deletingId.value = id
    const res: any = await request.delete(`/user/ukey/${id}`)
    if (res.code === 0) {

      fetchUkeys()
    }
  } catch (error: any) {
    toast.error(error.msg || '删除失败')
  } finally {
    deletingId.value = ''
  }
}

import * as clipboard from 'clipboard-polyfill'

function copyToClipboard(text: string) {
  if (!text) return

  clipboard.writeText(text).then(() => {
  }).catch((err) => {
    console.warn("Clipboard polyfill failed:", err)
    toast.error('复制失败，请长按下方配置文本手动复制')
  })
}

// 复制具体 JSON 配置并安装快捷指令
function copyAndInstall(host: string, ukey: string) {
  const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent) || (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1)
  if (!isIOS) {
    alert('此功能只适用于 iPhone 或 iPad 的 Safari 浏览器。\n如果当前是在电脑端，请使用上方「复制配置」并在手机上保存。')
    return
  }
  const jsonStr = JSON.stringify({ host, ukey }, null, 2)
  copyToClipboard(jsonStr)
  
  // 弹出自定义弹窗
  showInstallModal.value = true
}

function goToInstall() {
  showInstallModal.value = false
  window.location.href = SHORTCUT_LINK
}

onMounted(() => {
  fetchUkeys()
})
</script>

<template>
  <div class="ukey-page">
    <TopNavBar title="系统设置" />
    
    <div class="page-content">
      <div class="info-card">
        <h3>环境部署</h3>
        <p>用于外部生态部署（如苹果 iOS 快捷指令）对接后端的专属凭证。<strong>目前仅支持建立 1 个部署环境连接</strong>。可以随时查看、复制与删除凭证。</p>
      </div>

      <div class="key-list">
        <div class="list-header">
          <h2>环境凭证</h2>
          <button class="btn-create" @click="createUkey" :disabled="ukeys.length > 0 || creating" :class="{ disabled: ukeys.length > 0 || creating }">
            <Plus :size="18" v-if="!creating" /> {{ creating ? '创建中...' : '创建部署环境' }}
          </button>
        </div>

        <div v-if="loading" class="loading">加载中...</div>
        <div v-else-if="ukeys.length === 0" class="empty-state">
          <Key :size="48" style="opacity: 0.2; margin-bottom: 12px;" />
          <p>暂无任何环境部署</p>
        </div>
        
        <div v-else class="key-card" v-for="k in ukeys" :key="k.id">
          <div class="card-header">
            <div class="key-name">{{ k.name }}</div>
            <button class="btn-delete" @click="deleteUkey(k.id)" :disabled="deletingId === k.id">
              <Trash2 :size="18" v-if="deletingId !== k.id" />
              <span v-else style="font-size: 0.8rem;">...</span>
            </button>
          </div>

          <div style="font-size: 0.85rem; color: var(--text-secondary); margin-bottom: 6px;">请求地址 (URL)</div>
          <div class="secret-box" style="margin-bottom: 16px;">
            <code>{{ k.host }}</code>
          </div>

          <div style="font-size: 0.85rem; color: var(--text-secondary); margin-bottom: 6px;">请求头 (Authorization)</div>
          <div class="secret-box" style="margin-bottom: 24px;">
            <code>{{ k.full_token }}</code>
          </div>

          <div style="margin-top: 20px;">
            <button class="btn-primary" @click="copyAndInstall(k.host, k.full_token)" style="width: 100%; display: flex; align-items: center; justify-content: center; gap: 6px;">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
              复制配置并添加快捷指令
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 自定义弹窗 -->
    <div class="modal-overlay" v-if="showInstallModal">
      <div class="modal-content">
        <div class="modal-body">
          复制成功，请将复制的内容粘贴到对应的输入框中并添加快捷指令
        </div>
        <div class="modal-actions">
          <button class="btn-cancel" @click="showInstallModal = false">取消</button>
          <button class="btn-confirm" @click="goToInstall">立即前往</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ukey-page {
  min-height: 100vh;
  background-color: var(--bg-base);
}

.page-content {
  padding: 16px;
  padding-top: calc(env(safe-area-inset-top) + 70px);
  padding-bottom: calc(env(safe-area-inset-bottom) + 20px);
}

.info-card {
  background: white;
  padding: 20px;
  border-radius: 16px;
  margin-bottom: 24px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.03);
}
.info-card h3 {
  margin: 0 0 8px 0;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--text-primary);
}
.info-card p {
  margin: 0;
  font-size: 0.9rem;
  color: var(--text-secondary);
  line-height: 1.6;
}

.secret-alert {
  background: white;
  border: 1px solid rgba(26, 188, 156, 0.4);
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 8px 24px rgba(26, 188, 156, 0.1);
}
.alert-header {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--primary);
  font-weight: 700;
  font-size: 1.1rem;
  margin-bottom: 12px;
}
.warning-text {
  font-size: 0.9rem;
  color: #e67e22;
  margin-bottom: 20px;
  font-weight: 500;
}
.secret-box {
  display: flex;
  background: var(--bg-input);
  border-radius: 8px;
  padding: 12px;
  align-items: center;
  margin-bottom: 24px;
}
.secret-box code {
  flex: 1;
  font-family: monospace;
  font-size: 0.95rem;
  color: var(--text-primary);
  word-break: break-all;
}
.json-textarea {
  width: 100%;
  height: 80px;
  background: transparent;
  border: none;
  font-family: monospace;
  font-size: 0.95rem;
  color: var(--text-primary);
  resize: none;
  outline: none;
}
.copy-btn {
  background: none;
  border: none;
  color: var(--primary);
  padding: 8px;
  cursor: pointer;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease-out;
}
.modal-content {
  background: white;
  width: 85%;
  max-width: 320px;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 10px 30px rgba(0,0,0,0.1);
  display: flex;
  flex-direction: column;
}
.modal-body {
  padding: 30px 20px;
  font-size: 1rem;
  color: var(--text-primary);
  text-align: center;
  line-height: 1.5;
}
.modal-actions {
  display: flex;
  border-top: 1px solid rgba(0,0,0,0.05);
}
.modal-actions button {
  flex: 1;
  padding: 16px 0;
  border: none;
  background: transparent;
  font-size: 1.05rem;
  cursor: pointer;
}
.btn-cancel {
  border-right: 1px solid rgba(0,0,0,0.05) !important;
  color: var(--text-secondary);
}
.btn-confirm {
  color: var(--primary);
  font-weight: 600;
}
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.btn-primary {
  width: 100%;
  background: var(--primary);
  color: white;
  border: none;
  padding: 14px;
  border-radius: 12px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
}
.btn-install {
  display: block;
  width: 100%;
  text-align: center;
  padding: 14px;
  border-radius: 12px;
  font-size: 1rem;
  font-weight: 600;
  color: var(--primary);
  background: var(--primary-soft);
  border: 1.5px solid var(--primary);
  text-decoration: none;
  cursor: pointer;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.list-header h2 {
  font-size: 1.1rem;
  color: var(--text-primary);
  margin: 0;
}
.btn-create {
  display: flex;
  align-items: center;
  gap: 4px;
  background: var(--primary);
  color: white;
  border: none;
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
}
.btn-create.disabled {
  background: var(--text-muted);
  cursor: not-allowed;
}

.loading, .empty-state {
  text-align: center;
  padding: 40px 0;
  color: var(--text-muted);
}
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.key-card {
  background: white;
  padding: 24px;
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.key-name {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--text-primary);
}
.secret-box {
  display: flex;
  background: var(--bg-input);
  border-radius: 8px;
  padding: 12px;
  align-items: center;
}
.secret-box code {
  flex: 1;
  font-family: monospace;
  font-size: 0.95rem;
  color: var(--text-primary);
  word-break: break-all;
}
.btn-delete {
  background: rgba(231, 76, 60, 0.1);
  color: var(--danger);
  border: none;
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
}
.btn-shortcut {
  background: linear-gradient(135deg, #007aff, #5856d6);
  border: none;
  color: white;
}
.btn-shortcut:active {
  opacity: 0.8;
}
</style>

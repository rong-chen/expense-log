<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Key, Plus, Trash2 } from 'lucide-vue-next'
import TopNavBar from '@/components/layout/TopNavBar.vue'
import request from '@/api'
import { toast } from 'vue-sonner'

const router = useRouter()
const ukeys = ref<any[]>([])
const loading = ref(true)

// 获取后的快捷指令配置参数
const hostUrl = ref('')
const ukeyFull = ref('')

// iOS 快捷指令安装链接
const SHORTCUT_LINK = 'https://www.icloud.com/shortcuts/4b81141048f94fe18f414354722de1f1'

const configJson = computed(() => {
  if (!hostUrl.value || !ukeyFull.value) return ''
  return JSON.stringify({
    host: hostUrl.value,
    ukey: ukeyFull.value
  }, null, 2)
})


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
    // 默认名称
    const res: any = await request.post('/user/ukey', { name: 'iOS 快捷指令' })
    if (res.code === 0) {
      hostUrl.value = res.data.host
      ukeyFull.value = res.data.ukey
      toast.success('凭证创建成功！')
      fetchUkeys()
    } else {
      toast.error(res.msg || '创建失败')
    }
  } catch (error: any) {
    toast.error(error.msg || '网络或未知错误')
  }
}

async function deleteUkey(id: string) {
  if (!confirm('确定要删除此凭证吗？关联的自动化指令将立即失效！')) return
  try {
    const res: any = await request.delete(`/user/ukey/${id}`)
    if (res.code === 0) {
      toast.success('删除成功')
      fetchUkeys()
    }
  } catch (error: any) {
    toast.error(error.msg || '删除失败')
  }
}

import * as clipboard from 'clipboard-polyfill'

function copyToClipboard(text: string) {
  if (!text) return

  clipboard.writeText(text).then(() => {
    toast.success('已复制到剪贴板！')
  }).catch((err) => {
    console.warn("Clipboard polyfill failed:", err)
    toast.error('复制失败，请长按下方配置文本手动复制')
  })
}

// 仅复制 JSON 配置
function copyConfig() {
  copyToClipboard(configJson.value)
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
        <p>用于外部生态部署（如苹果 iOS 快捷指令）对接后端的专属凭证。<strong>目前仅支持建立 1 个部署环境连接</strong>。为了安全，凭证仅在生成时显示一次明文。</p>
      </div>

      <div v-if="ukeyFull" class="secret-alert">
        <div class="alert-header">
          <Key class="icon" :size="20"/>
          <span>您的环境部署密钥已签发 (仅显示一次)</span>
        </div>
        
        <div style="font-size: 0.85rem; color: var(--text-secondary); margin-bottom: 6px;">请求地址 (URL)</div>
        <div class="secret-box" style="margin-bottom: 16px;">
          <code>{{ hostUrl }}</code>
        </div>

        <div style="font-size: 0.85rem; color: var(--text-secondary); margin-bottom: 6px;">请求头 (Authorization)</div>
        <div class="secret-box" style="margin-bottom: 24px;">
          <code>{{ ukeyFull }}</code>
        </div>

        <div style="display: flex; flex-direction: column; gap: 12px;">
          <button class="btn-primary" @click="copyConfig">复制配置到剪贴板</button>
          <button @click="ukeyFull = ''" style="padding: 14px 20px; border-radius: 12px; background: rgba(0,0,0,0.05); border: none; font-size: 1rem; font-weight: 600; color: var(--text-secondary); cursor: pointer;">关闭</button>
        </div>
      </div>

      <div class="key-list" v-else>
        <div class="list-header">
          <h2>环境凭证</h2>
          <button class="btn-create" @click="createUkey" :disabled="ukeys.length > 0" :class="{ disabled: ukeys.length > 0 }">
            <Plus :size="18" /> 创建部署环境
          </button>
        </div>

        <div v-if="loading" class="loading">加载中...</div>
        <div v-else-if="ukeys.length === 0" class="empty-state">
          <Key :size="48" style="opacity: 0.2; margin-bottom: 12px;" />
          <p>暂无任何环境部署</p>
        </div>
        
        <div v-else class="key-card" v-for="k in ukeys" :key="k.id">
          <div class="key-info">
            <div class="key-name">{{ k.name }}</div>
            <div class="key-value">{{ k.secret_key }}</div>
            <div class="key-date">创建于 {{ new Date(k.created_at).toLocaleString() }}</div>
          </div>
          <button class="btn-delete" @click="deleteUkey(k.id)">
            <Trash2 :size="18" />
          </button>
        </div>
      </div>

      <!-- 常驻：导入快捷指令入口 -->
      <a :href="SHORTCUT_LINK" class="btn-install" target="_blank" style="margin-top: 24px;">导入 iOS 快捷指令</a>
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
  padding: 20px;
  border-radius: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);
}
.key-info {
  flex: 1;
}
.key-name {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 6px;
}
.key-value {
  font-family: monospace;
  font-size: 0.9rem;
  color: var(--primary);
  opacity: 0.8;
  margin-bottom: 8px;
  word-break: break-all;
}
.key-date {
  font-size: 0.8rem;
  color: var(--text-muted);
}
.btn-delete {
  background: rgba(231, 76, 60, 0.1);
  color: var(--danger);
  border: none;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
}
</style>

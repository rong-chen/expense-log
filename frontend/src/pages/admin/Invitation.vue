<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus, Copy, Check, RotateCcw } from 'lucide-vue-next'
import TopNavBar from '@/components/layout/TopNavBar.vue'
import request from '@/api'
import { toast } from 'vue-sonner'

interface InvitationCode {
  ID: string
  code: string
  role: string
  is_used: boolean
  used_by: string
  CreatedAt: string
}

const loading = ref(true)
const list = ref<InvitationCode[]>([])
const generating = ref(false)
const generateCount = ref(1)
const targetRole = ref('user')

async function fetchList() {
  loading.value = true
  try {
    const res: any = await request.get('/invitation/list')
    if (res.code === 0) list.value = res.data || []
  } catch (err) {
    toast.error('获取邀请码列表失败')
  } finally {
    loading.value = false
  }
}

async function handleGenerate() {
  if (generateCount.value < 1 || generateCount.value > 100) {
    toast.error('生成数量需在 1-100 之间')
    return
  }

  generating.value = true
  try {
    const res: any = await request.post('/invitation/generate', { 
      count: generateCount.value,
      role: targetRole.value
    })
    if (res.code === 0) {
      toast.success(`成功生成 ${res.data.codes.length} 个邀请码`)
      fetchList()
    }
  } catch (err) {
    toast.error('请求失败')
  } finally {
    generating.value = false
  }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text).then(() => toast.success('已复制')).catch(() => toast.error('复制失败'))
}

onMounted(fetchList)
</script>

<template>
  <div class="invitation-page">
    <TopNavBar title="邀请码管理" back-to="/admin" />

    <div class="content">
      <div class="section-header">
        <h3>生成新邀请码</h3>
        <p>新用户注册必须持有一个有效的邀请码。生成的码可以是一次性的。</p>
      </div>

      <!-- 生成区域 -->
      <div class="card generate-card">
        <div class="input-group">
          <input type="number" v-model.number="generateCount" min="1" max="100" class="form-control count-input" />
          <select v-model="targetRole" class="form-control role-select">
            <option value="user">普通用户</option>
            <option value="admin">管理员</option>
          </select>
          <button class="btn btn-primary" @click="handleGenerate" :disabled="generating">
            <Plus :size="18" v-if="!generating" />
            <RotateCcw :size="18" class="animate-spin" v-else />
            {{ generating ? '生成' : '立即生成' }}
          </button>
        </div>
      </div>

      <!-- 列表区域 -->
      <div class="card list-card mt-24">
        <div class="list-header">
          <h4>已生成的邀请码 ({{ list.length }})</h4>
          <button class="refresh-btn" @click="fetchList" :disabled="loading">
            <RotateCcw :size="16" :class="{ 'animate-spin': loading }" />
          </button>
        </div>

        <div v-if="list.length === 0" class="empty-state">暂无邀请码</div>
        <div v-else class="invitation-list">
          <div v-for="item in list" :key="item.ID" class="invitation-item" :class="{ used: item.is_used }">
            <div class="item-main">
              <div class="item-code">{{ item.code }}</div>
              <div class="item-meta">
                <span class="badge" :class="item.is_used ? 'badge-gray' : 'badge-green'">{{ item.is_used ? '已使用' : '未使用' }}</span>
                <span class="badge" :class="item.role === 'admin' ? 'badge-orange' : 'badge-blue'">{{ item.role === 'admin' ? '管理员' : '普通' }}</span>
                <span class="time">{{ new Date(item.CreatedAt).toLocaleDateString() }}</span>
              </div>
            </div>
            
            <div class="item-actions">
              <button v-if="!item.is_used" class="icon-btn" @click="copyToClipboard(item.code)">
                <Copy :size="18" />
              </button>
              <Check v-else :size="18" class="text-success" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.invitation-page {
  min-height: 100vh;
  background: var(--bg-body);
  padding: 16px;
  padding-top: calc(72px + env(safe-area-inset-top));
}

.content {
  max-width: 600px;
  margin: 0 auto;
}

.section-header {
  margin-bottom: 20px;
}
.section-header h3 {
  font-size: 1.1rem;
  font-weight: 800;
  margin: 0 0 6px 0;
  color: var(--text-primary);
}
.section-header p {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin: 0;
}

.card {
  background: white;
  border-radius: 20px;
  padding: 20px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);
}

.mt-24 { margin-top: 24px; }

.input-group {
  display: flex;
  gap: 12px;
}
.count-input { flex: 0 0 70px; text-align: center; }
.role-select { flex: 1; }
.form-control {
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 10px;
  font-size: 0.95rem;
  outline: none;
}
.btn-primary {
  background: var(--primary);
  color: white;
  border: none;
  border-radius: 12px;
  padding: 10px 16px;
  font-weight: 700;
  display: flex; align-items: center; gap: 8px;
}

.list-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 16px;
}
.list-header h4 { margin: 0; font-size: 1rem; }

.invitation-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.invitation-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px;
  background: #f8fafc;
  border-radius: 16px;
}
.invitation-item.used { opacity: 0.6; }

.item-main { display: flex; flex-direction: column; gap: 4px; }
.item-code { font-family: monospace; font-weight: 800; font-size: 1.1rem; letter-spacing: 1px; }

.item-meta { display: flex; align-items: center; gap: 8px; }
.badge { font-size: 0.65rem; font-weight: 700; padding: 2px 6px; border-radius: 6px; }
.badge-green { background: #dcfce7; color: #166534; }
.badge-gray { background: #f1f5f9; color: #64748b; }
.badge-orange { background: #ffedd5; color: #9a3412; }
.badge-blue { background: #dbeafe; color: #1e40af; }
.time { font-size: 0.75rem; color: #94a3b8; }

.icon-btn { background: white; border: 1px solid #e2e8f0; padding: 8px; border-radius: 10px; cursor: pointer; display: flex; }
.text-success { color: #10b981; }
.refresh-btn { background: none; border: none; color: #94a3b8; }
.animate-spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.empty-state { text-align: center; padding: 40px; color: #94a3b8; }
</style>

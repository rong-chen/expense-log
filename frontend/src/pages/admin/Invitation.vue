<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus, Users, Copy, Check, RotateCcw } from 'lucide-vue-next'
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
    if (res.code === 0) {
      list.value = res.data || []
    }
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
    } else {
      toast.error(res.msg || '生成失败')
    }
  } catch (err) {
    toast.error('请求失败')
  } finally {
    generating.value = false
  }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text).then(() => {
    toast.success('邀请码已复制')
  }).catch(() => {
    toast.error('复制失败')
  })
}

onMounted(() => {
  fetchList()
})
</script>

<template>
  <div class="admin-page">
    <TopNavBar title="管理后台 - 邀请码" />

    <div class="content">
      <div class="section-header">
        <h3>邀请码管理</h3>
        <p>生成并管理注册所需的邀请码。只有持有邀请码的用户才能注册本系统。</p>
      </div>

      <!-- 生成区域 -->
      <div class="card generate-card">
        <div class="form-group">
          <label>批量生成数量</label>
          <div class="input-group">
            <input 
              type="number" 
              v-model.number="generateCount" 
              min="1" 
              max="100" 
              class="form-control"
              placeholder="数量"
            />
            <select v-model="targetRole" class="form-control role-select">
              <option value="user">普通用户</option>
              <option value="admin">管理员</option>
            </select>
            <button 
              class="btn btn-primary" 
              @click="handleGenerate" 
              :disabled="generating"
            >
              <Plus :size="18" v-if="!generating" />
              <RotateCcw :size="18" class="animate-spin" v-else />
              {{ generating ? '生成中...' : '生成邀请码' }}
            </button>
          </div>
        </div>
      </div>

      <!-- 列表区域 -->
      <div class="card list-card">
        <div class="list-header">
          <h2>已生成的邀请码</h2>
          <button class="refresh-btn" @click="fetchList" :disabled="loading">
            <RotateCcw :size="16" :class="{ 'animate-spin': loading }" />
          </button>
        </div>

        <div v-if="loading && list.length === 0" class="loading-state">
          加载中...
        </div>

        <div v-else-if="list.length === 0" class="empty-state">
          <p>暂无任何邀请码</p>
        </div>

        <div v-else class="invitation-list">
          <div v-for="item in list" :key="item.ID" class="invitation-item" :class="{ used: item.is_used }">
            <div class="item-main">
              <div class="item-code">{{ item.code }}</div>
              <div class="item-meta">
                <span class="status-badge" :class="item.is_used ? 'badge-used' : 'badge-unused'">
                  {{ item.is_used ? '已使用' : '未使用' }}
                </span>
                <span class="role-badge" :class="item.role === 'admin' ? 'badge-admin' : 'badge-user'">
                  {{ item.role === 'admin' ? '管理员' : '普通用户' }}
                </span>
                <span class="time">{{ new Date(item.CreatedAt).toLocaleDateString() }}</span>
              </div>
              <div v-if="item.is_used" class="item-user">
                <Users :size="12" />
                <span>使用者 ID: {{ item.used_by.substring(0, 8) }}...</span>
              </div>
            </div>
            
            <div class="item-actions" v-if="!item.is_used">
              <button class="icon-btn" @click="copyToClipboard(item.code)">
                <Copy :size="18" />
              </button>
            </div>
            <div class="item-actions" v-else>
              <Check :size="18" class="text-success" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.admin-page {
  padding: 16px;
  padding-top: calc(72px + env(safe-area-inset-top));
  max-width: 600px;
  margin: 0 auto;
  min-height: 100vh;
  background: var(--bg-body);
}

.content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-header {
  margin-bottom: 8px;
}
.section-header h3 {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 6px 0;
}
.section-header p {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin: 0;
  line-height: 1.5;
}

.card {
  background: white;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);
}

.generate-card .form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.generate-card label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-secondary);
}

.input-group {
  display: flex;
  gap: 12px;
}

.form-control {
  flex: 0 0 80px;
  border: 1px solid rgba(0,0,0,0.08);
  border-radius: 12px;
  padding: 10px 14px;
  font-size: 1rem;
  background: #fafafa;
  outline: none;
}

.role-select {
  flex: 0 0 110px;
}

.btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px;
  border-radius: 12px;
  border: none;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: var(--primary);
  color: white;
}

.btn-primary:active { transform: scale(0.98); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.list-header h2 {
  font-size: 1rem;
  font-weight: 700;
  margin: 0;
}

.refresh-btn {
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 4px;
}

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
  background: #fafafa;
  border-radius: 14px;
  transition: all 0.2s;
}

.invitation-item.used {
  opacity: 0.6;
}

.item-main {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.item-code {
  font-family: 'Monaco', 'Consolas', monospace;
  font-size: 1.1rem;
  font-weight: 800;
  letter-spacing: 1px;
  color: var(--text-primary);
}

.item-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-badge {
  font-size: 0.7rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 6px;
}

.badge-unused { background: #e3f2fd; color: #1e88e5; }
.badge-used { background: #eeeeee; color: #757575; }

.role-badge {
  font-size: 0.7rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 6px;
}
.badge-admin { background: rgba(230, 126, 34, 0.12); color: #e67e22; }
.badge-user { background: rgba(26, 188, 156, 0.12); color: #1abc9c; }

.time {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.item-user {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-top: 2px;
}

.icon-btn {
  background: white;
  border: 1px solid rgba(0,0,0,0.06);
  padding: 8px;
  border-radius: 10px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(0,0,0,0.02);
}

.icon-btn:active { transform: scale(0.9); }

.text-success { color: #10b981; }

.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.loading-state, .empty-state {
  padding: 40px;
  text-align: center;
  color: var(--text-secondary);
  font-size: 0.9rem;
}
</style>

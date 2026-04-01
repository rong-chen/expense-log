<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { 
  Plus, Users, Copy, Check, RotateCcw, 
  Activity, Database, Smartphone, 
  Mail, CalendarClock, UserPlus
} from 'lucide-vue-next'
import TopNavBar from '@/components/layout/TopNavBar.vue'
import request from '@/api'
import { toast } from 'vue-sonner'

// --- Types ---
interface InvitationCode {
  ID: string
  code: string
  role: string
  is_used: boolean
  used_by: string
  CreatedAt: string
}

interface User {
  ID: string
  phone: string
  nickname: string
  role: string
  last_login: number
  CreatedAt: string
}

interface Stats {
  total_users: number
  total_bills: number
  total_email_accounts: number
  total_recurring_tasks: number
}

// --- State ---
const activeTab = ref('invitations') // 'invitations', 'users', 'monitoring'
const loading = ref(false)

// Invitations State
const invitationList = ref<InvitationCode[]>([])
const generating = ref(false)
const generateCount = ref(1)
const targetRole = ref('user')

// Users State
const userList = ref<User[]>([])
const totalUsers = ref(0)
const userPage = ref(1)

// Stats State
const stats = ref<Stats | null>(null)

// --- Methods ---

// Invitations
async function fetchInvitations() {
  loading.value = true
  try {
    const res: any = await request.get('/invitation/list')
    if (res.code === 0) invitationList.value = res.data || []
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
      fetchInvitations()
    }
  } catch (err) {
    toast.error('生成失败')
  } finally {
    generating.value = false
  }
}

const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text).then(() => toast.success('已复制')).catch(() => toast.error('复制失败'))
}

// Users
async function fetchUsers() {
  loading.value = true
  try {
    const res: any = await request.get('/admin/users', { params: { page: userPage.value, pageSize: 20 } })
    if (res.code === 0) {
      userList.value = res.data.users || []
      totalUsers.value = res.data.total
    }
  } catch (err) {
    toast.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

async function toggleUserRole(user: User) {
  const newRole = user.role === 'admin' ? 'user' : 'admin'
  try {
    const res: any = await request.post('/admin/role', { user_id: user.ID, role: newRole })
    if (res.code === 0) {
      toast.success('角色更新成功')
      fetchUsers()
    }
  } catch (err) {
    toast.error('角色更新失败')
  }
}

// Monitoring
async function fetchStats() {
  loading.value = true
  try {
    const res: any = await request.get('/admin/stats')
    if (res.code === 0) stats.value = res.data
  } catch (err) {
    toast.error('获取统计数据失败')
  } finally {
    loading.value = false
  }
}

// --- Lifecycle ---
onMounted(() => {
  switchTab('invitations')
})

function switchTab(tab: string) {
  activeTab.value = tab
  if (tab === 'invitations') fetchInvitations()
  if (tab === 'users') fetchUsers()
  if (tab === 'monitoring') fetchStats()
}

const formatTime = (ts: number) => {
  if (!ts) return '从未登录'
  return new Date(ts * 1000).toLocaleString()
}
</script>

<template>
  <div class="admin-dashboard">
    <TopNavBar title="管理后台" />

    <!-- Tab Switching -->
    <div class="tabs-container">
      <div class="tabs">
        <button 
          :class="['tab-item', { active: activeTab === 'invitations' }]" 
          @click="switchTab('invitations')"
        >
          <UserPlus :size="18" />
          <span>邀请码</span>
        </button>
        <button 
          :class="['tab-item', { active: activeTab === 'users' }]" 
          @click="switchTab('users')"
        >
          <Users :size="18" />
          <span>用户管理</span>
        </button>
        <button 
          :class="['tab-item', { active: activeTab === 'monitoring' }]" 
          @click="switchTab('monitoring')"
        >
          <Activity :size="18" />
          <span>监控统计</span>
        </button>
      </div>
    </div>

    <div class="content">
      <!-- --- Invitations Tab --- -->
      <div v-if="activeTab === 'invitations'" class="tab-content">
        <div class="section-header">
          <h3>邀请码生成</h3>
          <p>新用户注册必须持有有效的邀请码。</p>
        </div>

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
              {{ generating ? '生成中...' : '生成' }}
            </button>
          </div>
        </div>

        <div class="card list-card mt-16">
          <div class="list-header">
            <h4>邀请码列表 ({{ invitationList.length }})</h4>
            <button @click="fetchInvitations" :disabled="loading" class="refresh-btn">
              <RotateCcw :size="14" :class="{ 'animate-spin': loading }" />
            </button>
          </div>

          <div v-if="invitationList.length === 0" class="empty-state">暂无邀请码</div>
          <div v-else class="invitation-list">
            <div v-for="item in invitationList" :key="item.ID" class="invitation-item" :class="{ used: item.is_used }">
              <div class="item-info">
                <span class="code-text">{{ item.code }}</span>
                <div class="badges">
                  <span :class="['badge', item.is_used ? 'badge-gray' : 'badge-green']">{{ item.is_used ? '已使用' : '未使用' }}</span>
                  <span :class="['badge', item.role === 'admin' ? 'badge-orange' : 'badge-blue']">{{ item.role === 'admin' ? '管理员' : '普通用户' }}</span>
                </div>
              </div>
              <button v-if="!item.is_used" @click="copyToClipboard(item.code)" class="icon-btn"><Copy :size="16" /></button>
              <Check v-else :size="16" class="text-success" />
            </div>
          </div>
        </div>
      </div>

      <!-- --- Users Management Tab --- -->
      <div v-if="activeTab === 'users'" class="tab-content animate-fade-in">
        <div class="section-header">
          <h3>注册用户列表</h3>
          <p>共 {{ totalUsers }} 位注册成员。</p>
        </div>

        <div class="card list-card">
          <div v-if="userList.length === 0" class="empty-state">暂无用户</div>
          <div v-else class="user-list">
            <div v-for="user in userList" :key="user.ID" class="user-item">
              <div class="user-main">
                <div class="user-header">
                  <span class="phone-text">{{ user.phone }}</span>
                  <span :class="['badge', user.role === 'admin' ? 'badge-orange' : 'badge-blue']">{{ user.role === 'admin' ? '管理员' : '用户' }}</span>
                </div>
                <div class="user-detail">
                  <span>最后登录: {{ formatTime(user.last_login) }}</span>
                </div>
              </div>
              <button @click="toggleUserRole(user)" class="role-toggle-btn">
                {{ user.role === 'admin' ? '降级' : '设为管理' }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- --- Monitoring Stats Tab --- -->
      <div v-if="activeTab === 'monitoring' && stats" class="tab-content animate-fade-in">
        <div class="section-header">
          <h3>系统运行概览</h3>
          <p>实时统计系统资源使用情况。</p>
        </div>

        <div class="stats-grid">
          <div class="stats-card">
            <div class="stats-icon bg-blue"><Users :size="24" /></div>
            <div class="stats-info">
              <span class="stats-label">总注册用户</span>
              <span class="stats-value">{{ stats.total_users }}</span>
            </div>
          </div>
          <div class="stats-card">
            <div class="stats-icon bg-green"><Database :size="24" /></div>
            <div class="stats-info">
              <span class="stats-label">总账单记录</span>
              <span class="stats-value">{{ stats.total_bills }}</span>
            </div>
          </div>
          <div class="stats-card">
            <div class="stats-icon bg-purple"><CalendarClock :size="24" /></div>
            <div class="stats-info">
              <span class="stats-label">周期定账任务</span>
              <span class="stats-value">{{ stats.total_recurring_tasks }}</span>
            </div>
          </div>
          <div class="stats-card">
            <div class="stats-icon bg-orange"><Mail :size="24" /></div>
            <div class="stats-info">
              <span class="stats-label">同步邮箱数量</span>
              <span class="stats-value">{{ stats.total_email_accounts }}</span>
            </div>
          </div>
        </div>

        <div class="card info-card mt-16">
          <h4>系统状态</h4>
          <div class="status-line">
            <Activity :size="16" class="text-success" />
            <span>核心服务运行正常</span>
          </div>
          <div class="status-line">
            <Smartphone :size="16" />
            <span>客户端并发接入稳定</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.admin-dashboard {
  min-height: 100vh;
  background: var(--bg-body);
  padding-top: calc(72px + env(safe-area-inset-top));
}

/* Tabs */
.tabs-container {
  background: white;
  padding: 8px 16px;
  position: sticky;
  top: calc(64px + env(safe-area-inset-top));
  z-index: 10;
  box-shadow: 0 4px 12px rgba(0,0,0,0.02);
  margin-bottom: 16px;
}
.tabs {
  display: flex;
  background: #f1f5f9;
  border-radius: 12px;
  padding: 4px;
}
.tab-item {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 4px;
  border: none;
  background: transparent;
  color: #64748b;
  font-size: 0.85rem;
  font-weight: 600;
  border-radius: 8px;
  transition: all 0.2s;
  cursor: pointer;
}
.tab-item.active {
  background: white;
  color: var(--primary);
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
}

.content {
  padding: 0 16px 32px;
  max-width: 600px;
  margin: 0 auto;
}

.section-header {
  margin: 16px 0 12px;
}
.section-header h3 {
  font-size: 1.1rem;
  font-weight: 800;
  margin: 0 0 4px;
}
.section-header p {
  font-size: 0.8rem;
  color: #64748b;
  margin: 0;
}

.card {
  background: white;
  border-radius: 20px;
  padding: 20px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);
}
.mt-16 { margin-top: 16px; }

/* Invitations */
.input-group {
  display: flex;
  gap: 10px;
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
  gap: 10px;
}
.invitation-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #f8fafc;
  border-radius: 14px;
}
.item-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.code-text {
  font-family: monospace;
  font-weight: 800;
  font-size: 1.1rem;
  letter-spacing: 1px;
}
.badges { display: flex; gap: 6px; }
.badge {
  font-size: 0.65rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 6px;
}
.badge-green { background: #dcfce7; color: #166534; }
.badge-gray { background: #f1f5f9; color: #64748b; }
.badge-orange { background: #ffedd5; color: #9a3412; }
.badge-blue { background: #dbeafe; color: #1e40af; }

/* User List */
.user-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.user-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 16px;
  border-bottom: 1px solid #f1f5f9;
}
.user-item:last-child { border-bottom: none; padding-bottom: 0; }
.user-header { display: flex; align-items: center; gap: 10px; margin-bottom: 4px; }
.phone-text { font-weight: 700; font-size: 1rem; }
.user-detail { font-size: 0.75rem; color: #94a3b8; }
.role-toggle-btn {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  padding: 6px 12px;
  border-radius: 8px;
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748b;
}

/* Monitoring Stats */
.stats-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.stats-card {
  background: white;
  border-radius: 20px;
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);
}
.stats-icon {
  width: 48px;
  height: 48px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}
.bg-blue { background: #3b82f6; }
.bg-green { background: #10b981; }
.bg-purple { background: #8b5cf6; }
.bg-orange { background: #f59e0b; }
.stats-info { display: flex; flex-direction: column; }
.stats-label { font-size: 0.7rem; color: #64748b; font-weight: 600; }
.stats-value { font-size: 1.25rem; font-weight: 900; color: #0f172a; }

.status-line {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 10px;
  font-size: 0.85rem;
  color: #334155;
  font-weight: 500;
}

.empty-state { text-align: center; padding: 40px; color: #94a3b8; font-size: 0.9rem; }
.refresh-btn { background: none; border: none; color: #94a3b8; cursor: pointer; }
.animate-spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@keyframes fadeIn { from { opacity: 0; transform: translateY(5px); } to { opacity: 1; transform: translateY(0); } }
.animate-fade-in { animation: fadeIn 0.3s ease-out forwards; }
</style>

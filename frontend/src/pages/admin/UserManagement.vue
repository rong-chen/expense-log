<script setup lang="ts">
import { ref, onMounted } from 'vue'
import TopNavBar from '@/components/layout/TopNavBar.vue'
import { adminApi } from '@/api'
import { toast } from 'vue-sonner'

interface User {
  ID: string
  phone: string
  nickname: string
  role: string
  last_login: number
  CreatedAt: string
}

const loading = ref(false)
const userList = ref<User[]>([])
const totalUsers = ref(0)
const userPage = ref(1)

async function fetchUsers() {
  loading.value = true
  try {
    const res: any = await adminApi.listUsers({ page: userPage.value, pageSize: 20 })
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
    const res: any = await adminApi.updateRole({ user_id: user.ID, role: newRole })
    if (res.code === 0) {
      toast.success('角色更新成功')
      fetchUsers()
    }
  } catch (err) {
    toast.error('角色更新失败')
  }
}

async function resetPassword(user: User) {
  const newPassword = window.prompt(`请输入用户 ${user.phone} 的新密码 (最少 6 位):`)
  if (newPassword === null) return // 取消
  if (newPassword.length < 6) {
    toast.error('密码长度不足 6 位')
    return
  }

  try {
    const res: any = await adminApi.resetPassword({ user_id: user.ID, password: newPassword })
    if (res.code === 0) {
      toast.success('密码重置成功')
    }
  } catch (err) {
    toast.error('密码重置失败')
  }
}

function formatTime(ts: number) {
  if (!ts) return '从未登录'
  return new Date(ts * 1000).toLocaleString()
}

onMounted(fetchUsers)
</script>

<template>
  <div class="user-management-page">
    <TopNavBar title="用户管理" back-to="/admin" />

    <div class="content">
      <div class="section-header">
        <h3>注册用户列表</h3>
        <p>查看并管理本系统的所有成员（当前：{{ totalUsers }}人）</p>
      </div>

      <div class="card list-card">
        <div class="list-container">
          <div v-if="loading && userList.length === 0" class="loading-state">加载中...</div>
          <div v-else-if="userList.length === 0" class="empty-state">暂无用户</div>
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
              <div class="user-actions">
                <button @click="resetPassword(user)" class="action-btn text-orange">
                  重置密码
                </button>
                <button @click="toggleUserRole(user)" class="action-btn">
                  {{ user.role === 'admin' ? '设为普通' : '设为管理' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.user-management-page {
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

.user-list {
  display: flex;
  flex-direction: column;
}

.user-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  border-bottom: 1px solid #f1f5f9;
}
.user-item:last-child { border-bottom: none; }

.user-main {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.user-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.phone-text {
  font-weight: 700;
  font-size: 1rem;
  color: var(--text-primary);
}

.user-detail {
  font-size: 0.75rem;
  color: #94a3b8;
}

.badge {
  font-size: 0.65rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 6px;
}
.badge-orange { background: #ffedd5; color: #9a3412; }
.badge-blue { background: #dbeafe; color: #1e40af; }

.user-actions {
  display: flex;
  gap: 8px;
}
.action-btn {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  padding: 6px 12px;
  border-radius: 10px;
  font-size: 0.75rem;
  font-weight: 700;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}
.action-btn:active {
  background: #f1f5f9;
}
.text-orange {
  color: #e67e22;
  border-color: rgba(230, 126, 34, 0.2);
  background: rgba(230, 126, 34, 0.05);
}
.text-orange:active {
  background: rgba(230, 126, 34, 0.1);
}

.loading-state, .empty-state {
  padding: 40px;
  text-align: center;
  color: #94a3b8;
}
</style>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Activity, Database, Users, CalendarClock, Mail, Smartphone } from 'lucide-vue-next'
import TopNavBar from '@/components/layout/TopNavBar.vue'
import request from '@/api'
import { toast } from 'vue-sonner'

interface Stats {
  total_users: number
  total_bills: number
  total_email_accounts: number
  total_recurring_tasks: number
}

const loading = ref(false)
const stats = ref<Stats | null>(null)

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

onMounted(fetchStats)
</script>

<template>
  <div class="system-stats-page">
    <TopNavBar title="系统监控" back-to="/admin" />

    <div class="content">
      <div class="section-header">
        <h3>运行概览</h3>
        <p>实时掌握系统的资源使用情况与活跃度</p>
      </div>

      <div class="stats-grid" v-if="stats">
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

      <div class="card info-card mt-24">
        <h4>系统服务状态</h4>
        <div class="status-list">
          <div class="status-item">
            <Activity :size="16" class="text-success" />
            <span>核心 API 服务：运行正常</span>
          </div>
          <div class="status-item">
            <Smartphone :size="16" />
            <span>Web/PWA 接入：稳定</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.system-stats-page {
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
  margin-bottom: 24px;
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

.stats-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.stats-card {
  background: white;
  border-radius: 20px;
  padding: 20px;
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
  flex-shrink: 0;
}

.bg-blue { background: #3b82f6; }
.bg-green { background: #10b981; }
.bg-purple { background: #8b5cf6; }
.bg-orange { background: #f59e0b; }

.stats-info {
  display: flex;
  flex-direction: column;
}

.stats-label {
  font-size: 0.65rem;
  color: var(--text-secondary);
  font-weight: 700;
  text-transform: uppercase;
  margin-bottom: 2px;
}

.stats-value {
  font-size: 1.35rem;
  font-weight: 900;
  color: var(--text-primary);
  line-height: 1.1;
}

.card {
  background: white;
  border-radius: 20px;
  padding: 20px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);
}

.mt-24 { margin-top: 24px; }

.info-card h4 {
  font-size: 0.95rem;
  font-weight: 700;
  margin: 0 0 16px 0;
  color: var(--text-primary);
}

.status-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-primary);
}

.text-success { color: #10b981; }
</style>

<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { useLedgerStore } from '@/stores/ledger'
import {
  FileText, Bell, PenLine, ArrowRight, CalendarClock, ChevronDown
} from 'lucide-vue-next'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart, PieChart, LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent, TitleComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { ref, onMounted } from 'vue'
import { billApi } from '@/api'
import { toast } from 'vue-sonner'

use([
  CanvasRenderer, BarChart, LineChart, PieChart,
  GridComponent, TooltipComponent, LegendComponent, TitleComponent,
])

const auth = useAuthStore()
const ledgerStore = useLedgerStore()
const router = useRouter()

const stats = ref({
  monthExpense: 0,
  lastMonthExpense: 0,
  monthIncome: 0,
  billCount: 0,
  pendingEmail: 0
})

const recentBills = ref<any[]>([])

// ECharts 配置
const trendOption = ref({
  tooltip: { trigger: 'axis', backgroundColor: 'rgba(255,255,255,0.9)' },
  grid: { left: '4%', right: '5%', bottom: '15%', containLabel: true },
  xAxis: { type: 'category', data: [] as string[], axisLine: { lineStyle: { color: '#bdc3c7' } }, boundaryGap: false },
  yAxis: { type: 'value', splitLine: { lineStyle: { type: 'dashed', color: '#E2E8F0' } } },
  series: [{
    name: '支出', type: 'line', smooth: true, showSymbol: false,
    areaStyle: {
      color: {
        type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
        colorStops: [{ offset: 0, color: 'rgba(37, 99, 235, 0.3)' }, { offset: 1, color: 'rgba(37, 99, 235, 0.02)' }]
      }
    },
    itemStyle: { color: '#2563EB' },
    data: [] as number[]
  }]
})

const pieOption = ref({
  tooltip: { trigger: 'item' },
  legend: { bottom: 0, padding: 0 },
  series: [{
    name: '支出分类', type: 'pie', radius: ['45%', '70%'], center: ['50%', '45%'],
    avoidLabelOverlap: false,
    itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
    label: { show: false, position: 'center' },
    emphasis: { label: { show: true, fontSize: 18, fontWeight: 'bold' } },
    labelLine: { show: false },
    data: [] as any[]
  }]
})

const hasTrendData = ref(false)
const hasCategoryData = ref(false)

async function fetchAnalytics() {
  try {
    const [dashRes, trendRes, catRes]: any = await Promise.all([
      billApi.getDashboardStats(),
      billApi.getTrendStats(),
      billApi.getCategoryStats()
    ])

    if (dashRes.code === 0) {
      stats.value.monthExpense = dashRes.data?.month_expense || 0
      stats.value.lastMonthExpense = dashRes.data?.last_month_expense || 0
      stats.value.monthIncome = dashRes.data?.month_income || 0
      stats.value.billCount = dashRes.data?.bill_count || 0
      stats.value.pendingEmail = dashRes.data?.pending_email || 0
    }
    if (trendRes.code === 0) {
      const d = trendRes.data || []
      trendOption.value.xAxis.data = d.map((i: any) => i.month)
      trendOption.value.series[0].data = d.map((i: any) => i.expense)
      hasTrendData.value = d.some((i: any) => i.expense > 0)
    }
    if (catRes.code === 0) {
      const d = catRes.data || []
      pieOption.value.series[0].data = d.map((i: any) => ({ name: i.name, value: i.value }))
      hasCategoryData.value = d.length > 0
    }
  } catch (err) {
    console.error('Failed to load analytics:', err)
  }
}

async function fetchRecentBills() {
  try {
    const now = new Date()
    const dateStr = `${now.getFullYear()}-${(now.getMonth() + 1).toString().padStart(2, '0')}`
    const res: any = await billApi.getBillList({ page: 1, size: 5, date: dateStr })
    if (res.code === 0) {
      recentBills.value = res.data?.list || []
    }
  } catch (e) {
    console.error('Failed to load recent bills:', e)
  }
}

function formatDate(dateStr: string) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

onMounted(async () => {
  if (!auth.user) auth.fetchUserInfo()
  await ledgerStore.fetchLedgers()
  fetchAnalytics()
  fetchRecentBills()
  
  // 监听全局账本变化
  window.addEventListener('ledger-changed', () => {
    fetchAnalytics()
    fetchRecentBills()
  })
})

function openEditPage(bill: any) {
  if (bill.user_id !== auth.user?.id) {
    toast.error('这是他人记录的账单，您无权修改')
    return
  }
  router.push('/bill/edit/' + bill.ID)
}
</script>

<template>
  <div class="home-page">
    <!-- 蓝色顶部 Header -->
    <div class="blue-header">
      <div class="header-row">
        <div class="ledger-selector" @click="router.push('/ledger/select')">
          <span class="ledger-name">{{ ledgerStore.currentLedger?.name || '个人主账本' }}</span>
          <ChevronDown :size="16" class="dropdown-icon" />
        </div>
        <button class="header-icon-btn"><Bell :size="20" /></button>
      </div>
      <div class="stats-row">
        <div class="stat-block">
          <span class="stat-label">本月支出</span>
          <span class="stat-value-lg">¥ {{ stats.monthExpense.toFixed(2) }}</span>
        </div>
        <div class="stat-block align-end">
          <span class="stat-label">本月收入</span>
          <span class="stat-value-sm">¥ {{ stats.monthIncome.toFixed(2) }}</span>
        </div>
      </div>
    </div>

    <!-- 内容区 -->
    <div class="content-area">
      <!-- 快捷操作栏 -->
      <div class="quick-actions">
        <div class="action-card" @click="router.push('/bill/add')">
          <div class="action-icon" style="color: var(--primary);">
            <PenLine :size="20" />
          </div>
          <span class="action-label">手动记账</span>
        </div>
        <div class="action-card" @click="router.push('/recurring')">
          <div class="action-icon" style="color: #D97706;">
            <CalendarClock :size="20" />
          </div>
          <span class="action-label">周期账单</span>
        </div>
      </div>

      <!-- 最近账单 -->
      <div class="section">
        <div class="section-header">
          <h2>最近账单</h2>
          <span class="view-all" @click="router.push('/bills')">
            查看全部 <ArrowRight :size="14" style="vertical-align: middle;" />
          </span>
        </div>
        <div v-if="recentBills.length > 0" class="recent-list">
          <div
            v-for="bill in recentBills" :key="bill.ID"
            class="recent-item"
            @click="openEditPage(bill)"
          >
            <div class="recent-info">
              <div class="recent-merchant">{{ bill.merchant || '未识别商户' }}</div>
              <div class="recent-meta">
                {{ formatDate(bill.transaction_date) }}<span v-if="bill.category"> · {{ bill.category }}</span>
                <span v-if="bill.user_id !== auth.user?.id" style="color: var(--primary);"> · 他人</span>
              </div>
            </div>
            <div class="recent-amount" :class="{ refund: bill.category === '退款' }">
              {{ bill.category === '退款' ? '' : '-' }}¥{{ bill.amount?.toFixed(2) }}
            </div>
          </div>
        </div>
        <div v-else class="empty-hint">
          <FileText :size="28" style="opacity: 0.3; margin-bottom: 8px;" />
          <p>暂无账单记录，试试手动记账</p>
        </div>
      </div>

      <!-- 图表区 -->
      <div class="section">
        <div class="section-header">
          <h2>收支趋势</h2>
        </div>
        <div class="card chart-card">
          <div v-if="!hasTrendData" class="chart-empty">
            <FileText :size="36" style="opacity: 0.4; margin-bottom: 8px;" />
            <p>暂无趋势数据</p>
          </div>
          <v-chart v-else class="chart" :option="trendOption" autoresize />
        </div>
      </div>
      <div class="section">
        <div class="section-header">
          <h2>分类支出</h2>
        </div>
        <div class="card chart-card">
          <div v-if="!hasCategoryData" class="chart-empty">
            <FileText :size="36" style="opacity: 0.4; margin-bottom: 8px;" />
            <p>暂无分类数据</p>
          </div>
          <v-chart v-else class="chart" :option="pieOption" autoresize />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.home-page {
  min-height: 100vh;
  background: var(--bg-base);
}

.blue-header {
  background: var(--primary);
  padding: calc(env(safe-area-inset-top) + 16px) 20px 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.ledger-selector {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
}
.dropdown-icon { color: rgba(255,255,255,0.67); }
.header-icon-btn {
  background: none;
  border: none;
  color: #fff;
  cursor: pointer;
  padding: 4px;
}
.stats-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}
.stat-block { display: flex; flex-direction: column; gap: 4px; }
.stat-block.align-end { align-items: flex-end; }
.stat-label { font-size: 12px; color: rgba(255,255,255,0.73); }
.stat-value-lg { font-size: 26px; font-weight: 800; color: #fff; }
.stat-value-sm { font-size: 16px; font-weight: 700; color: #fff; }

.content-area {
  padding: 16px;
  padding-bottom: 80px;
  max-width: 600px;
  margin: 0 auto;
  background: var(--bg-base);
}

.quick-actions {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}
.action-card {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}
.action-card:active { transform: scale(0.97); }
.action-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.action-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.section { margin-bottom: 20px; }
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.section-header h2 {
  font-size: 15px;
  font-weight: 700;
  margin: 0;
  color: var(--text-primary);
}
.view-all {
  font-size: 13px;
  color: var(--primary);
  cursor: pointer;
  font-weight: 500;
}

.recent-list {
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
  background: var(--bg-card);
}
.recent-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px;
  cursor: pointer;
  transition: background 0.15s;
}
.recent-item:not(:last-child) {
  border-bottom: 1px solid var(--border);
}
.recent-item:active { background: var(--bg-base); }
.recent-info { flex: 1; min-width: 0; }
.recent-merchant {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.recent-meta {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 2px;
}
.recent-amount {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  flex-shrink: 0;
  margin-left: 16px;
}
.recent-amount.refund {
  color: var(--success);
  text-decoration: line-through;
  opacity: 0.6;
}

.empty-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 30px 20px;
  color: var(--text-secondary);
  text-align: center;
}
.empty-hint p { font-size: 13px; margin: 0; }

.chart-card { padding: 16px; }
.chart { width: 100%; height: 240px; }
.chart-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 0;
  color: var(--text-secondary);
}
.chart-empty p { margin: 0; font-size: 13px; }
</style>

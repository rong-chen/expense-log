<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { useLedgerStore } from '@/stores/ledger'
import {
  FileText, Camera, Bell, PenLine, ArrowRight, CalendarClock, ChevronDown, Check, Plus, ScanLine
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
const activeTab = ref<'text' | 'chart'>('text')
const cameraInput = ref<HTMLInputElement | null>(null)
const isUploading = ref(false)
const showLedgerDrawer = ref(false)

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
  yAxis: { type: 'value', splitLine: { lineStyle: { type: 'dashed', color: '#f0ebe4' } } },
  series: [{
    name: '支出', type: 'line', smooth: true, showSymbol: false,
    areaStyle: {
      color: {
        type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
        colorStops: [{ offset: 0, color: 'rgba(230,126,34,0.4)' }, { offset: 1, color: 'rgba(230,126,34,0.05)' }]
      }
    },
    itemStyle: { color: '#e67e22' },
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

function triggerCamera() {
  cameraInput.value?.click()
}

function handleImageUpload(e: Event) {
  const target = e.target as HTMLInputElement
  if (!target.files || target.files.length === 0) return
  
  const file = target.files[0]
  target.value = '' // 立即清空，允许连续触发拍照

  isUploading.value = true
  billApi.uploadImage(file).then((res: any) => {
    if (res === 'success' || res?.code === 0) {
      fetchAnalytics()
      fetchRecentBills()
    } else {
      toast.error('识别失败: ' + (res?.msg || '未知错误'))
    }
  }).catch((err) => {
    console.error('Upload failed:', err)
    toast.error('上传失败，请检查网络后重试')
  }).finally(() => {
    isUploading.value = false
  })
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

function switchLedger(id: string) {
  ledgerStore.setCurrentLedger(id)
  showLedgerDrawer.value = false
}

async function copyInviteCode(code: string) {
  try {
    await navigator.clipboard.writeText(code)
    toast.success('邀请码已复制到剪贴板')
  } catch (err) {
    toast.error('复制失败，请手动记录')
  }
}

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
    <!-- 全屏上传 Loading -->
    <div v-if="isUploading" class="global-upload-overlay">
      <div class="spinner-container">
        <div class="loader-spinner"></div>
        <p>正在智能识别账单...</p>
      </div>
    </div>

    <!-- 顶部 Header -->
    <div class="home-header">
      <div class="header-logo" @click="showLedgerDrawer = true" style="cursor: pointer;">
        <img src="/favicon.png" alt="logo" class="logo-img" />
        <div class="ledger-selector">
          <span class="ledger-name">{{ ledgerStore.currentLedger?.name || '易账' }}</span>
          <ChevronDown :size="16" class="dropdown-icon" />
        </div>
      </div>
      <div class="header-actions">
        <button class="header-icon-btn"><Bell :size="22" /></button>
      </div>
    </div>

    <!-- 账本切换抽屉 (简易底盘) -->
    <div v-if="showLedgerDrawer" class="ledger-overlay" @click="showLedgerDrawer = false">
      <div class="ledger-drawer" @click.stop>
        <div class="drawer-header">切换账本</div>
        <div class="ledger-list">
          <div 
            v-for="l in ledgerStore.allLedgers" 
            :key="l.ID" 
            class="ledger-item"
            :class="{ active: l.ID === ledgerStore.currentLedgerId }"
            @click="switchLedger(l.ID)"
          >
            <div class="ledger-info">
              <span class="ledger-item-name">{{ l.name }}</span>
              <div style="display: flex; gap: 8px; align-items: center;">
                <span class="ledger-item-type">{{ l.type === 'personal' ? '个人' : '共享' }}</span>
                <span v-if="l.invite_code" class="ledger-invite-code" @click.stop="copyInviteCode(l.invite_code)">
                  邀请码: {{ l.invite_code }} (点击复制)
                </span>
              </div>
            </div>
            <Check v-if="l.ID === ledgerStore.currentLedgerId" :size="20" class="active-icon" />
          </div>
        </div>
        <div class="drawer-actions">
          <button class="drawer-btn" @click="router.push('/ledger/create')">
            <Plus :size="18" /> 创建账本
          </button>
          <button class="drawer-btn" @click="router.push('/ledger/join')">
            <ScanLine :size="18" /> 加入账本
          </button>
        </div>
      </div>
    </div>

    <!-- 快捷操作栏 -->
    <div class="quick-actions">
      <div class="action-card card" @click="router.push('/bill/add')">
        <div class="action-icon" style="background: rgba(26,188,156,0.12); color: var(--primary);">
          <PenLine :size="22" />
        </div>
        <span class="action-label">手动记账</span>
      </div>
      <div class="action-card card" @click="triggerCamera">
        <div class="action-icon" style="background: rgba(230,126,34,0.12); color: #e67e22;">
          <Camera :size="22" />
        </div>
        <span class="action-label">拍照识别</span>
      </div>
      <input type="file" ref="cameraInput" accept="image/*" @change="handleImageUpload" style="display: none" />
      <div class="action-card card" @click="router.push('/recurring')">
        <div class="action-icon" style="background: rgba(231,76,60,0.12); color: #e74c3c;">
          <CalendarClock :size="22" />
        </div>
        <span class="action-label">周期账单</span>
      </div>
    </div>

    <!-- 切换 Tab -->
    <div class="tab-switch">
      <button :class="['tab-btn', { active: activeTab === 'text' }]" @click="activeTab = 'text'">数据总览</button>
      <button :class="['tab-btn', { active: activeTab === 'chart' }]" @click="activeTab = 'chart'">图表分析</button>
    </div>

    <!-- 文字统计视图 -->
    <div v-if="activeTab === 'text'">
      <div class="stats-grid">
        <div class="stat-card card">
          <div class="stat-body">
            <span class="stat-label">本月支出</span>
            <span class="stat-value">¥ {{ stats.monthExpense.toFixed(2) }}</span>
          </div>
        </div>
        <div class="stat-card card">
          <div class="stat-body">
            <span class="stat-label">本月账单数</span>
            <span class="stat-value">{{ stats.billCount }} <span>笔</span></span>
          </div>
        </div>
        <div class="stat-card card span-2">
          <div class="stat-body">
            <span class="stat-label">环比上月</span>
            <div class="compare-row">
              <span class="stat-value">¥ {{ stats.lastMonthExpense.toFixed(2) }}</span>
              <span 
                v-if="stats.lastMonthExpense > 0" 
                class="compare-badge" 
                :class="stats.monthExpense > stats.lastMonthExpense ? 'up' : 'down'"
              >
                {{ stats.monthExpense > stats.lastMonthExpense ? '↑' : '↓' }}
                {{ Math.abs(((stats.monthExpense - stats.lastMonthExpense) / stats.lastMonthExpense) * 100).toFixed(1) }}%
              </span>
              <span v-else class="compare-badge neutral">暂无数据</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 最近账单预览 -->
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
            class="recent-item card"
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
          <p>暂无账单记录，试试手动记账或拍照录入</p>
        </div>
      </div>
    </div>

    <!-- 图表视图 -->
    <div v-else>
      <div class="card chart-card">
        <h2 class="chart-title">收支趋势</h2>
        <div v-if="!hasTrendData" class="chart-empty">
          <FileText :size="36" style="opacity: 0.4; margin-bottom: 8px;" />
          <p>暂无趋势数据</p>
        </div>
        <v-chart v-else class="chart" :option="trendOption" autoresize />
      </div>
      <div class="card chart-card">
        <h2 class="chart-title">分类支出</h2>
        <div v-if="!hasCategoryData" class="chart-empty">
          <FileText :size="36" style="opacity: 0.4; margin-bottom: 8px;" />
          <p>暂无分类数据</p>
        </div>
        <v-chart v-else class="chart" :option="pieOption" autoresize />
      </div>
    </div>
  </div>
</template>

<style scoped>
.home-page {
  padding: 16px;
  max-width: 600px;
  margin: 0 auto;
  padding-top: calc(16px + env(safe-area-inset-top));
}

/* Header */
.home-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.header-logo { display: flex; align-items: center; gap: 8px; font-size: 1.4rem; font-weight: 800; color: var(--primary); letter-spacing: -0.5px; }
.logo-img { width: 28px; height: 28px; border-radius: 6px; box-shadow: 0 4px 12px rgba(230,126,34,0.2); }
.header-actions { display: flex; gap: 8px; }
.header-icon-btn {
  width: 40px; height: 40px; border-radius: 12px; border: none;
  background: white; color: var(--text-secondary);
  display: flex; justify-content: center; align-items: center;
  cursor: pointer; box-shadow: 0 2px 8px rgba(0,0,0,0.04); transition: all 0.2s;
}
.header-icon-btn:active { transform: scale(0.92); background: var(--primary-soft); color: var(--primary); }

/* 快捷操作栏 */
.quick-actions {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 20px;
}
.action-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 18px 12px;
  cursor: pointer;
  transition: all 0.2s;
}
.action-card:active {
  transform: scale(0.96);
}
.action-icon {
  width: 48px;
  height: 48px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.action-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-primary);
}

/* 切换 Tab */
.tab-switch {
  display: flex;
  background: rgba(0,0,0,0.04);
  border-radius: 12px;
  padding: 3px;
  margin-bottom: 20px;
}
.tab-btn {
  flex: 1;
  padding: 10px 0;
  border: none;
  background: transparent;
  border-radius: 10px;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.25s;
}
.tab-btn.active {
  background: white;
  color: var(--primary);
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
}

/* 文字统计 */
.stats-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; margin-bottom: 24px; }
.stat-card { display: flex; align-items: center; padding: 16px; min-width: 0; }
.stat-body { display: flex; flex-direction: column; flex: 1; min-width: 0; }
.stat-label { font-size: 0.8rem; color: var(--text-secondary); margin-bottom: 2px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.stat-value { font-size: 1.15rem; font-weight: 700; color: var(--text-primary); white-space: normal; word-break: break-all; }
.span-2 { grid-column: span 2; }
.compare-row { display: flex; align-items: center; gap: 10px; margin-top: 2px; }
.compare-badge {
  font-size: 0.8rem;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: 20px;
}
.compare-badge.up { background: rgba(231,76,60,0.1); color: #e74c3c; }
.compare-badge.down { background: rgba(39,174,96,0.1); color: #27ae60; }
.compare-badge.neutral { background: rgba(0,0,0,0.05); color: var(--text-secondary); }

/* 最近账单 */
.section { margin-bottom: 24px; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 14px; }
.section-header h2 { font-size: 1.1rem; font-weight: 700; margin: 0; color: var(--text-primary); }
.view-all { font-size: 0.85rem; color: var(--primary); cursor: pointer; font-weight: 500; }

.recent-list { display: flex; flex-direction: column; gap: 10px; }
.recent-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  cursor: pointer;
  transition: all 0.15s;
}
.recent-item:active { transform: scale(0.98); }
.recent-info { flex: 1; min-width: 0; }
.recent-merchant {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.recent-meta {
  font-size: 0.78rem;
  color: var(--text-secondary);
  margin-top: 3px;
}
.recent-amount {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--text-primary);
  flex-shrink: 0;
  margin-left: 16px;
}
.recent-amount.refund {
  color: #27ae60;
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
.empty-hint p { font-size: 0.85rem; margin: 0; max-width: 220px; }

/* 图表视图 */
.chart-card { padding: 20px; margin-bottom: 16px; }
.chart-title { font-size: 1.1rem; font-weight: 600; margin: 0 0 16px 0; color: var(--text-primary); }
.chart { width: 100%; height: 260px; }
.chart-empty {
  display: flex; flex-direction: column; align-items: center;
  justify-content: center; padding: 40px 0; color: var(--text-secondary);
}
.chart-empty p { margin: 0; font-size: 0.9rem; }

/* 全局上传 Loading 遮罩层 */
.global-upload-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  z-index: 9999;
  display: flex;
  justify-content: center;
  align-items: center;
}
.dark .global-upload-overlay {
  background: rgba(20, 18, 24, 0.85);
}

.spinner-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  animation: fadeIn 0.3s ease-out;
}

.spinner-container p {
  color: var(--primary);
  font-weight: 600;
  font-size: 0.95rem;
  margin: 0;
  letter-spacing: 0.5px;
}

.loader-spinner {
  width: 48px;
  height: 48px;
  border: 4px solid var(--primary-soft);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
@keyframes fadeIn {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}
/* 账本选择器样式 */
.ledger-selector {
  display: flex;
  align-items: center;
  gap: 4px;
  background: rgba(0, 0, 0, 0.04);
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 0.9rem;
  font-weight: 500;
  transition: background 0.2s;
}
.ledger-selector:active {
  background: rgba(0, 0, 0, 0.08);
}
.ledger-name {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.dropdown-icon {
  opacity: 0.6;
}

/* 账本抽屉样式 */
.ledger-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 2000;
  animation: fadeIn 0.2s ease-out;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
}
.ledger-drawer {
  background: var(--surface);
  border-radius: 24px 24px 0 0;
  padding: 24px 16px env(safe-area-inset-bottom);
  animation: slideUp 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.drawer-header {
  font-size: 1.1rem;
  font-weight: 600;
  text-align: center;
  margin-bottom: 20px;
}
.ledger-list {
  max-height: 50vh;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.ledger-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: var(--background);
  border-radius: 16px;
  transition: background 0.2s;
}
.ledger-item:active {
  background: rgba(0, 0, 0, 0.05);
}
.ledger-item.active {
  background: rgba(44, 62, 80, 0.05);
  border: 1px solid rgba(44, 62, 80, 0.2);
}
.ledger-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.ledger-item-name {
  font-weight: 600;
  font-size: 1rem;
}
.ledger-item-type {
  font-size: 0.8rem;
  color: var(--text-secondary);
  background: rgba(0, 0, 0, 0.04);
  padding: 2px 8px;
  border-radius: 10px;
  width: fit-content;
}
.ledger-invite-code {
  font-size: 0.8rem;
  color: var(--primary);
  background: rgba(26, 188, 156, 0.1);
  padding: 2px 8px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
}
.ledger-invite-code:active {
  background: rgba(26, 188, 156, 0.2);
}
.active-icon {
  color: var(--primary);
}
.drawer-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
}
.drawer-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px;
  border: none;
  border-radius: 12px;
  background: rgba(44, 62, 80, 0.05);
  color: var(--primary);
  font-weight: 600;
  font-size: 0.95rem;
}
@keyframes slideUp {
  from { transform: translateY(100%); }
  to { transform: translateY(0); }
}
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>

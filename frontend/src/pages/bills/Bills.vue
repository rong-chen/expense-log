<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { billApi } from '@/api'
import { FileText, Search } from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import { toast } from 'vue-sonner'

const router = useRouter()
const auth = useAuthStore()

const bills = ref<any[]>([])
const loading = ref(true)
const loadingMore = ref(false)
const page = ref(1)
const size = ref(15) // 改为小批量，启用无限滚动
const hasMore = ref(false)
const totalCount = ref(0)

const keyword = ref('')
const category = ref('')

// 本地时区获取当前月份避免 UTC 跨天误差
const now = new Date()
const currentMonth = `${now.getFullYear()}-${(now.getMonth() + 1).toString().padStart(2, '0')}`
const date = ref(currentMonth)

const CATEGORIES = ['餐饮', '交通', '购物', '娱乐', '生活缴费', '转账', '其他']

async function fetchBills(isAppend = false) {
  if (!isAppend) loading.value = true
  else loadingMore.value = true

  try {
    const res: any = await billApi.getBillList({ 
      page: page.value, 
      size: size.value,
      keyword: keyword.value,
      category: category.value,
      date: date.value
    })
    if (res.code === 0) {
      const newList = res.data.list || []
      if (isAppend) {
        bills.value.push(...newList)
      } else {
        bills.value = newList
      }
      totalCount.value = res.data.total
      hasMore.value = bills.value.length < res.data.total
    }
  } catch (err) {
    console.error('Failed to load bills:', err)
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

let filterTimer: any = null
function onFilterChange() {
  // 防御 iOS 等系统原生的清除/还原按钮，若文本被置空则自动倒退回当前时间，避免拉取出全库脏数据
  if (!date.value) {
    const now = new Date()
    date.value = `${now.getFullYear()}-${(now.getMonth() + 1).toString().padStart(2, '0')}`
  }

  clearTimeout(filterTimer)
  filterTimer = setTimeout(() => {
    page.value = 1
    fetchBills(false) // 重新从第一页查起
  }, 400)
}

function formatDate(dateStr: string) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const loadMoreTrigger = ref<HTMLElement | null>(null)
let observer: IntersectionObserver | null = null

onMounted(() => {
  fetchBills(false)

  // 绑定 IntersectionObserver 进行无限滚动监听
  observer = new IntersectionObserver((entries) => {
    if (entries[0].isIntersecting && hasMore.value && !loading.value && !loadingMore.value) {
      page.value++
      fetchBills(true)
    }
  }, { rootMargin: '100px' })

  // 延时保证 DOM 挂载完毕
  setTimeout(() => {
    if (loadMoreTrigger.value) {
      observer?.observe(loadMoreTrigger.value)
    }
  }, 500)
})

function openEditPage(bill: any) {
  if (bill.user_id !== auth.user?.uid) {
    toast.error('这是他人记录的账单，您无权修改')
    return
  }
  router.push('/bill/edit/' + bill.ID)
}
</script>

<template>
  <div class="bills-page">
    <div class="filter-bar">
      <!-- 搜索框 -->
      <div class="search-box">
        <Search :size="16" class="search-icon" />
        <input type="text" v-model="keyword" @input="onFilterChange" placeholder="搜索商户名或备注" class="search-input" />
      </div>
      
      <!-- 筛选组合 -->
      <div class="filter-group">
        <select v-model="category" @change="onFilterChange" class="filter-select">
          <option value="">全部分类</option>
          <option v-for="c in CATEGORIES" :key="c" :value="c">{{ c }}</option>
        </select>
        <input type="month" v-model="date" @change="onFilterChange" class="filter-date" />
      </div>
    </div>

    <div class="list-summary" v-if="!loading">
      符合条件的账单共 <strong>{{ totalCount }}</strong> 笔
    </div>

    <!-- 骨架屏加载状态 -->
    <div v-if="loading" class="skeleton-list">
      <div v-for="i in 5" :key="i" class="skeleton-item card">
        <div class="skeleton-main">
          <div class="skeleton-line skeleton w-100"></div>
          <div class="skeleton-line skeleton w-60"></div>
        </div>
        <div class="skeleton-amount skeleton"></div>
      </div>
    </div>

    <!-- 真实数据 -->
    <div v-else-if="bills.length > 0" class="bill-list">
      <div v-for="bill in bills" :key="bill.ID" class="bill-card card" @click="openEditPage(bill)" style="cursor: pointer;">
        
        <!-- 商户名与备注 -->
        <div class="bill-info">
          <div class="bill-merchant text-truncate">{{ bill.merchant || '未识别商户' }}</div>
          <div class="bill-meta text-truncate">
            <span class="meta-item">{{ formatDate(bill.transaction_date) }}</span>
            <span class="meta-item" v-if="bill.category">· {{ bill.category }}</span>
            <span class="meta-item" v-if="bill.user_id !== auth.user?.uid" style="color: var(--primary);">· 他人</span>
          </div>
          <div class="bill-remark text-truncate">
            {{ bill.remark || '（暂无备注）' }}
          </div>
        </div>

        <!-- 右侧金额：绝对保护不被挤压 -->
        <div class="bill-amount" :class="{ refund: bill.category === '退款' }">
          {{ bill.category === '退款' ? '' : '-' }}¥{{ Number(bill.amount).toFixed(2) }}
        </div>

      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <FileText :size="48" style="opacity: 0.3; margin-bottom: 16px;" />
      <h3>尚无明细</h3>
      <p>您的所有账单分析都会归档在这里</p>
    </div>

    <!-- 无限滚动底部触发器 -->
    <div ref="loadMoreTrigger" class="load-more-trigger">
      <div v-if="loadingMore" class="loading-spinner">加载更多数据中...</div>
    </div>
  </div>
</template>

<style scoped>
.bills-page {
  padding: 16px;
  max-width: 600px;
  margin: 0 auto;
  padding-top: calc(16px + env(safe-area-inset-top));
  padding-bottom: 24px;
}

/* 过滤工具栏 */
.filter-bar {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 20px;
}

.search-box {
  position: relative;
  display: flex;
  align-items: center;
}
.search-icon {
  position: absolute;
  left: 14px;
  color: var(--text-muted);
}
.search-input {
  width: 100%;
  background: white;
  border: 1px solid rgba(0,0,0,0.06);
  padding: 14px 14px 14px 40px;
  border-radius: 12px;
  font-size: 0.95rem;
  outline: none;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(0,0,0,0.01);
}
.search-input:focus {
  border-color: var(--primary);
  box-shadow: 0 4px 12px rgba(26, 188, 156, 0.1);
}

.filter-group {
  display: flex;
  gap: 12px;
}
.filter-select, .filter-date {
  flex: 1;
  background: white;
  border: 1px solid rgba(0,0,0,0.06);
  padding: 12px;
  border-radius: 12px;
  font-size: 0.9rem;
  color: var(--text-primary);
  outline: none;
  box-shadow: 0 2px 8px rgba(0,0,0,0.01);
}

.list-summary {
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin-bottom: 16px;
  padding: 0 4px;
}
.list-summary strong {
  color: var(--primary);
  font-weight: 700;
  margin: 0 2px;
}

/* 账单卡片核心布局：解决长文本换行痛点 */
.bill-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.bill-card {
  display: flex;
  align-items: center;
  padding: 16px;
  border-radius: 16px;
  transition: active 0.2s;
}
.bill-card:active {
  transform: scale(0.98);
  background: var(--bg-card-hover);
}

.bill-info {
  flex: 1;
  min-width: 0; /* CSS Flex 文本截断核心：允许子元素宽度收缩到比内容还小 */
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 4px;
}

/* 这个工具类至关重要，解决无限溢出 */
.text-truncate {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bill-merchant {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.2;
}

.bill-meta {
  font-size: 0.75rem;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  gap: 8px;
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 3px;
}

.bill-remark {
  font-size: 0.75rem;
  color: var(--text-tertiary);
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.load-more-trigger {
  height: 20px;
  width: 100%;
}
.loading-spinner {
  text-align: center;
  font-size: 0.8rem;
  color: var(--text-tertiary);
  padding: 10px 0;
}

.bill-amount {
  font-size: 1.15rem;
  font-weight: 800;
  color: var(--text-primary);
  margin-left: 12px;
  flex-shrink: 0;
  max-width: 40%;
  text-align: right;
  word-break: break-all;
}
.bill-amount.refund {
  color: #27ae60;
  text-decoration: line-through;
  opacity: 0.6;
}

/* 骨架屏 */
.skeleton-list { display: flex; flex-direction: column; gap: 12px; }
.skeleton-item { display: flex; align-items: center; padding: 16px; border-radius: 16px; }
.skeleton-main { flex: 1; display: flex; flex-direction: column; gap: 8px; }
.skeleton-line { height: 14px; border-radius: 7px; }
.w-100 { width: 100%; }
.w-60 { width: 60%; }
.skeleton-amount { width: 60px; height: 20px; border-radius: 10px; margin-left: 12px; flex-shrink: 0; }

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  color: var(--text-secondary);
  text-align: center;
}
.empty-state h3 { font-size: 1.15rem; margin-bottom: 8px; }
.empty-state p { font-size: 0.9rem; max-width: 240px; }
</style>

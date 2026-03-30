<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Plus, Trash2, Power } from 'lucide-vue-next'
import TopNavBar from '@/components/layout/TopNavBar.vue'
import request from '@/api'
import { toast } from 'vue-sonner'

const router = useRouter()

interface RecurringBill {
  ID: string
  amount: number
  merchant: string
  category: string
  remark: string
  day_of_month: number
  is_active: boolean
  last_exec_at: string | null
}

const loading = ref(true)
const list = ref<RecurringBill[]>([])
const deletingId = ref('')

async function fetchList() {
  loading.value = true
  try {
    const res: any = await request.get('/recurring')
    if (res.code === 0) {
      list.value = res.data || []
    }
  } catch {
    toast.error('获取周期账单失败')
  } finally {
    loading.value = false
  }
}

async function deleteItem(id: string) {
  if (!confirm('确定要删除此周期账单吗？')) return
  deletingId.value = id
  try {
    const res: any = await request.delete(`/recurring/${id}`)
    if (res.code === 0) {
      fetchList()
    } else {
      toast.error(res.msg || '删除失败')
    }
  } catch {
    toast.error('删除失败')
  } finally {
    deletingId.value = ''
  }
}

async function toggleActive(id: string) {
  try {
    const res: any = await request.patch(`/recurring/${id}/toggle`)
    if (res.code === 0) {
      fetchList()
    } else {
      toast.error(res.msg || '操作失败')
    }
  } catch {
    toast.error('操作失败')
  }
}

onMounted(() => {
  fetchList()
})
</script>

<template>
  <div class="recurring-page">
    <TopNavBar title="周期账单" />

    <div class="content">
      <div class="section-header">
        <h3>订阅与固定支出</h3>
        <p>设置每月固定扣款项目，系统将在指定日期自动生成账单记录。</p>
      </div>

      <div class="card-area">
        <div class="card-header">
          <h2>周期列表</h2>
          <button class="btn-add" @click="router.push('/recurring/add')">
            <Plus :size="18" /> 新增
          </button>
        </div>

        <div v-if="loading" class="loading">加载中...</div>

        <div v-else-if="list.length === 0" class="empty">
          <p>暂无周期账单</p>
        </div>

        <div v-else class="list">
          <div v-for="item in list" :key="item.ID" class="recurring-card" :class="{ inactive: !item.is_active }">
            <div class="card-top">
              <div class="card-info" @click="router.push('/recurring/edit/' + item.ID)">
                <div class="card-name">{{ item.merchant }}</div>
                <div class="card-meta">
                  <span class="tag">{{ item.category || '未分类' }}</span>
                  <span class="day">每月 {{ item.day_of_month }} 号</span>
                </div>
                <div class="card-remark" v-if="item.remark">{{ item.remark }}</div>
              </div>
              <div class="card-amount">¥{{ item.amount.toFixed(2) }}</div>
            </div>

            <div class="card-bottom">
              <div class="card-status">
                <span v-if="item.last_exec_at" class="last-exec">上次执行: {{ item.last_exec_at }}</span>
                <span v-else class="last-exec">尚未执行过</span>
              </div>
              <div class="card-actions">
                <button class="action-btn toggle-btn" @click="toggleActive(item.ID)" :title="item.is_active ? '暂停' : '启用'">
                  <Power :size="16" />
                  {{ item.is_active ? '暂停' : '启用' }}
                </button>
                <button class="action-btn delete-btn" @click="deleteItem(item.ID)" :disabled="deletingId === item.ID">
                  <Trash2 :size="16" />
                  {{ deletingId === item.ID ? '删除中...' : '删除' }}
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
.recurring-page {
  padding: 16px;
  padding-top: calc(72px + env(safe-area-inset-top));
  max-width: 600px;
  margin: 0 auto;
  min-height: 100vh;
  background: var(--bg-body);
}

.section-header {
  margin-bottom: 20px;
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

.card-area {
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);
  padding: 20px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.card-header h2 {
  font-size: 1.05rem;
  font-weight: 700;
  margin: 0;
}
.btn-add {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 16px;
  background: var(--primary);
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
}
.btn-add:active { transform: scale(0.95); }

.loading, .empty {
  text-align: center;
  padding: 40px 0;
  color: var(--text-secondary);
}

.list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.recurring-card {
  background: #fafafa;
  border-radius: 14px;
  padding: 16px;
  transition: opacity 0.2s;
}
.recurring-card.inactive {
  opacity: 0.5;
}
.card-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}
.card-info {
  flex: 1;
  cursor: pointer;
}
.card-name {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 6px;
}
.card-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}
.tag {
  font-size: 0.75rem;
  color: white;
  background: var(--primary);
  padding: 2px 8px;
  border-radius: 10px;
}
.day {
  font-size: 0.8rem;
  color: var(--text-secondary);
}
.card-remark {
  font-size: 0.8rem;
  color: var(--text-secondary);
  margin-top: 6px;
}
.card-amount {
  font-size: 1.3rem;
  font-weight: 800;
  color: var(--primary);
  white-space: nowrap;
  margin-left: 12px;
}

.card-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed rgba(0,0,0,0.06);
}
.card-status { flex: 1; }
.last-exec {
  font-size: 0.75rem;
  color: var(--text-muted, #95a5a6);
}
.card-actions {
  display: flex;
  gap: 8px;
}
.action-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border-radius: 8px;
  border: none;
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}
.toggle-btn {
  background: rgba(41, 128, 185, 0.1);
  color: #2980b9;
}
.delete-btn {
  background: rgba(231, 76, 60, 0.1);
  color: var(--danger);
}
.action-btn:active { transform: scale(0.93); }
</style>

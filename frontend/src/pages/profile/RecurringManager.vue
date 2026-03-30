<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus, Trash2, Calendar, Power } from 'lucide-vue-next'
import TopNavBar from '@/components/layout/TopNavBar.vue'
import request from '@/api'
import { toast } from 'vue-sonner'

const CATEGORIES = ['餐饮', '交通', '购物', '娱乐', '生活缴费', '转账', '医疗', '其他']

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
const showForm = ref(false)
const saving = ref(false)
const deletingId = ref('')

const form = ref({
  amount: '',
  merchant: '',
  category: '生活缴费',
  remark: '',
  day_of_month: 1
})

const editingId = ref<string | null>(null)

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

function openAdd() {
  editingId.value = null
  form.value = { amount: '', merchant: '', category: '生活缴费', remark: '', day_of_month: 1 }
  showForm.value = true
}

function openEdit(item: RecurringBill) {
  editingId.value = item.ID
  form.value = {
    amount: String(item.amount),
    merchant: item.merchant,
    category: item.category || '其他',
    remark: item.remark || '',
    day_of_month: item.day_of_month
  }
  showForm.value = true
}

async function submitForm() {
  const amt = Number(form.value.amount)
  if (!amt || amt <= 0) {
    toast.error('请输入有效金额')
    return
  }
  if (!form.value.merchant.trim()) {
    toast.error('请输入项目名称')
    return
  }

  saving.value = true
  try {
    const payload = {
      amount: amt,
      merchant: form.value.merchant.trim(),
      category: form.value.category,
      remark: form.value.remark.trim(),
      day_of_month: form.value.day_of_month
    }

    let res: any
    if (editingId.value) {
      res = await request.put(`/recurring/${editingId.value}`, payload)
    } else {
      res = await request.post('/recurring', payload)
    }

    if (res.code === 0) {
      showForm.value = false
      fetchList()
    } else {
      toast.error(res.msg || '保存失败')
    }
  } catch {
    toast.error('网络或未知错误')
  } finally {
    saving.value = false
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

// 生成 1-31 的日期选项
const dayOptions = Array.from({ length: 31 }, (_, i) => i + 1)

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
          <button class="btn-add" @click="openAdd" :disabled="saving">
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
              <div class="card-info" @click="openEdit(item)">
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

    <!-- 新增/编辑弹窗 -->
    <div class="modal-overlay" v-if="showForm" @click.self="showForm = false">
      <div class="modal-content form-modal">
        <h3>{{ editingId ? '编辑周期账单' : '新增周期账单' }}</h3>

        <div class="form-group amount-group">
          <label>每期金额 (¥)</label>
          <input type="number" step="0.01" v-model="form.amount" class="form-control amount-input" placeholder="0.00" inputmode="decimal" />
        </div>

        <div class="form-group">
          <label>项目名称</label>
          <input type="text" v-model="form.merchant" class="form-control" placeholder="例如：房租、Apple Music" />
        </div>

        <div class="form-group">
          <label>分类</label>
          <div class="category-grid">
            <button
              v-for="c in CATEGORIES" :key="c"
              :class="['category-chip', { active: form.category === c }]"
              @click="form.category = c"
            >{{ c }}</button>
          </div>
        </div>

        <div class="form-group">
          <label><Calendar :size="14" style="vertical-align: middle; margin-right: 4px;" />每月扣款日</label>
          <select v-model.number="form.day_of_month" class="form-control">
            <option v-for="d in dayOptions" :key="d" :value="d">{{ d }} 号</option>
          </select>
        </div>

        <div class="form-group">
          <label>备注</label>
          <textarea v-model="form.remark" class="form-control remark-textarea" placeholder="可选" rows="2"></textarea>
        </div>

        <div class="form-actions">
          <button class="btn btn-cancel" @click="showForm = false">取消</button>
          <button class="btn btn-primary" @click="submitForm" :disabled="saving">
            {{ saving ? '保存中...' : '保存' }}
          </button>
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

/* 弹窗 */
.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.4);
  display: flex;
  justify-content: center;
  align-items: flex-end;
  z-index: 1000;
  animation: fadeIn 0.2s ease-out;
}
.form-modal {
  background: white;
  width: 100%;
  max-width: 500px;
  border-radius: 24px 24px 0 0;
  padding: 28px 24px calc(24px + env(safe-area-inset-bottom));
  max-height: 90vh;
  overflow-y: auto;
  animation: slideUp 0.3s ease-out;
}
.form-modal h3 {
  font-size: 1.15rem;
  font-weight: 700;
  margin: 0 0 20px 0;
  text-align: center;
}

.form-group {
  display: flex;
  flex-direction: column;
  margin-bottom: 16px;
}
.form-group label {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-bottom: 8px;
  font-weight: 600;
}
.form-control {
  width: 100%;
  border: 1px solid rgba(0,0,0,0.08);
  border-radius: 12px;
  padding: 14px 16px;
  font-size: 1rem;
  font-family: inherit;
  background: #fafafa;
  outline: none;
  color: var(--text-primary);
  transition: all 0.2s;
  -webkit-appearance: none;
  appearance: none;
}
.form-control:focus {
  border-color: var(--primary);
  background: white;
  box-shadow: 0 4px 12px rgba(26, 188, 156, 0.1);
}

.amount-group {
  border-bottom: 1px dashed rgba(0,0,0,0.1);
  padding-bottom: 16px;
}
.amount-input {
  font-size: 2rem !important;
  font-weight: 800;
  color: var(--primary) !important;
  height: auto !important;
  padding: 8px 0 !important;
  background: transparent !important;
  border: none !important;
  border-radius: 0 !important;
}

.category-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.category-chip {
  padding: 6px 14px;
  border-radius: 20px;
  border: 1.5px solid rgba(0,0,0,0.08);
  background: #fafafa;
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}
.category-chip.active {
  background: var(--primary);
  color: white;
  border-color: var(--primary);
}
.category-chip:active { transform: scale(0.95); }

.remark-textarea { resize: none; }

.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}
.btn {
  flex: 1;
  padding: 16px;
  border-radius: 14px;
  font-size: 1rem;
  font-weight: 700;
  cursor: pointer;
  border: none;
  text-align: center;
  transition: all 0.15s;
}
.btn-cancel {
  background: #f0f0f0;
  color: var(--text-secondary);
}
.btn-primary {
  background: var(--primary);
  color: white;
  box-shadow: 0 6px 20px rgba(26, 188, 156, 0.3);
}
.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.btn:active { transform: scale(0.98); }

@keyframes fadeIn { from { opacity: 0 } to { opacity: 1 } }
@keyframes slideUp { from { transform: translateY(100%) } to { transform: translateY(0) } }
</style>

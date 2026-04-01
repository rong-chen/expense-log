<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Calendar } from 'lucide-vue-next'
import TopNavBar from '@/components/layout/TopNavBar.vue'
import request from '@/api'
import { toast } from 'vue-sonner'

const router = useRouter()
const route = useRoute()

const CATEGORIES = ['餐饮', '交通', '购物', '娱乐', '生活缴费', '转账', '医疗', '其他']
const dayOptions = Array.from({ length: 31 }, (_, i) => i + 1)

const editingId = ref<string | null>(null)
const saving = ref(false)
const pageTitle = ref('新增周期账单')

const form = ref({
  amount: '',
  merchant: '',
  category: '生活缴费',
  remark: '',
  day_of_month: 1,
  execute_now: false
})

// 如果有 id 参数，说明是编辑模式
async function loadEditData(id: string) {
  try {
    const res: any = await request.get('/recurring')
    if (res.code === 0) {
      const item = (res.data || []).find((r: any) => r.ID === id)
      if (item) {
        editingId.value = id
        pageTitle.value = '编辑周期账单'
        form.value = {
          amount: String(item.amount),
          merchant: item.merchant,
          category: item.category || '其他',
          remark: item.remark || '',
          day_of_month: item.day_of_month,
          execute_now: false
        }
      }
    }
  } catch {
    toast.error('加载数据失败')
  }
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
      day_of_month: form.value.day_of_month,
      execute_now: form.value.execute_now
    }

    let res: any
    if (editingId.value) {
      res = await request.put(`/recurring/${editingId.value}`, payload)
    } else {
      res = await request.post('/recurring', payload)
    }

    if (res.code === 0) {
      router.back()
    } else {
      toast.error(res.msg || '保存失败')
    }
  } catch {
    toast.error('网络或未知错误')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  const id = route.params.id as string
  if (id) {
    loadEditData(id)
  }
})
</script>

<template>
  <div class="add-page">
    <TopNavBar :title="pageTitle" />

    <div class="form-container">
      <div class="card form-card">
        <div class="form-group amount-group">
          <label>每期金额 (¥)</label>
          <input
            type="number"
            step="0.01"
            v-model="form.amount"
            class="form-control amount-input"
            placeholder="0.00"
            inputmode="decimal"
          />
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
          <textarea v-model="form.remark" class="form-control remark-textarea" placeholder="可选" rows="3"></textarea>
        </div>

        <div class="form-group row-group" v-if="!editingId">
          <div class="checkbox-wrapper">
            <input type="checkbox" id="execute_now" v-model="form.execute_now" class="checkbox-input" />
            <label for="execute_now" class="checkbox-label">立即生成本月账单记录</label>
          </div>
          <p class="help-text">选中后，系统将立即为您补录一笔本月（{{ new Date().getMonth() + 1 }}月）的消费记录。</p>
        </div>
      </div>

      <div class="bottom-actions">
        <button class="btn btn-primary" @click="submitForm" :disabled="saving">
          {{ saving ? '保存中...' : '确认保存' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.add-page {
  padding: 16px;
  padding-top: calc(72px + env(safe-area-inset-top));
  max-width: 600px;
  margin: 0 auto;
  min-height: 100vh;
  background: var(--bg-body);
  display: flex;
  flex-direction: column;
}

.form-container {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.form-card {
  padding: 24px 20px;
  border-radius: 20px;
  margin-bottom: 30px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
}
.amount-group {
  border-bottom: 1px dashed rgba(0,0,0,0.1);
  padding-bottom: 20px;
}
.amount-input {
  font-size: 2.2rem !important;
  font-weight: 800;
  color: var(--primary) !important;
  height: auto !important;
  padding: 10px 0 !important;
  background: transparent !important;
  border: none !important;
  border-radius: 0 !important;
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
  text-align: left;
}
.form-control:focus {
  border-color: var(--primary);
  background: white;
  box-shadow: 0 4px 12px rgba(26, 188, 156, 0.1);
}

.category-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
.category-chip {
  padding: 8px 16px;
  border-radius: 20px;
  border: 1.5px solid rgba(0,0,0,0.08);
  background: #fafafa;
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}
.category-chip.active {
  background: var(--primary);
  color: white;
  border-color: var(--primary);
  box-shadow: 0 4px 12px rgba(26, 188, 156, 0.25);
}
.category-chip:active {
  transform: scale(0.95);
}

.remark-textarea { resize: none; }

.bottom-actions {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding-bottom: calc(24px + env(safe-area-inset-bottom));
  margin-top: auto;
}

.btn {
  width: 100%;
  padding: 16px;
  border-radius: 16px;
  font-size: 1.05rem;
  font-weight: 700;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
  text-align: center;
}
.btn-primary {
  background: var(--primary);
  color: white;
  box-shadow: 0 6px 20px rgba(26, 188, 156, 0.3);
}
.btn-primary:active {
  transform: scale(0.98);
  box-shadow: 0 2px 10px rgba(26, 188, 156, 0.2);
}
.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.checkbox-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}
.checkbox-input {
  width: 18px;
  height: 18px;
  cursor: pointer;
}
.checkbox-label {
  font-size: 0.95rem !important;
  color: var(--text-primary) !important;
  margin-bottom: 0 !important;
  cursor: pointer;
}
.help-text {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-top: 4px;
}
</style>

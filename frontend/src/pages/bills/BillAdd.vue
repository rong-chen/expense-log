<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { billApi } from '@/api'
import { toast } from 'vue-sonner'
import TopNavBar from '@/components/layout/TopNavBar.vue'

const router = useRouter()

const CATEGORIES = ['餐饮', '交通', '购物', '娱乐', '生活缴费', '转账', '医疗', '其他']

function getNowDatetime() {
  const d = new Date()
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const form = ref({
  amount: '',
  merchant: '',
  category: '其他',
  remark: '',
  created_at: getNowDatetime()
})

const saving = ref(false)

function onDateChange() {
  if (!form.value.created_at) {
    form.value.created_at = getNowDatetime()
    toast.success('已自动恢复为当前时间')
  }
}

async function submitBill() {
  const amt = Number(form.value.amount)
  if (!amt || amt <= 0) {
    toast.error('请输入有效金额')
    return
  }
  if (!form.value.merchant.trim()) {
    toast.error('请输入商户名称')
    return
  }

  saving.value = true
  try {
    // 使用 updateBill 的方式不行，需要一个创建接口
    // 我们复用手动创建逻辑：调后端的 manual create
    const res: any = await billApi.createBill({
      amount: amt,
      merchant: form.value.merchant.trim(),
      category: form.value.category,
      remark: form.value.remark.trim(),
      created_at: form.value.created_at
    })
    
    if (res.code === 0) {
      toast.success('账单录入成功！')
      router.back()
    } else {
      toast.error(res.msg || '录入失败')
    }
  } catch (err) {
    toast.error('网络请求失败')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="add-page">
    <TopNavBar title="手动记账" />

    <div class="form-container">
      <div class="card form-card">
        <div class="form-group amount-group">
          <label>金额 (¥)</label>
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
          <label>商户名称</label>
          <input type="text" v-model="form.merchant" class="form-control" placeholder="例如：美团外卖、星巴克" />
        </div>

        <div class="form-group">
          <label>交易分类</label>
          <div class="category-grid">
            <button 
              v-for="c in CATEGORIES" :key="c" 
              :class="['category-chip', { active: form.category === c }]"
              @click="form.category = c"
            >{{ c }}</button>
          </div>
        </div>

        <div class="form-group">
          <label>记账时间</label>
          <input type="datetime-local" v-model="form.created_at" @change="onDateChange" class="form-control" />
        </div>

        <div class="form-group">
          <label>备注记事</label>
          <textarea 
            v-model="form.remark" 
            class="form-control remark-textarea" 
            placeholder="可选，记录一些细节..." 
            rows="3"
          ></textarea>
        </div>
      </div>

      <div class="bottom-actions">
        <button class="btn btn-primary" @click="submitBill" :disabled="saving">
          {{ saving ? '正在录入...' : '确认录入' }}
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
input[type="datetime-local"] {
  text-align: left !important;
}
input[type="datetime-local"]::-webkit-date-and-time-value {
  text-align: left !important;
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
</style>

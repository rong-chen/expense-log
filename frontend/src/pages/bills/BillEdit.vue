<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { billApi, tagApi } from '@/api'
import { toast } from 'vue-sonner'
import TopNavBar from '@/components/layout/TopNavBar.vue'

const route = useRoute()
const router = useRouter()

const billID = route.params.id as string
const loading = ref(true)
const saving = ref(false)
const deleting = ref(false)

const CATEGORIES = ['餐饮', '交通', '购物', '娱乐', '生活缴费', '转账', '医疗', '其他']

const editForm = ref({
  amount: 0,
  merchant: '',
  category: '',
  remark: '',
  created_at: ''
})

async function fetchBillDetail() {
  try {
    const res: any = await billApi.getBillDetail(billID)
    if (res.code === 0) {
      const bill = res.data
      const d = new Date(bill.transaction_date)
      const pad = (n: number) => n.toString().padStart(2, '0')
      const localDatetime = `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
      
      editForm.value = {
        amount: bill.amount,
        merchant: bill.merchant,
        category: bill.category,
        remark: bill.remark || '',
        created_at: localDatetime
      }
      
      // 直接从账单对象中提取聚合的标签（后端已修复，直接支持）
      if (bill.tags && Array.isArray(bill.tags)) {
        selectedTagIDs.value = bill.tags.map((t: any) => t.ID)
      }
    } else {
      toast.error('获取账单失败')
      router.back()
    }
  } catch (e) {
    toast.error('网络请求失败')
    router.back()
  } finally {
    loading.value = false
  }
}

// 标签相关
interface TagItem {
  ID: string
  name: string
  color: string
}
const allTags = ref<TagItem[]>([])
const selectedTagIDs = ref<string[]>([])

async function fetchTags() {
  try {
    const tagsRes: any = await tagApi.list()
    if (tagsRes.code === 0) allTags.value = tagsRes.data || []
  } catch (e) {
    console.error('Failed to load user tags:', e)
  }
}

function toggleTag(tagID: string) {
  const idx = selectedTagIDs.value.indexOf(tagID)
  if (idx >= 0) {
    selectedTagIDs.value.splice(idx, 1)
  } else {
    selectedTagIDs.value.push(tagID)
  }
}

onMounted(() => {
  fetchBillDetail()
  fetchTags()
})

function onDateChange() {
  // 防御 iOS 等系统原生的清除/还原按钮，若文本被置空则自动倒退回当前时间
  if (!editForm.value.created_at) {
    const d = new Date()
    const pad = (n: number) => n.toString().padStart(2, '0')
    editForm.value.created_at = `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`

  }
}

async function confirmSave() {
  if (!editForm.value.created_at) {
    onDateChange()
  }
  saving.value = true
  try {
    const res: any = await billApi.updateBill(billID, {
      amount: Number(editForm.value.amount),
      merchant: editForm.value.merchant,
      category: editForm.value.category,
      remark: editForm.value.remark.trim(),
      created_at: editForm.value.created_at
    })
    
    if (res.code !== 0) {
      toast.error(res.msg || '保存失败')
      return
    }

    // 保存标签关联（独立处理，不影响账单保存结果）
    if (selectedTagIDs.value.length > 0 || allTags.value.length > 0) {
      try {
        const tagRes: any = await tagApi.setBillTags(billID, selectedTagIDs.value)
        if (tagRes?.code !== 0) {
          toast.error('部分标签保存失败')
        }
      } catch (tagErr: any) {
        toast.error('网络错误，标签未保存')
      }
    }

    toast.success('保存成功')
    router.back()
  } catch (err) {
    console.error('[Bill Save Error]', err)
    toast.error('保存请求失败')
  } finally {
    saving.value = false
  }
}

async function confirmDelete() {
  if (!confirm('确定要永久删除这笔账单吗？删除后将不可恢复。')) return

  deleting.value = true
  try {
    const res: any = await billApi.deleteBill(billID)
    if (res.code === 0) {

      router.back()
    } else {
      toast.error(res.msg || '删除失败')
    }
  } catch (err) {
    toast.error('删除请求失败')
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <div class="edit-page">
    <TopNavBar title="详情编辑" />

    <!-- 骨架屏加载状态 -->
    <div v-if="loading" class="skeleton-form">
      <div class="skeleton-item card">
        <div class="skeleton-line skeleton w-100"></div>
        <div class="skeleton-line skeleton w-60"></div>
        <div class="skeleton-line skeleton w-100" style="margin-top:20px;"></div>
        <div class="skeleton-line skeleton w-60"></div>
        <div class="skeleton-line skeleton w-100" style="margin-top:20px;"></div>
      </div>
    </div>

    <div v-else class="form-container">
      <div class="card form-card">
        <div class="form-group amount-group">
          <label>金额 (¥)</label>
          <input type="number" step="0.01" v-model="editForm.amount" class="form-control amount-input" />
        </div>
        
        <div class="form-group">
          <label>商户名称</label>
          <input type="text" v-model="editForm.merchant" class="form-control" placeholder="未识别商户" />
        </div>

        <div class="form-group">
          <label>交易分类</label>
          <select v-model="editForm.category" class="form-control">
            <option v-for="c in CATEGORIES" :key="c" :value="c">{{ c }}</option>
          </select>
        </div>

        <div class="form-group">
          <label>记账时间</label>
          <input type="datetime-local" v-model="editForm.created_at" @change="onDateChange" class="form-control" />
        </div>

        <div class="form-group">
          <label>标签</label>
          <div class="tag-selector" v-if="allTags.length > 0">
            <button
              v-for="tag in allTags" :key="tag.ID"
              :class="['tag-chip', { active: selectedTagIDs.includes(tag.ID) }]"
              :style="selectedTagIDs.includes(tag.ID) ? { background: tag.color, borderColor: tag.color, color: '#fff' } : { borderColor: tag.color + '40', color: tag.color }"
              @click="toggleTag(tag.ID)"
            >{{ tag.name }}</button>
          </div>
          <div v-else class="tag-empty-hint">暂无标签，可在「我的 → 标签管理」中创建</div>
        </div>

        <div class="form-group">
          <label>备注记事</label>
          <textarea 
            v-model="editForm.remark" 
            class="form-control remark-textarea" 
            placeholder="支持多行记事..." 
            rows="4"
          ></textarea>
        </div>
      </div>

      <div class="bottom-actions">
        <button class="btn btn-primary" @click="confirmSave" :disabled="saving || deleting">
          {{ saving ? '保存中...' : '保存并返回' }}
        </button>
        <button class="btn btn-danger" @click="confirmDelete" :disabled="saving || deleting">
          {{ deleting ? '删除中...' : '永久删除此账单' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.edit-page {
  padding: 16px;
  padding-top: calc(72px + env(safe-area-inset-top));
  max-width: 600px;
  margin: 0 auto;
  min-height: 100vh;
  background: var(--bg-base);
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
  text-align: left;
}
input[type="datetime-local"] {
  text-align: left !important;
  /* 不要加 display: block/flex，破坏原生控件宽度结算 */
}
input[type="datetime-local"]::-webkit-datetime-edit {
  text-align: left !important;
  padding: 0 !important;
}
input[type="datetime-local"]::-webkit-date-and-time-value {
  /* iOS 15+ 特供原生时间文字层强制左对齐 */
  text-align: left !important;
  justify-content: flex-start !important;
}
input[type="datetime-local"]::-webkit-datetime-edit-fields-wrapper {
  display: flex;
  justify-content: flex-start;
  padding: 0 !important;
  width: 100%;
}
.form-control:focus {
  border-color: var(--primary);
  background: white;
  box-shadow: 0 4px 12px rgba(26, 188, 156, 0.1);
}

.remark-textarea {
  resize: none;
}

.tag-selector {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.tag-chip {
  padding: 6px 14px;
  border-radius: 20px;
  border: 1.5px solid;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  background: transparent;
}
.tag-chip.active {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}
.tag-chip:active {
  transform: scale(0.95);
}
.tag-empty-hint {
  font-size: 0.85rem;
  color: var(--text-secondary);
  opacity: 0.7;
}

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

.btn-danger {
  background: white;
  color: #dc3545;
  border: 1px solid rgba(220, 53, 69, 0.2);
}
.btn-danger:active {
  background: rgba(220, 53, 69, 0.05);
}

.skeleton-form { display: flex; flex-direction: column; gap: 20px; padding: 20px 0; }
.skeleton-item { padding: 30px 20px; border-radius: 20px; display: flex; flex-direction: column; gap: 12px; }
.skeleton-line { height: 20px; border-radius: 10px; }
.w-100 { width: 100%; }
.w-60 { width: 60%; }
</style>

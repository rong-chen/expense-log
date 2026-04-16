<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ledgerApi } from '@/api'
import TopNavBar from '@/components/layout/TopNavBar.vue'
import { toast } from 'vue-sonner'
import { useLedgerStore } from '@/stores/ledger'

const router = useRouter()
const ledgerStore = useLedgerStore()
const form = ref({
  name: '',
  description: ''
})
const loading = ref(false)

async function submit() {
  if (!form.value.name.trim()) {
    toast.error('请输入账本名称')
    return
  }
  loading.value = true
  try {
    const res: any = await ledgerApi.create(form.value)
    if (res.code === 0) {
      toast.success('账本创建成功！')
      await ledgerStore.fetchLedgers()
      ledgerStore.setCurrentLedger(res.data.ID)
      router.back()
    } else {
      toast.error(res.msg || '创建失败')
    }
  } catch (e: any) {
    toast.error('网络异常')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page-container">
    <TopNavBar title="创建共享账本" />
    <div class="content-body">
      <div class="header-desc">
        <h2>新建家庭/团队账本</h2>
        <p>创建后，您可以邀请家人或朋友一同记账。</p>
      </div>
      
      <div class="form-group">
        <label>账本名称</label>
        <input v-model="form.name" type="text" placeholder="例如：2026年广州家庭账本" class="custom-input" />
      </div>

      <div class="form-group">
        <label>账本描述 (选填)</label>
        <textarea v-model="form.description" rows="3" placeholder="描述一下这个账本的用途..." class="custom-input"></textarea>
      </div>

      <button class="primary-btn submit-btn" :disabled="loading" @click="submit">
        {{ loading ? '创建中...' : '立即创建' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.page-container {
  min-height: 100vh;
  background: var(--background);
}
.content-body {
  padding: 80px 20px 20px;
}
.header-desc {
  margin-bottom: 32px;
}
.header-desc h2 {
  font-size: 1.5rem;
  margin-bottom: 8px;
}
.header-desc p {
  color: var(--text-secondary);
  font-size: 0.95rem;
}
.form-group {
  margin-bottom: 24px;
}
.form-group label {
  display: block;
  font-weight: 500;
  margin-bottom: 8px;
}
.custom-input {
  width: 100%;
  box-sizing: border-box;
  padding: 14px 16px;
  border-radius: 12px;
  border: 1px solid rgba(0,0,0,0.1);
  background: var(--surface);
  font-size: 1rem;
  transition: all 0.2s;
  outline: none;
}
.custom-input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(26,188,156,0.1);
}
.submit-btn {
  width: 100%;
  margin-top: 20px;
  padding: 14px;
  font-size: 1.1rem;
}
</style>

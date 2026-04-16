<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ledgerApi } from '@/api'
import TopNavBar from '@/components/layout/TopNavBar.vue'
import { toast } from 'vue-sonner'
import { useLedgerStore } from '@/stores/ledger'

const router = useRouter()
const ledgerStore = useLedgerStore()
const inviteCode = ref('')
const loading = ref(false)

async function submit() {
  if (inviteCode.value.trim().length !== 6) {
    toast.error('请输入有效的 6 位邀请码')
    return
  }
  loading.value = true
  try {
    const res: any = await ledgerApi.join({ invite_code: inviteCode.value.toUpperCase() })
    if (res.code === 0) {
      toast.success('加入账本成功！')
      await ledgerStore.fetchLedgers()
      ledgerStore.setCurrentLedger(res.data.ID)
      router.back()
    } else {
      toast.error(res.msg || '加入失败')
    }
  } catch (e: any) {
    toast.error('网络异常或加入失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page-container">
    <TopNavBar title="加入账本" />
    <div class="content-body">
      <div class="header-desc">
        <h2>输入邀请码</h2>
        <p>向账本创建者索要 6 位字母数字构成的邀请码以加入共享。</p>
      </div>
      
      <div class="code-input-container">
        <input 
          v-model="inviteCode" 
          type="text" 
          maxlength="6" 
          placeholder="XXXXXX" 
          class="code-input"
          style="text-transform: uppercase;"
        />
      </div>

      <button class="primary-btn submit-btn" :disabled="loading || inviteCode.length !== 6" @click="submit">
        {{ loading ? '验证中...' : '确认加入' }}
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
  text-align: center;
}
.header-desc h2 {
  font-size: 1.5rem;
  margin-bottom: 12px;
}
.header-desc p {
  color: var(--text-secondary);
  font-size: 0.95rem;
}
.code-input-container {
  margin-bottom: 32px;
  display: flex;
  justify-content: center;
}
.code-input {
  width: 100%;
  max-width: 280px;
  text-align: center;
  font-size: 2rem;
  font-weight: bold;
  letter-spacing: 8px;
  padding: 16px;
  border-radius: 16px;
  border: 2px solid rgba(0,0,0,0.1);
  background: var(--surface);
  transition: all 0.2s;
  outline: none;
}
.code-input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 4px rgba(26,188,156,0.1);
}
.submit-btn {
  width: 100%;
  padding: 14px;
  font-size: 1.1rem;
}
.submit-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>

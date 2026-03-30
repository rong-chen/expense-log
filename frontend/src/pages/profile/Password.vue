<script setup lang="ts">
import TopNavBar from '@/components/layout/TopNavBar.vue'
import { ref } from 'vue'
import { toast } from 'vue-sonner'

import { authApi } from '@/api'
import { useRouter } from 'vue-router'

const oldPassword = ref('')
const newPassword = ref('')
const loading = ref(false)
const router = useRouter()

async function submit() {
  if (!oldPassword.value || !newPassword.value) {
    toast.error('请填写完整密码')
    return
  }
  loading.value = true
  try {
    const res: any = await authApi.updatePassword({
      old_password: oldPassword.value,
      new_password: newPassword.value
    })
    if (res.code === 0) {

      oldPassword.value = ''
      newPassword.value = ''
      setTimeout(() => router.back(), 1500)
    } else {
      toast.error(res.message || '修改失败，原密码可能错误')
    }
  } catch (err: any) {
    toast.error(err.message || '网络或服务器错误')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="password-page">
    <TopNavBar title="修改密码" />
    <div class="page-content">
      <div class="card form-card">
        <div class="form-group">
          <label>原密码</label>
          <input type="password" v-model="oldPassword" class="form-input" placeholder="请输入当前密码" />
        </div>
        <div class="form-group">
          <label>新密码</label>
          <input type="password" v-model="newPassword" class="form-input" placeholder="请输入新密码" />
        </div>
        <button class="btn btn-primary w-100 mt-4" :disabled="loading" @click="submit">
          {{ loading ? '修改中...' : '确认修改' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.password-page {
  background: var(--bg-base);
  min-height: 100vh;
}
.page-content {
  padding: 16px;
  padding-top: calc(70px + env(safe-area-inset-top));
}
.form-card {
  padding: 24px;
}
.form-group {
  margin-bottom: 20px;
}
.form-group label {
  display: block;
  font-size: 0.85rem;
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--text-secondary);
}
.form-input {
  width: 100%;
  padding: 14px 16px;
  font-size: 1rem;
  border: 1.5px solid var(--border, rgba(0,0,0,0.08));
  border-radius: 12px;
  background: var(--bg-base, #faf8f5);
  color: var(--text-primary);
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
  -webkit-appearance: none;
}
.form-input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-soft);
}
.form-input::placeholder {
  color: var(--text-muted, #bbb);
}
.mt-4 {
  margin-top: 24px;
}
</style>

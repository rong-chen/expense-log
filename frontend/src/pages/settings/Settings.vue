<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { emailApi } from '@/api'
import { useAuthStore } from '@/stores/auth'
import TopNavBar from '@/components/layout/TopNavBar.vue'
import {
  Mail, Plus, Trash2, MailCheck, CheckCircle2, Sparkles, AlertTriangle
} from 'lucide-vue-next'

const auth = useAuthStore()

interface EmailAccount {
  id: string
  host: string
  username: string
  port: number
  tls: boolean
  folder: string
  enabled: boolean
}

const accounts = ref<EmailAccount[]>([])
const showBindForm = ref(false)
const loading = ref(false)
const error = ref('')
const success = ref('')

const form = ref({
  host: 'imap.qq.com',
  port: 993,
  username: '',
  password: '',
  tls: true,
  folder: 'INBOX',
})

const showAdvanced = ref(false)
const helpObject = ref<{url: string, title: string} | null>(null)

watch(() => form.value.username, (val) => {
  if (!val.includes('@')) {
    helpObject.value = null
    return
  }
  const domain = val.split('@')[1].toLowerCase()
  
  const rules: Record<string, {host: string, port: number}> = {
    'qq.com': { host: 'imap.qq.com', port: 993 },
    'foxmail.com': { host: 'imap.qq.com', port: 993 },
    '163.com': { host: 'imap.163.com', port: 993 },
    '126.com': { host: 'imap.126.com', port: 993 },
    'gmail.com': { host: 'imap.gmail.com', port: 993 },
    'outlook.com': { host: 'outlook.office365.com', port: 993 },
    'hotmail.com': { host: 'outlook.office365.com', port: 993 },
    'icloud.com': { host: 'imap.mail.me.com', port: 993 },
  }

  const helps: Record<string, {url: string, title: string}> = {
    'qq.com': { url: 'https://service.mail.qq.com/detail/0/75', title: '教程：如何获取 QQ 邮箱授权码？' },
    'foxmail.com': { url: 'https://service.mail.qq.com/detail/0/75', title: '教程：如何获取 Foxmail 授权码？' },
    '163.com': { url: 'https://help.mail.163.com/faqDetail.do?code=d7c39a39f1c30ce2e4b4006c9f323a23', title: '教程：如何获取 163 邮箱授权码？' },
    '126.com': { url: 'https://help.mail.126.com/faqDetail.do?code=d7c39a39f1c30ce2df584fbbc9d59218', title: '教程：如何获取 126 邮箱授权码？' },
    'gmail.com': { url: 'https://support.google.com/mail/answer/185833?hl=zh-Hans', title: '教程：如何生成 Gmail 应用专用密码？' },
    'outlook.com': { url: 'https://support.microsoft.com/zh-cn/account-billing/%E4%BD%BF%E7%94%A8%E5%BA%94%E7%94%A8%E5%AF%86%E7%A0%81%E7%99%BB%E5%BD%95-5896ed9b-4263-e681-128a-a6f2979a7944', title: '教程：如何生成 Outlook 应用密码？' },
  }

  if (rules[domain]) {
    form.value.host = rules[domain].host
    form.value.port = rules[domain].port
    helpObject.value = helps[domain] || null
    // 若识别成功，建议收起高级设置
    if (showAdvanced.value && !form.value.password) {
      showAdvanced.value = false
    }
  } else {
    // 无法识别，自动展开高级设置让用户手动填
    showAdvanced.value = true
  }
})

async function fetchAccounts() {
  try {
    const res: any = await emailApi.getAccounts()
    if (res.code === 0) {
      accounts.value = res.data || []
    }
  } catch { /* ignore */ }
}

async function handleBind() {
  if (!form.value.username || !form.value.password) {
    error.value = '请填写邮箱和授权码'
    return
  }
  loading.value = true
  error.value = ''
  success.value = ''
  try {
    const res: any = await emailApi.bind(form.value)
    if (res.code === 0) {
      success.value = '邮箱绑定成功！'
      showBindForm.value = false
      form.value.username = ''
      form.value.password = ''
      await fetchAccounts()
    } else {
      error.value = res.message || '绑定失败'
    }
  } catch (err: any) {
    error.value = err.message || '网络或连接错误'
  } finally {
    loading.value = false
  }
}

async function handleDelete(id: string) {
  if (!confirm('确定要解绑这个邮箱吗？')) return
  try {
    await emailApi.deleteAccount(id)
    await fetchAccounts()
  } catch { /* ignore */ }
}

onMounted(() => {
  fetchAccounts()
  if (!auth.user) auth.fetchUserInfo()
})
</script>

<template>
  <div class="settings-page">
    <TopNavBar title="邮箱记账配置" />

    <div class="page-content">
      <!-- 邮箱管理 -->
    <div class="section" style="animation-delay: 0.2s">
      <div class="section-header">
        <h2 class="section-title">
          <Mail :size="18" class="section-icon" />
          邮箱管理
        </h2>
        <button class="btn btn-primary" @click="showBindForm = !showBindForm" id="bind-email-btn">
          <Plus v-if="!showBindForm" :size="16" />
          {{ showBindForm ? '取消' : '绑定邮箱' }}
        </button>
      </div>

      <!-- 绑定表单 -->
      <div class="card bind-form" v-if="showBindForm">
        <div class="form-grid">
          <div class="input-group full-width">
            <label>邮箱地址</label>
            <input class="input" type="email" v-model="form.username" placeholder="输入邮箱后自动识别服务器配置" id="email-input" />
          </div>
          <div class="input-group full-width">
            <label>授权码（App 专用）</label>
            <p class="field-hint">
              <AlertTriangle :size="12" class="icon-inline" />
              请根据下方教程生成并填入“授权码/应用密码”。
            </p>
            <input class="input" type="password" v-model="form.password" placeholder="填入生成的专属授权码" id="email-password" />
            <div style="min-height: 28px; margin-top: 6px;">
              <transition name="fade">
                <a v-if="helpObject" :href="helpObject.url" target="_blank" class="help-link" style="margin-top: 0;">
                  <Sparkles :size="14" class="icon-inline" /> 
                  {{ helpObject.title }}
                </a>
              </transition>
            </div>
          </div>
        </div>

        <div class="advanced-toggle" @click="showAdvanced = !showAdvanced">
          <span class="auto-hint" v-if="!showAdvanced && form.host">
            <CheckCircle2 :size="14" class="icon-inline" /> 
            已识别服务器: {{ form.host }}
          </span>
          <span class="toggle-text">{{ showAdvanced ? '收起配置' : '手动配置 IMAP (高级)' }}</span>
        </div>

        <transition name="fade">
          <div class="form-grid advanced-fields" v-if="showAdvanced">
            <div class="input-group">
              <label>IMAP 服务器</label>
              <input class="input" v-model="form.host" placeholder="imap.exmail.qq.com" />
            </div>
            <div class="input-group">
              <label>端口</label>
              <input class="input" type="number" v-model.number="form.port" />
            </div>
          </div>
        </transition>

        <p class="error-msg" v-if="error">{{ error }}</p>
        <p class="success-msg" v-if="success">{{ success }}</p>

        <button class="btn btn-primary" :disabled="loading" @click="handleBind" id="bind-submit">
          {{ loading ? '验证连接中...' : '绑定邮箱' }}
        </button>
      </div>

      <!-- 已绑定列表 -->
      <div v-if="accounts.length" class="accounts-list">
        <div class="card account-item" v-for="acc in accounts" :key="acc.id">
          <div class="account-info">
            <div class="account-icon-wrap">
              <MailCheck :size="20" />
            </div>
            <div>
              <span class="account-email">{{ acc.username }}</span>
              <span class="account-host">{{ acc.host }}:{{ acc.port }}</span>
            </div>
          </div>
          <div class="account-actions">
            <span class="badge badge-success" v-if="acc.enabled">已启用</span>
            <button class="btn btn-danger btn-sm" @click="handleDelete(acc.id)">
              <Trash2 :size="14" />
              解绑
            </button>
          </div>
        </div>
      </div>

      <div class="card empty-inline" v-else-if="!showBindForm">
        <p>暂未绑定邮箱，绑定后可自动拉取发票和交易通知</p>
      </div>
    </div>
    </div>
  </div>
</template>

<style scoped>
.settings-page {
  background: var(--bg-body);
  min-height: 100vh;
}
.page-content {
  padding: 16px;
  padding-top: calc(70px + env(safe-area-inset-top));
}
.section {
  margin-bottom: 32px;
}
.section-title {
  font-size: 1.1rem;
  font-weight: 600;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.section-icon {
  color: var(--primary);
}
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.icon-inline {
  vertical-align: middle;
}

.bind-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 16px;
}
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}
.full-width {
  grid-column: 1 / -1;
}
.advanced-toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 8px;
  margin-bottom: 4px;
}
.auto-hint {
  font-size: 0.8rem;
  color: var(--success);
  display: flex;
  align-items: center;
  gap: 4px;
}
.toggle-text {
  font-size: 0.8rem;
  color: var(--primary);
  cursor: pointer;
  text-decoration: underline;
  text-underline-offset: 2px;
}
.advanced-fields {
  background: var(--bg-hover);
  padding: 16px;
  border-radius: 12px;
  border: 1px dashed var(--border-light);
}
.field-hint {
  font-size: 0.75rem;
  color: var(--warning-text, #d97706);
  margin-top: -6px;
  margin-bottom: 4px;
  line-height: 1.4;
  display: flex;
  align-items: center;
  gap: 4px;
}
.help-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
  font-size: 0.8rem;
  color: var(--primary);
  text-decoration: none;
  background: var(--primary-soft);
  padding: 4px 10px;
  border-radius: 6px;
  transition: all 0.2s ease;
}
.help-link:hover {
  background: var(--primary);
  color: white;
}
.error-msg {
  color: var(--danger); font-size: 0.85rem;
  padding: 8px 12px; background: var(--danger-soft); border-radius: 8px;
}
.success-msg {
  color: var(--success); font-size: 0.85rem;
  padding: 8px 12px; background: var(--success-soft); border-radius: 8px;
}

.accounts-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.account-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
}
.account-info {
  display: flex;
  align-items: center;
  gap: 12px;
}
.account-icon-wrap {
  width: 40px; height: 40px; border-radius: 12px;
  background: var(--primary-soft); color: var(--primary);
  display: flex; align-items: center; justify-content: center;
}
.account-email { font-weight: 600; display: block; }
.account-host { font-size: 0.8rem; color: var(--text-tertiary); }
.account-actions { display: flex; align-items: center; gap: 10px; }
.btn-sm { padding: 6px 14px; font-size: 0.8rem; gap: 6px; }

.empty-inline {
  text-align: center;
  padding: 32px;
  color: var(--text-secondary);
}

@media (max-width: 600px) {
  .form-grid { grid-template-columns: 1fr; }
}
</style>

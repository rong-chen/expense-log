<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useLedgerStore } from '@/stores/ledger'
import { Check, Plus, ScanLine, ChevronRight, Layout, Users } from 'lucide-vue-next'
import TopNavBar from '@/components/layout/TopNavBar.vue'
import { toast } from 'vue-sonner'

const router = useRouter()
const ledgerStore = useLedgerStore()

function selectLedger(id: string) {
  ledgerStore.setCurrentLedger(id)
  toast.success('已切换账本')
  router.back()
}

async function copyInviteCode(code: string) {
  try {
    await navigator.clipboard.writeText(code)
    toast.success('邀请码已复制')
  } catch (err) {
    toast.error('复制失败')
  }
}
</script>

<template>
  <div class="ledger-select-page">
    <TopNavBar title="选择账本" />

    <div class="content">
      <div class="section-label">当前空间</div>
      
      <div class="ledger-list">
        <div 
          v-for="l in ledgerStore.allLedgers" 
          :key="l.ID"
          class="ledger-card"
          :class="{ active: l.ID === ledgerStore.currentLedgerId }"
          @click="selectLedger(l.ID)"
        >
          <div class="ledger-main">
            <div class="ledger-icon" :class="l.type">
              <Layout v-if="l.type === 'personal'" :size="20" />
              <Users v-else :size="20" />
            </div>
            <div class="ledger-info">
              <div class="ledger-top">
                <span class="ledger-name">{{ l.name }}</span>
                <span class="type-tag" :class="l.type">
                  {{ l.type === 'personal' ? '私有' : '共享' }}
                </span>
              </div>
              <p class="ledger-desc">{{ l.description || '暂无描述' }}</p>
              
              <div v-if="l.invite_code" class="invite-section" @click.stop="copyInviteCode(l.invite_code)">
                <span class="invite-label">邀请码:</span>
                <span class="invite-code">{{ l.invite_code }}</span>
                <span class="copy-tip">(点击复制)</span>
              </div>
            </div>
          </div>
          <div class="selection-indicator">
            <div v-if="l.ID === ledgerStore.currentLedgerId" class="check-circle">
              <Check :size="16" stroke-width="3" />
            </div>
            <ChevronRight v-else :size="20" class="arrow-icon" />
          </div>
        </div>
      </div>

      <div class="actions-section">
        <button class="action-btn primary" @click="router.push('/ledger/create')">
          <Plus :size="20" />
          <span>创建新账本</span>
        </button>
        <button class="action-btn secondary" @click="router.push('/ledger/join')">
          <ScanLine :size="20" />
          <span>加入已有账本</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ledger-select-page {
  min-height: 100vh;
  background: var(--bg-base);
  padding-bottom: 40px;
}

.content {
  padding: 84px 16px 24px;
  max-width: 600px;
  margin: 0 auto;
}

.section-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 12px;
  padding-left: 4px;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.ledger-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 32px;
}

.ledger-card {
  background: var(--bg-card);
  border: 1.5px solid var(--border-light);
  border-radius: 20px;
  padding: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  cursor: pointer;
  position: relative;
  overflow: hidden;
}

.ledger-card:active {
  transform: scale(0.98);
}

.ledger-card.active {
  border-color: var(--primary);
  background: var(--primary-soft);
  box-shadow: 0 4px 16px rgba(230, 126, 34, 0.08);
}

.ledger-main {
  display: flex;
  gap: 16px;
  align-items: center;
  flex: 1;
}

.ledger-icon {
  width: 48px;
  height: 48px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.ledger-icon.personal {
  background: rgba(52, 152, 219, 0.1);
  color: #3498db;
}

.ledger-icon.shared {
  background: rgba(230, 126, 34, 0.1);
  color: var(--primary);
}

.ledger-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.ledger-top {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ledger-name {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--text-primary);
}

.type-tag {
  font-size: 0.65rem;
  font-weight: 800;
  padding: 2px 6px;
  border-radius: 6px;
  text-transform: uppercase;
}

.type-tag.personal {
  background: #dbeafe;
  color: #1e40af;
}

.type-tag.shared {
  background: #ffedd5;
  color: #9a3412;
}

.ledger-desc {
  font-size: 0.8rem;
  color: var(--text-secondary);
  line-height: 1.4;
}

.invite-section {
  margin-top: 6px;
  font-size: 0.75rem;
  display: flex;
  align-items: center;
  gap: 4px;
  background: rgba(0, 0, 0, 0.03);
  padding: 4px 8px;
  border-radius: 6px;
  width: fit-content;
}

.invite-label {
  color: var(--text-secondary);
}

.invite-code {
  font-weight: 700;
  color: var(--primary);
  letter-spacing: 0.5px;
}

.copy-tip {
  color: var(--text-tertiary);
  font-size: 0.65rem;
}

.selection-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
}

.check-circle {
  width: 24px;
  height: 24px;
  background: var(--primary);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(230, 126, 34, 0.3);
}

.arrow-icon {
  color: var(--text-tertiary);
}

.actions-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.action-btn {
  width: 100%;
  padding: 16px;
  border-radius: 18px;
  border: none;
  font-size: 1rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn.primary {
  background: linear-gradient(135deg, var(--primary), var(--primary-light));
  color: white;
  box-shadow: 0 4px 12px rgba(230, 126, 34, 0.2);
}

.action-btn.secondary {
  background: var(--bg-card);
  color: var(--text-primary);
  border: 1.5px solid var(--border);
}

.action-btn:active {
  transform: scale(0.97);
}
</style>

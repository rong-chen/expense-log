import { defineStore } from 'pinia'
import { ledgerApi } from '@/api'

export interface Ledger {
  ID: string
  name: string
  description: string
  owner_id: string
  type: string
  invite_code: string
}

export const useLedgerStore = defineStore('ledger', {
  state: () => ({
    ledgers: [] as Ledger[],
    currentLedgerId: localStorage.getItem('currentLedgerId') || '',
    loading: false
  }),
  actions: {
    async fetchLedgers() {
      this.loading = true
      try {
        const res: any = await ledgerApi.list()
        if (res.code === 0 && res.data) {
          this.ledgers = res.data
          // 如果当前没有选定，或者是无效的账本，选中列表第一个（通常是个人默认账本）
          if (!this.currentLedgerId || !this.ledgers.find(l => l.ID === this.currentLedgerId)) {
            if (this.ledgers.length > 0) {
              this.setCurrentLedger(this.ledgers[0].ID)
            }
          }
        }
      } catch (e) {
        console.error('获取账本失败:', e)
      } finally {
        this.loading = false
      }
    },
    setCurrentLedger(id: string) {
      if (id !== this.currentLedgerId) {
        this.currentLedgerId = id
        localStorage.setItem('currentLedgerId', id)
        // 触发重新加载页面的逻辑，或者使用 EventEmitter 让其他组件刷新
        window.dispatchEvent(new Event('ledger-changed'))
      }
    }
  },
  getters: {
    currentLedger(state): Ledger | undefined {
      return state.ledgers.find(l => l.ID === state.currentLedgerId)
    }
  }
})

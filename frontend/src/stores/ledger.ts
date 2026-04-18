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
          // 只有当存了一个非空 ID 且该 ID 已经在物理列表中失效时，才进行回退
          if (this.currentLedgerId !== '' && !this.ledgers.find(l => l.ID === this.currentLedgerId)) {
            if (this.ledgers.length > 0) {
              this.setCurrentLedger(this.ledgers[0].ID)
            } else {
              this.setCurrentLedger('') 
            }
          }
        } else {
          // 接口都没返回数据说明没有任何账本
          this.ledgers = []
          if (this.currentLedgerId) {
            this.setCurrentLedger('')
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
    // 聚合所有物理账本和虚拟的“个人私有空间”
    allLedgers(state): Ledger[] {
      const personalVirtual: Ledger = {
        ID: '',
        name: '个人主账本',
        description: '您的私有记账空间，不与他人共享',
        owner_id: '',
        type: 'personal',
        invite_code: ''
      }
      return [personalVirtual, ...state.ledgers]
    },
    currentLedger(state): Ledger | undefined {
      if (!state.currentLedgerId) {
        return {
          ID: '',
          name: '个人主账本',
          description: '您的私有记账空间',
          owner_id: '',
          type: 'personal',
          invite_code: ''
        }
      }
      return state.ledgers.find(l => l.ID === state.currentLedgerId)
    }
  }
})

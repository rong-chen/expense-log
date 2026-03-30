import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  const isDark = ref(false)

  function toggle() {
    // 禁用暗色模式切换
  }

  // 强制亮色
  document.documentElement.classList.remove('dark')
  localStorage.setItem('theme', 'light')

  return { isDark, toggle }
})

import { computed } from 'vue'
import { useAppStore } from '@/store/modules/app'

export function useTheme() {
  const appStore = useAppStore()

  const isDark = computed(() => appStore.isDark)
  const theme = computed(() => appStore.theme)

  function toggleTheme() {
    appStore.toggleTheme()
  }

  function setTheme(theme: 'light' | 'dark') {
    appStore.setTheme(theme)
  }

  return {
    isDark,
    theme,
    toggleTheme,
    setTheme,
  }
}

import { defineStore } from 'pinia'

interface AppState {
  sidebar: {
    opened: boolean
    withoutAnimation: boolean
  }
  device: string
  theme: 'light' | 'dark'
}

const STORAGE_KEY = 'go-admin-theme'

function getInitialTheme(): 'light' | 'dark' {
  const saved = localStorage.getItem(STORAGE_KEY) as 'light' | 'dark' | null
  return saved || 'light'
}

// Sync theme to DOM on store init - uses Element Plus dark mode (class="dark" on <html>)
const initialTheme = getInitialTheme()
if (initialTheme === 'dark') {
  document.documentElement.classList.add('dark')
} else {
  document.documentElement.classList.remove('dark')
}

export const useAppStore = defineStore('app', {
  state: (): AppState => ({
    sidebar: {
      opened: true,
      withoutAnimation: false,
    },
    device: 'desktop',
    theme: initialTheme,
  }),

  getters: {
    isDark: (state) => state.theme === 'dark',
  },

  actions: {
    toggleSidebar() {
      this.sidebar.opened = !this.sidebar.opened
      this.sidebar.withoutAnimation = false
    },

    closeSidebar(withoutAnimation: boolean) {
      this.sidebar.opened = false
      this.sidebar.withoutAnimation = withoutAnimation
    },

    toggleDevice(device: string) {
      this.device = device
    },

    _syncTheme() {
      if (this.theme === 'dark') {
        document.documentElement.classList.add('dark')
      } else {
        document.documentElement.classList.remove('dark')
      }
      localStorage.setItem(STORAGE_KEY, this.theme)
    },

    toggleTheme() {
      this.theme = this.theme === 'light' ? 'dark' : 'light'
      this._syncTheme()
    },

    setTheme(theme: 'light' | 'dark') {
      this.theme = theme
      this._syncTheme()
    },
  },
})

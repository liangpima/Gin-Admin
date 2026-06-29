import { ref, onMounted, onUnmounted } from 'vue'

const width = ref(window.innerWidth)
const isMobile = ref(width.value < 768)
const isTablet = ref(width.value >= 768 && width.value < 1024)
const isDesktop = ref(width.value >= 1024)

let listenerCount = 0

function update() {
  width.value = window.innerWidth
  isMobile.value = width.value < 768
  isTablet.value = width.value >= 768 && width.value < 1024
  isDesktop.value = width.value >= 1024
}

export function useResponsive() {
  onMounted(() => {
    if (listenerCount === 0) {
      window.addEventListener('resize', update)
    }
    listenerCount++
  })

  onUnmounted(() => {
    listenerCount--
    if (listenerCount <= 0) {
      window.removeEventListener('resize', update)
      listenerCount = 0
    }
  })

  return { width, isMobile, isTablet, isDesktop }
}

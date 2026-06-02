import { ref, onMounted, onUnmounted, computed } from 'vue'

export function useMobile() {
  const windowWidth = ref(window.innerWidth)
  const isMobile = computed(() => windowWidth.value <= 768)
  const isTablet = computed(() => windowWidth.value > 768 && windowWidth.value <= 1024)
  const isDesktop = computed(() => windowWidth.value > 1024)

  const handleResize = () => {
    windowWidth.value = window.innerWidth
  }

  onMounted(() => {
    window.addEventListener('resize', handleResize)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', handleResize)
  })

  return {
    windowWidth,
    isMobile,
    isTablet,
    isDesktop
  }
}
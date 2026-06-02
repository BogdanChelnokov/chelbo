import { ref, onMounted, onUnmounted } from 'vue'

export function useKeyboard() {
  const keyboardHeight = ref(0)
  const isKeyboardOpen = ref(false)

  const handleResize = () => {
    // На мобильных устройствах при открытии клавиатуры окно сжимается
    const viewportHeight = window.visualViewport?.height || window.innerHeight
    const windowHeight = window.innerHeight
    const diff = windowHeight - viewportHeight
    
    if (diff > 150) { // Клавиатура открыта
      keyboardHeight.value = diff
      isKeyboardOpen.value = true
    } else {
      keyboardHeight.value = 0
      isKeyboardOpen.value = false
    }
  }

  onMounted(() => {
    if ('visualViewport' in window) {
      window.visualViewport?.addEventListener('resize', handleResize)
    }
    window.addEventListener('resize', handleResize)
  })

  onUnmounted(() => {
    if ('visualViewport' in window) {
      window.visualViewport?.removeEventListener('resize', handleResize)
    }
    window.removeEventListener('resize', handleResize)
  })

  return {
    keyboardHeight,
    isKeyboardOpen
  }
}
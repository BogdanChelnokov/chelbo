<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from './stores/authStore'

const authStore = useAuthStore()

onMounted(() => {
  authStore.checkAuth()
})
</script>

<style>
/* Добавь в глобальные стили */
.highlight-message {
  animation: highlight 1s ease;
}

@keyframes highlight {
  0% {
    background: transparent;
  }
  50% {
    background: var(--primary-light);
  }
  100% {
    background: transparent;
  }
}
</style>
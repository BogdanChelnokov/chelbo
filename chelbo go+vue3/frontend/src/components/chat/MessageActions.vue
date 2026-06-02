<template>
  <div class="message-actions" :class="{ 'show': isOpen }">
    <div class="actions-menu">
      <button class="action-btn" @click="handleReply">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="3 15 9 9 15 15" />
          <line x1="21" y1="9" x2="21" y2="21" />
        </svg>
        Ответить
      </button>
      
      <button class="action-btn" @click="handleForward">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="13 3 22 12 13 21" />
          <polyline points="2 3 11 12 2 21" />
        </svg>
        Переслать
      </button>
      
      <button class="action-btn danger" @click="handleDelete">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="3 6 5 6 21 6" />
          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
          <line x1="10" y1="11" x2="10" y2="17" />
          <line x1="14" y1="11" x2="14" y2="17" />
        </svg>
        Удалить
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const emit = defineEmits(['reply', 'forward', 'delete'])

const isOpen = ref(false)

function handleReply() {
  emit('reply')
  isOpen.value = false
}

function handleForward() {
  emit('forward')
  isOpen.value = false
}

function handleDelete() {
  if (confirm('Удалить сообщение?')) {
    emit('delete')
  }
  isOpen.value = false
}

defineExpose({ isOpen })
</script>

<style scoped>
.message-actions {
  position: absolute;
  top: 50%;
  right: 10px;
  transform: translateY(-50%);
  z-index: 10;
  opacity: 0;
  visibility: hidden;
  transition: all 0.2s;
}

.message-actions.show {
  opacity: 1;
  visibility: visible;
}

.actions-menu {
  background: var(--bg-primary);
  border-radius: 12px;
  box-shadow: var(--shadow-lg);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-width: 140px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  color: var(--text-primary);
  transition: background 0.2s;
  text-align: left;
}

.action-btn:hover {
  background: var(--bg-hover);
}

.action-btn.danger {
  color: var(--danger-color);
}

.action-btn.danger:hover {
  background: #FEE2E2;
}

@media (max-width: 768px) {
  .actions-menu {
    min-width: 120px;
  }
  
  .action-btn {
    padding: 8px 12px;
    font-size: 13px;
  }
}
</style>
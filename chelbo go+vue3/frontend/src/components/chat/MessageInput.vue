<template>
  <div class="message-input-container">
    <div v-if="replyTo" class="reply-indicator">
      <div class="reply-info">
        <div class="reply-label">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 15 9 9 15 15" />
            <line x1="21" y1="9" x2="21" y2="21" />
          </svg>
          <span>Ответ {{ replyTo.sender_id === authStore.user?.id ? 'себе' : getSenderName() }}</span>
        </div>
        <span class="reply-text">{{ truncateText(replyTo.text || 'Вложение', 50) }}</span>
      </div>
      <button class="cancel-reply" @click="cancelReply">×</button>
    </div>
    
    <div class="message-input-wrapper">
      <div class="attach-wrapper">
        <button class="attach-btn btn-icon" @click.stop="toggleAttachMenu">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21.44 11.05l-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48" />
          </svg>
        </button>
        
        <!-- Меню вложений -->
        <div v-if="showAttachMenu" class="attach-menu" @click.stop>
          <button class="attach-option" @click="triggerFileUpload">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18" />
              <line x1="8" y1="2" x2="8" y2="22" />
              <line x1="16" y1="2" x2="16" y2="22" />
              <line x1="2" y1="8" x2="22" y2="8" />
              <line x1="2" y1="16" x2="22" y2="16" />
            </svg>
            <span>Файл</span>
          </button>
          <button class="attach-option" @click="startVoiceRecording">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="3" />
              <path d="M19 12a7 7 0 0 1-14 0" />
              <line x1="12" y1="19" x2="12" y2="22" />
            </svg>
            <span>Голосовое</span>
          </button>
          <button class="attach-option" @click="startVideoRecording">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polygon points="23 7 16 12 23 17 23 7" />
              <rect x="1" y="5" width="15" height="14" rx="2" ry="2" />
            </svg>
            <span>Видео</span>
          </button>
        </div>
      </div>
      
      <input
        ref="inputRef"
        v-model="message"
        type="text"
        :placeholder="replyTo ? 'Напишите ответ...' : placeholder"
        class="message-input"
        @keydown.enter.prevent="send"
        @input="onTyping"
      />
      
      <button class="send-btn btn-primary" :disabled="!message.trim()" @click="send">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="22" y1="2" x2="11" y2="13" />
          <polygon points="22 2 15 22 11 13 2 9 22 2" />
        </svg>
      </button>
    </div>
    
    <input ref="fileInput" type="file" style="display: none" @change="handleFileSelect" />
    
    <!-- Voice Recorder Modal -->
    <div v-if="showVoiceRecorder" class="modal-overlay" @click.self="showVoiceRecorder = false">
      <div class="modal-content recorder-modal">
        <VoiceRecorder @send="sendVoiceMessage" />
      </div>
    </div>
    
    <!-- Video Recorder Modal -->
    <div v-if="showVideoRecorder" class="modal-overlay" @click.self="showVideoRecorder = false">
      <div class="modal-content recorder-modal">
        <VideoRecorder @send="sendVideoMessage" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useAuthStore } from '@/stores/authStore'
import type { Message } from '@/types/chat'
import VoiceRecorder from './VoiceRecorder.vue'
import VideoRecorder from './VideoRecorder.vue'

const props = defineProps<{
  placeholder?: string
  replyTo?: Message | null
}>()

const emit = defineEmits(['send', 'sendFile', 'sendVoice', 'sendVideo', 'typing', 'cancel-reply'])

const authStore = useAuthStore()
const message = ref('')
const inputRef = ref<HTMLInputElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const showAttachMenu = ref(false)
const showVoiceRecorder = ref(false)
const showVideoRecorder = ref(false)
let typingTimeout: ReturnType<typeof setTimeout> | null = null

function send() {
  if (!message.value.trim()) return
  emit('send', message.value.trim())
  message.value = ''
  nextTick(() => inputRef.value?.focus())
}

function onTyping() {
  emit('typing', true)
  if (typingTimeout) clearTimeout(typingTimeout)
  typingTimeout = setTimeout(() => emit('typing', false), 1000)
}

function cancelReply() {
  emit('cancel-reply')
}

function getSenderName(): string {
  if (!props.replyTo) return ''
  return props.replyTo.sender_id === authStore.user?.id ? 'Вы' : 'Собеседник'
}

function truncateText(text: string, maxLength: number): string {
  if (!text) return 'Вложение'
  if (text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

function toggleAttachMenu() {
  showAttachMenu.value = !showAttachMenu.value
}

function triggerFileUpload() {
  showAttachMenu.value = false
  fileInput.value?.click()
}

function handleFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  if (input.files?.length) {
    emit('sendFile', input.files[0])
  }
  input.value = ''
}

function startVoiceRecording() {
  showAttachMenu.value = false
  showVoiceRecorder.value = true
}

function sendVoiceMessage(blob: Blob) {
  emit('sendVoice', blob)
  showVoiceRecorder.value = false
}

function startVideoRecording() {
  showAttachMenu.value = false
  showVideoRecorder.value = true
}

function sendVideoMessage(blob: Blob) {
  emit('sendVideo', blob)
  showVideoRecorder.value = false
}

// Закрываем меню при клике вне
function handleClickOutside(event: MouseEvent) {
  const target = event.target as HTMLElement
  if (!target.closest('.attach-wrapper')) {
    showAttachMenu.value = false
  }
}

onMounted(() => {
  inputRef.value?.focus()
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  if (typingTimeout) clearTimeout(typingTimeout)
})
</script>

<style scoped>
.message-input-container {
  padding: 16px 20px;
  background: var(--bg-primary);
  border-top: 1px solid var(--border-color);
  flex-shrink: 0;
  position: relative;
}

.reply-indicator {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-tertiary);
  border-radius: 14px;
  padding: 10px 14px;
  margin-bottom: 12px;
  border-left: 3px solid var(--primary-color);
}

.reply-info {
  flex: 1;
  overflow: hidden;
}

.reply-label {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}

.reply-label svg {
  width: 12px;
  height: 12px;
  stroke: var(--primary-color);
}

.reply-label span {
  font-size: 12px;
  font-weight: 600;
  color: var(--primary-color);
}

.reply-text {
  font-size: 12px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.cancel-reply {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: var(--text-muted);
  padding: 0 4px;
}

.message-input-wrapper {
  display: flex;
  align-items: center;
  gap: 12px;
  background: var(--bg-tertiary);
  border-radius: 24px;
  padding: 8px 16px;
  position: relative;
}

.attach-wrapper {
  position: relative;
}

.attach-btn {
  color: var(--text-secondary);
  flex-shrink: 0;
}

.attach-btn:hover {
  color: var(--primary-color);
}

.attach-menu {
  position: absolute;
  bottom: 45px;
  left: 0;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  z-index: 100;
  min-width: 150px;
  box-shadow: var(--shadow-lg);
}

.attach-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  background: none;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  color: var(--text-primary);
  transition: background 0.2s;
  width: 100%;
  text-align: left;
}

.attach-option:hover {
  background: var(--bg-hover);
}

.attach-option svg {
  width: 20px;
  height: 20px;
  stroke: var(--primary-color);
}

.message-input {
  flex: 1;
  background: none;
  border: none;
  padding: 8px 0;
  font-size: 14px;
  outline: none;
  color: var(--text-primary);
}

.message-input::placeholder {
  color: var(--text-muted);
}

.send-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.recorder-modal {
  background: transparent;
  box-shadow: none;
  padding: 0;
  max-width: 90%;
  width: auto;
}

@media (max-width: 768px) {
  .message-input-container {
    padding: 10px 12px;
  }
  
  .message-input-wrapper {
    padding: 6px 12px;
  }
  
  .message-input {
    font-size: 16px;
  }
  
  .send-btn {
    width: 32px;
    height: 32px;
  }
  
  .attach-menu {
    bottom: 40px;
    min-width: 130px;
  }
  
  .attach-option {
    padding: 8px 10px;
    font-size: 13px;
  }
}
</style>
<template>
  <div class="voice-recorder">
    <div v-if="!recording && !audioBlob" class="recorder-start">
      <button class="record-btn" @click="startRecording">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10" />
          <circle cx="12" cy="12" r="3" />
        </svg>
        <span>Начать запись</span>
      </button>
    </div>

    <div v-if="recording" class="recording-active">
      <div class="recording-wave">
        <span></span><span></span><span></span><span></span><span></span>
      </div>
      <div class="recording-info">
        <div class="recording-dot"></div>
        <span class="recording-time">{{ formatTime(recordingTime) }}</span>
      </div>
      <div class="recording-actions">
        <button class="stop-btn" @click="stopRecording">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="6" y="6" width="12" height="12" />
          </svg>
          <span>Остановить</span>
        </button>
        <button class="cancel-btn" @click="cancelRecording">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
          <span>Отмена</span>
        </button>
      </div>
    </div>

    <div v-if="audioBlob && !recording" class="recorder-preview">
      <div class="audio-preview">
        <audio ref="audioPlayer" :src="audioUrl" controls></audio>
      </div>
      <div class="preview-actions">
        <button class="send-btn" @click="sendAudio">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="22" y1="2" x2="11" y2="13" />
            <polygon points="22 2 15 22 11 13 2 9 22 2" />
          </svg>
          <span>Отправить</span>
        </button>
        <button class="cancel-btn" @click="clearAudio">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
          <span>Отмена</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onUnmounted } from 'vue'

const emit = defineEmits(['send'])

const recording = ref(false)
const audioBlob = ref<Blob | null>(null)
const audioUrl = ref('')
const mediaRecorder = ref<MediaRecorder | null>(null)
const audioChunks = ref<Blob[]>([])
const recordingTime = ref(0)
let timeInterval: ReturnType<typeof setInterval> | null = null
let stream: MediaStream | null = null

function formatTime(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
}

async function startRecording() {
  try {
    stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    
    mediaRecorder.value = new MediaRecorder(stream)
    audioChunks.value = []
    
    mediaRecorder.value.ondataavailable = (event) => {
      if (event.data.size > 0) {
        audioChunks.value.push(event.data)
      }
    }
    
    mediaRecorder.value.onstop = () => {
      const blob = new Blob(audioChunks.value, { type: 'audio/webm' })
      audioBlob.value = blob
      audioUrl.value = URL.createObjectURL(blob)
      recording.value = false
      
      if (timeInterval) {
        clearInterval(timeInterval)
        timeInterval = null
      }
      
      if (stream) {
        stream.getTracks().forEach(track => track.stop())
        stream = null
      }
    }
    
    mediaRecorder.value.start(1000)
    recording.value = true
    recordingTime.value = 0
    
    timeInterval = setInterval(() => {
      recordingTime.value++
    }, 1000)
    
  } catch (error) {
    console.error('Failed to start recording:', error)
    alert('Не удалось получить доступ к микрофону. Разрешите доступ к микрофону в браузере.')
  }
}

function stopRecording() {
  if (mediaRecorder.value && mediaRecorder.value.state === 'recording') {
    mediaRecorder.value.stop()
  }
}

function cancelRecording() {
  if (mediaRecorder.value && mediaRecorder.value.state === 'recording') {
    mediaRecorder.value.stop()
  }
  
  if (timeInterval) {
    clearInterval(timeInterval)
    timeInterval = null
  }
  
  recording.value = false
  recordingTime.value = 0
  
  if (stream) {
    stream.getTracks().forEach(track => track.stop())
    stream = null
  }
}

function clearAudio() {
  audioBlob.value = null
  audioUrl.value = ''
}

function sendAudio() {
  if (audioBlob.value) {
    emit('send', audioBlob.value)
    clearAudio()
  }
}

onUnmounted(() => {
  if (timeInterval) {
    clearInterval(timeInterval)
  }
  if (stream) {
    stream.getTracks().forEach(track => track.stop())
  }
  if (mediaRecorder.value && mediaRecorder.value.state === 'recording') {
    mediaRecorder.value.stop()
  }
})
</script>

<style scoped>
.voice-recorder {
  background: var(--bg-primary);
  border-radius: 20px;
  padding: 24px;
  min-width: 300px;
  text-align: center;
}

.recorder-start {
  display: flex;
  justify-content: center;
}

.record-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 24px;
  background: var(--primary-light);
  border: 2px solid var(--primary-color);
  border-radius: 50%;
  color: var(--primary-color);
  transition: all 0.2s;
}

.record-btn span {
  font-size: 14px;
  white-space: nowrap;
}

.record-btn:hover {
  background: var(--primary-color);
  color: white;
  transform: scale(1.05);
}

.recording-active {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

.recording-wave {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  height: 60px;
}

.recording-wave span {
  width: 4px;
  background: var(--primary-color);
  border-radius: 2px;
  animation: wave 0.5s ease-in-out infinite;
}

.recording-wave span:nth-child(1) { animation-delay: 0s; height: 20px; }
.recording-wave span:nth-child(2) { animation-delay: 0.1s; height: 35px; }
.recording-wave span:nth-child(3) { animation-delay: 0.2s; height: 50px; }
.recording-wave span:nth-child(4) { animation-delay: 0.3s; height: 35px; }
.recording-wave span:nth-child(5) { animation-delay: 0.4s; height: 20px; }

@keyframes wave {
  0%, 100% { transform: scaleY(1); }
  50% { transform: scaleY(0.5); }
}

.recording-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.recording-dot {
  width: 12px;
  height: 12px;
  background: red;
  border-radius: 50%;
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.recording-time {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
}

.recording-actions {
  display: flex;
  gap: 20px;
  margin-top: 10px;
}

.stop-btn, .cancel-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  border-radius: 12px;
  font-size: 14px;
  transition: all 0.2s;
}

.stop-btn {
  background: var(--primary-color);
  color: white;
}

.stop-btn:hover {
  background: var(--primary-hover);
  transform: scale(1.02);
}

.cancel-btn {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}

.cancel-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.recorder-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

.audio-preview {
  width: 100%;
}

.audio-preview audio {
  width: 100%;
  height: 48px;
  border-radius: 24px;
}

.preview-actions {
  display: flex;
  gap: 20px;
}

.send-btn, .preview-actions .cancel-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border-radius: 12px;
  font-size: 14px;
}

.send-btn {
  background: var(--primary-color);
  color: white;
}

.send-btn:hover {
  background: var(--primary-hover);
}

@media (max-width: 768px) {
  .voice-recorder {
    padding: 20px;
    min-width: 280px;
  }
  
  .recording-time {
    font-size: 20px;
  }
  
  .stop-btn, .cancel-btn {
    padding: 10px 16px;
  }
}
</style>
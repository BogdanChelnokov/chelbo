<template>
  <div class="settings-page">
    <div class="settings-container">
      <div class="settings-header">
        <button class="back-btn btn-icon" @click="router.push('/')">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="19" y1="12" x2="5" y2="12" />
            <polyline points="12 19 5 12 12 5" />
          </svg>
        </button>
        <h1>Настройки</h1>
      </div>
      
      <div class="settings-content">
        <div class="settings-section">
          <h3>Профиль</h3>
          
          <div class="avatar-section">
            <div class="avatar-preview">
              <img v-if="authStore.user?.avatar_url" :src="authStore.user.avatar_url" alt="Avatar" />
              <span v-else>{{ authStore.user?.name?.charAt(0).toUpperCase() }}</span>
            </div>
            <button class="btn btn-secondary" @click="triggerAvatarUpload">
              Изменить аватар
            </button>
          </div>
          
          <div class="form-group">
            <label>Имя</label>
            <input v-model="form.name" type="text" class="form-input" />
          </div>
          
          <div class="form-group">
            <label>О себе</label>
            <textarea v-model="form.bio" rows="4" class="form-input" placeholder="Расскажите о себе..."></textarea>
          </div>
          
          <div class="form-group">
            <label>Email</label>
            <input :value="authStore.user?.email" type="email" class="form-input" disabled />
          </div>
          
          <button class="btn btn-primary save-btn" @click="saveProfile">
            Сохранить изменения
          </button>
        </div>
        
        <div class="settings-section">
          <h3>Аккаунт</h3>
          <button class="btn btn-danger logout-btn" @click="handleLogout">
            Выйти из аккаунта
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'
import { useWebSocketStore } from '@/stores/websocketStore'
import { authApi } from '@/api/auth'

const router = useRouter()
const authStore = useAuthStore()
const wsStore = useWebSocketStore()

const form = reactive({
  name: authStore.user?.name || '',
  bio: authStore.user?.bio || ''
})

const avatarInput = ref<HTMLInputElement | null>(null)

function triggerAvatarUpload() {
  avatarInput.value?.click()
}

async function saveProfile() {
  try {
    const response = await authApi.updateProfile({
      name: form.name,
      bio: form.bio
    })
    
    if (response.data.data) {
      authStore.updateUser(response.data.data)
      alert('Профиль успешно обновлен')
    }
  } catch (error) {
    console.error('Failed to update profile:', error)
    alert('Ошибка обновления профиля')
  }
}

async function handleLogout() {
  await authStore.logout()
  wsStore.disconnect()
  router.push('/login')
}
</script>

<style scoped>
.settings-page {
  width: 100vw;
  height: 100vh;
  background: var(--bg-secondary);
  overflow-y: auto;
}

.settings-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 40px 20px;
}

.settings-header {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 40px;
}

.back-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--bg-primary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.settings-header h1 {
  font-size: 28px;
  font-weight: 600;
}

.settings-content {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.settings-section {
  background: var(--bg-primary);
  border-radius: 16px;
  padding: 24px;
  box-shadow: var(--shadow-sm);
}

.settings-section h3 {
  font-size: 18px;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-color);
}

.avatar-section {
  text-align: center;
  margin-bottom: 24px;
}

.avatar-preview {
  width: 120px;
  height: 120px;
  margin: 0 auto 16px;
  border-radius: 50%;
  background: var(--primary-color);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 48px;
  font-weight: 600;
  overflow: hidden;
}

.avatar-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: var(--text-secondary);
}

.form-input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s;
}

.form-input:focus {
  border-color: var(--primary-color);
  outline: none;
  box-shadow: 0 0 0 3px var(--primary-light);
}

textarea.form-input {
  resize: vertical;
  font-family: inherit;
}

.save-btn {
  width: 100%;
  margin-top: 8px;
}

.logout-btn {
  width: 100%;
}

.settings-page {
  background: var(--bg-secondary);
}

.settings-section {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
}

.settings-section h3 {
  border-bottom: 1px solid var(--border-color);
}

.form-input {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
}

.form-input:focus {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px var(--primary-light);
}
</style>
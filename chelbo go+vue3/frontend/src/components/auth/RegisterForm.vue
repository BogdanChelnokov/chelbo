<template>
  <div class="register-form">
    <h2 class="form-title">Регистрация в Chelbo</h2>
    
    <div v-if="errorMessage" class="error-message">
      {{ errorMessage }}
    </div>
    
    <form @submit.prevent="handleSubmit">
      <div class="form-group">
        <label for="name">Имя</label>
        <input
          id="name"
          v-model="form.name"
          type="text"
          placeholder="Иван Иванов"
          required
          :disabled="loading"
        />
      </div>
      
      <div class="form-group">
        <label for="email">Email</label>
        <input
          id="email"
          v-model="form.email"
          type="email"
          placeholder="your@email.com"
          required
          :disabled="loading"
        />
      </div>
      
      <div class="form-group">
        <label for="password">Пароль</label>
        <input
          id="password"
          v-model="form.password"
          type="password"
          placeholder="••••••"
          required
          :disabled="loading"
        />
        <p class="hint">Минимум 6 символов, включая заглавные, строчные, цифры и спецсимволы</p>
      </div>
      
      <button type="submit" class="btn btn-primary btn-submit" :disabled="loading">
        {{ loading ? 'Регистрация...' : 'Зарегистрироваться' }}
      </button>
    </form>
    
    <p class="form-footer">
      Уже есть аккаунт?
      <router-link to="/login" class="link">Войти</router-link>
    </p>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'

const router = useRouter()
const authStore = useAuthStore()

const form = reactive({
  name: '',
  email: '',
  password: ''
})

const loading = ref(false)
const errorMessage = ref('')

async function handleSubmit() {
  loading.value = true
  errorMessage.value = ''
  
  if (form.name.length < 2) {
    errorMessage.value = 'Имя должно содержать минимум 2 символа'
    loading.value = false
    return
  }
  
  if (!form.email.includes('@') || !form.email.includes('.')) {
    errorMessage.value = 'Введите корректный email'
    loading.value = false
    return
  }
  
  if (form.password.length < 6) {
    errorMessage.value = 'Пароль должен содержать минимум 6 символов'
    loading.value = false
    return
  }
  
  try {
    const result = await authStore.register(form)
    
    if (result.success) {
      router.push('/')
    } else {
      errorMessage.value = result.error || 'Ошибка регистрации. Попробуйте позже.'
    }
  } catch (error: any) {
    console.error('Registration error:', error)
    errorMessage.value = error.response?.data?.error || 'Ошибка соединения с сервером'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.register-form {
  background: var(--bg-primary);
  padding: 40px;
  border-radius: 16px;
  box-shadow: var(--shadow-xl);
  width: 100%;
  max-width: 400px;
}

.form-title {
  font-size: 28px;
  font-weight: 600;
  margin-bottom: 32px;
  text-align: center;
  color: var(--text-primary);
}

.error-message {
  background: rgba(157, 19, 6, 0.15);
  color: var(--primary-color);
  padding: 12px;
  border-radius: 8px;
  margin-bottom: 20px;
  font-size: 14px;
  text-align: center;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
}

.form-group input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s;
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.form-group input:focus {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px var(--primary-light);
  outline: none;
}

.form-group input::placeholder {
  color: var(--text-muted);
}

.form-group input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.hint {
  margin-top: 6px;
  font-size: 12px;
  color: var(--text-muted);
}

.btn-submit {
  width: 100%;
  margin-top: 8px;
  padding: 12px;
  font-size: 16px;
}

.btn-submit:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.form-footer {
  margin-top: 24px;
  text-align: center;
  font-size: 14px;
  color: var(--text-secondary);
}

.link {
  color: var(--primary-color);
  font-weight: 500;
  text-decoration: none;
}

.link:hover {
  text-decoration: underline;
}

/* Мобильная адаптация */
@media (max-width: 768px) {
  .register-form {
    padding: 24px;
    max-width: 90%;
  }
  
  .form-title {
    font-size: 24px;
    margin-bottom: 24px;
  }
  
  .form-group input {
    padding: 10px 14px;
  }
  
  .btn-submit {
    padding: 10px;
  }
}
</style>
<template>
  <div class="login-page">
    <div class="login-container">
      <div class="brand">
        <div class="logo">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2">
            <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z" />
          </svg>
        </div>
        <h1>Chelbo</h1>
        <p>Российский кроссплатформенный мессенджер</p>
      </div>
      <LoginForm />
    </div>
  </div>
</template>

<script setup lang="ts">
import LoginForm from '@/components/auth/LoginForm.vue'
</script>

<style scoped>
.login-page {
  width: 100vw;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-secondary);
}

.login-container {
  width: 100%;
  max-width: 1100px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 60px;
  padding: 20px;
}

.brand {
  color: var(--text-primary);
  flex: 1;
}

.logo {
  width: 80px;
  height: 80px;
  background: var(--primary-color);
  border-radius: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
}

.logo svg {
  stroke: white;
}

.brand h1 {
  font-size: 48px;
  font-weight: 700;
  margin-bottom: 12px;
  color: var(--text-primary);
}

.brand p {
  font-size: 18px;
  color: var(--text-secondary);
}

@media (max-width: 768px) {
  .login-container {
    flex-direction: column;
    justify-content: center;
    text-align: center;
  }
  
  .logo {
    margin: 0 auto 24px;
  }
}
</style>
<template>
  <div class="not-found-page">
    <div class="not-found-content">
      <div class="error-code">404</div>
      <h1>Страница не найдена</h1>
      <p>Извините, запрашиваемая страница не существует</p>
      <router-link to="/" class="btn btn-primary">
        Вернуться на главную
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.not-found-page {
  width: 100vw;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.not-found-content {
  text-align: center;
  color: white;
  padding: 40px;
}

.error-code {
  font-size: 120px;
  font-weight: 700;
  margin-bottom: 20px;
  text-shadow: 4px 4px 0 rgba(0, 0, 0, 0.1);
}

.not-found-content h1 {
  font-size: 32px;
  margin-bottom: 12px;
}

.not-found-content p {
  font-size: 16px;
  opacity: 0.9;
  margin-bottom: 32px;
}

.not-found-content .btn {
  display: inline-block;
}

.not-found-page {
  background: var(--bg-secondary);
}

.error-code {
  color: var(--primary-color);
}
</style>
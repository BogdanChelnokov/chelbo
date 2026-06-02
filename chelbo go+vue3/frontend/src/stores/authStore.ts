import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import type { User, LoginRequest, RegisterRequest } from '@/types/user'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<User | null>(null)

  const isAuthenticated = computed(() => !!token.value && !!user.value)

  async function login(credentials: LoginRequest) {
    try {
      const response = await authApi.login(credentials)
      console.log('Login response:', response.data)
      
      if (response.data.success && response.data.data) {
        const { token: authToken, user: userData } = response.data.data
        
        token.value = authToken
        user.value = userData
        
        localStorage.setItem('token', authToken)
        
        return { success: true }
      } else {
        // В случае ошибки, response.data может содержать поле error
        const errorMsg = (response.data as any).error || 'Ошибка входа'
        return { 
          success: false, 
          error: errorMsg
        }
      }
    } catch (error: any) {
      console.error('Login API error:', error)
      const errorMsg = error.response?.data?.error || error.message || 'Ошибка соединения с сервером'
      return { 
        success: false, 
        error: errorMsg
      }
    }
  }

  async function register(userData: RegisterRequest) {
    try {
      const response = await authApi.register(userData)
      console.log('Register response:', response.data)
      
      if (response.data.success && response.data.data) {
        const { token: authToken, user: newUser } = response.data.data
        
        token.value = authToken
        user.value = newUser
        
        localStorage.setItem('token', authToken)
        
        return { success: true }
      } else {
        // В случае ошибки, response.data может содержать поле error
        const errorMsg = (response.data as any).error || 'Ошибка регистрации'
        return { 
          success: false, 
          error: errorMsg
        }
      }
    } catch (error: any) {
      console.error('Register API error:', error)
      let errorMsg = 'Ошибка соединения с сервером'
      
      if (error.response?.data?.error) {
        errorMsg = error.response.data.error
      } else if (error.response?.data?.message) {
        errorMsg = error.response.data.message
      } else if (error.message) {
        errorMsg = error.message
      }
      
      return { 
        success: false, 
        error: errorMsg
      }
    }
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      token.value = null
      user.value = null
      localStorage.removeItem('token')
    }
  }

  async function checkAuth() {
    if (!token.value) return false
    
    try {
      const response = await authApi.getMe()
      if (response.data.success && response.data.data) {
        user.value = response.data.data
        return true
      }
      return false
    } catch (error) {
      console.error('Check auth error:', error)
      token.value = null
      user.value = null
      localStorage.removeItem('token')
      return false
    }
  }

  function updateUser(updatedUser: User) {
    user.value = updatedUser
  }

  return {
    token,
    user,
    isAuthenticated,
    login,
    register,
    logout,
    checkAuth,
    updateUser
  }
})
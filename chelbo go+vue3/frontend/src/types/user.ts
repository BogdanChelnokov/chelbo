export interface User {
  id: number
  email: string
  name: string
  avatar_url?: string
  bio: string
  last_seen: string
  is_online?: boolean
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  name: string
}

export interface UpdateProfileRequest {
  name?: string
  bio?: string
}

export interface ApiResponse<T = any> {
  success: boolean
  message?: string
  data?: T
  error?: string
}
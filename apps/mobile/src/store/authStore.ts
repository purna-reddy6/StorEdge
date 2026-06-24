import { create } from 'zustand'
import { setAuthToken } from '../utils/api'

interface User {
  id: string
  phone: string
  name: string
  role: string
}

interface AuthState {
  user: User | null
  token: string | null
  login: (user: User, token: string) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  login: (user, token) => {
    setAuthToken(token)
    set({ user, token })
  },
  logout: () => {
    setAuthToken(null)
    set({ user: null, token: null })
  },
}))

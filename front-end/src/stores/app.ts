import { create } from 'zustand'

type State = {
  apiUrl: string
  token: string
  sessionId: string
  refreshToken: string
  csrfToken: string
}

type Actions = {
  setApiUrl: (v: string) => void
  setToken: (v: string) => void
  setSessionId: (v: string) => void
  setRefreshToken: (v: string) => void
  setCsrfToken: (v: string) => void
  logout: () => void
}

export const useAppStore = create<State & Actions>((set) => ({
  apiUrl: import.meta.env.VITE_API_URL || 'http://localhost:8000',
  token: '',
  sessionId: '',
  refreshToken: '',
  csrfToken: '',
  setApiUrl: (v) => set({ apiUrl: v }),
  setToken: (v) => set({ token: v }),
  setSessionId: (v) => set({ sessionId: v }),
  setRefreshToken: (v) => set({ refreshToken: v }),
  setCsrfToken: (v) => set({ csrfToken: v }),
  logout: () => set({ token: '', sessionId: '', refreshToken: '' }),
}))

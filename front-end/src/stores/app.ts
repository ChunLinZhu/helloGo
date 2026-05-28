import { create } from 'zustand'
import { persist } from 'zustand/middleware'

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

export const useAppStore = create<State & Actions>()(
  persist(
    (set) => ({
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
    }),
    {
      name: 'helloGo-auth', // localStorage key
      partialize: (state) => ({
        // 只持久化认证数据，apiUrl 每次从 .env 读取
        token: state.token,
        sessionId: state.sessionId,
        refreshToken: state.refreshToken,
        csrfToken: state.csrfToken,
      }),
    }
  )
)

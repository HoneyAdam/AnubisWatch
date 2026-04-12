import { create } from 'zustand'
import { api } from '../api/client'
import type { Soul, ApiResponse } from '../api/client'

interface SoulStore {
  souls: Soul[]
  pagination: { total: number; has_more: boolean } | null
  loading: boolean
  error: string | null
  fetchSouls: () => Promise<void>
  createSoul: (soul: Omit<Soul, 'id' | 'created_at' | 'updated_at' | 'updated_at'>) => Promise<Soul | null>
  updateSoul: (id: string, soul: Partial<Soul>) => Promise<Soul | null>
  deleteSoul: (id: string) => Promise<void>
}

export const useSoulStore = create<SoulStore>((set) => ({
  souls: [],
  pagination: null,
  loading: false,
  error: null,

  fetchSouls: async () => {
    set({ loading: true, error: null })
    try {
      const result = await api.get<ApiResponse<Soul[]>>('/souls')
      if (result) {
        set({ souls: result.data, pagination: result.pagination ?? null, loading: false })
      }
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Unknown error', loading: false })
    }
  },

  createSoul: async (soul) => {
    set({ loading: true, error: null })
    try {
      const result = await api.post<Soul>('/souls', soul)
      if (result) {
        set((state) => ({ souls: [...state.souls, result], loading: false }))
      }
      return result ?? null
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Unknown error', loading: false })
      return null
    }
  },

  updateSoul: async (id, soul) => {
    set({ loading: true, error: null })
    try {
      const result = await api.put<Soul>(`/souls/${id}`, soul)
      if (result) {
        set((state) => ({
          souls: state.souls.map((s) => (s.id === id ? result : s)),
          loading: false,
        }))
      }
      return result ?? null
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Unknown error', loading: false })
      return null
    }
  },

  deleteSoul: async (id) => {
    set({ loading: true, error: null })
    try {
      await api.delete(`/souls/${id}`)
      set((state) => ({
        souls: state.souls.filter((s) => s.id !== id),
        loading: false,
      }))
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Unknown error', loading: false })
    }
  },
}))

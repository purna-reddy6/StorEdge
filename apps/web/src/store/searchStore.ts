import { create } from 'zustand'
import type { Warehouse, SearchParams } from '../types'

interface SearchState {
  params: SearchParams
  results: Warehouse[]
  selectedWarehouse: Warehouse | null
  isSearching: boolean
  setParams: (p: Partial<SearchParams>) => void
  setResults: (r: Warehouse[]) => void
  selectWarehouse: (w: Warehouse | null) => void
  setSearching: (v: boolean) => void
}

export const useSearchStore = create<SearchState>((set) => ({
  params: {
    latitude: 27.1767,
    longitude: 78.0081,
    radiusKm: 50,
    requiredPallets: 10,
    needsColdChain: false,
  },
  results: [],
  selectedWarehouse: null,
  isSearching: false,
  setParams: (p) => set((s) => ({ params: { ...s.params, ...p } })),
  setResults: (results) => set({ results }),
  selectWarehouse: (selectedWarehouse) => set({ selectedWarehouse }),
  setSearching: (isSearching) => set({ isSearching }),
}))

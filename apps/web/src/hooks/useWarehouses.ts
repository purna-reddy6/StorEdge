import { useQuery, useMutation, useQueryClient } from 'react-query'
import api from '../utils/api'
import type { Warehouse, SearchParams, Booking } from '../types'

export function useWarehouseSearch(params: SearchParams, enabled: boolean) {
  return useQuery<Warehouse[]>(
    ['warehouses', 'search', params],
    async () => {
      const { data } = await api.get('/warehouses/search', { params: {
        lat: params.latitude,
        lng: params.longitude,
        radius_km: params.radiusKm,
        pallets: params.requiredPallets,
        cold_chain: params.needsColdChain,
        min_temp: params.minTemp,
        max_temp: params.maxTemp,
        max_price: params.maxPriceInr,
      }})
      return data.warehouses
    },
    { enabled, keepPreviousData: true },
  )
}

export function useWarehouse(id: string | undefined) {
  return useQuery<Warehouse>(
    ['warehouses', id],
    async () => {
      const { data } = await api.get(`/warehouses/${id}`)
      return data
    },
    { enabled: !!id },
  )
}

export function useCreateBooking() {
  const qc = useQueryClient()
  return useMutation(
    async (body: {
      warehouseId: string
      palletCount: number
      commodity: string
      inwardDate: string
      expectedOutwardDate: string
    }) => {
      const { data } = await api.post('/bookings', body)
      return data as Booking
    },
    { onSuccess: () => qc.invalidateQueries('bookings') },
  )
}

export function useBookings() {
  return useQuery<Booking[]>('bookings', async () => {
    const { data } = await api.get('/bookings')
    return data.bookings
  })
}

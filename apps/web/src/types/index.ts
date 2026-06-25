export type UserRole = 'farmer' | 'trader' | 'operator' | 'admin' | 'logistics'

export interface User {
  id: string
  phone: string
  name: string
  role: UserRole
  kycStatus: 'pending' | 'verified' | 'rejected'
}

export type WarehouseType =
  | 'cold_storage'
  | 'dry_warehouse'
  | 'silo'
  | 'controlled_atmosphere'
  | 'refrigerated_transport'
  | 'agri_processing'

export interface Warehouse {
  id: string
  name: string
  address_line1: string
  city: string
  state: string
  pincode: string
  latitude: number
  longitude: number
  type: WarehouseType
  total_pallet_capacity: number
  available_pallet_slots: number
  base_price_per_pallet_inr: number
  price_per_pallet_per_day_inr: number
  rating: number
  total_reviews: number
  wdra_status: string
  apmc_licensed: boolean
  min_temperature_celsius?: number
  max_temperature_celsius?: number
  distance_km?: number
  match_score?: number
  estimated_monthly_cost_inr?: number
}

export interface SearchParams {
  latitude: number
  longitude: number
  radiusKm: number
  requiredPallets: number
  needsColdChain: boolean
  minTemp?: number
  maxTemp?: number
  maxPriceInr?: number
}

export type BookingStatus = 'pending' | 'confirmed' | 'active' | 'completed' | 'cancelled'

export interface Booking {
  id: string
  booking_number: string
  warehouse_id: string
  warehouse_name: string
  farmer_name: string
  commodity_type: string
  pallet_count: number
  start_date: string
  end_date: string
  total_amount_inr: number
  commission_amount_inr: number
  payout_amount_inr: number
  status: BookingStatus
}

export interface IoTAlert {
  id: string
  warehouse_id: string
  sensor_id: string
  alert_type: string
  severity: 'info' | 'warning' | 'critical'
  message: string
  is_resolved: boolean
  resolved_at?: string
  created_at: string
}

export interface OccupancyStat {
  date: string
  occupancyPct: number
  revenue: number
}

export interface EnwrReceipt {
  id: string
  receiptNumber: string
  warehouseId: string
  commodity: string
  quantityKg: number
  marketValueInr: number
  maxLoanAmountInr: number
  status: 'draft' | 'issued' | 'pledged' | 'released' | 'expired'
  expiryDate: string
  issuedAt?: string
}

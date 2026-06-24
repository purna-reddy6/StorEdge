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
  address: string
  city: string
  state: string
  latitude: number
  longitude: number
  warehouseType: WarehouseType
  totalCapacityMt: number
  availablePallets: number
  totalPallets: number
  basePricePerMtPerMonth: number
  dynamicPrice?: number
  rating: number
  reviewCount: number
  wdraRegistered: boolean
  apmcLicensed: boolean
  minTemperatureCelsius?: number
  maxTemperatureCelsius?: number
  distanceKm?: number
  matchScore?: number
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
  bookingNumber: string
  warehouseId: string
  warehouseName: string
  farmerName: string
  commodity: string
  palletCount: number
  inwardDate: string
  expectedOutwardDate: string
  totalAmountInr: number
  commissionInr: number
  status: BookingStatus
}

export interface IoTAlert {
  id: string
  warehouseId: string
  sensorId: string
  alertType: string
  severity: 'info' | 'warning' | 'critical'
  message: string
  resolvedAt?: string
  createdAt: string
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
  issueDate: string
}

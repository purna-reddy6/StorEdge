export const formatINR = (amount: number): string =>
  new Intl.NumberFormat('en-IN', { style: 'currency', currency: 'INR', maximumFractionDigits: 0 }).format(amount)

export const formatDate = (iso: string): string =>
  new Date(iso).toLocaleDateString('en-IN', { day: '2-digit', month: 'short', year: 'numeric' })

export const formatOccupancy = (pct: number): string => `${Math.round(pct * 100)}%`

export const warehouseTypeLabel: Record<string, string> = {
  cold_storage: 'Cold Storage',
  ambient: 'Ambient / Dry',
  dry_warehouse: 'Dry Warehouse',
  industrial: 'Industrial',
  self_storage: 'Self Storage',
  pharmaceutical: 'Pharma Cold Chain',
  bonded: 'Bonded Warehouse',
  hazmat: 'HAZMAT Certified',
  controlled_atmosphere: 'CA Store',
  retail_backroom: 'Retail Backroom',
  silo: 'Grain Silo',
}

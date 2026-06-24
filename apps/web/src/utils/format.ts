export const formatINR = (amount: number): string =>
  new Intl.NumberFormat('en-IN', { style: 'currency', currency: 'INR', maximumFractionDigits: 0 }).format(amount)

export const formatDate = (iso: string): string =>
  new Date(iso).toLocaleDateString('en-IN', { day: '2-digit', month: 'short', year: 'numeric' })

export const formatOccupancy = (pct: number): string => `${Math.round(pct * 100)}%`

export const warehouseTypeLabel: Record<string, string> = {
  cold_storage: 'Cold Storage',
  dry_warehouse: 'Dry Warehouse',
  silo: 'Silo',
  controlled_atmosphere: 'CA Store',
  refrigerated_transport: 'Reefer Transport',
  agri_processing: 'Agri Processing',
}

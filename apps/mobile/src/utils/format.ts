export const formatINR = (amount: number): string =>
  '₹' + amount.toLocaleString('en-IN')

export const formatDate = (iso: string): string => {
  const d = new Date(iso)
  return `${d.getDate()} ${d.toLocaleString('en-IN', { month: 'short' })} ${d.getFullYear()}`
}

export const warehouseTypeLabel: Record<string, string> = {
  cold_storage: 'Cold Storage',
  dry_warehouse: 'Dry Warehouse',
  silo: 'Silo',
  controlled_atmosphere: 'CA Store',
  refrigerated_transport: 'Reefer',
  agri_processing: 'Processing',
}

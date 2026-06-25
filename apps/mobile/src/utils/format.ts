export const formatINR = (amount: number): string =>
  '₹' + amount.toLocaleString('en-IN')

export const formatDate = (iso: string): string => {
  const d = new Date(iso)
  return `${d.getDate()} ${d.toLocaleString('en-IN', { month: 'short' })} ${d.getFullYear()}`
}

export const warehouseTypeLabel: Record<string, string> = {
  cold_storage: 'Cold Storage',
  ambient: 'Ambient / Dry',
  dry_warehouse: 'Dry Warehouse',
  industrial: 'Industrial',
  self_storage: 'Self Storage',
  pharmaceutical: 'Pharma',
  bonded: 'Bonded',
  hazmat: 'HAZMAT',
  controlled_atmosphere: 'CA Store',
  retail_backroom: 'Retail Backroom',
  silo: 'Grain Silo',
}

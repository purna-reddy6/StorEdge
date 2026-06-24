import { useQuery } from 'react-query'
import api from '../utils/api'
import { formatDate } from '../utils/format'

interface PalletItem {
  id: string
  palletCode: string
  commodity: string
  quantityKg: number
  slotPosition: string
  inwardDate: string
  expectedOutwardDate: string
  currentTemperatureCelsius?: number
  currentHumidityPct?: number
  spoilageRiskScore?: number
  spoilageRiskLevel?: string
}

function usePallets() {
  return useQuery<PalletItem[]>('pallets', () =>
    api.get('/inventory/pallets').then((r) => r.data.items)
  )
}

const riskColors: Record<string, string> = {
  safe: 'badge-green',
  low: 'badge-green',
  medium: 'badge-yellow',
  high: 'badge-red',
  critical: 'badge-red',
}

export default function InventoryPage() {
  const { data: pallets = [], isLoading } = usePallets()

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Inventory — Pallet Register</h1>

      {isLoading && <p className="text-gray-500">Loading inventory…</p>}

      <div className="card overflow-hidden p-0">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b border-gray-200">
            <tr>
              {['Pallet Code', 'Commodity', 'Qty (kg)', 'Slot', 'Inward', 'Out (Expected)', 'Temp', 'Risk'].map(h => (
                <th key={h} className="text-left text-xs font-medium text-gray-500 uppercase tracking-wide px-4 py-3">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {pallets.map((p) => (
              <tr key={p.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 font-mono text-xs text-gray-600">{p.palletCode}</td>
                <td className="px-4 py-3 capitalize">{p.commodity}</td>
                <td className="px-4 py-3">{p.quantityKg.toLocaleString()}</td>
                <td className="px-4 py-3 font-mono text-xs bg-gray-50 text-gray-700">{p.slotPosition}</td>
                <td className="px-4 py-3 text-gray-600">{formatDate(p.inwardDate)}</td>
                <td className="px-4 py-3 text-gray-600">{formatDate(p.expectedOutwardDate)}</td>
                <td className="px-4 py-3">
                  {p.currentTemperatureCelsius !== undefined ? (
                    <span className={p.currentTemperatureCelsius > 8 ? 'text-red-600 font-medium' : 'text-gray-700'}>
                      {p.currentTemperatureCelsius.toFixed(1)}°C
                    </span>
                  ) : '—'}
                </td>
                <td className="px-4 py-3">
                  {p.spoilageRiskLevel ? (
                    <span className={riskColors[p.spoilageRiskLevel] ?? 'badge-yellow'}>
                      {p.spoilageRiskLevel}
                    </span>
                  ) : '—'}
                </td>
              </tr>
            ))}
            {pallets.length === 0 && !isLoading && (
              <tr>
                <td colSpan={8} className="px-4 py-8 text-center text-gray-400">
                  No pallets in inventory
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}

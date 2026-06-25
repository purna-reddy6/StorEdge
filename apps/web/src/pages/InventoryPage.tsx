import { useQuery } from 'react-query'
import api from '../utils/api'
import { formatDate } from '../utils/format'

interface PalletItem {
  id: string
  bookingId: string
  commodity: string
  weightKg: number
  slotPosition: string
  inwardDate?: string
  expectedOutwardDate?: string
  currentTemperatureCelsius?: number
  releaseStatus?: string
}

function usePallets() {
  return useQuery<PalletItem[]>('pallets', () =>
    api.get('/inventory/pallets').then((r) => r.data.items)
  )
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
              {['Slot', 'Commodity', 'Weight (kg)', 'Booking', 'Inward', 'Out (Expected)', 'Temp', 'Release'].map(h => (
                <th key={h} className="text-left text-xs font-medium text-gray-500 uppercase tracking-wide px-4 py-3">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {pallets.map((p) => (
              <tr key={p.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 font-mono text-xs text-gray-600">{p.slotPosition || '—'}</td>
                <td className="px-4 py-3 capitalize">{p.commodity}</td>
                <td className="px-4 py-3">{p.weightKg.toLocaleString()}</td>
                <td className="px-4 py-3 font-mono text-xs text-gray-400 truncate max-w-24">{p.bookingId.slice(0,8)}…</td>
                <td className="px-4 py-3 text-gray-600">{p.inwardDate ? formatDate(p.inwardDate) : '—'}</td>
                <td className="px-4 py-3 text-gray-600">{p.expectedOutwardDate ? formatDate(p.expectedOutwardDate) : '—'}</td>
                <td className="px-4 py-3">
                  {p.currentTemperatureCelsius != null ? (
                    <span className={p.currentTemperatureCelsius > 8 ? 'text-red-600 font-medium' : 'text-gray-700'}>
                      {p.currentTemperatureCelsius.toFixed(1)}°C
                    </span>
                  ) : '—'}
                </td>
                <td className="px-4 py-3">
                  {p.releaseStatus ? (
                    <span className="badge-yellow capitalize">{p.releaseStatus.replace('_', ' ')}</span>
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

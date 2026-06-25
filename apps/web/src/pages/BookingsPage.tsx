import { useBookings } from '../hooks/useWarehouses'
import { formatINR, formatDate } from '../utils/format'
import type { BookingStatus } from '../types'

const statusConfig: Record<BookingStatus, { label: string; cls: string }> = {
  pending:   { label: 'Pending',   cls: 'badge-yellow' },
  confirmed: { label: 'Confirmed', cls: 'badge-green' },
  active:    { label: 'Active',    cls: 'badge-green' },
  completed: { label: 'Completed', cls: 'bg-gray-100 text-gray-700 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium' as const },
  cancelled: { label: 'Cancelled', cls: 'badge-red' },
}

export default function BookingsPage() {
  const { data: bookings = [], isLoading, error } = useBookings()

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">My Bookings</h1>

      {isLoading && <p className="text-gray-500">Loading bookings…</p>}
      {!!error && <p className="text-red-500">Failed to load bookings.</p>}

      {bookings.length === 0 && !isLoading && (
        <div className="card text-center py-12">
          <p className="text-gray-500">No bookings yet.</p>
          <a href="/search" className="btn-primary mt-4 inline-flex">Find a Warehouse</a>
        </div>
      )}

      <div className="space-y-4">
        {bookings.map((b) => {
          const sc = statusConfig[b.status]
          return (
            <div key={b.id} className="card">
              <div className="flex justify-between items-start">
                <div>
                  <div className="flex items-center gap-3">
                    <span className="font-mono text-sm font-medium text-gray-600">{b.booking_number}</span>
                    <span className={sc.cls}>{sc.label}</span>
                  </div>
                  <h3 className="font-semibold text-gray-900 mt-1">{b.warehouse_name}</h3>
                  <p className="text-sm text-gray-500 mt-0.5 capitalize">{b.commodity_type} · {b.pallet_count} pallets</p>
                </div>
                <div className="text-right">
                  <div className="text-lg font-bold text-brand-600">{formatINR(b.total_amount_inr)}</div>
                  <div className="text-xs text-gray-400">commission {formatINR(b.commission_amount_inr)}</div>
                </div>
              </div>

              <div className="mt-4 flex gap-6 text-sm text-gray-600">
                <div>
                  <span className="text-gray-400">Inward</span>
                  <div className="font-medium">{formatDate(b.start_date)}</div>
                </div>
                <div>
                  <span className="text-gray-400">Expected Outward</span>
                  <div className="font-medium">{formatDate(b.end_date)}</div>
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

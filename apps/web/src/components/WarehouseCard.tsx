import { StarIcon, MapPinIcon, CheckBadgeIcon } from '@heroicons/react/24/solid'
import { ArrowRightIcon } from '@heroicons/react/24/outline'
import { useNavigate } from 'react-router-dom'
import type { Warehouse } from '../types'
import { warehouseTypeLabel } from '../utils/format'

interface Props {
  warehouse: Warehouse
  selected?: boolean
  onSelect?: () => void
}

export default function WarehouseCard({ warehouse: w, selected, onSelect }: Props) {
  const navigate = useNavigate()
  const occupancyPct = Math.round((1 - w.available_pallet_slots / w.total_pallet_capacity) * 100)

  return (
    <div
      className={`card cursor-pointer transition-all ${selected ? 'ring-2 ring-brand-500' : 'hover:shadow-md'}`}
      onClick={onSelect}
    >
      <div className="flex justify-between items-start gap-2">
        <div className="min-w-0">
          <h3 className="font-semibold text-gray-900 text-sm truncate">{w.name}</h3>
          <div className="flex items-center gap-1 text-xs text-gray-500 mt-0.5">
            <MapPinIcon className="h-3 w-3 flex-shrink-0" />
            <span className="truncate">{w.city}</span>
            {w.distance_km !== undefined && (
              <span className="ml-1 text-gray-400">· {w.distance_km.toFixed(1)} km</span>
            )}
          </div>
        </div>
        <div className="text-right flex-shrink-0">
          <div className="text-sm font-bold text-brand-600">
            ₹{w.price_per_pallet_per_day_inr.toFixed(0)}
          </div>
          <div className="text-xs text-gray-400">/pallet/day</div>
        </div>
      </div>

      <div className="mt-3 flex items-center gap-3 text-xs text-gray-600">
        <span className="bg-gray-100 rounded px-1.5 py-0.5">{warehouseTypeLabel[w.type]}</span>
        {w.gst_registered && (
          <span className="flex items-center gap-0.5 text-green-700">
            <CheckBadgeIcon className="h-3.5 w-3.5" /> GST Verified
          </span>
        )}
        {w.min_temperature_celsius !== undefined && (
          <span className="text-blue-600">
            {w.min_temperature_celsius}°–{w.max_temperature_celsius}°C
          </span>
        )}
      </div>

      <div className="mt-3 flex items-center justify-between">
        <div>
          <div className="flex items-center gap-1">
            <StarIcon className="h-3.5 w-3.5 text-yellow-400" />
            <span className="text-xs font-medium">{w.rating.toFixed(1)}</span>
            <span className="text-xs text-gray-400">({w.total_reviews})</span>
          </div>
          <div className="text-xs text-gray-500 mt-0.5">
            {w.available_pallet_slots} pallets free · {occupancyPct}% occupied
          </div>
        </div>

        {w.match_score !== undefined && (
          <div className="text-right">
            <div className="text-xs text-gray-400">Match</div>
            <div className={`text-sm font-bold ${w.match_score > 0.7 ? 'text-green-600' : 'text-orange-500'}`}>
              {Math.round(w.match_score * 100)}%
            </div>
          </div>
        )}
      </div>

      <button
        className="mt-4 btn-primary w-full justify-center text-xs py-1.5"
        onClick={(e) => { e.stopPropagation(); navigate(`/warehouses/${w.id}`) }}
      >
        View & Book <ArrowRightIcon className="h-3.5 w-3.5 ml-1" />
      </button>
    </div>
  )
}

import { StarIcon, MapPinIcon, CheckBadgeIcon } from '@heroicons/react/24/solid'
import { ArrowRightIcon } from '@heroicons/react/24/outline'
import { useNavigate } from 'react-router-dom'
import type { Warehouse } from '../types'
import { formatINR, warehouseTypeLabel } from '../utils/format'

interface Props {
  warehouse: Warehouse
  selected?: boolean
  onSelect?: () => void
}

export default function WarehouseCard({ warehouse: w, selected, onSelect }: Props) {
  const navigate = useNavigate()
  const occupancyPct = Math.round((1 - w.availablePallets / w.totalPallets) * 100)

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
            {w.distanceKm !== undefined && (
              <span className="ml-1 text-gray-400">· {w.distanceKm.toFixed(1)} km</span>
            )}
          </div>
        </div>
        <div className="text-right flex-shrink-0">
          <div className="text-sm font-bold text-brand-600">
            {formatINR(w.dynamicPrice ?? w.basePricePerMtPerMonth)}
          </div>
          <div className="text-xs text-gray-400">/MT/mo</div>
        </div>
      </div>

      <div className="mt-3 flex items-center gap-3 text-xs text-gray-600">
        <span className="bg-gray-100 rounded px-1.5 py-0.5">{warehouseTypeLabel[w.warehouseType]}</span>
        {w.wdraRegistered && (
          <span className="flex items-center gap-0.5 text-green-700">
            <CheckBadgeIcon className="h-3.5 w-3.5" /> WDRA
          </span>
        )}
        {w.minTemperatureCelsius !== undefined && (
          <span className="text-blue-600">
            {w.minTemperatureCelsius}°–{w.maxTemperatureCelsius}°C
          </span>
        )}
      </div>

      <div className="mt-3 flex items-center justify-between">
        <div>
          <div className="flex items-center gap-1">
            <StarIcon className="h-3.5 w-3.5 text-yellow-400" />
            <span className="text-xs font-medium">{w.rating.toFixed(1)}</span>
            <span className="text-xs text-gray-400">({w.reviewCount})</span>
          </div>
          <div className="text-xs text-gray-500 mt-0.5">
            {w.availablePallets} pallets free · {occupancyPct}% occupied
          </div>
        </div>

        {w.matchScore !== undefined && (
          <div className="text-right">
            <div className="text-xs text-gray-400">Match</div>
            <div className={`text-sm font-bold ${w.matchScore > 0.7 ? 'text-green-600' : 'text-orange-500'}`}>
              {Math.round(w.matchScore * 100)}%
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

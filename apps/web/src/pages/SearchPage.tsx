import { useState, useCallback } from 'react'
import Map, { Marker, Popup, NavigationControl } from 'react-map-gl'
import { MagnifyingGlassIcon, AdjustmentsHorizontalIcon } from '@heroicons/react/24/outline'
import { MapPinIcon } from '@heroicons/react/24/solid'
import { useSearchStore } from '../store/searchStore'
import { useWarehouseSearch } from '../hooks/useWarehouses'
import WarehouseCard from '../components/WarehouseCard'
import type { Warehouse } from '../types'

const MAPBOX_TOKEN = import.meta.env.VITE_MAPBOX_TOKEN as string

export default function SearchPage() {
  const { params, setParams, selectedWarehouse, selectWarehouse } = useSearchStore()
  const [searchEnabled, setSearchEnabled] = useState(false)
  const [showFilters, setShowFilters] = useState(false)

  const { data: warehouses = [], isFetching } = useWarehouseSearch(params, searchEnabled)

  const handleSearch = useCallback(() => setSearchEnabled(true), [])

  const handleMapClick = useCallback((e: mapboxgl.MapMouseEvent) => {
    setParams({ latitude: e.lngLat.lat, longitude: e.lngLat.lng })
    setSearchEnabled(false)
  }, [setParams])

  return (
    <div className="flex h-full">
      {/* Left panel */}
      <div className="w-96 flex flex-col border-r border-gray-200 bg-white">
        {/* Search controls */}
        <div className="p-4 border-b border-gray-100 space-y-3">
          <div className="flex gap-2">
            <button
              className="btn-primary flex-1 justify-center"
              onClick={handleSearch}
              disabled={isFetching}
            >
              <MagnifyingGlassIcon className="h-4 w-4 mr-1.5" />
              {isFetching ? 'Searching…' : 'Search Area'}
            </button>
            <button
              className="btn-secondary px-3"
              onClick={() => setShowFilters(!showFilters)}
            >
              <AdjustmentsHorizontalIcon className="h-4 w-4" />
            </button>
          </div>

          {showFilters && (
            <div className="space-y-3 pt-2 border-t border-gray-100">
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="block text-xs font-medium text-gray-600 mb-1">Radius (km)</label>
                  <input
                    type="number"
                    className="input"
                    value={params.radiusKm}
                    min={5}
                    max={200}
                    onChange={(e) => setParams({ radiusKm: +e.target.value })}
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-gray-600 mb-1">Pallets Needed</label>
                  <input
                    type="number"
                    className="input"
                    value={params.requiredPallets}
                    min={1}
                    onChange={(e) => setParams({ requiredPallets: +e.target.value })}
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Max Price (₹/pallet/day)</label>
                <input
                  type="number"
                  className="input"
                  value={params.maxPriceInr ?? ''}
                  placeholder="No limit"
                  onChange={(e) => setParams({ maxPriceInr: e.target.value ? +e.target.value : undefined })}
                />
              </div>

              <label className="flex items-center gap-2 text-sm text-gray-700">
                <input
                  type="checkbox"
                  checked={params.needsColdChain}
                  onChange={(e) => setParams({ needsColdChain: e.target.checked })}
                  className="rounded border-gray-300 text-brand-600"
                />
                Temperature-controlled (cold chain)
              </label>

              {params.needsColdChain && (
                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="block text-xs font-medium text-gray-600 mb-1">Min Temp (°C)</label>
                    <input
                      type="number"
                      className="input"
                      value={params.minTemp ?? ''}
                      onChange={(e) => setParams({ minTemp: e.target.value ? +e.target.value : undefined })}
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-gray-600 mb-1">Max Temp (°C)</label>
                    <input
                      type="number"
                      className="input"
                      value={params.maxTemp ?? ''}
                      onChange={(e) => setParams({ maxTemp: e.target.value ? +e.target.value : undefined })}
                    />
                  </div>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Results list */}
        <div className="flex-1 overflow-y-auto p-4 space-y-3">
          {!searchEnabled && (
            <p className="text-sm text-gray-500 text-center py-8">
              Click on the map to set location, then search.
            </p>
          )}
          {searchEnabled && warehouses.length === 0 && !isFetching && (
            <p className="text-sm text-gray-500 text-center py-8">
              No warehouses found in this area. Try expanding the radius.
            </p>
          )}
          {warehouses.map((w) => (
            <WarehouseCard
              key={w.id}
              warehouse={w}
              selected={selectedWarehouse?.id === w.id}
              onSelect={() => selectWarehouse(w)}
            />
          ))}
        </div>

        <div className="p-3 border-t border-gray-100 text-xs text-gray-400 text-center">
          {warehouses.length > 0 && `${warehouses.length} warehouses found`}
        </div>
      </div>

      {/* Map */}
      <div className="flex-1 relative">
        <Map
          mapboxAccessToken={MAPBOX_TOKEN}
          initialViewState={{
            longitude: params.longitude,
            latitude: params.latitude,
            zoom: 10,
          }}
          style={{ width: '100%', height: '100%' }}
          mapStyle="mapbox://styles/mapbox/light-v11"
          onClick={handleMapClick}
        >
          <NavigationControl position="top-right" />

          {/* Search center pin */}
          <Marker longitude={params.longitude} latitude={params.latitude} anchor="bottom">
            <div className="w-4 h-4 bg-blue-500 rounded-full border-2 border-white shadow-lg" />
          </Marker>

          {/* Warehouse markers */}
          {warehouses.map((w) => (
            <Marker
              key={w.id}
              longitude={w.longitude}
              latitude={w.latitude}
              anchor="bottom"
              onClick={(e) => { e.originalEvent.stopPropagation(); selectWarehouse(w) }}
            >
              <WarehouseMarker warehouse={w} selected={selectedWarehouse?.id === w.id} />
            </Marker>
          ))}

          {/* Popup for selected warehouse */}
          {selectedWarehouse && (
            <Popup
              longitude={selectedWarehouse.longitude}
              latitude={selectedWarehouse.latitude}
              anchor="top"
              closeOnClick={false}
              onClose={() => selectWarehouse(null)}
              className="z-10"
            >
              <div className="p-1 min-w-48">
                <div className="font-semibold text-sm">{selectedWarehouse.name}</div>
                <div className="text-xs text-gray-500">{selectedWarehouse.city}</div>
                <div className="text-sm font-bold text-brand-600 mt-1">
                  ₹{selectedWarehouse.price_per_pallet_per_day_inr.toFixed(0)}/pallet/day
                </div>
              </div>
            </Popup>
          )}
        </Map>
      </div>
    </div>
  )
}

function WarehouseMarker({ warehouse: w, selected }: { warehouse: Warehouse; selected: boolean }) {
  return (
    <div className={`flex flex-col items-center ${selected ? 'scale-125' : ''} transition-transform`}>
      <div
        className={`px-2 py-1 rounded-full text-xs font-bold shadow-md ${
          selected ? 'bg-brand-600 text-white' : 'bg-white text-brand-700 border border-brand-300'
        }`}
      >
        ₹{Math.round(w.price_per_pallet_per_day_inr)}
      </div>
      <MapPinIcon className={`h-4 w-4 -mt-0.5 ${selected ? 'text-brand-600' : 'text-brand-400'}`} />
    </div>
  )
}

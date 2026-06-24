import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import {
  StarIcon,
  CheckBadgeIcon,
  MapPinIcon,
  CalendarIcon,
  CubeIcon,
} from '@heroicons/react/24/solid'
import { useWarehouse, useCreateBooking } from '../hooks/useWarehouses'
import { formatINR, warehouseTypeLabel } from '../utils/format'

interface BookingForm {
  commodity: string
  palletCount: number
  inwardDate: string
  expectedOutwardDate: string
}

const COMMISION_RATE = 0.15

export default function WarehouseDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: warehouse, isLoading, error } = useWarehouse(id)
  const { mutateAsync: createBooking, isLoading: isBooking } = useCreateBooking()
  const [booked, setBooked] = useState(false)

  const { register, handleSubmit, watch, formState: { errors } } = useForm<BookingForm>({
    defaultValues: { commodity: 'potato', palletCount: 10 },
  })

  const palletCount = watch('palletCount') || 0
  const inDate = watch('inwardDate')
  const outDate = watch('expectedOutwardDate')

  const durationDays = inDate && outDate
    ? Math.max(1, Math.ceil((new Date(outDate).getTime() - new Date(inDate).getTime()) / 86400000))
    : 30

  const pricePerMonth = warehouse?.dynamicPrice ?? warehouse?.basePricePerMtPerMonth ?? 0
  const totalAmount = Math.round((pricePerMonth / 30) * durationDays * palletCount * 0.5)
  const commission = Math.round(totalAmount * COMMISION_RATE)

  const onSubmit = async (form: BookingForm) => {
    if (!id) return
    await createBooking({ warehouseId: id, ...form })
    setBooked(true)
  }

  if (isLoading) return <div className="p-8 text-gray-500">Loading…</div>
  if (error || !warehouse) return <div className="p-8 text-red-600">Warehouse not found.</div>
  if (booked) {
    return (
      <div className="p-8 text-center">
        <div className="text-4xl mb-4">✓</div>
        <h2 className="text-xl font-bold text-green-700">Booking Confirmed!</h2>
        <p className="text-gray-600 mt-2">Your booking has been created. The operator will confirm shortly.</p>
        <button className="btn-primary mt-6" onClick={() => navigate('/bookings')}>
          View My Bookings
        </button>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto p-6 space-y-6">
      {/* Header */}
      <div className="card">
        <div className="flex justify-between items-start">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">{warehouse.name}</h1>
            <div className="flex items-center gap-1.5 text-gray-500 mt-1">
              <MapPinIcon className="h-4 w-4" />
              <span>{warehouse.address}, {warehouse.city}, {warehouse.state}</span>
            </div>
          </div>
          <div className="text-right">
            <div className="text-2xl font-bold text-brand-600">
              {formatINR(pricePerMonth)}
            </div>
            <div className="text-sm text-gray-400">/MT/month</div>
          </div>
        </div>

        <div className="mt-4 flex flex-wrap gap-3 text-sm">
          <span className="bg-gray-100 rounded-full px-3 py-1">{warehouseTypeLabel[warehouse.warehouseType]}</span>
          {warehouse.wdraRegistered && (
            <span className="flex items-center gap-1 text-green-700 bg-green-50 rounded-full px-3 py-1">
              <CheckBadgeIcon className="h-4 w-4" /> WDRA Registered
            </span>
          )}
          {warehouse.apmcLicensed && (
            <span className="bg-blue-50 text-blue-700 rounded-full px-3 py-1">APMC Licensed</span>
          )}
          {warehouse.minTemperatureCelsius !== undefined && (
            <span className="bg-cyan-50 text-cyan-700 rounded-full px-3 py-1">
              {warehouse.minTemperatureCelsius}°C – {warehouse.maxTemperatureCelsius}°C
            </span>
          )}
        </div>

        <div className="mt-4 grid grid-cols-3 gap-4">
          <div className="text-center">
            <div className="flex items-center justify-center gap-1">
              <StarIcon className="h-4 w-4 text-yellow-400" />
              <span className="font-bold">{warehouse.rating.toFixed(1)}</span>
            </div>
            <div className="text-xs text-gray-500">{warehouse.reviewCount} reviews</div>
          </div>
          <div className="text-center">
            <div className="flex items-center justify-center gap-1">
              <CubeIcon className="h-4 w-4 text-brand-500" />
              <span className="font-bold">{warehouse.availablePallets}</span>
            </div>
            <div className="text-xs text-gray-500">pallets available</div>
          </div>
          <div className="text-center">
            <div className="flex items-center justify-center gap-1">
              <CalendarIcon className="h-4 w-4 text-purple-500" />
              <span className="font-bold">{warehouse.totalCapacityMt} MT</span>
            </div>
            <div className="text-xs text-gray-500">total capacity</div>
          </div>
        </div>
      </div>

      {/* Booking form */}
      <div className="card">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Book Storage</h2>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Commodity</label>
              <select
                className="input"
                {...register('commodity', { required: true })}
              >
                <option value="potato">Potato</option>
                <option value="onion">Onion</option>
                <option value="garlic">Garlic</option>
                <option value="fruits">Fruits</option>
                <option value="vegetables">Vegetables</option>
                <option value="grains">Grains</option>
                <option value="pharma">Pharma</option>
                <option value="fmcg">FMCG</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Pallets Required</label>
              <input
                type="number"
                className="input"
                min={1}
                max={warehouse.availablePallets}
                {...register('palletCount', { required: true, min: 1, max: warehouse.availablePallets, valueAsNumber: true })}
              />
              {errors.palletCount && <p className="text-xs text-red-500 mt-1">Max {warehouse.availablePallets} pallets</p>}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Inward Date</label>
              <input
                type="date"
                className="input"
                {...register('inwardDate', { required: true })}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Expected Outward</label>
              <input
                type="date"
                className="input"
                {...register('expectedOutwardDate', { required: true })}
              />
            </div>
          </div>

          {/* Price breakdown */}
          <div className="bg-gray-50 rounded-lg p-4 space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-600">Storage ({durationDays} days × {palletCount} pallets)</span>
              <span>{formatINR(totalAmount)}</span>
            </div>
            <div className="flex justify-between text-gray-500">
              <span>Platform commission (15%)</span>
              <span>–{formatINR(commission)}</span>
            </div>
            <div className="flex justify-between font-semibold border-t border-gray-200 pt-2">
              <span>Operator receives</span>
              <span className="text-brand-600">{formatINR(totalAmount - commission)}</span>
            </div>
          </div>

          <button
            type="submit"
            className="btn-primary w-full justify-center"
            disabled={isBooking}
          >
            {isBooking ? 'Creating booking…' : `Confirm Booking — ${formatINR(totalAmount)}`}
          </button>
        </form>
      </div>
    </div>
  )
}

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
  commodity_type: string
  pallet_count: number
  start_date: string
  end_date: string
}

const COMMISION_RATE = 0.10

export default function WarehouseDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: warehouse, isLoading, error } = useWarehouse(id)
  const { mutateAsync: createBooking, isLoading: isBooking } = useCreateBooking()
  const [booked, setBooked] = useState(false)

  const { register, handleSubmit, watch, formState: { errors } } = useForm<BookingForm>({
    defaultValues: { commodity_type: 'ecommerce_fmcg', pallet_count: 10 },
  })

  const palletCount = watch('pallet_count') || 0
  const inDate = watch('start_date')
  const outDate = watch('end_date')

  const durationDays = inDate && outDate
    ? Math.max(1, Math.ceil((new Date(outDate).getTime() - new Date(inDate).getTime()) / 86400000))
    : 30

  const pricePerPalletDay = warehouse?.price_per_pallet_per_day_inr ?? 0
  const totalAmount = Math.round(pricePerPalletDay * durationDays * palletCount)
  const commission = Math.round(totalAmount * COMMISION_RATE)

  const onSubmit = async (form: BookingForm) => {
    if (!id) return
    await createBooking({
      warehouseId: id,
      palletCount: form.pallet_count,
      commodity: form.commodity_type,
      inwardDate: form.start_date,
      expectedOutwardDate: form.end_date,
    })
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
              <span>{warehouse.address_line1}, {warehouse.city}, {warehouse.state}</span>
            </div>
          </div>
          <div className="text-right">
            <div className="text-2xl font-bold text-brand-600">
              ₹{pricePerPalletDay.toFixed(0)}
            </div>
            <div className="text-sm text-gray-400">/pallet/day</div>
          </div>
        </div>

        <div className="mt-4 flex flex-wrap gap-3 text-sm">
          <span className="bg-gray-100 rounded-full px-3 py-1">{warehouseTypeLabel[warehouse.type]}</span>
          {warehouse.gst_registered && (
            <span className="flex items-center gap-1 text-green-700 bg-green-50 rounded-full px-3 py-1">
              <CheckBadgeIcon className="h-4 w-4" /> GST Verified
            </span>
          )}
          {warehouse.min_temperature_celsius !== undefined && (
            <span className="bg-cyan-50 text-cyan-700 rounded-full px-3 py-1">
              {warehouse.min_temperature_celsius}°C – {warehouse.max_temperature_celsius}°C
            </span>
          )}
        </div>

        <div className="mt-4 grid grid-cols-3 gap-4">
          <div className="text-center">
            <div className="flex items-center justify-center gap-1">
              <StarIcon className="h-4 w-4 text-yellow-400" />
              <span className="font-bold">{warehouse.rating.toFixed(1)}</span>
            </div>
            <div className="text-xs text-gray-500">{warehouse.total_reviews} reviews</div>
          </div>
          <div className="text-center">
            <div className="flex items-center justify-center gap-1">
              <CubeIcon className="h-4 w-4 text-brand-500" />
              <span className="font-bold">{warehouse.available_pallet_slots}</span>
            </div>
            <div className="text-xs text-gray-500">pallets available</div>
          </div>
          <div className="text-center">
            <div className="flex items-center justify-center gap-1">
              <CalendarIcon className="h-4 w-4 text-purple-500" />
              <span className="font-bold">{warehouse.total_pallet_capacity}</span>
            </div>
            <div className="text-xs text-gray-500">total capacity (pallets)</div>
          </div>
        </div>
      </div>

      {/* Booking form */}
      <div className="card">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Book Storage</h2>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Goods / Commodity</label>
              <select
                className="input"
                {...register('commodity_type', { required: true })}
              >
                <optgroup label="Food & Agriculture">
                  <option value="fruits_vegetables">Fruits &amp; Vegetables</option>
                  <option value="grains_pulses">Grains &amp; Pulses</option>
                  <option value="dairy_frozen">Dairy &amp; Frozen Foods</option>
                </optgroup>
                <optgroup label="Industrial & Manufacturing">
                  <option value="electronics">Electronics &amp; Components</option>
                  <option value="auto_parts">Automobile Parts</option>
                  <option value="machinery">Machinery &amp; Equipment</option>
                  <option value="chemicals">Chemicals &amp; Raw Materials</option>
                  <option value="furniture">Furniture &amp; White Goods</option>
                </optgroup>
                <optgroup label="Retail & E-Commerce">
                  <option value="ecommerce_fmcg">FMCG / E-Commerce</option>
                  <option value="apparel_textiles">Apparel &amp; Textiles</option>
                  <option value="pharmaceutical">Pharmaceutical</option>
                </optgroup>
                <optgroup label="Personal Storage">
                  <option value="household_goods">Household Goods</option>
                  <option value="documents_archives">Documents &amp; Archives</option>
                  <option value="other">Other</option>
                </optgroup>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Pallets Required</label>
              <input
                type="number"
                className="input"
                min={1}
                max={warehouse.available_pallet_slots}
                {...register('pallet_count', { required: true, min: 1, max: warehouse.available_pallet_slots, valueAsNumber: true })}
              />
              {errors.pallet_count && <p className="text-xs text-red-500 mt-1">Max {warehouse.available_pallet_slots} pallets</p>}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Inward Date</label>
              <input
                type="date"
                className="input"
                {...register('start_date', { required: true })}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Expected Outward</label>
              <input
                type="date"
                className="input"
                {...register('end_date', { required: true })}
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
              <span>Platform fee (10%)</span>
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

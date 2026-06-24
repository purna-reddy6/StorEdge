import { useQuery } from 'react-query'
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  BarChart, Bar, Legend,
} from 'recharts'
import { formatINR, formatDate, formatOccupancy } from '../utils/format'
import api from '../utils/api'
import type { OccupancyStat, Booking, IoTAlert } from '../types'

function useDashboard() {
  const occupancy = useQuery<OccupancyStat[]>('operator-occupancy', () =>
    api.get('/operator/occupancy').then((r) => r.data.stats)
  )
  const bookings = useQuery<Booking[]>('operator-bookings', () =>
    api.get('/operator/bookings').then((r) => r.data.bookings)
  )
  const alerts = useQuery<IoTAlert[]>('operator-alerts', () =>
    api.get('/operator/alerts').then((r) => r.data.alerts)
  )
  return { occupancy, bookings, alerts }
}

export default function OperatorDashboard() {
  const { occupancy, bookings, alerts } = useDashboard()

  const totalRevenue = (occupancy.data ?? []).reduce((s, d) => s + d.revenue, 0)
  const avgOccupancy =
    (occupancy.data ?? []).reduce((s, d) => s + d.occupancyPct, 0) /
    Math.max(1, occupancy.data?.length ?? 1)

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">Operator Dashboard</h1>

      {/* KPI row */}
      <div className="grid grid-cols-4 gap-4">
        <KpiCard label="30-Day Revenue" value={formatINR(totalRevenue)} sub="after platform commission" />
        <KpiCard label="Avg Occupancy" value={formatOccupancy(avgOccupancy / 100)} sub="last 30 days" />
        <KpiCard label="Active Bookings" value={String(bookings.data?.filter(b => b.status === 'active').length ?? '—')} sub="tenants in-facility" />
        <KpiCard
          label="Open Alerts"
          value={String(alerts.data?.filter(a => !a.resolvedAt).length ?? '—')}
          sub="IoT / sensor"
          accent={alerts.data?.some(a => !a.resolvedAt && a.severity === 'critical') ? 'red' : 'default'}
        />
      </div>

      {/* Charts row */}
      <div className="grid grid-cols-2 gap-6">
        <div className="card">
          <h2 className="text-sm font-semibold text-gray-700 mb-4">Occupancy Rate (30 days)</h2>
          {occupancy.data ? (
            <ResponsiveContainer width="100%" height={200}>
              <AreaChart data={occupancy.data}>
                <defs>
                  <linearGradient id="occ" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#16a34a" stopOpacity={0.2} />
                    <stop offset="95%" stopColor="#16a34a" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#f3f4f6" />
                <XAxis dataKey="date" tick={{ fontSize: 10 }} tickFormatter={(d) => d.slice(5)} />
                <YAxis tick={{ fontSize: 10 }} tickFormatter={(v) => `${v}%`} domain={[0, 100]} />
                <Tooltip formatter={(v: number) => [`${v}%`, 'Occupancy']} />
                <Area type="monotone" dataKey="occupancyPct" stroke="#16a34a" fill="url(#occ)" strokeWidth={2} />
              </AreaChart>
            </ResponsiveContainer>
          ) : <Skeleton />}
        </div>

        <div className="card">
          <h2 className="text-sm font-semibold text-gray-700 mb-4">Daily Revenue (₹)</h2>
          {occupancy.data ? (
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={occupancy.data}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f3f4f6" />
                <XAxis dataKey="date" tick={{ fontSize: 10 }} tickFormatter={(d) => d.slice(5)} />
                <YAxis tick={{ fontSize: 10 }} tickFormatter={(v) => `₹${(v / 1000).toFixed(0)}k`} />
                <Tooltip formatter={(v: number) => [formatINR(v), 'Revenue']} />
                <Legend />
                <Bar dataKey="revenue" fill="#16a34a" name="Revenue" radius={[3, 3, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          ) : <Skeleton />}
        </div>
      </div>

      {/* Recent bookings & alerts */}
      <div className="grid grid-cols-2 gap-6">
        <div className="card">
          <h2 className="text-sm font-semibold text-gray-700 mb-3">Recent Bookings</h2>
          {(bookings.data ?? []).slice(0, 5).map((b) => (
            <div key={b.id} className="flex justify-between items-center py-2 border-b border-gray-50 last:border-0">
              <div>
                <div className="text-sm font-medium text-gray-900">{b.farmerName}</div>
                <div className="text-xs text-gray-500 capitalize">{b.commodity} · {b.palletCount} pallets</div>
              </div>
              <div className="text-right">
                <div className="text-sm font-semibold">{formatINR(b.totalAmountInr - b.commissionInr)}</div>
                <div className="text-xs text-gray-400">{formatDate(b.inwardDate)}</div>
              </div>
            </div>
          ))}
        </div>

        <div className="card">
          <h2 className="text-sm font-semibold text-gray-700 mb-3">IoT Alerts</h2>
          {(alerts.data ?? []).slice(0, 5).map((a) => (
            <div key={a.id} className="flex items-start gap-3 py-2 border-b border-gray-50 last:border-0">
              <AlertDot severity={a.severity} />
              <div className="min-w-0">
                <div className="text-sm font-medium text-gray-900 capitalize">
                  {a.alertType.replace(/_/g, ' ')}
                </div>
                <div className="text-xs text-gray-500 truncate">{a.message}</div>
                <div className="text-xs text-gray-400">{formatDate(a.createdAt)}</div>
              </div>
              {!a.resolvedAt && (
                <span className="badge-red flex-shrink-0">Open</span>
              )}
            </div>
          ))}
          {(alerts.data ?? []).length === 0 && (
            <p className="text-sm text-gray-400 text-center py-4">No active alerts</p>
          )}
        </div>
      </div>
    </div>
  )
}

function KpiCard({ label, value, sub, accent = 'default' }: {
  label: string; value: string; sub: string; accent?: 'red' | 'default'
}) {
  return (
    <div className="card">
      <div className="text-xs font-medium text-gray-500 uppercase tracking-wide">{label}</div>
      <div className={`text-2xl font-bold mt-1 ${accent === 'red' ? 'text-red-600' : 'text-gray-900'}`}>{value}</div>
      <div className="text-xs text-gray-400 mt-0.5">{sub}</div>
    </div>
  )
}

function AlertDot({ severity }: { severity: IoTAlert['severity'] }) {
  const cls = severity === 'critical' ? 'bg-red-500' : severity === 'warning' ? 'bg-yellow-400' : 'bg-blue-400'
  return <div className={`w-2 h-2 rounded-full flex-shrink-0 mt-1.5 ${cls}`} />
}

function Skeleton() {
  return <div className="h-48 bg-gray-100 rounded animate-pulse" />
}

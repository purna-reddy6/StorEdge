import { useQuery, useMutation, useQueryClient } from 'react-query'
import { BellAlertIcon, CheckCircleIcon } from '@heroicons/react/24/outline'
import api from '../utils/api'
import { formatDate } from '../utils/format'
import type { IoTAlert } from '../types'

function useAlerts() {
  return useQuery<IoTAlert[]>('iot-alerts', () =>
    api.get('/iot/alerts').then((r) => r.data.alerts),
    { refetchInterval: 30_000 }
  )
}

const severityConfig = {
  info:     { cls: 'bg-blue-50 border-blue-200',  dot: 'bg-blue-400',   label: 'Info' },
  warning:  { cls: 'bg-yellow-50 border-yellow-200', dot: 'bg-yellow-400', label: 'Warning' },
  critical: { cls: 'bg-red-50 border-red-200',    dot: 'bg-red-500',    label: 'Critical' },
}

export default function AlertsPage() {
  const qc = useQueryClient()
  const { data: alerts = [], isLoading } = useAlerts()
  const { mutate: resolve } = useMutation(
    (alertId: string) => api.patch(`/iot/alerts/${alertId}/resolve`),
    { onSuccess: () => qc.invalidateQueries('iot-alerts') }
  )

  const open = alerts.filter(a => !a.resolvedAt)
  const resolved = alerts.filter(a => a.resolvedAt)

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">IoT Alerts</h1>
        {open.length > 0 && (
          <span className="badge-red px-3 py-1 text-sm">
            {open.length} open alert{open.length > 1 ? 's' : ''}
          </span>
        )}
      </div>

      {isLoading && <p className="text-gray-500">Loading alerts…</p>}

      {open.length === 0 && !isLoading && (
        <div className="card flex items-center gap-3 text-green-700">
          <CheckCircleIcon className="h-6 w-6" />
          <span className="font-medium">All clear — no open alerts</span>
        </div>
      )}

      {/* Open alerts */}
      {open.length > 0 && (
        <div className="space-y-3">
          <h2 className="text-sm font-semibold text-gray-600 uppercase tracking-wide">Open</h2>
          {open.map((a) => <AlertRow key={a.id} alert={a} onResolve={() => resolve(a.id)} />)}
        </div>
      )}

      {/* Resolved */}
      {resolved.length > 0 && (
        <div className="space-y-3">
          <h2 className="text-sm font-semibold text-gray-600 uppercase tracking-wide">Resolved</h2>
          {resolved.slice(0, 10).map((a) => <AlertRow key={a.id} alert={a} />)}
        </div>
      )}
    </div>
  )
}

function AlertRow({ alert: a, onResolve }: { alert: IoTAlert; onResolve?: () => void }) {
  const sc = severityConfig[a.severity]
  return (
    <div className={`border rounded-lg p-4 ${sc.cls}`}>
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-start gap-3">
          <div className={`w-2.5 h-2.5 rounded-full flex-shrink-0 mt-1.5 ${sc.dot}`} />
          <div>
            <div className="flex items-center gap-2">
              <span className="font-semibold text-gray-900 capitalize">
                {a.alertType.replace(/_/g, ' ')}
              </span>
              <span className="text-xs text-gray-500">{sc.label}</span>
            </div>
            <p className="text-sm text-gray-700 mt-0.5">{a.message}</p>
            <div className="text-xs text-gray-500 mt-1 flex gap-3">
              <span>Sensor: {a.sensorId}</span>
              <span>{formatDate(a.createdAt)}</span>
              {a.resolvedAt && <span className="text-green-600">Resolved {formatDate(a.resolvedAt)}</span>}
            </div>
          </div>
        </div>

        {!a.resolvedAt && onResolve && (
          <button
            className="btn-secondary text-xs py-1 px-2 flex-shrink-0"
            onClick={onResolve}
          >
            <BellAlertIcon className="h-3.5 w-3.5 mr-1" /> Resolve
          </button>
        )}
      </div>
    </div>
  )
}

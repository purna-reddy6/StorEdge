import { useQuery } from 'react-query'
import { CheckCircleIcon, DocumentTextIcon } from '@heroicons/react/24/solid'
import api from '../utils/api'
import { formatINR, formatDate } from '../utils/format'
import type { EnwrReceipt } from '../types'

function useReceipts() {
  return useQuery<EnwrReceipt[]>('enwrs', () =>
    api.get('/financing/receipts').then((r) => r.data.receipts)
  )
}

const statusConfig = {
  draft:   { label: 'Draft',   cls: 'badge-yellow' },
  issued:  { label: 'Issued',  cls: 'badge-green' },
  pledged: { label: 'Pledged', cls: 'bg-purple-100 text-purple-800 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium' },
  released:{ label: 'Released',cls: 'bg-gray-100 text-gray-700 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium' },
  expired: { label: 'Expired', cls: 'badge-red' },
}

export default function FinancingPage() {
  const { data: receipts = [], isLoading } = useReceipts()

  const totalMarketValue = receipts.reduce((s, r) => s + r.marketValueInr, 0)
  const totalLoanEligible = receipts.reduce((s, r) => s + r.maxLoanAmountInr, 0)
  const activeReceipts = receipts.filter(r => r.status === 'issued' || r.status === 'pledged')

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">e-NWR & Financing</h1>

      {/* Info banner */}
      <div className="bg-brand-50 border border-brand-200 rounded-xl p-4 flex gap-4">
        <CheckCircleIcon className="h-6 w-6 text-brand-600 flex-shrink-0 mt-0.5" />
        <div>
          <p className="text-sm font-semibold text-brand-800">
            PSL Loan Limit: ₹75 lakh per e-NWR (vs ₹50L for paper NWR)
          </p>
          <p className="text-sm text-brand-700 mt-1">
            Electronic Negotiable Warehouse Receipts issued via NERL repository. LTV ratio: 70% of market value.
            Origination fee: 1.5%. Processed within 24 hours.
          </p>
        </div>
      </div>

      {/* Summary stats */}
      <div className="grid grid-cols-3 gap-4">
        <div className="card">
          <div className="text-xs text-gray-500 uppercase tracking-wide">Total Market Value</div>
          <div className="text-xl font-bold text-gray-900 mt-1">{formatINR(totalMarketValue)}</div>
          <div className="text-xs text-gray-400 mt-0.5">across {receipts.length} receipts</div>
        </div>
        <div className="card">
          <div className="text-xs text-gray-500 uppercase tracking-wide">Max Loan Eligible</div>
          <div className="text-xl font-bold text-brand-600 mt-1">{formatINR(totalLoanEligible)}</div>
          <div className="text-xs text-gray-400 mt-0.5">at 70% LTV</div>
        </div>
        <div className="card">
          <div className="text-xs text-gray-500 uppercase tracking-wide">Active Receipts</div>
          <div className="text-xl font-bold text-gray-900 mt-1">{activeReceipts.length}</div>
          <div className="text-xs text-gray-400 mt-0.5">issued or pledged</div>
        </div>
      </div>

      {/* Receipts table */}
      <div className="card">
        <h2 className="text-sm font-semibold text-gray-700 mb-4 flex items-center gap-2">
          <DocumentTextIcon className="h-4 w-4" /> Warehouse Receipts (e-NWR)
        </h2>

        {isLoading && <p className="text-gray-400 text-sm">Loading receipts…</p>}

        <div className="space-y-3">
          {receipts.map((r) => {
            const sc = statusConfig[r.status]
            return (
              <div key={r.id} className="border border-gray-100 rounded-lg p-4">
                <div className="flex justify-between items-start">
                  <div>
                    <div className="flex items-center gap-3">
                      <span className="font-mono text-sm text-gray-700">{r.receiptNumber}</span>
                      <span className={sc.cls}>{sc.label}</span>
                    </div>
                    <div className="text-sm font-semibold text-gray-900 mt-1 capitalize">
                      {r.commodity} — {r.quantityKg.toLocaleString()} kg
                    </div>
                    <div className="text-xs text-gray-500 mt-0.5">
                      Issued: {r.issuedAt ? formatDate(r.issuedAt) : ""} · Expires: {formatDate(r.expiryDate)}
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm font-bold text-gray-900">{formatINR(r.marketValueInr)}</div>
                    <div className="text-xs text-brand-600 mt-0.5">Loan up to {formatINR(r.maxLoanAmountInr)}</div>
                  </div>
                </div>

                {r.status === 'issued' && (
                  <div className="mt-3 flex gap-2">
                    <button className="btn-primary text-xs py-1.5 px-3">Apply for Loan</button>
                    <button className="btn-secondary text-xs py-1.5 px-3">Download e-NWR</button>
                  </div>
                )}
              </div>
            )
          })}

          {receipts.length === 0 && !isLoading && (
            <p className="text-sm text-gray-400 text-center py-6">
              No e-NWR receipts yet. Book storage and request a warehouse receipt after inward.
            </p>
          )}
        </div>
      </div>
    </div>
  )
}

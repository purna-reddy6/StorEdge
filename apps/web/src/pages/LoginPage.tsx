import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import api from '../utils/api'

type Step = 'phone' | 'otp'

export default function LoginPage() {
  const [step, setStep] = useState<Step>('phone')
  const [phone, setPhone] = useState('')
  const [otp, setOtp] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [devOtp, setDevOtp] = useState('')
  const { login } = useAuthStore()
  const navigate = useNavigate()

  const fullPhone = `+91${phone}`

  const sendOtp = async () => {
    setError('')
    setLoading(true)
    try {
      const { data } = await api.post('/auth/otp/send', { phone: fullPhone })
      if (data.dev_otp) setDevOtp(data.dev_otp)
      setStep('otp')
    } catch {
      setError('Failed to send OTP. Check the phone number.')
    } finally {
      setLoading(false)
    }
  }

  const verifyOtp = async () => {
    setError('')
    setLoading(true)
    try {
      const { data } = await api.post('/auth/otp/verify', { phone: fullPhone, otp })
      login(data.user, data.token)
      navigate('/search')
    } catch {
      setError('Invalid or expired OTP. Try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-brand-50 to-white flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-brand-700">StorEdge</h1>
          <p className="text-gray-600 mt-2">On-Demand Storage Marketplace</p>
        </div>

        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-6">
            {step === 'phone' ? 'Sign in with your phone' : 'Enter your OTP'}
          </h2>

          {step === 'phone' ? (
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Mobile Number
                </label>
                <div className="flex">
                  <span className="inline-flex items-center px-3 rounded-l-lg border border-r-0 border-gray-300 bg-gray-50 text-gray-500 text-sm">
                    +91
                  </span>
                  <input
                    type="tel"
                    className="input rounded-l-none flex-1"
                    placeholder="9876543210"
                    maxLength={10}
                    value={phone}
                    onChange={(e) => setPhone(e.target.value.replace(/\D/g, ''))}
                    onKeyDown={(e) => e.key === 'Enter' && phone.length === 10 && sendOtp()}
                  />
                </div>
              </div>
              {error && <p className="text-sm text-red-600">{error}</p>}
              <button
                className="btn-primary w-full justify-center"
                onClick={sendOtp}
                disabled={loading || phone.length !== 10}
              >
                {loading ? 'Sending…' : 'Send OTP'}
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              <p className="text-sm text-gray-600">
                OTP sent to <span className="font-medium">+91 {phone}</span>
              </p>
              {devOtp && (
                <div className="bg-yellow-50 border border-yellow-200 rounded-lg px-4 py-2 text-center">
                  <p className="text-xs text-yellow-700 font-medium">Dev mode — OTP:</p>
                  <p className="text-2xl font-mono font-bold text-yellow-800 tracking-widest">{devOtp}</p>
                </div>
              )}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">6-digit OTP</label>
                <input
                  type="text"
                  className="input text-center text-2xl tracking-widest"
                  placeholder="••••••"
                  maxLength={6}
                  value={otp}
                  onChange={(e) => setOtp(e.target.value.replace(/\D/g, ''))}
                  onKeyDown={(e) => e.key === 'Enter' && otp.length === 6 && verifyOtp()}
                  autoFocus
                />
              </div>
              {error && <p className="text-sm text-red-600">{error}</p>}
              <button
                className="btn-primary w-full justify-center"
                onClick={verifyOtp}
                disabled={loading || otp.length !== 6}
              >
                {loading ? 'Verifying…' : 'Verify & Sign In'}
              </button>
              <button
                className="text-sm text-brand-600 hover:underline w-full text-center"
                onClick={() => { setStep('phone'); setOtp(''); setError('') }}
              >
                Change phone number
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

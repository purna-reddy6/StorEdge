import React, { useState } from 'react'
import {
  View, Text, TextInput, TouchableOpacity, StyleSheet,
  KeyboardAvoidingView, Platform, ActivityIndicator, Alert,
} from 'react-native'
import { useAuthStore } from '../store/authStore'
import api from '../utils/api'

type Step = 'phone' | 'otp'

export default function LoginScreen() {
  const [step, setStep] = useState<Step>('phone')
  const [phone, setPhone] = useState('')
  const [otp, setOtp] = useState('')
  const [loading, setLoading] = useState(false)
  const { login } = useAuthStore()

  const sendOtp = async () => {
    setLoading(true)
    try {
      await api.post('/auth/otp/send', { phone: `+91${phone}` })
      setStep('otp')
    } catch {
      Alert.alert('Error', 'Failed to send OTP. Check your phone number.')
    } finally {
      setLoading(false)
    }
  }

  const verifyOtp = async () => {
    setLoading(true)
    try {
      const { data } = await api.post('/auth/otp/verify', { phone: `+91${phone}`, otp })
      login(data.user, data.token)
    } catch {
      Alert.alert('Invalid OTP', 'The code you entered is wrong or expired.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <KeyboardAvoidingView style={styles.container} behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
      <View style={styles.inner}>
        <Text style={styles.logo}>StorEdge</Text>
        <Text style={styles.tagline}>Agricultural Warehouse Marketplace</Text>

        <View style={styles.card}>
          {step === 'phone' ? (
            <>
              <Text style={styles.label}>Mobile Number</Text>
              <View style={styles.phoneRow}>
                <View style={styles.prefix}>
                  <Text style={styles.prefixText}>+91</Text>
                </View>
                <TextInput
                  style={[styles.input, styles.phoneInput]}
                  placeholder="9876543210"
                  keyboardType="phone-pad"
                  maxLength={10}
                  value={phone}
                  onChangeText={(t) => setPhone(t.replace(/\D/g, ''))}
                />
              </View>

              <TouchableOpacity
                style={[styles.btn, phone.length !== 10 && styles.btnDisabled]}
                onPress={sendOtp}
                disabled={loading || phone.length !== 10}
              >
                {loading ? <ActivityIndicator color="#fff" /> : <Text style={styles.btnText}>Send OTP</Text>}
              </TouchableOpacity>
            </>
          ) : (
            <>
              <Text style={styles.label}>Enter the 6-digit OTP sent to +91 {phone}</Text>
              <TextInput
                style={[styles.input, styles.otpInput]}
                placeholder="••••••"
                keyboardType="number-pad"
                maxLength={6}
                value={otp}
                onChangeText={(t) => setOtp(t.replace(/\D/g, ''))}
                autoFocus
              />

              <TouchableOpacity
                style={[styles.btn, otp.length !== 6 && styles.btnDisabled]}
                onPress={verifyOtp}
                disabled={loading || otp.length !== 6}
              >
                {loading ? <ActivityIndicator color="#fff" /> : <Text style={styles.btnText}>Verify & Sign In</Text>}
              </TouchableOpacity>

              <TouchableOpacity onPress={() => { setStep('phone'); setOtp('') }}>
                <Text style={styles.changePhone}>Change phone number</Text>
              </TouchableOpacity>
            </>
          )}
        </View>
      </View>
    </KeyboardAvoidingView>
  )
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f0fdf4' },
  inner: { flex: 1, justifyContent: 'center', padding: 24 },
  logo: { fontSize: 32, fontWeight: '700', color: '#16a34a', textAlign: 'center' },
  tagline: { fontSize: 14, color: '#6b7280', textAlign: 'center', marginTop: 4, marginBottom: 32 },
  card: { backgroundColor: '#fff', borderRadius: 16, padding: 24, shadowColor: '#000', shadowOpacity: 0.08, shadowRadius: 12, elevation: 4 },
  label: { fontSize: 14, fontWeight: '500', color: '#374151', marginBottom: 8 },
  phoneRow: { flexDirection: 'row', marginBottom: 16 },
  prefix: { backgroundColor: '#f3f4f6', borderWidth: 1, borderColor: '#d1d5db', borderRightWidth: 0, borderTopLeftRadius: 8, borderBottomLeftRadius: 8, paddingHorizontal: 12, justifyContent: 'center' },
  prefixText: { fontSize: 14, color: '#6b7280' },
  input: { borderWidth: 1, borderColor: '#d1d5db', borderRadius: 8, paddingHorizontal: 12, paddingVertical: 10, fontSize: 16, color: '#111827', backgroundColor: '#fff' },
  phoneInput: { flex: 1, borderTopLeftRadius: 0, borderBottomLeftRadius: 0 },
  otpInput: { textAlign: 'center', fontSize: 24, letterSpacing: 8, marginBottom: 16 },
  btn: { backgroundColor: '#16a34a', borderRadius: 8, paddingVertical: 14, alignItems: 'center', marginTop: 4 },
  btnDisabled: { opacity: 0.5 },
  btnText: { color: '#fff', fontSize: 16, fontWeight: '600' },
  changePhone: { textAlign: 'center', color: '#16a34a', marginTop: 16, fontSize: 14 },
})

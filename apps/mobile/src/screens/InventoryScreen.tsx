import React, { useEffect, useState } from 'react'
import {
  View, Text, StyleSheet, FlatList, RefreshControl,
  ActivityIndicator, Alert, TouchableOpacity, Modal, TextInput,
} from 'react-native'
import api from '../utils/api'
import { formatDate } from '../utils/format'

interface PalletItem {
  id: string; commodity: string
  weightKg: number; slotPosition: string
  inwardDate?: string; expectedOutwardDate?: string
  currentTemperatureCelsius?: number
  releaseStatus?: string
}

const riskColor: Record<string, string> = {
  safe: '#16a34a', low: '#16a34a', medium: '#f59e0b', high: '#ef4444', critical: '#991b1b',
}

export default function InventoryScreen() {
  const [pallets, setPallets] = useState<PalletItem[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [otpModal, setOtpModal] = useState<{ palletId: string; slot: string } | null>(null)
  const [otp, setOtp] = useState('')
  const [otpLoading, setOtpLoading] = useState(false)

  const load = async (refresh = false) => {
    if (refresh) setRefreshing(true)
    try {
      const { data } = await api.get('/inventory/pallets')
      setPallets(data.items ?? [])
    } catch {
      Alert.alert('Error', 'Could not load inventory.')
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => { load() }, [])

  const requestRelease = async (palletId: string) => {
    try {
      await api.post(`/inventory/pallets/${palletId}/release/initiate`)
      Alert.alert('OTP Sent', 'A 6-digit OTP has been sent to your WhatsApp. Enter it to authorize the release.')
      load()
    } catch {
      Alert.alert('Error', 'Failed to initiate release.')
    }
  }

  const submitOtp = async () => {
    if (!otpModal) return
    setOtpLoading(true)
    try {
      await api.post(`/inventory/pallets/${otpModal.palletId}/release/authorize`, { otp })
      Alert.alert('Release Authorized', 'Stock release has been authorized. The operator will prepare your goods.')
      setOtpModal(null)
      setOtp('')
      load()
    } catch {
      Alert.alert('Invalid OTP', 'The OTP you entered is wrong or expired.')
    } finally {
      setOtpLoading(false)
    }
  }

  if (loading) return <View style={styles.center}><ActivityIndicator color="#16a34a" size="large" /></View>

  return (
    <>
      <FlatList
        style={styles.container}
        data={pallets}
        keyExtractor={(p) => p.id}
        refreshControl={<RefreshControl refreshing={refreshing} onRefresh={() => load(true)} />}
        contentContainerStyle={{ padding: 16, gap: 12 }}
        ListHeaderComponent={<Text style={styles.heading}>My Inventory</Text>}
        ListEmptyComponent={<Text style={styles.empty}>No pallets in storage yet.</Text>}
        renderItem={({ item: p }) => (
          <View style={styles.card}>
            <View style={styles.topRow}>
              <Text style={styles.palletCode}>{p.slotPosition || '—'}</Text>
              <Text style={styles.slot}>{p.commodity.toUpperCase()}</Text>
            </View>
            <Text style={styles.commodity}>{p.commodity} — {p.weightKg.toLocaleString()} kg</Text>
            <View style={styles.datesRow}>
              {p.inwardDate && <Text style={styles.date}>In: {formatDate(p.inwardDate)}</Text>}
              {p.expectedOutwardDate && <Text style={styles.date}>Out: {formatDate(p.expectedOutwardDate)}</Text>}
              {p.currentTemperatureCelsius != null && (
                <Text style={styles.date}>🌡 {p.currentTemperatureCelsius.toFixed(1)}°C</Text>
              )}
            </View>

            <View style={styles.bottomRow}>
              {(p.releaseStatus === 'pending_otp' || p.releaseStatus === 'otp_sent') ? (
                <TouchableOpacity
                  style={styles.otpBtn}
                  onPress={() => { setOtpModal({ palletId: p.id, slot: p.slotPosition }); setOtp('') }}
                >
                  <Text style={styles.otpBtnText}>Enter OTP to Release</Text>
                </TouchableOpacity>
              ) : !p.releaseStatus && (
                <TouchableOpacity
                  style={styles.releaseBtn}
                  onPress={() => requestRelease(p.id)}
                >
                  <Text style={styles.releaseBtnText}>Request Release</Text>
                </TouchableOpacity>
              )}
            </View>
          </View>
        )}
      />

      {/* OTP Modal */}
      <Modal visible={!!otpModal} transparent animationType="slide">
        <View style={styles.modalOverlay}>
          <View style={styles.modalCard}>
            <Text style={styles.modalTitle}>Authorize Stock Release</Text>
            <Text style={styles.modalSub}>Slot: {otpModal?.slot}</Text>
            <Text style={styles.modalDesc}>Enter the 6-digit OTP sent to your WhatsApp to authorize release without visiting the warehouse.</Text>
            <TextInput
              style={styles.otpInput}
              placeholder="••••••"
              keyboardType="number-pad"
              maxLength={6}
              value={otp}
              onChangeText={(t) => setOtp(t.replace(/\D/g, ''))}
              autoFocus
            />
            <TouchableOpacity
              style={[styles.modalBtn, otp.length !== 6 && styles.btnDisabled]}
              onPress={submitOtp}
              disabled={otpLoading || otp.length !== 6}
            >
              {otpLoading ? <ActivityIndicator color="#fff" /> : <Text style={styles.modalBtnText}>Authorize Release</Text>}
            </TouchableOpacity>
            <TouchableOpacity onPress={() => setOtpModal(null)}>
              <Text style={styles.cancel}>Cancel</Text>
            </TouchableOpacity>
          </View>
        </View>
      </Modal>
    </>
  )
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f9fafb' },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center' },
  heading: { fontSize: 20, fontWeight: '700', color: '#111827', marginBottom: 4 },
  empty: { textAlign: 'center', color: '#9ca3af', marginTop: 48 },
  card: { backgroundColor: '#fff', borderRadius: 12, padding: 16 },
  topRow: { flexDirection: 'row', justifyContent: 'space-between' },
  palletCode: { fontFamily: 'monospace', fontSize: 13, fontWeight: '600', color: '#374151' },
  slot: { fontSize: 12, color: '#6b7280', backgroundColor: '#f3f4f6', paddingHorizontal: 8, paddingVertical: 2, borderRadius: 6 },
  commodity: { fontSize: 15, fontWeight: '600', color: '#111827', marginTop: 6, textTransform: 'capitalize' },
  datesRow: { flexDirection: 'row', gap: 16, marginTop: 6 },
  date: { fontSize: 12, color: '#6b7280' },
  bottomRow: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', marginTop: 12 },
  riskBadge: { borderRadius: 20, paddingHorizontal: 10, paddingVertical: 4 },
  riskText: { fontSize: 12, fontWeight: '600', textTransform: 'capitalize' },
  releaseBtn: { backgroundColor: '#16a34a', borderRadius: 8, paddingHorizontal: 14, paddingVertical: 8 },
  releaseBtnText: { color: '#fff', fontSize: 12, fontWeight: '600' },
  otpBtn: { backgroundColor: '#2563eb', borderRadius: 8, paddingHorizontal: 14, paddingVertical: 8 },
  otpBtnText: { color: '#fff', fontSize: 12, fontWeight: '600' },
  modalOverlay: { flex: 1, backgroundColor: 'rgba(0,0,0,0.5)', justifyContent: 'flex-end' },
  modalCard: { backgroundColor: '#fff', borderTopLeftRadius: 20, borderTopRightRadius: 20, padding: 24 },
  modalTitle: { fontSize: 18, fontWeight: '700', color: '#111827' },
  modalSub: { fontSize: 13, color: '#6b7280', marginTop: 4 },
  modalDesc: { fontSize: 13, color: '#374151', marginTop: 12, lineHeight: 20 },
  otpInput: { borderWidth: 1, borderColor: '#d1d5db', borderRadius: 10, paddingVertical: 14, textAlign: 'center', fontSize: 28, letterSpacing: 10, marginTop: 20, marginBottom: 16 },
  modalBtn: { backgroundColor: '#16a34a', borderRadius: 10, paddingVertical: 14, alignItems: 'center' },
  btnDisabled: { opacity: 0.5 },
  modalBtnText: { color: '#fff', fontSize: 16, fontWeight: '700' },
  cancel: { textAlign: 'center', color: '#6b7280', marginTop: 16, fontSize: 14 },
})

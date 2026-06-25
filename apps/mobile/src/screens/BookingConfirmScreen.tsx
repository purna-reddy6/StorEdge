import React, { useState } from 'react'
import {
  View, Text, StyleSheet, ScrollView, TextInput,
  TouchableOpacity, ActivityIndicator, Alert,
} from 'react-native'
import { useRoute, useNavigation } from '@react-navigation/native'
import type { NativeStackNavigationProp, RouteProp } from '@react-navigation/native-stack'
import api from '../utils/api'
import { formatINR } from '../utils/format'
import type { SearchStackParamList } from '../navigation/types'

const COMMODITIES = ['potato', 'onion', 'garlic', 'fruits', 'vegetables', 'grains', 'pharma', 'fmcg']

export default function BookingConfirmScreen() {
  const route = useRoute<RouteProp<SearchStackParamList, 'BookingConfirm'>>()
  const nav = useNavigation<NativeStackNavigationProp<SearchStackParamList>>()

  const [commodity, setCommodity] = useState('potato')
  const [pallets, setPallets] = useState('10')
  const [inwardDate, setInwardDate] = useState('')
  const [outwardDate, setOutwardDate] = useState('')
  const [loading, setLoading] = useState(false)

  const palletCount = parseInt(pallets, 10) || 0
  const days = inwardDate && outwardDate
    ? Math.max(1, Math.ceil((new Date(outwardDate).getTime() - new Date(inwardDate).getTime()) / 86400000))
    : 30
  const estimatedTotal = Math.round((2500 / 30) * days * palletCount * 0.5)
  const commission = Math.round(estimatedTotal * 0.15)

  const submit = async () => {
    if (!inwardDate || !outwardDate) {
      Alert.alert('Required', 'Please enter both inward and expected outward dates (YYYY-MM-DD).')
      return
    }
    setLoading(true)
    try {
      const { data } = await api.post('/bookings', {
        warehouse_id: route.params.warehouseId,
        commodity_type: commodity,
        pallet_count: palletCount,
        start_date: inwardDate,
        end_date: outwardDate,
      })
      nav.replace('BookingSuccess', { bookingNumber: data.booking?.booking_number ?? data.booking?.id ?? 'CONFIRMED' })
    } catch {
      Alert.alert('Booking failed', 'Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <ScrollView style={styles.container}>
      <View style={styles.card}>
        <Text style={styles.sectionTitle}>Commodity</Text>
        <View style={styles.chips}>
          {COMMODITIES.map((c) => (
            <TouchableOpacity
              key={c}
              style={[styles.chip, commodity === c && styles.chipActive]}
              onPress={() => setCommodity(c)}
            >
              <Text style={[styles.chipText, commodity === c && styles.chipTextActive]}>{c}</Text>
            </TouchableOpacity>
          ))}
        </View>
      </View>

      <View style={styles.card}>
        <Text style={styles.sectionTitle}>Quantity & Dates</Text>

        <Text style={styles.label}>Number of Pallets</Text>
        <TextInput
          style={styles.input}
          keyboardType="number-pad"
          value={pallets}
          onChangeText={setPallets}
          placeholder="e.g. 20"
        />

        <Text style={styles.label}>Inward Date (YYYY-MM-DD)</Text>
        <TextInput
          style={styles.input}
          placeholder="2026-07-01"
          value={inwardDate}
          onChangeText={setInwardDate}
        />

        <Text style={styles.label}>Expected Outward Date</Text>
        <TextInput
          style={styles.input}
          placeholder="2026-10-01"
          value={outwardDate}
          onChangeText={setOutwardDate}
        />
      </View>

      <View style={styles.card}>
        <Text style={styles.sectionTitle}>Price Breakdown</Text>
        <Row label={`Storage (${days} days × ${palletCount} pallets)`} value={formatINR(estimatedTotal)} />
        <Row label="Platform commission (15%)" value={`−${formatINR(commission)}`} muted />
        <View style={styles.divider} />
        <Row label="Operator receives" value={formatINR(estimatedTotal - commission)} bold />
        <Text style={styles.note}>* Estimate based on average ₹2,500/MT/mo. Actual price confirmed on booking.</Text>
      </View>

      <TouchableOpacity style={styles.submitBtn} onPress={submit} disabled={loading}>
        {loading ? <ActivityIndicator color="#fff" /> : <Text style={styles.submitBtnText}>Confirm Booking — {formatINR(estimatedTotal)}</Text>}
      </TouchableOpacity>
    </ScrollView>
  )
}

function Row({ label, value, muted, bold }: { label: string; value: string; muted?: boolean; bold?: boolean }) {
  return (
    <View style={styles.row}>
      <Text style={[styles.rowLabel, muted && styles.muted]}>{label}</Text>
      <Text style={[styles.rowValue, muted && styles.muted, bold && styles.bold]}>{value}</Text>
    </View>
  )
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f9fafb', padding: 16 },
  card: { backgroundColor: '#fff', borderRadius: 12, padding: 16, marginBottom: 12 },
  sectionTitle: { fontSize: 14, fontWeight: '600', color: '#374151', marginBottom: 12 },
  chips: { flexDirection: 'row', flexWrap: 'wrap', gap: 8 },
  chip: { backgroundColor: '#f3f4f6', borderRadius: 20, paddingHorizontal: 12, paddingVertical: 6, borderWidth: 1, borderColor: '#e5e7eb' },
  chipActive: { backgroundColor: '#dcfce7', borderColor: '#16a34a' },
  chipText: { fontSize: 12, color: '#374151', textTransform: 'capitalize' },
  chipTextActive: { color: '#15803d', fontWeight: '600' },
  label: { fontSize: 13, color: '#374151', fontWeight: '500', marginBottom: 6, marginTop: 12 },
  input: { borderWidth: 1, borderColor: '#d1d5db', borderRadius: 8, paddingHorizontal: 12, paddingVertical: 10, fontSize: 14, color: '#111827' },
  row: { flexDirection: 'row', justifyContent: 'space-between', paddingVertical: 4 },
  rowLabel: { fontSize: 13, color: '#374151' },
  rowValue: { fontSize: 13, color: '#374151' },
  muted: { color: '#9ca3af' },
  bold: { fontWeight: '700', color: '#16a34a' },
  divider: { height: 1, backgroundColor: '#e5e7eb', marginVertical: 8 },
  note: { fontSize: 11, color: '#9ca3af', marginTop: 8 },
  submitBtn: { backgroundColor: '#16a34a', borderRadius: 12, paddingVertical: 16, alignItems: 'center', marginBottom: 32 },
  submitBtnText: { color: '#fff', fontSize: 15, fontWeight: '700' },
})

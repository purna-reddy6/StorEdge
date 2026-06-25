import React, { useEffect, useState } from 'react'
import { View, Text, StyleSheet, FlatList, ActivityIndicator, RefreshControl, TouchableOpacity, Alert } from 'react-native'
import api from '../utils/api'
import { formatINR, formatDate } from '../utils/format'

interface Booking {
  id: string; booking_number: string; warehouse_name: string
  commodity_type: string; pallet_count: number; start_date: string
  end_date: string; total_amount_inr: number
  commission_amount_inr: number; status: string
}

const statusColor: Record<string, string> = {
  pending: '#f59e0b', confirmed: '#16a34a', active: '#16a34a',
  completed: '#9ca3af', cancelled: '#ef4444',
}

export default function BookingsScreen() {
  const [bookings, setBookings] = useState<Booking[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)

  const load = async (refresh = false) => {
    if (refresh) setRefreshing(true)
    try {
      const { data } = await api.get('/bookings')
      setBookings(data.bookings ?? [])
    } catch {
      Alert.alert('Error', 'Could not load bookings.')
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => { load() }, [])

  if (loading) return <View style={styles.center}><ActivityIndicator color="#16a34a" size="large" /></View>

  return (
    <FlatList
      style={styles.container}
      data={bookings}
      keyExtractor={(b) => b.id}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={() => load(true)} />}
      contentContainerStyle={{ padding: 16, gap: 12 }}
      ListHeaderComponent={<Text style={styles.heading}>My Bookings</Text>}
      ListEmptyComponent={<Text style={styles.empty}>No bookings yet. Find a warehouse and book!</Text>}
      renderItem={({ item: b }) => (
        <View style={styles.card}>
          <View style={styles.topRow}>
            <Text style={styles.bookingNum}>{b.booking_number}</Text>
            <View style={[styles.statusBadge, { backgroundColor: `${statusColor[b.status]}20` }]}>
              <Text style={[styles.statusText, { color: statusColor[b.status] }]}>{b.status}</Text>
            </View>
          </View>
          <Text style={styles.warehouseName}>{b.warehouse_name}</Text>
          <Text style={styles.meta}>{b.commodity_type} · {b.pallet_count} pallets</Text>
          <View style={styles.datesRow}>
            <DateLabel label="In" date={b.start_date} />
            <Text style={styles.arrow}>→</Text>
            <DateLabel label="Out" date={b.end_date} />
            <View style={styles.spacer} />
            <View>
              <Text style={styles.amount}>{formatINR(b.total_amount_inr)}</Text>
              <Text style={styles.commission}>Commission: {formatINR(b.commission_amount_inr)}</Text>
            </View>
          </View>
        </View>
      )}
    />
  )
}

function DateLabel({ label, date }: { label: string; date: string }) {
  return (
    <View>
      <Text style={{ fontSize: 10, color: '#9ca3af' }}>{label}</Text>
      <Text style={{ fontSize: 12, fontWeight: '500', color: '#374151' }}>{formatDate(date)}</Text>
    </View>
  )
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f9fafb' },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center' },
  heading: { fontSize: 20, fontWeight: '700', color: '#111827', marginBottom: 4 },
  empty: { textAlign: 'center', color: '#9ca3af', marginTop: 48 },
  card: { backgroundColor: '#fff', borderRadius: 12, padding: 16 },
  topRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  bookingNum: { fontFamily: 'monospace', fontSize: 12, color: '#6b7280' },
  statusBadge: { borderRadius: 20, paddingHorizontal: 8, paddingVertical: 3 },
  statusText: { fontSize: 11, fontWeight: '600', textTransform: 'capitalize' },
  warehouseName: { fontSize: 16, fontWeight: '600', color: '#111827', marginTop: 6 },
  meta: { fontSize: 12, color: '#6b7280', marginTop: 2, textTransform: 'capitalize' },
  datesRow: { flexDirection: 'row', alignItems: 'center', marginTop: 12, gap: 10 },
  arrow: { color: '#9ca3af' },
  spacer: { flex: 1 },
  amount: { fontSize: 15, fontWeight: '700', color: '#16a34a', textAlign: 'right' },
  commission: { fontSize: 10, color: '#9ca3af', textAlign: 'right' },
})

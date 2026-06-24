import React from 'react'
import { View, Text, StyleSheet, ScrollView, TouchableOpacity, ActivityIndicator, Alert } from 'react-native'
import { useRoute, useNavigation } from '@react-navigation/native'
import type { NativeStackNavigationProp, RouteProp } from '@react-navigation/native-stack'
import api from '../utils/api'
import { formatINR, warehouseTypeLabel } from '../utils/format'
import type { SearchStackParamList } from '../navigation/types'
import { useEffect, useState } from 'react'

interface Warehouse {
  id: string; name: string; address: string; city: string; state: string
  warehouseType: string; totalCapacityMt: number; availablePallets: number
  totalPallets: number; basePricePerMtPerMonth: number; dynamicPrice?: number
  rating: number; reviewCount: number; wdraRegistered: boolean; apmcLicensed: boolean
  minTemperatureCelsius?: number; maxTemperatureCelsius?: number
}

export default function WarehouseDetailScreen() {
  const route = useRoute<RouteProp<SearchStackParamList, 'WarehouseDetail'>>()
  const nav = useNavigation<NativeStackNavigationProp<SearchStackParamList>>()
  const [warehouse, setWarehouse] = useState<Warehouse | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get(`/warehouses/${route.params.warehouseId}`)
      .then((r) => setWarehouse(r.data))
      .catch(() => Alert.alert('Error', 'Failed to load warehouse.'))
      .finally(() => setLoading(false))
  }, [route.params.warehouseId])

  if (loading) return <View style={styles.center}><ActivityIndicator color="#16a34a" size="large" /></View>
  if (!warehouse) return <View style={styles.center}><Text>Warehouse not found.</Text></View>

  const price = warehouse.dynamicPrice ?? warehouse.basePricePerMtPerMonth
  const occupancy = Math.round((1 - warehouse.availablePallets / warehouse.totalPallets) * 100)

  return (
    <ScrollView style={styles.container}>
      <View style={styles.header}>
        <View style={styles.headerTop}>
          <Text style={styles.name}>{warehouse.name}</Text>
          <View>
            <Text style={styles.price}>{formatINR(price)}</Text>
            <Text style={styles.priceSub}>/MT/month</Text>
          </View>
        </View>
        <Text style={styles.address}>{warehouse.address}, {warehouse.city}, {warehouse.state}</Text>

        <View style={styles.badges}>
          <View style={styles.badge}><Text style={styles.badgeText}>{warehouseTypeLabel[warehouse.warehouseType] ?? warehouse.warehouseType}</Text></View>
          {warehouse.wdraRegistered && <View style={[styles.badge, styles.badgeGreen]}><Text style={styles.badgeGreenText}>✓ WDRA</Text></View>}
          {warehouse.apmcLicensed && <View style={[styles.badge, styles.badgeBlue]}><Text style={styles.badgeBluetText}>APMC</Text></View>}
        </View>
      </View>

      <View style={styles.statsRow}>
        <Stat label="Rating" value={`⭐ ${warehouse.rating.toFixed(1)}`} sub={`${warehouse.reviewCount} reviews`} />
        <Stat label="Available" value={`${warehouse.availablePallets}`} sub="pallets" />
        <Stat label="Capacity" value={`${warehouse.totalCapacityMt} MT`} sub={`${occupancy}% occupied`} />
      </View>

      {warehouse.minTemperatureCelsius !== undefined && (
        <View style={styles.tempCard}>
          <Text style={styles.tempLabel}>🌡 Cold Chain: {warehouse.minTemperatureCelsius}°C – {warehouse.maxTemperatureCelsius}°C</Text>
        </View>
      )}

      <TouchableOpacity
        style={styles.bookBtn}
        onPress={() => nav.navigate('BookingConfirm', { warehouseId: warehouse.id })}
      >
        <Text style={styles.bookBtnText}>Book This Warehouse</Text>
      </TouchableOpacity>
    </ScrollView>
  )
}

function Stat({ label, value, sub }: { label: string; value: string; sub: string }) {
  return (
    <View style={styles.stat}>
      <Text style={styles.statLabel}>{label}</Text>
      <Text style={styles.statValue}>{value}</Text>
      <Text style={styles.statSub}>{sub}</Text>
    </View>
  )
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f9fafb' },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center' },
  header: { backgroundColor: '#fff', padding: 20, marginBottom: 8 },
  headerTop: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'flex-start' },
  name: { fontSize: 18, fontWeight: '700', color: '#111827', flex: 1, marginRight: 12 },
  price: { fontSize: 20, fontWeight: '700', color: '#16a34a', textAlign: 'right' },
  priceSub: { fontSize: 11, color: '#9ca3af', textAlign: 'right' },
  address: { fontSize: 13, color: '#6b7280', marginTop: 6 },
  badges: { flexDirection: 'row', flexWrap: 'wrap', gap: 6, marginTop: 12 },
  badge: { backgroundColor: '#f3f4f6', borderRadius: 20, paddingHorizontal: 10, paddingVertical: 4 },
  badgeText: { fontSize: 12, color: '#374151' },
  badgeGreen: { backgroundColor: '#dcfce7' },
  badgeGreenText: { fontSize: 12, color: '#15803d', fontWeight: '500' },
  badgeBlue: { backgroundColor: '#dbeafe' },
  badgeBluetText: { fontSize: 12, color: '#1d4ed8' },
  statsRow: { flexDirection: 'row', backgroundColor: '#fff', marginBottom: 8 },
  stat: { flex: 1, alignItems: 'center', paddingVertical: 16, borderRightWidth: 1, borderColor: '#f3f4f6' },
  statLabel: { fontSize: 11, color: '#9ca3af', marginBottom: 4 },
  statValue: { fontSize: 16, fontWeight: '700', color: '#111827' },
  statSub: { fontSize: 11, color: '#6b7280', marginTop: 2 },
  tempCard: { backgroundColor: '#eff6ff', marginHorizontal: 16, borderRadius: 10, padding: 12, marginBottom: 8 },
  tempLabel: { fontSize: 14, color: '#1e40af', fontWeight: '500' },
  bookBtn: { backgroundColor: '#16a34a', margin: 16, borderRadius: 12, paddingVertical: 16, alignItems: 'center' },
  bookBtnText: { color: '#fff', fontSize: 16, fontWeight: '700' },
})

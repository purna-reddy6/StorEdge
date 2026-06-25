import React, { useState } from 'react'
import { View, Text, TextInput, TouchableOpacity, StyleSheet, FlatList, ActivityIndicator, Alert } from 'react-native'
import MapView, { Marker, Callout } from 'react-native-maps'
import { useNavigation } from '@react-navigation/native'
import type { NativeStackNavigationProp } from '@react-navigation/native-stack'
import api from '../utils/api'
import { formatINR, warehouseTypeLabel } from '../utils/format'
import type { SearchStackParamList } from '../navigation/types'

interface Warehouse {
  id: string
  name: string
  city: string
  latitude: number
  longitude: number
  type: string
  available_pallet_slots: number
  base_price_per_pallet_inr: number
  price_per_pallet_per_day_inr: number
  rating: number
  gst_registered: boolean
  distance_km?: number
  match_score?: number
}

const INITIAL_REGION = {
  latitude: 27.1767, longitude: 78.0081,
  latitudeDelta: 0.8, longitudeDelta: 0.8,
}

export default function SearchMapScreen() {
  const nav = useNavigation<NativeStackNavigationProp<SearchStackParamList>>()
  const [region, setRegion] = useState(INITIAL_REGION)
  const [pallets, setPallets] = useState('10')
  const [radius, setRadius] = useState('50')
  const [results, setResults] = useState<Warehouse[]>([])
  const [loading, setLoading] = useState(false)
  const [selected, setSelected] = useState<Warehouse | null>(null)

  const search = async () => {
    setLoading(true)
    setSelected(null)
    try {
      const { data } = await api.get('/warehouses/search', {
        params: {
          lat: region.latitude, lng: region.longitude,
          radius_km: radius, pallets,
        },
      })
      setResults(data.warehouses ?? [])
    } catch {
      Alert.alert('Search failed', 'Could not reach the server. Make sure the backend is running.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <View style={styles.container}>
      {/* Search bar */}
      <View style={styles.searchBar}>
        <TextInput
          style={styles.input}
          placeholder="Pallets needed"
          keyboardType="number-pad"
          value={pallets}
          onChangeText={setPallets}
        />
        <TextInput
          style={styles.input}
          placeholder="Radius (km)"
          keyboardType="number-pad"
          value={radius}
          onChangeText={setRadius}
        />
        <TouchableOpacity style={styles.searchBtn} onPress={search} disabled={loading}>
          {loading ? <ActivityIndicator color="#fff" size="small" /> : <Text style={styles.searchBtnText}>Search</Text>}
        </TouchableOpacity>
      </View>

      {/* Map */}
      <MapView
        style={styles.map}
        region={region}
        onRegionChangeComplete={setRegion}
      >
        {results.map((w) => (
          <Marker
            key={w.id}
            coordinate={{ latitude: w.latitude, longitude: w.longitude }}
            onPress={() => setSelected(w)}
            pinColor={selected?.id === w.id ? '#16a34a' : '#ef4444'}
          >
            <Callout onPress={() => nav.navigate('WarehouseDetail', { warehouseId: w.id })}>
              <View style={styles.callout}>
                <Text style={styles.calloutName}>{w.name}</Text>
                <Text style={styles.calloutPrice}>₹{w.price_per_pallet_per_day_inr.toFixed(0)}/pallet/day</Text>
                <Text style={styles.calloutLink}>Tap to view →</Text>
              </View>
            </Callout>
          </Marker>
        ))}
      </MapView>

      {/* Bottom result list */}
      <View style={styles.list}>
        <Text style={styles.listHeader}>
          {results.length > 0 ? `${results.length} warehouses found` : 'Set location on map and search'}
        </Text>
        <FlatList
          data={results}
          keyExtractor={(i) => i.id}
          horizontal
          showsHorizontalScrollIndicator={false}
          contentContainerStyle={{ paddingHorizontal: 8 }}
          renderItem={({ item: w }) => (
            <TouchableOpacity
              style={[styles.card, selected?.id === w.id && styles.cardSelected]}
              onPress={() => {
                setSelected(w)
                setRegion({ ...region, latitude: w.latitude, longitude: w.longitude })
              }}
            >
              <Text style={styles.cardName} numberOfLines={1}>{w.name}</Text>
              <Text style={styles.cardType}>{warehouseTypeLabel[w.type] ?? w.type}</Text>
              <Text style={styles.cardPrice}>₹{w.price_per_pallet_per_day_inr.toFixed(0)}/pallet/day</Text>
              <Text style={styles.cardPallets}>⬡ {w.available_pallet_slots} pallets free</Text>
              {w.gst_registered && <Text style={styles.wdra}>✓ GST Verified</Text>}
              <TouchableOpacity
                style={styles.bookBtn}
                onPress={() => nav.navigate('WarehouseDetail', { warehouseId: w.id })}
              >
                <Text style={styles.bookBtnText}>View & Book</Text>
              </TouchableOpacity>
            </TouchableOpacity>
          )}
        />
      </View>
    </View>
  )
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#fff' },
  searchBar: { flexDirection: 'row', gap: 8, padding: 12, backgroundColor: '#fff', borderBottomWidth: 1, borderColor: '#e5e7eb' },
  input: { flex: 1, borderWidth: 1, borderColor: '#d1d5db', borderRadius: 8, paddingHorizontal: 10, paddingVertical: 8, fontSize: 14 },
  searchBtn: { backgroundColor: '#16a34a', borderRadius: 8, paddingHorizontal: 16, justifyContent: 'center' },
  searchBtnText: { color: '#fff', fontWeight: '600', fontSize: 14 },
  map: { flex: 1 },
  callout: { padding: 8, minWidth: 140 },
  calloutName: { fontWeight: '600', fontSize: 13 },
  calloutPrice: { color: '#16a34a', fontWeight: '700', marginTop: 2 },
  calloutLink: { color: '#6b7280', fontSize: 11, marginTop: 4 },
  list: { backgroundColor: '#fff', paddingTop: 8, paddingBottom: 12, borderTopWidth: 1, borderColor: '#e5e7eb' },
  listHeader: { fontSize: 12, color: '#6b7280', paddingHorizontal: 16, marginBottom: 8 },
  card: { backgroundColor: '#f9fafb', borderRadius: 12, borderWidth: 1, borderColor: '#e5e7eb', padding: 12, width: 160, marginHorizontal: 4 },
  cardSelected: { borderColor: '#16a34a', backgroundColor: '#f0fdf4' },
  cardName: { fontWeight: '600', fontSize: 13, color: '#111827' },
  cardType: { fontSize: 11, color: '#6b7280', marginTop: 2 },
  cardPrice: { fontSize: 14, fontWeight: '700', color: '#16a34a', marginTop: 6 },
  cardPallets: { fontSize: 11, color: '#6b7280', marginTop: 2 },
  wdra: { fontSize: 11, color: '#15803d', fontWeight: '500', marginTop: 2 },
  bookBtn: { marginTop: 8, backgroundColor: '#16a34a', borderRadius: 6, paddingVertical: 6, alignItems: 'center' },
  bookBtnText: { color: '#fff', fontSize: 12, fontWeight: '600' },
})

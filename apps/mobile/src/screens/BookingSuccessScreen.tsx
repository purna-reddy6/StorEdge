import React from 'react'
import { View, Text, StyleSheet, TouchableOpacity } from 'react-native'
import { useRoute, useNavigation } from '@react-navigation/native'
import type { NativeStackNavigationProp, RouteProp } from '@react-navigation/native-stack'
import type { SearchStackParamList } from '../navigation/types'

export default function BookingSuccessScreen() {
  const route = useRoute<RouteProp<SearchStackParamList, 'BookingSuccess'>>()
  const nav = useNavigation<NativeStackNavigationProp<SearchStackParamList>>()

  return (
    <View style={styles.container}>
      <View style={styles.check}><Text style={styles.checkMark}>✓</Text></View>
      <Text style={styles.title}>Booking Confirmed!</Text>
      <Text style={styles.sub}>Booking #{route.params.bookingNumber}</Text>
      <Text style={styles.desc}>
        The warehouse operator will confirm your slot within 2 hours.
        You will receive an SMS when the booking is active.
      </Text>

      <View style={styles.infoCard}>
        <Text style={styles.infoTitle}>Stock Release — How it Works</Text>
        <Text style={styles.infoText}>
          When you are ready to collect your goods, request a stock release from the Inventory tab.
          You will receive a 6-digit OTP on WhatsApp to authorize the release — no travel required.
        </Text>
      </View>

      <TouchableOpacity style={styles.btn} onPress={() => nav.navigate('SearchMap')}>
        <Text style={styles.btnText}>Back to Search</Text>
      </TouchableOpacity>
    </View>
  )
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f0fdf4', alignItems: 'center', justifyContent: 'center', padding: 32 },
  check: { width: 80, height: 80, borderRadius: 40, backgroundColor: '#16a34a', alignItems: 'center', justifyContent: 'center', marginBottom: 24 },
  checkMark: { fontSize: 40, color: '#fff' },
  title: { fontSize: 24, fontWeight: '700', color: '#166534', textAlign: 'center' },
  sub: { fontSize: 14, color: '#6b7280', marginTop: 8, textAlign: 'center' },
  desc: { fontSize: 14, color: '#374151', textAlign: 'center', marginTop: 16, lineHeight: 22 },
  infoCard: { backgroundColor: '#fff', borderRadius: 12, padding: 16, marginTop: 24, width: '100%' },
  infoTitle: { fontSize: 13, fontWeight: '600', color: '#1f2937', marginBottom: 8 },
  infoText: { fontSize: 13, color: '#6b7280', lineHeight: 20 },
  btn: { marginTop: 32, backgroundColor: '#16a34a', borderRadius: 12, paddingHorizontal: 32, paddingVertical: 14 },
  btnText: { color: '#fff', fontSize: 15, fontWeight: '700' },
})

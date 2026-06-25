import React, { useEffect, useState } from 'react'
import { View, Text, StyleSheet, FlatList, RefreshControl, ActivityIndicator, Alert, TouchableOpacity } from 'react-native'
import api from '../utils/api'
import { formatINR, formatDate } from '../utils/format'

interface EnwrReceipt {
  id: string; receiptNumber: string; commodity: string
  quantityKg: number; marketValueInr: number; maxLoanAmountInr: number
  status: string; expiryDate: string; issuedAt?: string
}

const statusColor: Record<string, string> = {
  draft: '#f59e0b', issued: '#16a34a', pledged: '#7c3aed',
  released: '#9ca3af', expired: '#ef4444',
}

export default function FinancingScreen() {
  const [receipts, setReceipts] = useState<EnwrReceipt[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)

  const load = async (refresh = false) => {
    if (refresh) setRefreshing(true)
    try {
      const { data } = await api.get('/financing/receipts')
      setReceipts(data.receipts ?? [])
    } catch {
      Alert.alert('Error', 'Could not load receipts.')
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => { load() }, [])

  const applyForLoan = async (receiptId: string) => {
    Alert.alert(
      'Apply for Loan',
      'This will submit a loan application backed by your e-NWR to partner banks. Proceed?',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Apply', onPress: async () => {
            try {
              await api.post(`/financing/receipts/${receiptId}/loans`)
              Alert.alert('Application Submitted', 'Your loan application has been submitted. The bank will contact you within 24 hours.')
              load()
            } catch {
              Alert.alert('Error', 'Failed to submit loan application.')
            }
          },
        },
      ],
    )
  }

  if (loading) return <View style={styles.center}><ActivityIndicator color="#16a34a" size="large" /></View>

  const totalValue = receipts.reduce((s, r) => s + r.marketValueInr, 0)
  const totalLoan = receipts.reduce((s, r) => s + r.maxLoanAmountInr, 0)

  return (
    <FlatList
      style={styles.container}
      data={receipts}
      keyExtractor={(r) => r.id}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={() => load(true)} />}
      contentContainerStyle={{ padding: 16, gap: 12 }}
      ListHeaderComponent={
        <View>
          <Text style={styles.heading}>e-NWR & Financing</Text>

          <View style={styles.infoCard}>
            <Text style={styles.infoTitle}>PSL Limit: ₹75 Lakh per e-NWR</Text>
            <Text style={styles.infoText}>
              Electronic Negotiable Warehouse Receipts issued via NERL.
              Loan eligibility: 70% of commodity market value. Origination fee: 1.5%.
            </Text>
          </View>

          {receipts.length > 0 && (
            <View style={styles.summaryRow}>
              <View style={styles.summaryCard}>
                <Text style={styles.summaryLabel}>Market Value</Text>
                <Text style={styles.summaryValue}>{formatINR(totalValue)}</Text>
              </View>
              <View style={styles.summaryCard}>
                <Text style={styles.summaryLabel}>Loan Eligible</Text>
                <Text style={[styles.summaryValue, { color: '#16a34a' }]}>{formatINR(totalLoan)}</Text>
              </View>
            </View>
          )}
        </View>
      }
      ListEmptyComponent={<Text style={styles.empty}>No warehouse receipts yet. Complete an inward to generate an e-NWR.</Text>}
      renderItem={({ item: r }) => (
        <View style={styles.card}>
          <View style={styles.topRow}>
            <Text style={styles.receiptNum}>{r.receiptNumber}</Text>
            <View style={[styles.statusBadge, { backgroundColor: `${statusColor[r.status]}20` }]}>
              <Text style={[styles.statusText, { color: statusColor[r.status] }]}>{r.status}</Text>
            </View>
          </View>
          <Text style={styles.commodity}>{r.commodity} — {r.quantityKg.toLocaleString()} kg</Text>

          <View style={styles.valuesRow}>
            <View>
              <Text style={styles.valLabel}>Market Value</Text>
              <Text style={styles.valAmount}>{formatINR(r.marketValueInr)}</Text>
            </View>
            <View>
              <Text style={styles.valLabel}>Max Loan</Text>
              <Text style={[styles.valAmount, { color: '#16a34a' }]}>{formatINR(r.maxLoanAmountInr)}</Text>
            </View>
          </View>

          <Text style={styles.expiry}>
            {r.issuedAt ? `Issued: ${formatDate(r.issuedAt)} · ` : ''}Expires: {formatDate(r.expiryDate)}
          </Text>

          {r.status === 'issued' && (
            <TouchableOpacity style={styles.loanBtn} onPress={() => applyForLoan(r.id)}>
              <Text style={styles.loanBtnText}>Apply for Loan</Text>
            </TouchableOpacity>
          )}
        </View>
      )}
    />
  )
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f9fafb' },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center' },
  heading: { fontSize: 20, fontWeight: '700', color: '#111827', marginBottom: 12 },
  infoCard: { backgroundColor: '#dcfce7', borderRadius: 12, padding: 14, marginBottom: 12 },
  infoTitle: { fontSize: 14, fontWeight: '700', color: '#15803d' },
  infoText: { fontSize: 13, color: '#166534', marginTop: 6, lineHeight: 20 },
  summaryRow: { flexDirection: 'row', gap: 12, marginBottom: 4 },
  summaryCard: { flex: 1, backgroundColor: '#fff', borderRadius: 10, padding: 12 },
  summaryLabel: { fontSize: 11, color: '#9ca3af' },
  summaryValue: { fontSize: 16, fontWeight: '700', color: '#111827', marginTop: 4 },
  empty: { textAlign: 'center', color: '#9ca3af', marginTop: 48, lineHeight: 22 },
  card: { backgroundColor: '#fff', borderRadius: 12, padding: 16 },
  topRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  receiptNum: { fontFamily: 'monospace', fontSize: 12, color: '#6b7280' },
  statusBadge: { borderRadius: 20, paddingHorizontal: 8, paddingVertical: 3 },
  statusText: { fontSize: 11, fontWeight: '600', textTransform: 'capitalize' },
  commodity: { fontSize: 15, fontWeight: '600', color: '#111827', marginTop: 6, textTransform: 'capitalize' },
  valuesRow: { flexDirection: 'row', gap: 32, marginTop: 12 },
  valLabel: { fontSize: 11, color: '#9ca3af' },
  valAmount: { fontSize: 16, fontWeight: '700', color: '#111827', marginTop: 2 },
  expiry: { fontSize: 11, color: '#9ca3af', marginTop: 8 },
  loanBtn: { backgroundColor: '#16a34a', borderRadius: 8, paddingVertical: 10, alignItems: 'center', marginTop: 12 },
  loanBtnText: { color: '#fff', fontSize: 13, fontWeight: '700' },
})

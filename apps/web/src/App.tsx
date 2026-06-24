import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './store/authStore'
import Layout from './components/Layout'
import LoginPage from './pages/LoginPage'
import SearchPage from './pages/SearchPage'
import WarehouseDetailPage from './pages/WarehouseDetailPage'
import BookingsPage from './pages/BookingsPage'
import OperatorDashboard from './pages/OperatorDashboard'
import InventoryPage from './pages/InventoryPage'
import FinancingPage from './pages/FinancingPage'
import AlertsPage from './pages/AlertsPage'

function RequireAuth({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((s) => s.token)
  return token ? <>{children}</> : <Navigate to="/login" replace />
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route
        path="/"
        element={
          <RequireAuth>
            <Layout />
          </RequireAuth>
        }
      >
        <Route index element={<Navigate to="/search" replace />} />
        <Route path="search" element={<SearchPage />} />
        <Route path="warehouses/:id" element={<WarehouseDetailPage />} />
        <Route path="bookings" element={<BookingsPage />} />
        <Route path="operator" element={<OperatorDashboard />} />
        <Route path="inventory" element={<InventoryPage />} />
        <Route path="financing" element={<FinancingPage />} />
        <Route path="alerts" element={<AlertsPage />} />
      </Route>
    </Routes>
  )
}

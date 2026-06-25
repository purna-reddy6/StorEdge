import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import {
  MagnifyingGlassIcon,
  CalendarDaysIcon,
  BuildingOfficeIcon,
  ArchiveBoxIcon,
  BellAlertIcon,
  ArrowRightOnRectangleIcon,
} from '@heroicons/react/24/outline'
import { useAuthStore } from '../store/authStore'

const navItems = [
  { to: '/search', label: 'Find Storage', icon: MagnifyingGlassIcon },
  { to: '/bookings', label: 'My Bookings', icon: CalendarDaysIcon },
  { to: '/inventory', label: 'Inventory', icon: ArchiveBoxIcon },
  { to: '/operator', label: 'Host Dashboard', icon: BuildingOfficeIcon },
  { to: '/alerts', label: 'IoT Alerts', icon: BellAlertIcon },
]

export default function Layout() {
  const { user, logout } = useAuthStore()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <div className="flex h-screen bg-gray-50">
      {/* Sidebar */}
      <aside className="w-64 bg-white border-r border-gray-200 flex flex-col">
        <div className="p-6 border-b border-gray-200">
          <h1 className="text-xl font-bold text-brand-700">StorEdge</h1>
          <p className="text-xs text-gray-500 mt-1">Storage Marketplace</p>
        </div>

        <nav className="flex-1 py-4 space-y-1 px-3">
          {navItems.map(({ to, label, icon: Icon }) => (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-brand-50 text-brand-700'
                    : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                }`
              }
            >
              <Icon className="h-5 w-5 flex-shrink-0" />
              {label}
            </NavLink>
          ))}
        </nav>

        <div className="p-4 border-t border-gray-200">
          <div className="text-sm font-medium text-gray-900 truncate">{user?.name}</div>
          <div className="text-xs text-gray-500 capitalize mt-0.5">{user?.role}</div>
          <button
            onClick={handleLogout}
            className="mt-3 flex items-center gap-2 text-xs text-gray-500 hover:text-red-600 transition-colors"
          >
            <ArrowRightOnRectangleIcon className="h-4 w-4" />
            Sign out
          </button>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}

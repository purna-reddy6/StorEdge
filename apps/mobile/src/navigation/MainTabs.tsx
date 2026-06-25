import React from 'react'
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs'
import { createNativeStackNavigator } from '@react-navigation/native-stack'
import SearchMapScreen from '../screens/SearchMapScreen'
import WarehouseDetailScreen from '../screens/WarehouseDetailScreen'
import BookingConfirmScreen from '../screens/BookingConfirmScreen'
import BookingSuccessScreen from '../screens/BookingSuccessScreen'
import BookingsScreen from '../screens/BookingsScreen'
import InventoryScreen from '../screens/InventoryScreen'
import ProfileScreen from '../screens/ProfileScreen'
import type { MainTabParamList, SearchStackParamList } from './types'

const Tab = createBottomTabNavigator<MainTabParamList>()
const SearchStack = createNativeStackNavigator<SearchStackParamList>()

function SearchNavigator() {
  return (
    <SearchStack.Navigator>
      <SearchStack.Screen name="SearchMap" component={SearchMapScreen} options={{ title: 'Find Warehouses' }} />
      <SearchStack.Screen name="WarehouseDetail" component={WarehouseDetailScreen} options={{ title: 'Warehouse Details' }} />
      <SearchStack.Screen name="BookingConfirm" component={BookingConfirmScreen} options={{ title: 'Confirm Booking' }} />
      <SearchStack.Screen name="BookingSuccess" component={BookingSuccessScreen} options={{ headerShown: false }} />
    </SearchStack.Navigator>
  )
}

export default function MainTabs() {
  return (
    <Tab.Navigator
      screenOptions={{
        tabBarActiveTintColor: '#16a34a',
        tabBarInactiveTintColor: '#9ca3af',
        tabBarStyle: { paddingBottom: 4, height: 60 },
        headerShown: false,
      }}
    >
      <Tab.Screen name="Search" component={SearchNavigator} options={{ title: 'Find' }} />
      <Tab.Screen name="Bookings" component={BookingsScreen} options={{ title: 'Bookings' }} />
      <Tab.Screen name="Inventory" component={InventoryScreen} options={{ title: 'Inventory' }} />
<Tab.Screen name="Profile" component={ProfileScreen} options={{ title: 'Profile' }} />
    </Tab.Navigator>
  )
}

export type RootStackParamList = {
  Login: undefined
  Main: undefined
}

export type MainTabParamList = {
  Search: undefined
  Bookings: undefined
  Inventory: undefined
  Financing: undefined
  Profile: undefined
}

export type SearchStackParamList = {
  SearchMap: undefined
  WarehouseDetail: { warehouseId: string }
  BookingConfirm: { warehouseId: string }
  BookingSuccess: { bookingNumber: string }
}

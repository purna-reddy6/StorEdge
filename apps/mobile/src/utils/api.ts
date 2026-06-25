import axios from 'axios'

// In dev: point to local backend; in prod: configure via env
const BASE_URL = __DEV__ ? 'http://10.0.2.2:8080/api/v1' : 'https://api.storedge.in/api/v1'

const api = axios.create({ baseURL: BASE_URL, timeout: 10_000 })

let _token: string | null = null

export function setAuthToken(token: string | null) {
  _token = token
}

api.interceptors.request.use((config) => {
  if (_token) config.headers.Authorization = `Bearer ${_token}`
  return config
})

export default api

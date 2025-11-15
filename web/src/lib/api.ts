import axios, { AxiosInstance } from 'axios'

const apiBase = import.meta.env.VITE_API_BASE_URL || 'http://localhost:3001'

export const api: AxiosInstance = axios.create({
  baseURL: apiBase,
  headers: {
    'Content-Type': 'application/json'
  }
})

export const fetchHosts = () => api.get('/hosts')
export const fetchHostDetail = (hostId: string) => api.get(`/hosts/${hostId}`)
export const fetchPackages = (hostId: string) => api.get('/apps', { params: { hostId } })
export const fetchCISResults = (hostId: string) => api.get('/cis-results', { params: { hostId } })

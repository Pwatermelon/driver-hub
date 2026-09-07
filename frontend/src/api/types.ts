export type Role = 'driver' | 'company'

export interface Driver {
  id: string
  hub_id: string
  email?: string
  phone?: string
  last_name: string
  first_name: string
  middle_name?: string
  birth_date?: string
  snils?: string
  inn?: string
  experience_years: number
  categories: string[]
  verification_status: string
  verified_at?: string
  license?: DriverLicense
}

export interface DriverLicense {
  id: string
  series: string
  number: string
  issue_date?: string
  expiry_date?: string
  categories: string[]
  issuer?: string
  source: string
  verified: boolean
}

export interface Company {
  id: string
  name: string
  inn: string
  ogrn?: string
  email: string
  phone?: string
  verified: boolean
}

export interface LicenseDraft {
  last_name: string
  first_name: string
  middle_name: string
  birth_date: string
  series: string
  number: string
  issue_date: string
  expiry_date: string
  categories: string[]
  issuer: string
  confidence: number
  notes?: string
  source: string
}

export interface DriverDossier {
  driver: Driver
  recommendations: Array<{ id: string; company_name?: string; text: string; rating: number }>
  complaints: Array<{ id: string; category: string; text: string; severity: string }>
  blacklist_entries: Array<{ id: string; company_name?: string; reason: string; active: boolean }>
  accidents: Array<{ id: string; description: string; fault: string; damage_level: string }>
  fines: Array<{ id: string; article?: string; amount: number; paid: boolean }>
  blacklisted_by_viewer: boolean
  grant?: { id: string; purpose: string; active: boolean }
}

export interface PublicCard {
  hub_id: string
  full_name_masked: string
  categories: string[]
  experience_years: number
  verification_status: string
  has_active_suspension: boolean
  blacklist_count: number
  complaint_count: number
  accident_count: number
}

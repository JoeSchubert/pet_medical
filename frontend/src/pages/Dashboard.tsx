import { Link } from 'react-router-dom'
import { Icon } from '@iconify/react'
import { useEffect, useState } from 'react'
import { petsApi, weightsApi, type Pet } from '../api/client'
import { useAuth } from '../contexts/AuthContext'
import { useTranslation } from '../i18n/context'
import { getSpeciesEmoji } from '../utils/speciesEmoji'

function weightToDisplay(lbs: number, unit: 'lbs' | 'kg'): { value: number; unit: string } {
  if (unit === 'kg') return { value: Math.round((lbs / 2.20462) * 100) / 100, unit: 'kg' }
  return { value: lbs, unit: 'lbs' }
}

function formatDateOnly(apiDate?: string | null): string {
  if (!apiDate) return ''
  const d = apiDate.length === 10 ? `${apiDate}T12:00:00` : apiDate
  return new Date(d).toLocaleDateString(undefined, { dateStyle: 'medium' })
}

function ageYearsAndMonths(dateOfBirth?: string | null): { years: number; months: number } | null {
  if (!dateOfBirth) return null
  const birth = new Date(dateOfBirth.length === 10 ? `${dateOfBirth}T12:00:00` : dateOfBirth)
  const today = new Date()
  if (today < birth) return null
  let months = (today.getFullYear() - birth.getFullYear()) * 12 + (today.getMonth() - birth.getMonth())
  if (today.getDate() < birth.getDate()) months -= 1
  if (months < 0) return null
  return { years: Math.floor(months / 12), months: months % 12 }
}

type SortOption = 'name' | 'species' | 'age'

export default function Dashboard() {
  const { user } = useAuth()
  const { t } = useTranslation()
  const weightUnit = user?.weight_unit ?? 'lbs'
  const [pets, setPets] = useState<Pet[]>([])
  const [latestWeight, setLatestWeight] = useState<Record<string, number>>({})
  const [loading, setLoading] = useState(true)
  const [sortBy, setSortBy] = useState<SortOption>('name')
  const [search, setSearch] = useState('')

  useEffect(() => {
    petsApi
      .list()
      .then((data) => setPets(Array.isArray(data) ? data : []))
      .catch(() => setPets([]))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    if (pets.length === 0) return
    const map: Record<string, number> = {}
    Promise.all(
      pets.map((pet) =>
        weightsApi.list(pet.id).then((entries) => {
          const latest = entries[0]
          if (latest != null) map[pet.id] = latest.weight_lbs
        })
      )
    ).then(() => setLatestWeight(map))
  }, [pets])

  const petImageUrl = (pet: Pet) => {
    if (!pet.photo_url) return null
    return pet.photo_url.startsWith('http') ? pet.photo_url : `${window.location.origin}${pet.photo_url}`
  }

  const searchLower = search.trim().toLowerCase()
  const filteredPets = searchLower
    ? pets.filter(
        (p) =>
          (p.name && p.name.toLowerCase().includes(searchLower)) ||
          (p.species && p.species.toLowerCase().includes(searchLower)) ||
          (p.breed && p.breed.toLowerCase().includes(searchLower)) ||
          (p.date_of_birth && ageYearsAndMonths(p.date_of_birth) && (() => {
            const age = ageYearsAndMonths(p.date_of_birth!)
            if (!age) return false
            const ageStr = age.years < 1 ? `${age.months}mo` : age.months === 0 ? `${age.years}y` : `${age.years}y ${age.months}mo`
            return ageStr.toLowerCase().includes(searchLower)
          })())
      )
    : pets

  const sortedPets = [...filteredPets].sort((a, b) => {
    if (sortBy === 'name') return (a.name || '').localeCompare(b.name || '')
    if (sortBy === 'species') return (a.species || '').localeCompare(b.species || '') || (a.name || '').localeCompare(b.name || '')
    if (sortBy === 'age') {
      const ageA = ageYearsAndMonths(a.date_of_birth)
      const ageB = ageYearsAndMonths(b.date_of_birth)
      const monthsA = ageA ? ageA.years * 12 + ageA.months : 0
      const monthsB = ageB ? ageB.years * 12 + ageB.months : 0
      return monthsA - monthsB
    }
    return 0
  })

  return (
    <div className="page" aria-label="Dashboard">
      <header className="page-header">
        <h1 id="dashboard-title" className="text-xl font-bold text-dark-primary flex items-center gap-2">
          <Icon icon="mdi:heart" className="text-red-400" width={28} height={28} aria-hidden />
          My Pets
        </h1>
      </header>

      {loading ? (
        <p className="text-dark-text-secondary" role="status" aria-live="polite">Loading…</p>
      ) : (
        <div className="card-panel">
          <div className="dashboard-pets-header">
            <h2 id="dashboard-pets-heading" className="text-lg font-bold text-dark-primary flex items-center gap-2">
              <Icon icon="mdi:paw" width={24} height={24} className="text-blue-400" aria-hidden />
              Your pets ({pets.length})
            </h2>
            <div className="dashboard-controls" style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem', alignItems: 'center' }} role="search" aria-label="Filter and sort pets">
              <input
                id="dashboard-search"
                type="text"
                placeholder="Search name, species, breed, age..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                style={{ minWidth: 200 }}
                aria-label="Search pets by name, species, breed, or age"
              />
              <select id="dashboard-sort" value={sortBy} onChange={(e) => setSortBy(e.target.value as SortOption)} style={{ width: 'auto' }} aria-label="Sort pets by">
                <option value="name">Sort by name</option>
                <option value="species">Sort by species</option>
                <option value="age">Sort by age</option>
              </select>
              <Link to="/pets/new" className="btn btn-primary btn-sm" style={{ marginLeft: 'auto' }}>
                <Icon icon="mdi:plus" width={18} height={18} aria-hidden />
                Add Pet
              </Link>
            </div>
          </div>
          {pets.length === 0 ? (
            <div className="empty-state">
              <p>No pets yet. Add your first pet to get started.</p>
              <Link to="/pets/new" className="btn btn-primary">
                <Icon icon="mdi:plus" width={20} height={20} />
                Add Pet
              </Link>
            </div>
          ) : (
            <>
              {filteredPets.length === 0 && (
                <p className="muted">No pets match your search.</p>
              )}
              <ul className="pet-grid">
                {sortedPets.map((pet) => (
                  <li key={pet.id}>
                    <Link to={`/pets/${pet.id}`} className="pet-card pet-card-horizontal" aria-label={`View ${pet.name || 'pet'}`}>
                      <div className="pet-card-image">
                        {petImageUrl(pet) ? (
                          <img src={petImageUrl(pet)!} alt="" role="presentation" />
                        ) : (
                          <span className="pet-card-placeholder" aria-hidden>{getSpeciesEmoji(pet.species)}</span>
                        )}
                      </div>
                      <div className="pet-card-body">
                        <h3 className="meta-line">
                          <Icon icon="mdi:tag" width={14} height={14} style={{ verticalAlign: 'middle', marginRight: 4 }} aria-hidden />
                          {pet.name}
                        </h3>
                        {pet.species && (
                          <p className="meta">
                            <Icon icon="mdi:dog" width={14} height={14} style={{ verticalAlign: 'middle', marginRight: 4 }} aria-hidden />
                            {pet.species}
                          </p>
                        )}
                        {pet.breed && (
                          <p className="meta">
                            <Icon icon="mdi:shape" width={14} height={14} style={{ verticalAlign: 'middle', marginRight: 4 }} aria-hidden />
                            {pet.breed}
                          </p>
                        )}
                        {pet.date_of_birth && (() => {
                          const age = ageYearsAndMonths(pet.date_of_birth)
                          const dateStr = formatDateOnly(pet.date_of_birth)
                          let ageStr: string | null = null
                          if (age) {
                            if (age.years < 1) {
                              ageStr = t('pet.ageMonths', { count: String(age.months) })
                            } else if (age.months === 0) {
                              ageStr = t('pet.ageYears', { years: String(age.years) })
                            } else {
                              ageStr = t('pet.ageYearsMonths', { years: String(age.years), months: String(age.months) })
                            }
                          }
                          return (
                            <>
                              <p className="meta">
                                <Icon icon="mdi:calendar" width={14} height={14} style={{ verticalAlign: 'middle', marginRight: 4 }} aria-hidden />
                                {dateStr}
                              </p>
                              {ageStr && (
                                <p className="meta">
                                  <Icon icon="mdi:calendar-clock" width={14} height={14} style={{ verticalAlign: 'middle', marginRight: 4 }} aria-hidden />
                                  {ageStr}
                                </p>
                              )}
                            </>
                          )
                        })()}
                        {latestWeight[pet.id] != null && (() => {
                          const d = weightToDisplay(latestWeight[pet.id], weightUnit)
                          return (
                            <p className="meta">
                              <Icon icon="mdi:scale-balance" width={14} height={14} style={{ verticalAlign: 'middle', marginRight: 4 }} aria-hidden />
                              {d.value} {d.unit}
                            </p>
                          )
                        })()}
                      </div>
                    </Link>
                  </li>
                ))}
              </ul>
            </>
          )}
        </div>
      )}
    </div>
  )
}

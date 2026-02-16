import { Link, useNavigate, useParams } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { customOptionsApi, petsApi, type CustomOptionsResponse, type Pet } from '../api/client'
import { useTranslation } from '../i18n/context'
import ComboBox from '../components/ComboBox'

function sortOptions(a: string[]): string[] {
  return [...a].sort((x, y) => x.localeCompare(y))
}

function toDateOnly(apiDate?: string | null): string {
  if (!apiDate) return ''
  if (apiDate.length === 10) return apiDate
  return apiDate.slice(0, 10)
}

export default function PetForm() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { t } = useTranslation()
  const isEdit = !!id
  const [pet, setPet] = useState<Partial<Pet> | null>({ name: '' })
  const [loading, setLoading] = useState(isEdit)
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [customOptions, setCustomOptions] = useState<CustomOptionsResponse | null>(null)

  useEffect(() => {
    if (isEdit && id) {
      petsApi.get(id).then(setPet).catch(() => setPet(null)).finally(() => setLoading(false))
    }
  }, [isEdit, id])

  useEffect(() => {
    customOptionsApi.get().then(setCustomOptions).catch(() => setCustomOptions(null))
  }, [])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!formPet.name?.trim()) return
    setSaving(true)
    try {
      const species = formPet.species ?? ''
      const breed = formPet.breed ?? ''
      const speciesOpts = customOptions?.species ?? []
      const breedOpts = species ? (customOptions?.breeds?.[species] ?? []) : []
      if (species && !speciesOpts.includes(species)) {
        await customOptionsApi.add('species', species)
      }
      if (breed && species && !breedOpts.includes(breed)) {
        await customOptionsApi.add('breed', breed, species)
      }
      if (isEdit && id) {
        await petsApi.update(id, formPet)
      } else {
        const created = await petsApi.create(formPet)
        navigate(`/pets/${created.id}`)
        return
      }
      navigate(`/pets/${id}`)
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!isEdit || !id) return
    const message =
      'Permanently delete this pet? All data—including vaccinations, weights, documents, photos, and uploaded files—will be deleted and cannot be recovered.'
    if (!window.confirm(message)) return
    setDeleting(true)
    try {
      await petsApi.delete(id)
      navigate('/', { replace: true })
    } catch {
      alert('Failed to delete pet.')
    } finally {
      setDeleting(false)
    }
  }

  if (loading) {
    return <div className="page"><p className="text-dark-text-secondary">Loading…</p></div>
  }
  if (isEdit && pet === null) {
    return (
      <div className="page">
        <p className="text-dark-text-secondary">Pet not found. <Link to="/" className="text-dark-accent hover:underline">Back to dashboard</Link></p>
      </div>
    )
  }
  const formPet = pet ?? { name: '' }
  const speciesOptions = sortOptions(customOptions?.species ?? [])
  const breedOptions = sortOptions(customOptions?.breeds?.[formPet.species ?? ''] ?? [])

  return (
    <div className="page">
      <header className="page-header" style={{ justifyContent: 'flex-start' }}>
        <h1>{isEdit ? 'Edit Pet' : 'Add Pet'}</h1>
      </header>
      <form onSubmit={handleSubmit} className="form">
        <label>
          Name *
          <input
            value={formPet.name ?? ''}
            onChange={(e) => setPet((p) => ({ ...p!, name: e.target.value }))}
            required
          />
        </label>
        <ComboBox
          label="Species"
          options={speciesOptions}
          value={formPet.species ?? ''}
          onChange={(v) => setPet((p) => ({ ...p!, species: v || undefined }))}
          placeholder="e.g. Dog, Cat, Bird..."
        />
        <ComboBox
          label="Breed"
          options={breedOptions}
          value={formPet.breed ?? ''}
          onChange={(v) => setPet((p) => ({ ...p!, breed: v || undefined }))}
          placeholder={formPet.species ? `e.g. ${breedOptions[0]}...` : 'Select species first'}
        />
        <label>
          Date of birth
          <input
            type="date"
            value={toDateOnly(formPet.date_of_birth) ?? ''}
            onChange={(e) => setPet((p) => ({ ...p!, date_of_birth: e.target.value || undefined }))}
          />
        </label>
        <div className="form-row-gender-fixed">
          <label>
            Gender
            <select
              value={formPet.gender ?? ''}
              onChange={(e) => setPet((p) => ({ ...p!, gender: e.target.value || undefined }))}
            >
              <option value="">—</option>
              <option value="female">Female</option>
              <option value="male">Male</option>
              <option value="other">Other</option>
            </select>
          </label>
          <label>
            {t('pet.fixedLabel')}
            <select
              value={formPet.fixed === true ? 'yes' : 'no'}
              onChange={(e) => setPet((p) => ({ ...p!, fixed: e.target.value === 'yes' }))}
            >
              <option value="no">{t('common.no')}</option>
              <option value="yes">{t('common.yes')}</option>
            </select>
          </label>
        </div>
        <label>
          Color
          <input
            value={formPet.color ?? ''}
            onChange={(e) => setPet((p) => ({ ...p!, color: e.target.value || undefined }))}
          />
        </label>
        <label>
          Microchip ID
          <input
            value={formPet.microchip_id ?? ''}
            onChange={(e) => setPet((p) => ({ ...p!, microchip_id: e.target.value || undefined }))}
          />
        </label>
        <label>
          Microchip company
          <input
            value={formPet.microchip_company ?? ''}
            onChange={(e) => setPet((p) => ({ ...p!, microchip_company: e.target.value || undefined }))}
            placeholder="e.g. HomeAgain, AKC"
          />
        </label>
        <label>
          Notes
          <textarea
            value={formPet.notes ?? ''}
            onChange={(e) => setPet((p) => ({ ...p!, notes: e.target.value || undefined }))}
            rows={3}
          />
        </label>
        <div className="form-actions">
          <button type="submit" className="btn btn-primary" disabled={saving}>
            {saving ? 'Saving…' : isEdit ? 'Save' : 'Add Pet'}
          </button>
          <button type="button" className="btn btn-secondary" onClick={() => navigate(-1)}>
            Cancel
          </button>
        </div>
      </form>
      {isEdit && id && (
        <section className="section" style={{ marginTop: '2rem', borderTop: '1px solid var(--dark-border)', paddingTop: '1.5rem' }}>
          <h2 className="text-lg font-bold text-dark-primary" style={{ marginBottom: '0.5rem' }}>Delete pet</h2>
          <p className="muted" style={{ marginBottom: '1rem' }}>
            This will permanently delete this pet and all associated data: vaccinations, weight history, documents, photos, and all uploaded files. This cannot be undone.
          </p>
          <button
            type="button"
            className="btn btn-secondary"
            style={{ background: 'var(--dark-danger, #dc2626)', color: '#fff', border: 'none' }}
            disabled={deleting}
            onClick={handleDelete}
          >
            {deleting ? 'Deleting…' : 'Delete pet permanently'}
          </button>
        </section>
      )}
    </div>
  )
}

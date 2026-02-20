import { Link, useParams } from 'react-router-dom'
import { Icon } from '@iconify/react'
import { useCallback, useEffect, useRef, useState } from 'react'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import {
  customOptionsApi,
  petsApi,
  vaccinationsApi,
  weightsApi,
  documentsApi,
  photosApi,
  type CustomOptionsResponse,
  type Pet,
  type Vaccination,
  type WeightEntry,
  type Document,
  type PetPhoto,
  type WeightUnit,
} from '../api/client'
import { useAuth } from '../contexts/AuthContext'
import { usePWAInstall } from '../contexts/PWAInstallContext'
import { useTranslation } from '../i18n/context'
import { addMonthsToDate, getDurationMonths } from '../data/vaccinations'
import ComboBox from '../components/ComboBox'
import { getSpeciesEmoji } from '../utils/speciesEmoji'
import { validateDocumentFile, validateImageFile } from '../lib/fileTypeValidation'

function weightToDisplay(lbs: number, unit: WeightUnit): { value: number; unit: string } {
  if (unit === 'kg') return { value: Math.round((lbs / 2.20462) * 100) / 100, unit: 'kg' }
  return { value: lbs, unit: 'lbs' }
}

function formatDateOnly(apiDate: string): string {
  if (!apiDate) return ''
  const d = apiDate.length === 10 ? `${apiDate}T12:00:00` : apiDate
  return new Date(d).toLocaleDateString(undefined, { dateStyle: 'medium' })
}

function todayStr(): string {
  return new Date().toISOString().slice(0, 10)
}

export default function PetDetail() {
  const { id } = useParams<{ id: string }>()
  const { user } = useAuth()
  const { t } = useTranslation()
  const [pet, setPet] = useState<Pet | null>(null)
  const [vaccinations, setVaccinations] = useState<Vaccination[]>([])
  const [weights, setWeights] = useState<WeightEntry[]>([])
  const [photos, setPhotos] = useState<PetPhoto[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<'vaccinations' | 'weights' | 'documents' | 'photos'>('vaccinations')

  const load = useCallback(() => {
    if (!id) return
    Promise.all([
      petsApi.get(id),
      vaccinationsApi.list(id),
      weightsApi.list(id),
      photosApi.list(id),
    ])
      .then(([p, v, w, ph]) => {
        setPet(p)
        setVaccinations(Array.isArray(v) ? v : [])
        setWeights(Array.isArray(w) ? w : [])
        setPhotos(Array.isArray(ph) ? ph : [])
      })
      .catch(() => setPet(null))
      .finally(() => setLoading(false))
  }, [id])

  useEffect(() => {
    load()
  }, [load])

  if (loading || !pet) {
    return (
      <div className="page">
        {loading ? (
          <p className="text-dark-text-secondary">{t('common.loading')}</p>
        ) : (
          <p className="text-dark-text-secondary">
            {t('pet.notFound')} <Link to="/" className="text-dark-accent hover:underline">{t('pet.backToDashboard')}</Link>
          </p>
        )}
      </div>
    )
  }

  return (
    <div className="page" aria-label={`Pet: ${pet.name}`}>
      <div className="pet-detail-layout">
        <aside className="pet-detail-sidebar" role="complementary" aria-label="Pet profile">
          <div className="pet-hero-card">
            <div className="pet-hero">
              <div className="pet-avatar large">
                {pet.photo_url ? (
                  <img
                    src={pet.photo_url.startsWith('http') ? pet.photo_url : `${window.location.origin}${pet.photo_url}`}
                    alt={pet.name ? `Photo of ${pet.name}` : ''}
                  />
                ) : (
                  <span aria-hidden>{getSpeciesEmoji(pet.species)}</span>
                )}
              </div>
              <div>
                <h1 id="pet-detail-name">{pet.name}</h1>
                {pet.species && <p className="meta">{pet.species}</p>}
                {pet.breed && <p className="meta">{pet.breed}</p>}
                {pet.date_of_birth && <p className="meta">Born {formatDateOnly(pet.date_of_birth)}</p>}
                {pet.fixed && (
                  <p className="meta">
                    {pet.gender === 'female' ? t('pet.spayed') : pet.gender === 'male' ? t('pet.neutered') : t('pet.spayedNeutered')}
                  </p>
                )}
              </div>
            </div>
          </div>
          <Link to={`/pets/${id}/edit`} className="btn btn-secondary" style={{ width: '100%', marginBottom: '1rem' }} aria-label={`Edit ${pet.name}`}>
            <Icon icon="mdi:pencil" width={18} height={18} aria-hidden />
            {t('pet.editPet')}
          </Link>
        </aside>

        <div className="pet-detail-main">
          <div className="tabs-wrap" role="tablist" aria-label="Pet data sections">
            <div className="tabs">
            <button
              role="tab"
              aria-selected={activeTab === 'vaccinations'}
              aria-controls="pet-tab-vaccinations"
              id="pet-tab-btn-vaccinations"
              className={activeTab === 'vaccinations' ? 'active' : ''}
              onClick={() => setActiveTab('vaccinations')}
            >
              <Icon icon="mdi:needle" width={20} height={20} aria-hidden />
              {t('pet.vaccinations')}
            </button>
            <button
              role="tab"
              aria-selected={activeTab === 'weights'}
              aria-controls="pet-tab-weights"
              id="pet-tab-btn-weights"
              className={activeTab === 'weights' ? 'active' : ''}
              onClick={() => setActiveTab('weights')}
            >
              <Icon icon="mdi:scale-balance" width={20} height={20} aria-hidden />
              {t('pet.weight')}
            </button>
            <button
              role="tab"
              aria-selected={activeTab === 'documents'}
              aria-controls="pet-tab-documents"
              id="pet-tab-btn-documents"
              className={activeTab === 'documents' ? 'active' : ''}
              onClick={() => setActiveTab('documents')}
            >
              <Icon icon="mdi:file-document-multiple" width={20} height={20} aria-hidden />
              {t('pet.documents')}
            </button>
            <button
              role="tab"
              aria-selected={activeTab === 'photos'}
              aria-controls="pet-tab-photos"
              id="pet-tab-btn-photos"
              className={activeTab === 'photos' ? 'active' : ''}
              onClick={() => setActiveTab('photos')}
            >
              <Icon icon="mdi:image-multiple" width={20} height={20} aria-hidden />
              {t('pet.photos')}
            </button>
            </div>
          </div>

          {activeTab === 'vaccinations' && (
            <div id="pet-tab-vaccinations" role="tabpanel" aria-labelledby="pet-tab-btn-vaccinations" tabIndex={0}>
              <VaccinationsSection petId={id!} pet={pet} list={vaccinations} currency={user?.currency ?? 'USD'} onUpdate={load} />
            </div>
          )}
          {activeTab === 'weights' && (
            <div id="pet-tab-weights" role="tabpanel" aria-labelledby="pet-tab-btn-weights" tabIndex={0}>
              <WeightsSection petId={id!} list={weights} weightUnit={user?.weight_unit ?? 'lbs'} onUpdate={load} />
            </div>
          )}
          {activeTab === 'documents' && (
            <div id="pet-tab-documents" role="tabpanel" aria-labelledby="pet-tab-btn-documents" tabIndex={0}>
              <DocumentsSection petId={id!} onUpdate={load} />
            </div>
          )}
          {activeTab === 'photos' && (
            <div id="pet-tab-photos" role="tabpanel" aria-labelledby="pet-tab-btn-photos" tabIndex={0}>
              <PhotosSection petId={id!} pet={pet} photos={photos} onUpdate={load} />
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

const UPLOADS_BASE = '/api/uploads'

function VaccinationsSection({
  petId,
  pet,
  list,
  currency,
  onUpdate,
}: {
  petId: string
  pet: Pet
  list: Vaccination[]
  currency: string
  onUpdate: () => void
}) {
  const { t } = useTranslation()
  const [adding, setAdding] = useState(false)
  const [name, setName] = useState('')
  const [administeredAt, setAdministeredAt] = useState(() => todayStr())
  const [nextDue, setNextDue] = useState('')
  const [costUsd, setCostUsd] = useState('')
  const [showExpiryHint, setShowExpiryHint] = useState(false)
  const [customOptions, setCustomOptions] = useState<CustomOptionsResponse | null>(null)

  const costLabel = t('pet.costLabel', { currency })
  const speciesKey = pet?.species ?? ''
  const vaccineOptions = (customOptions?.vaccinations?.[speciesKey] ?? []).slice().sort((a, b) => a.localeCompare(b))
  const vaccinationDurations = customOptions?.vaccination_durations ?? {}

  useEffect(() => {
    customOptionsApi.get().then(setCustomOptions).catch(() => setCustomOptions(null))
  }, [])

  useEffect(() => {
    if (!showExpiryHint) return
    const id = window.setTimeout(() => setShowExpiryHint(false), 6000)
    return () => window.clearTimeout(id)
  }, [showExpiryHint])

  function updateNextDueFromPreset(adminDate: string, vaccineName: string) {
    const duration =
      vaccinationDurations[speciesKey]?.[vaccineName] ??
      getDurationMonths(pet?.species ?? null, vaccineName)
    if (duration != null && adminDate) {
      setNextDue(addMonthsToDate(adminDate, duration))
      setShowExpiryHint(true)
    }
  }

  function handleNameChange(v: string) {
    setName(v)
    if (administeredAt) updateNextDueFromPreset(administeredAt, v)
  }

  function handleAdministeredChange(adminDate: string) {
    setAdministeredAt(adminDate)
    if (name) updateNextDueFromPreset(adminDate, name)
  }

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    await vaccinationsApi.create(petId, {
      name,
      administered_at: administeredAt,
      next_due: nextDue || undefined,
      cost_usd: costUsd ? parseFloat(costUsd) : undefined,
    })
    if (speciesKey && !vaccineOptions.includes(name)) {
      await customOptionsApi.add('vaccination', name, speciesKey)
      customOptionsApi.get().then(setCustomOptions).catch(() => {})
    }
    setName('')
    setAdministeredAt(todayStr())
    setNextDue('')
    setCostUsd('')
    setAdding(false)
    onUpdate()
  }

  async function handleDelete(id: string) {
    if (!confirm(t('pet.deleteVaccinationConfirm'))) return
    await vaccinationsApi.delete(petId, id)
    onUpdate()
  }

  function formatCost(amount: number): string {
    try {
      return new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(amount)
    } catch {
      return `${amount.toFixed(2)} ${currency}`
    }
  }

  return (
    <section className="section">
      <h2>{t('pet.vaccinations')}</h2>
      {!adding ? (
        <button className="btn btn-secondary" onClick={() => setAdding(true)}>
          <Icon icon="mdi:plus" width={18} height={18} />
          {t('pet.addVaccination')}
        </button>
      ) : (
        <>
          {showExpiryHint && (
            <div className="toast toast-info" role="status">
              <Icon icon="mdi:information-outline" width={20} height={20} aria-hidden />
              <span>{t('pet.expiryHint')}</span>
              <button
                type="button"
                className="toast-dismiss"
                onClick={() => setShowExpiryHint(false)}
                aria-label="Dismiss"
              >
                <Icon icon="mdi:close" width={18} height={18} />
              </button>
            </div>
          )}
        <form onSubmit={handleAdd} className="form-inline vaccinations-add-form">
          <span className="vacc-form-label">{t('common.name')}</span>
          <span className="vacc-form-label">{t('common.administered')}</span>
          <span className="vacc-form-label">{t('common.expires')}</span>
          <span className="vacc-form-label">{costLabel}</span>
          <span />
          <div className="vacc-form-combo">
            <ComboBox
              label=""
              options={vaccineOptions}
              value={name}
              onChange={handleNameChange}
              placeholder={t('common.name')}
              required
            />
          </div>
          <input
            type="date"
            value={administeredAt}
            onChange={(e) => handleAdministeredChange(e.target.value)}
            required
            className="input"
            title="Administered"
          />
          <input
            type="date"
            value={nextDue}
            onChange={(e) => setNextDue(e.target.value)}
            className="input"
            title="Expires"
          />
          <input
            type="number"
            step="0.01"
            min="0"
            placeholder={costLabel}
            value={costUsd}
            onChange={(e) => setCostUsd(e.target.value)}
            className="input"
          />
          <div className="vacc-form-actions">
            <button type="submit" className="btn btn-primary">{t('common.save')}</button>
            <button type="button" className="btn btn-secondary" onClick={() => setAdding(false)}>{t('common.cancel')}</button>
          </div>
        </form>
        </>
      )}
      <ul className="list">
        {list.map((v) => (
          <li key={v.id} className="list-item">
            <div>
              <strong>{v.name}</strong> — {formatDateOnly(v.administered_at)}
              {v.next_due && <span> ({t('pet.next')}: {formatDateOnly(v.next_due)})</span>}
              {v.cost_usd != null && <span> — {formatCost(v.cost_usd)}</span>}
            </div>
            <button type="button" className="btn btn-sm btn-danger" onClick={() => handleDelete(v.id)}>{t('common.delete')}</button>
          </li>
        ))}
      </ul>
    </section>
  )
}

function WeightsSection({
  petId,
  list,
  weightUnit,
  onUpdate,
}: {
  petId: string
  list: WeightEntry[]
  weightUnit: WeightUnit
  onUpdate: () => void
}) {
  const [adding, setAdding] = useState(false)
  const [weightValue, setWeightValue] = useState('')
  const [measuredAt, setMeasuredAt] = useState(() => todayStr())
  const [approximate, setApproximate] = useState(false)

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    const num = parseFloat(weightValue)
    const payload =
      weightUnit === 'kg'
        ? { weight_kg: num, entry_unit: 'kg' as const, measured_at: measuredAt, approximate }
        : { weight_lbs: num, entry_unit: 'lbs' as const, measured_at: measuredAt, approximate }
    await weightsApi.create(petId, payload)
    setWeightValue('')
    setMeasuredAt(todayStr())
    setApproximate(false)
    setAdding(false)
    onUpdate()
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this weight entry?')) return
    await weightsApi.delete(petId, id)
    onUpdate()
  }

  const unitLabel = weightUnit === 'kg' ? 'kg' : 'lbs'
  const chartData = [...list]
    .sort((a, b) => (a.measured_at < b.measured_at ? -1 : 1))
    .map((w) => {
      const d = weightToDisplay(w.weight_lbs, weightUnit)
      return { date: formatDateOnly(w.measured_at), weight: d.value, fullDate: w.measured_at, approximate: w.approximate }
    })

  return (
    <section className="section">
      <h2>Weight history</h2>
      {chartData.length > 0 && (
        <div className="weight-chart-wrap" style={{ height: 220, marginBottom: '1rem' }}>
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData} margin={{ top: 5, right: 10, left: 0, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--dark-border, #333)" />
              <XAxis dataKey="date" tick={{ fill: 'var(--dark-text-secondary, #999)', fontSize: 12 }} />
              <YAxis tick={{ fill: 'var(--dark-text-secondary, #999)', fontSize: 12 }} unit={` ${unitLabel}`} />
              <Tooltip
                contentStyle={{ background: 'var(--dark-card, #1e1e1e)', border: '1px solid var(--dark-border)' }}
                labelStyle={{ color: 'var(--dark-text)' }}
                formatter={(value: number | undefined) => [value != null ? `${value} ${unitLabel}` : '', 'Weight']}
                labelFormatter={(_, payload) => payload?.[0]?.payload?.date}
              />
              <Line
                type="monotone"
                dataKey="weight"
                stroke="var(--dark-accent, #7c9cbf)"
                strokeWidth={2}
                dot={(props) => {
                  const { cx, cy, payload } = props
                  const fill = payload?.approximate ? '#eab308' : '#22c55e'
                  return cx != null && cy != null ? <circle cx={cx} cy={cy} r={4} fill={fill} stroke="var(--dark-card, #1e1e1e)" strokeWidth={1} /> : null
                }}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}
      {!adding ? (
        <button className="btn btn-secondary" onClick={() => setAdding(true)}>
          <Icon icon="mdi:plus" width={18} height={18} />
          Add weight
        </button>
      ) : (
        <form onSubmit={handleAdd} className="form-inline">
          <input
            type="number"
            step="0.1"
            min="0"
            placeholder={`Weight (${unitLabel})`}
            value={weightValue}
            onChange={(e) => setWeightValue(e.target.value)}
            required
          />
          <input
            type="date"
            value={measuredAt}
            onChange={(e) => setMeasuredAt(e.target.value)}
            required
          />
          <label style={{ display: 'flex', alignItems: 'center', gap: 6, whiteSpace: 'nowrap' }}>
            <input type="checkbox" checked={approximate} onChange={(e) => setApproximate(e.target.checked)} />
            Approximate
          </label>
          <button type="submit" className="btn btn-primary">Save</button>
          <button type="button" className="btn btn-secondary" onClick={() => setAdding(false)}>Cancel</button>
        </form>
      )}
      <ul className="list">
        {list.map((w) => {
          const d = weightToDisplay(w.weight_lbs, weightUnit)
          return (
            <li key={w.id} className="list-item">
              <span><strong>{d.value} {d.unit}</strong>{w.approximate && ' (approx.)'} — {formatDateOnly(w.measured_at)}</span>
              <button type="button" className="btn btn-sm btn-danger" onClick={() => handleDelete(w.id)}>Delete</button>
            </li>
          )
        })}
      </ul>
    </section>
  )
}

function DocumentsSection({
  petId,
  onUpdate,
}: {
  petId: string
  onUpdate: () => void
}) {
  const { t } = useTranslation()
  const [sort, setSort] = useState<'date' | 'name'>('date')
  const [search, setSearch] = useState('')
  const [docList, setDocList] = useState<Document[]>([])
  const [uploading, setUploading] = useState(false)
  const [docName, setDocName] = useState('')
  const [docFile, setDocFile] = useState<File | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [editingDocId, setEditingDocId] = useState<string | null>(null)
  const [editingDocName, setEditingDocName] = useState('')

  const loadDocs = useCallback(() => {
    documentsApi.list(petId, { sort, search: search.trim() || undefined }).then(setDocList)
  }, [petId, sort, search])

  useEffect(() => {
    loadDocs()
  }, [loadDocs])

  async function handleUpload(e: React.FormEvent) {
    e.preventDefault()
    if (!docFile || !docName.trim()) return
    const validation = await validateDocumentFile(docFile)
    if (!validation.ok) {
      alert(validation.error)
      return
    }
    setUploading(true)
    try {
      await documentsApi.create(petId, docName.trim(), docFile)
      setDocName('')
      setDocFile(null)
      if (fileInputRef.current) fileInputRef.current.value = ''
      onUpdate()
      loadDocs()
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Upload failed')
    } finally {
      setUploading(false)
    }
  }

  const documentUrl = (d: Document) =>
    `${window.location.origin}${UPLOADS_BASE}/${d.file_path}`

  async function openDocument(d: Document) {
    const url = documentUrl(d)
    try {
      const res = await fetch(url, { credentials: 'include' })
      if (!res.ok) return
      const blob = await res.blob()
      const blobUrl = URL.createObjectURL(blob)
      const w = window.open(blobUrl, '_blank', 'noopener,noreferrer')
      if (w) setTimeout(() => URL.revokeObjectURL(blobUrl), 60000)
    } catch {
      window.open(url, '_blank', 'noopener,noreferrer')
    }
  }

  return (
    <section className="section">
      <h2>Documents</h2>
      <p className="muted" style={{ marginBottom: '1rem' }}>{t('pet.documentsOcrNote')}</p>
      <div className="form-inline" style={{ flexWrap: 'wrap', gap: '0.5rem', marginBottom: '1rem' }}>
        <input
          type="text"
          placeholder="Search by name or content"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="input"
          style={{ maxWidth: 200 }}
        />
        <select value={sort} onChange={(e) => setSort(e.target.value as 'date' | 'name')} className="input" style={{ width: 'auto' }}>
          <option value="date">Sort by date</option>
          <option value="name">Sort by name</option>
        </select>
      </div>
      <form onSubmit={handleUpload} className="form-inline documents-upload-form">
        <input
          type="text"
          placeholder="Document name"
          value={docName}
          onChange={(e) => setDocName(e.target.value)}
          className="input"
          required
        />
        <label className="file-input-label">
          <span className="file-input-button btn btn-secondary">
            <Icon icon="mdi:paperclip" width={18} height={18} />
            {docFile ? docFile.name : 'Choose file'}
          </span>
          <input
            ref={fileInputRef}
            type="file"
            className="file-input-hidden"
            accept=".pdf,.doc,.docx,.rtf,.odt,.png,.jpg,.jpeg"
            onChange={(e) => {
              const file = e.target.files?.[0] ?? null
              setDocFile(file)
              if (file && !docName.trim()) setDocName(file.name)
            }}
            required
          />
        </label>
        <button type="submit" className="btn btn-primary" disabled={uploading}>
          {uploading ? 'Uploading…' : 'Upload'}
        </button>
      </form>
      <ul className="list document-list">
        {docList.map((d) => (
          <li key={d.id} className="list-item document-list-item">
            <div className="document-item-info">
              {editingDocId === d.id ? (
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
                  <input
                    type="text"
                    value={editingDocName}
                    onChange={(e) => setEditingDocName(e.target.value)}
                    className="input"
                    style={{ minWidth: 160 }}
                    autoFocus
                  />
                  <button
                    type="button"
                    className="btn btn-sm btn-primary"
                    onClick={async () => {
                      if (!editingDocName.trim()) return
                      await documentsApi.update(petId, d.id, editingDocName.trim())
                      setEditingDocId(null)
                      loadDocs()
                    }}
                  >
                    Save
                  </button>
                  <button type="button" className="btn btn-sm btn-secondary" onClick={() => setEditingDocId(null)}>
                    Cancel
                  </button>
                </div>
              ) : (
                <>
                  <button
                    type="button"
                    className="document-item-name"
                    onClick={() => openDocument(d)}
                    title="View or download"
                  >
                    <Icon icon="mdi:file-document-outline" width={18} height={18} />
                    <span>{d.name}</span>
                  </button>
                  <button
                    type="button"
                    className="btn btn-sm btn-secondary"
                    onClick={() => {
                      setEditingDocId(d.id)
                      setEditingDocName(d.name)
                    }}
                    title="Edit name"
                  >
                    <Icon icon="mdi:pencil" width={14} height={14} />
                  </button>
                </>
              )}
              {d.created_at && <span className="document-item-date">{formatDateOnly(d.created_at)}</span>}
            </div>
            <button
              type="button"
              className="btn btn-sm btn-danger"
              onClick={async () => {
                if (!confirm('Delete this document?')) return
                await documentsApi.delete(petId, d.id)
                onUpdate()
                loadDocs()
              }}
            >
              Delete
            </button>
          </li>
        ))}
      </ul>
      {docList.length === 0 && <p className="muted">No documents. Upload a file and give it a name.</p>}
    </section>
  )
}

const IMAGE_ACCEPT = 'image/jpeg,image/png,image/gif,image/webp'

function PhotosSection({
  petId,
  pet,
  photos,
  onUpdate,
}: {
  petId: string
  pet: Pet
  photos: PetPhoto[]
  onUpdate: () => void
}) {
  const pwa = usePWAInstall()
  const isMobile = pwa?.isMobile ?? false
  const [uploading, setUploading] = useState(false)
  const cameraInputRef = useRef<HTMLInputElement>(null)
  const galleryInputRef = useRef<HTMLInputElement>(null)
  const photoUrl = (p: PetPhoto) => `${window.location.origin}${UPLOADS_BASE}/${p.file_path}`

  async function handleUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    const validation = await validateImageFile(file)
    if (!validation.ok) {
      alert(validation.error)
      e.target.value = ''
      return
    }
    setUploading(true)
    try {
      await photosApi.upload(petId, file)
      onUpdate()
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Upload failed')
    } finally {
      setUploading(false)
      e.target.value = ''
      cameraInputRef.current && (cameraInputRef.current.value = '')
      galleryInputRef.current && (galleryInputRef.current.value = '')
    }
  }

  async function setAvatar(photoId: string) {
    try {
      await photosApi.setAvatar(petId, photoId)
      onUpdate()
    } catch {
      alert('Failed to set profile picture')
    }
  }

  async function remove(photoId: string) {
    if (!confirm('Remove this photo?')) return
    try {
      await photosApi.delete(petId, photoId)
      onUpdate()
    } catch {
      alert('Delete failed')
    }
  }

  return (
    <section className="section">
      <h2>Photos</h2>
      <p className="muted">Upload images and choose one as the profile picture.</p>
      <div className="photos-upload-actions" style={{ marginBottom: '1rem', display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
        {isMobile && (
          <label className="btn btn-primary" style={{ cursor: 'pointer', margin: 0 }}>
            <Icon icon="mdi:camera" width={18} height={18} />
            {uploading ? ' Uploading…' : ' Take photo'}
            <input
              ref={cameraInputRef}
              type="file"
              accept={IMAGE_ACCEPT}
              capture="environment"
              hidden
              onChange={handleUpload}
              disabled={uploading}
            />
          </label>
        )}
        <label className="btn btn-secondary" style={{ cursor: 'pointer', margin: 0 }}>
          <Icon icon="mdi:image-multiple" width={18} height={18} />
          {uploading ? ' Uploading…' : (isMobile ? ' Choose from gallery' : ' Upload photo')}
          <input
            ref={galleryInputRef}
            type="file"
            accept={IMAGE_ACCEPT}
            hidden
            onChange={handleUpload}
            disabled={uploading}
          />
        </label>
      </div>
      <div className="photos-grid" style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(140px, 1fr))', gap: '1rem' }}>
        {photos.map((p) => (
          <div key={p.id} className="photo-card" style={{ position: 'relative', borderRadius: 8, overflow: 'hidden', background: 'var(--dark-card)' }}>
            <img src={photoUrl(p)} alt="" style={{ width: '100%', aspectRatio: '1', objectFit: 'cover', display: 'block' }} />
            <div style={{ padding: '0.5rem', display: 'flex', flexWrap: 'wrap', gap: '0.25rem' }}>
              <button
                type="button"
                className="btn btn-sm btn-primary"
                onClick={() => setAvatar(p.id)}
                title="Set as profile picture"
              >
                <Icon icon="mdi:account-circle" width={16} height={16} />
              </button>
              <button type="button" className="btn btn-sm btn-danger" onClick={() => remove(p.id)} title="Remove">
                <Icon icon="mdi:delete" width={16} height={16} />
              </button>
            </div>
            {pet.photo_url && photoUrl(p) === (pet.photo_url.startsWith('http') ? pet.photo_url : `${window.location.origin}${pet.photo_url}`) && (
              <span style={{ position: 'absolute', top: 4, right: 4, background: 'var(--dark-accent)', color: '#fff', fontSize: 10, padding: '2px 6px', borderRadius: 4 }}>Profile</span>
            )}
          </div>
        ))}
      </div>
      {photos.length === 0 && <p className="muted">No photos yet. Upload an image to get started.</p>}
    </section>
  )
}

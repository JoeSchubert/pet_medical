import { useEffect, useState } from 'react'
import { Icon } from '@iconify/react'
import { defaultOptionsApi, type DefaultOptionItem } from '../api/client'
import { useTranslation } from '../i18n/context'

type Tab = 'species' | 'breed' | 'vaccination'

export default function AdminDefaultOptions() {
  const { t } = useTranslation()
  const [list, setList] = useState<DefaultOptionItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [tab, setTab] = useState<Tab>('species')
  const [addValue, setAddValue] = useState('')
  const [addContext, setAddContext] = useState('')
  const [addDuration, setAddDuration] = useState<string>('')
  const [addSubmitting, setAddSubmitting] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')
  const [editContext, setEditContext] = useState('')
  const [editDuration, setEditDuration] = useState<string>('')
  const [listFilterSpecies, setListFilterSpecies] = useState('')

  function load() {
    setLoading(true)
    defaultOptionsApi
      .list()
      .then(setList)
      .catch(() => setError('Failed to load options'))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
  }, [])

  const filtered = list.filter((o) => o.option_type === tab)
  const listBySpecies = listFilterSpecies
    ? filtered.filter((o) => (o.context || '') === listFilterSpecies)
    : filtered
  const displayList = tab === 'species' ? filtered : listBySpecies
  const speciesList = list.filter((o) => o.option_type === 'species').map((o) => o.value)

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    const value = addValue.trim()
    if (!value) return
    const context = tab === 'species' ? '' : addContext.trim()
    if (tab !== 'species' && !context) {
      setError('Context (e.g. species) required for breed and vaccination')
      return
    }
    setAddSubmitting(true)
    setError('')
    try {
      const body: Omit<DefaultOptionItem, 'id'> = {
        option_type: tab,
        value,
        context,
        sort_order: filtered.length,
      }
      if (tab === 'vaccination' && addDuration.trim()) {
        const n = parseInt(addDuration, 10)
        if (!Number.isNaN(n) && n > 0) body.duration_months = n
      }
      await defaultOptionsApi.create(body)
      setAddValue('')
      setAddContext('')
      setAddDuration('')
      load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add')
    } finally {
      setAddSubmitting(false)
    }
  }

  async function handleUpdate(id: string) {
    const value = editValue.trim()
    if (!value) return
    setError('')
    try {
      const body: Partial<DefaultOptionItem> = { value, context: editContext.trim() }
      if (tab === 'vaccination' && editDuration.trim()) {
        const n = parseInt(editDuration, 10)
        body.duration_months = Number.isNaN(n) ? undefined : n
      } else if (tab === 'vaccination') {
        body.duration_months = undefined
      }
      await defaultOptionsApi.update(id, body)
      setEditingId(null)
      load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update')
    }
  }

  async function handleDelete(id: string) {
    if (!confirm(t('common.deleteConfirm') || 'Delete this item?')) return
    setError('')
    try {
      await defaultOptionsApi.delete(id)
      if (editingId === id) setEditingId(null)
      load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete')
    }
  }

  function startEdit(item: DefaultOptionItem) {
    setEditingId(item.id)
    setEditValue(item.value)
    setEditContext(item.context)
    setEditDuration(item.duration_months != null ? String(item.duration_months) : '')
  }

  if (loading) return <div className="page"><p className="text-dark-text-secondary">Loading…</p></div>

  return (
    <div className="page">
      <header className="page-header">
        <h1>{t('admin.defaultOptionsTitle')}</h1>
      </header>
      <p className="muted" style={{ marginBottom: '1rem' }}>
        {t('admin.defaultOptionsHelp')}
      </p>

      {error && <div className="toast toast-info" style={{ borderColor: 'var(--dark-accent)' }}>{error}</div>}

      <div className="admin-options-tabs">
        {(['species', 'breed', 'vaccination'] as const).map((tabKey) => (
          <button
            key={tabKey}
            type="button"
            className={`btn ${tab === tabKey ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setTab(tabKey)}
          >
            {tabKey === 'species' ? t('admin.tabSpecies') : tabKey === 'breed' ? t('admin.tabBreeds') : t('admin.tabVaccinations')}
          </button>
        ))}
      </div>

      <section className="section card-panel">
        <h2>
          {tab === 'species' ? t('admin.addSpecies') : tab === 'breed' ? t('admin.addBreed') : t('admin.addVaccination')}
        </h2>
        <form onSubmit={handleAdd} className="form-inline" style={{ flexWrap: 'wrap', gap: '0.5rem', alignItems: 'center' }}>
          {tab !== 'species' && (
            <select
              value={addContext}
              onChange={(e) => setAddContext(e.target.value)}
              className="input"
              required
              title="Species"
            >
              <option value="">— {t('admin.selectSpecies')} —</option>
              {speciesList.map((s) => (
                <option key={s} value={s}>{s}</option>
              ))}
            </select>
          )}
          <input
            type="text"
            placeholder={tab === 'species' ? 'e.g. Dog, Cat' : tab === 'breed' ? 'Breed name' : 'Vaccination name'}
            value={addValue}
            onChange={(e) => setAddValue(e.target.value)}
            className="input"
            style={{ minWidth: '12rem' }}
          />
          {tab === 'vaccination' && (
            <input
              type="number"
              min={1}
              placeholder={t('admin.durationMonths')}
              value={addDuration}
              onChange={(e) => setAddDuration(e.target.value)}
              className="input"
              style={{ width: '11rem' }}
            />
          )}
          <button type="submit" className="btn btn-primary" disabled={addSubmitting}>
            {addSubmitting ? t('common.saving') : t('common.add')}
          </button>
        </form>
      </section>

      <section className="section card-panel">
        <div className="admin-options-list-header">
          <h2 style={{ margin: 0 }}>{tab === 'species' ? t('admin.listSpecies') : tab === 'breed' ? t('admin.listBreeds') : t('admin.listVaccinations')}</h2>
          {tab !== 'species' && (
            <label className="admin-options-filter">
              <span className="text-dark-text-secondary">{t('admin.filterBySpecies')}</span>
              <select
                value={listFilterSpecies}
                onChange={(e) => setListFilterSpecies(e.target.value)}
                className="admin-options-filter-select"
              >
                <option value="">{t('admin.allSpecies')}</option>
                {speciesList.map((s) => (
                  <option key={s} value={s}>{s}</option>
                ))}
              </select>
            </label>
          )}
        </div>
        <ul className="list" style={{ listStyle: 'none', padding: 0, margin: 0 }}>
          {displayList
            .sort((a, b) => (a.context || '').localeCompare(b.context || '') || a.value.localeCompare(b.value))
            .map((item) => (
              <li key={item.id} className="list-item" style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap', marginBottom: '0.5rem' }}>
                {editingId === item.id ? (
                  <>
                    {tab !== 'species' && (
                      <input
                        type="text"
                        value={editContext}
                        onChange={(e) => setEditContext(e.target.value)}
                        className="input"
                        placeholder="Context"
                        style={{ width: '8rem' }}
                      />
                    )}
                    <input
                      type="text"
                      value={editValue}
                      onChange={(e) => setEditValue(e.target.value)}
                      className="input"
                      style={{ flex: 1, minWidth: '10rem' }}
                    />
                    {tab === 'vaccination' && (
                      <input
                        type="number"
                        min={1}
                        value={editDuration}
                        onChange={(e) => setEditDuration(e.target.value)}
                        className="input"
                        style={{ width: '5rem' }}
                        placeholder="Months"
                      />
                    )}
                    <button type="button" className="btn btn-primary" onClick={() => handleUpdate(item.id)}>
                      {t('common.save')}
                    </button>
                    <button type="button" className="btn btn-secondary" onClick={() => setEditingId(null)}>
                      {t('common.cancel')}
                    </button>
                  </>
                ) : (
                  <>
                    {tab !== 'species' && item.context && (
                      <span className="text-dark-text-secondary" style={{ minWidth: '6rem' }}>{item.context}</span>
                    )}
                    <span style={{ flex: 1 }}>{item.value}</span>
                    {tab === 'vaccination' && item.duration_months != null && (
                      <span className="text-dark-text-secondary">{item.duration_months} mo</span>
                    )}
                    <button type="button" className="btn btn-sm btn-secondary" onClick={() => startEdit(item)} aria-label={t('common.edit')}>
                      <Icon icon="mdi:pencil" width={16} height={16} />
                    </button>
                    <button type="button" className="btn btn-sm btn-danger" onClick={() => handleDelete(item.id)} aria-label={t('common.delete')}>
                      <Icon icon="mdi:delete-outline" width={16} height={16} />
                    </button>
                  </>
                )}
              </li>
            ))}
        </ul>
        {filtered.length === 0 && <p className="muted">{t('admin.noItems')}</p>}
        {filtered.length > 0 && displayList.length === 0 && tab !== 'species' && (
          <p className="muted">{t('admin.noItemsForSpecies')}</p>
        )}
      </section>
    </div>
  )
}

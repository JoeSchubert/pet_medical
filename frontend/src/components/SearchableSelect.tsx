import { useId } from 'react'

type Props = {
  options: string[]
  value: string
  onChange: (value: string) => void
  placeholder?: string
  label?: string
}

/**
 * Searchable select: type to filter options (datalist); can pick from list or enter custom value.
 */
export default function SearchableSelect({
  options,
  value,
  onChange,
  placeholder = 'Select or type...',
  label,
}: Props) {
  const id = useId()
  const filter = (value || '').toLowerCase()
  const suggested = filter
    ? options.filter((o) => o.toLowerCase().includes(filter))
    : options

  return (
    <label className="searchable-select-label">
      {label}
      <span className="searchable-select-wrap">
        <input
          list={id}
          type="text"
          placeholder={placeholder}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="input searchable-select-input"
        />
        <span className="searchable-select-arrow" aria-hidden>▼</span>
      </span>
      <datalist id={id}>
        {suggested.map((opt) => (
          <option key={opt} value={opt} />
        ))}
      </datalist>
    </label>
  )
}

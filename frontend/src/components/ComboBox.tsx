import { useId, useRef, useState, useEffect } from 'react'

type Props = {
  options: string[]
  value: string
  onChange: (value: string) => void
  placeholder?: string
  label?: string
  required?: boolean
}

/**
 * Combobox with dark-themed dropdown list (replaces native datalist for consistent styling).
 */
export default function ComboBox({
  options,
  value,
  onChange,
  placeholder = 'Select or type...',
  label,
  required,
}: Props) {
  const id = useId()
  const [open, setOpen] = useState(false)
  const [filter, setFilter] = useState(value)
  const containerRef = useRef<HTMLSpanElement>(null)

  const filtered =
    filter.trim() === ''
      ? options
      : options.filter((o) => o.toLowerCase().includes(filter.toLowerCase()))

  useEffect(() => {
    setFilter(value)
  }, [value])

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    if (open) {
      document.addEventListener('mousedown', handleClickOutside)
      return () => document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [open])

  const displayValue = open ? filter : value
  const showList = open && (filtered.length > 0 || filter.trim() !== '')

  return (
    <label className="searchable-select-label" htmlFor={id}>
      {label}
      <span className="searchable-select-wrap combobox-wrap" ref={containerRef}>
        <input
          id={id}
          type="text"
          placeholder={placeholder}
          value={displayValue}
          required={required}
          onChange={(e) => {
            setFilter(e.target.value)
            onChange(e.target.value)
            setOpen(true)
          }}
          onFocus={() => setOpen(true)}
          className="input searchable-select-input"
          autoComplete="off"
          role="combobox"
          aria-expanded={open}
          aria-haspopup="listbox"
          aria-autocomplete="list"
        />
        <span
          className="searchable-select-arrow combobox-arrow"
          aria-hidden
          onClick={() => setOpen(!open)}
          onKeyDown={(e) => e.key === 'Enter' && setOpen(!open)}
          role="button"
          tabIndex={-1}
        >
          ▼
        </span>
        {showList && (
          <ul className="combobox-list" role="listbox" aria-label={label ?? 'Options'}>
            {filtered.map((opt) => (
              <li
                key={opt}
                className="combobox-option"
                role="option"
                aria-selected={opt === value}
                onMouseDown={(e) => {
                  e.preventDefault()
                  onChange(opt)
                  setFilter(opt)
                  setOpen(false)
                }}
              >
                {opt}
              </li>
            ))}
            {filter.trim() !== '' && !filtered.includes(filter) && (
              <li
                className="combobox-option combobox-option-custom"
                role="option"
                onMouseDown={(e) => {
                  e.preventDefault()
                  onChange(filter)
                  setOpen(false)
                }}
              >
                Use &quot;{filter}&quot;
              </li>
            )}
          </ul>
        )}
      </span>
    </label>
  )
}

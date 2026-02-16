/**
 * Common vaccinations by species with typical duration (months) until next due.
 * Duration is used to auto-compute expiration from administered date.
 */
export interface VaccinationPreset {
  name: string
  durationMonths: number
}

export const VACCINATIONS_BY_SPECIES: Record<string, VaccinationPreset[]> = {
  Dog: [
    { name: 'Rabies', durationMonths: 12 },
    { name: 'Rabies (3-year)', durationMonths: 36 },
    { name: 'DHPP (Distemper, Hepatitis, Parvovirus, Parainfluenza)', durationMonths: 12 },
    { name: 'Bordetella (Kennel Cough)', durationMonths: 12 },
    { name: 'Leptospirosis', durationMonths: 12 },
    { name: 'Canine Influenza', durationMonths: 12 },
    { name: 'Lyme Disease', durationMonths: 12 },
    { name: 'Coronavirus', durationMonths: 12 },
  ],
  Cat: [
    { name: 'Rabies', durationMonths: 12 },
    { name: 'Rabies (3-year)', durationMonths: 36 },
    { name: 'FVRCP (Feline Distemper)', durationMonths: 12 },
    { name: 'FeLV (Feline Leukemia)', durationMonths: 12 },
    { name: 'Feline Herpesvirus', durationMonths: 12 },
    { name: 'Calicivirus', durationMonths: 12 },
  ],
  Bird: [
    { name: 'Polyomavirus', durationMonths: 12 },
    { name: 'Psittacosis', durationMonths: 12 },
    { name: 'Pacheco\'s Disease', durationMonths: 12 },
    { name: 'Avian Influenza (as recommended)', durationMonths: 12 },
  ],
  Rabbit: [
    { name: 'RHDV1 / RHDV2 (Rabbit Hemorrhagic Disease)', durationMonths: 12 },
    { name: 'Myxomatosis', durationMonths: 12 },
  ],
  Horse: [
    { name: 'Eastern/Western Encephalomyelitis', durationMonths: 12 },
    { name: 'Tetanus', durationMonths: 12 },
    { name: 'West Nile Virus', durationMonths: 12 },
    { name: 'Rabies', durationMonths: 12 },
    { name: 'Rhino (EHV-1/EHV-4)', durationMonths: 6 },
    { name: 'Influenza', durationMonths: 6 },
    { name: 'Strangles (optional)', durationMonths: 12 },
  ],
  Hamster: [],
  'Guinea Pig': [],
  Fish: [],
  Reptile: [],
  Other: [
    { name: 'Rabies (if applicable)', durationMonths: 12 },
    { name: 'Core vaccine (species-specific)', durationMonths: 12 },
  ],
}

export function getVaccinationPresetsForSpecies(species: string | null | undefined): VaccinationPreset[] {
  if (!species) return []
  return VACCINATIONS_BY_SPECIES[species] ?? VACCINATIONS_BY_SPECIES.Other ?? []
}

export function getVaccinationNamesForSpecies(species: string | null | undefined): string[] {
  return getVaccinationPresetsForSpecies(species).map((v) => v.name).sort((a, b) => a.localeCompare(b))
}

export function getDurationMonths(species: string | null | undefined, vaccineName: string): number | null {
  const presets = getVaccinationPresetsForSpecies(species)
  const found = presets.find((v) => v.name === vaccineName)
  return found ? found.durationMonths : null
}

/** Add months to a YYYY-MM-DD date string; returns YYYY-MM-DD */
export function addMonthsToDate(dateStr: string, months: number): string {
  const d = new Date(dateStr.length === 10 ? `${dateStr}T12:00:00` : dateStr)
  d.setMonth(d.getMonth() + months)
  return d.toISOString().slice(0, 10)
}

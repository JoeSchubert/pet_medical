/** Species emoji for avatar fallback (RobiPet-style) */
export function getSpeciesEmoji(species?: string | null): string {
  if (!species) return '🐾'
  const s = species.toLowerCase()
  if (s.includes('dog')) return '🐕'
  if (s.includes('cat')) return '🐱'
  if (s.includes('bird')) return '🦅'
  if (s.includes('rabbit')) return '🐰'
  if (s.includes('hamster')) return '🐹'
  if (s.includes('fish')) return '🐠'
  if (s.includes('guinea')) return '🐹'
  if (s.includes('reptile')) return '🦎'
  if (s.includes('horse')) return '🐴'
  return '🐾'
}

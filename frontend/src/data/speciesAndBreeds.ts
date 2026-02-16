/**
 * Common species and breeds (RobiPet-style). Users can also add custom values.
 * Breeds are keyed by species; "Other" allows any custom breed.
 * All lists are sorted alphabetically.
 */
const SPECIES_RAW = [
  'Bird',
  'Cat',
  'Dog',
  'Fish',
  'Guinea Pig',
  'Hamster',
  'Horse',
  'Other',
  'Rabbit',
  'Reptile',
]

export const COMMON_SPECIES: string[] = [...SPECIES_RAW].sort((a, b) => a.localeCompare(b))

const sort = (arr: string[]) => [...arr].sort((a, b) => a.localeCompare(b))

export const BREEDS_BY_SPECIES: Record<string, string[]> = {
  Dog: sort([
    'Australian Shepherd',
    'Beagle',
    'Boston Terrier',
    'Boxer',
    'Bulldog',
    'Cavalier King Charles Spaniel',
    'Chihuahua',
    'Dachshund',
    'Doberman Pinscher',
    'French Bulldog',
    'German Shepherd',
    'Golden Retriever',
    'Havanese',
    'Labrador Retriever',
    'Miniature Schnauzer',
    'Mixed breed',
    'Other',
    'Pembroke Welsh Corgi',
    'Poodle',
    'Rottweiler',
    'Shih Tzu',
    'Yorkshire Terrier',
  ]),
  Cat: sort([
    'Abyssinian',
    'Bengal',
    'Birman',
    'British Shorthair',
    'Domestic Longhair',
    'Domestic Shorthair',
    'Maine Coon',
    'Mixed breed',
    'Oriental',
    'Other',
    'Persian',
    'Ragdoll',
    'Russian Blue',
    'Scottish Fold',
    'Siamese',
    'Sphynx',
  ]),
  Bird: sort([
    'African Grey',
    'Budgerigar',
    'Canary',
    'Cockatiel',
    'Cockatoo',
    'Conure',
    'Finch',
    'Lovebird',
    'Macaw',
    'Other',
    'Parakeet',
  ]),
  Rabbit: sort([
    'Angora',
    'Dwarf',
    'Dutch',
    'Flemish Giant',
    'Lionhead',
    'Lop',
    'Mixed breed',
    'Other',
    'Rex',
  ]),
  Hamster: sort(['Chinese', 'Dwarf (Campbell)', 'Dwarf (Winter White)', 'Other', 'Roborovski', 'Syrian']),
  'Guinea Pig': sort(['Abyssinian', 'American', 'Mixed breed', 'Other', 'Peruvian', 'Silkie', 'Teddy']),
  Fish: sort(['Betta', 'Cichlid', 'Goldfish', 'Guppy', 'Koi', 'Other', 'Tetra']),
  Reptile: sort([
    'Ball Python',
    'Bearded Dragon',
    'Corn Snake',
    'Crested Gecko',
    'Leopard Gecko',
    'Other',
    'Tortoise',
  ]),
  Horse: sort([
    'Appaloosa',
    'Arabian',
    'Other',
    'Paint',
    'Pony',
    'Quarter Horse',
    'Thoroughbred',
    'Warmblood',
  ]),
  Other: ['Other'],
}

export function getBreedsForSpecies(species: string): string[] {
  if (!species) return []
  const breeds = BREEDS_BY_SPECIES[species]
  return breeds ? [...breeds] : ['Other']
}

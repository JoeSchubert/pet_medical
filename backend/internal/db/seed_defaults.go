package db

import (
	"gorm.io/gorm"
	"github.com/pet-medical/api/internal/models"
)

// SeedDefaultDropdownOptions inserts built-in species, breeds, and vaccinations if the table is empty.
func SeedDefaultDropdownOptions(gdb *gorm.DB) error {
	var count int64
	if err := gdb.Model(&models.DefaultDropdownOption{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	species := []string{"Bird", "Cat", "Dog", "Fish", "Guinea Pig", "Hamster", "Horse", "Other", "Rabbit", "Reptile"}
	for i, s := range species {
		if err := gdb.Create(&models.DefaultDropdownOption{OptionType: "species", Value: s, Context: "", SortOrder: i}).Error; err != nil {
			return err
		}
	}

	breedsBySpecies := map[string][]string{
		"Dog": {"Australian Shepherd", "Beagle", "Boston Terrier", "Boxer", "Bulldog", "Cavalier King Charles Spaniel", "Chihuahua", "Dachshund", "Doberman Pinscher", "French Bulldog", "German Shepherd", "Golden Retriever", "Havanese", "Labrador Retriever", "Miniature Schnauzer", "Mixed breed", "Other", "Pembroke Welsh Corgi", "Poodle", "Rottweiler", "Shih Tzu", "Yorkshire Terrier"},
		"Cat": {"Abyssinian", "Bengal", "Birman", "British Shorthair", "Domestic Longhair", "Domestic Shorthair", "Maine Coon", "Mixed breed", "Oriental", "Other", "Persian", "Ragdoll", "Russian Blue", "Scottish Fold", "Siamese", "Sphynx"},
		"Bird": {"African Grey", "Budgerigar", "Canary", "Cockatiel", "Cockatoo", "Conure", "Finch", "Lovebird", "Macaw", "Other", "Parakeet"},
		"Rabbit": {"Angora", "Dwarf", "Dutch", "Flemish Giant", "Lionhead", "Lop", "Mixed breed", "Other", "Rex"},
		"Hamster": {"Chinese", "Dwarf (Campbell)", "Dwarf (Winter White)", "Other", "Roborovski", "Syrian"},
		"Guinea Pig": {"Abyssinian", "American", "Mixed breed", "Other", "Peruvian", "Silkie", "Teddy"},
		"Fish": {"Betta", "Cichlid", "Goldfish", "Guppy", "Koi", "Other", "Tetra"},
		"Reptile": {"Ball Python", "Bearded Dragon", "Corn Snake", "Crested Gecko", "Leopard Gecko", "Other", "Tortoise"},
		"Horse": {"Appaloosa", "Arabian", "Other", "Paint", "Pony", "Quarter Horse", "Thoroughbred", "Warmblood"},
		"Other": {"Other"},
	}
	for speciesName, breeds := range breedsBySpecies {
		for i, b := range breeds {
			if err := gdb.Create(&models.DefaultDropdownOption{OptionType: "breed", Value: b, Context: speciesName, SortOrder: i}).Error; err != nil {
				return err
			}
		}
	}

	type vacc struct{ name string; months int }
	vaccsBySpecies := map[string][]vacc{
		"Dog":   {{"Rabies", 12}, {"Rabies (3-year)", 36}, {"DHPP (Distemper, Hepatitis, Parvovirus, Parainfluenza)", 12}, {"Bordetella (Kennel Cough)", 12}, {"Leptospirosis", 12}, {"Canine Influenza", 12}, {"Lyme Disease", 12}, {"Coronavirus", 12}},
		"Cat":   {{"Rabies", 12}, {"Rabies (3-year)", 36}, {"FVRCP (Feline Distemper)", 12}, {"FeLV (Feline Leukemia)", 12}, {"Feline Herpesvirus", 12}, {"Calicivirus", 12}},
		"Bird":  {{"Polyomavirus", 12}, {"Psittacosis", 12}, {"Pacheco's Disease", 12}, {"Avian Influenza (as recommended)", 12}},
		"Rabbit": {{"RHDV1 / RHDV2 (Rabbit Hemorrhagic Disease)", 12}, {"Myxomatosis", 12}},
		"Horse": {{"Eastern/Western Encephalomyelitis", 12}, {"Tetanus", 12}, {"West Nile Virus", 12}, {"Rabies", 12}, {"Rhino (EHV-1/EHV-4)", 6}, {"Influenza", 6}, {"Strangles (optional)", 12}},
		"Other": {{"Rabies (if applicable)", 12}, {"Core vaccine (species-specific)", 12}},
	}
	for speciesName, vaccs := range vaccsBySpecies {
		for i, v := range vaccs {
			dur := v.months
			if err := gdb.Create(&models.DefaultDropdownOption{OptionType: "vaccination", Value: v.name, Context: speciesName, SortOrder: i, DurationMonths: &dur}).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

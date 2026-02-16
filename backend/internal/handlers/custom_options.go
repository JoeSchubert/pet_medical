package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/pet-medical/api/internal/middleware"
	"github.com/pet-medical/api/internal/models"
	"gorm.io/gorm"
)

type CustomOptionsHandler struct {
	GORM *gorm.DB
}

// GetResponse is the JSON shape for GET /custom-options
type CustomOptionsGetResponse struct {
	Species              []string                  `json:"species"`
	Breeds               map[string][]string       `json:"breeds"`
	Vaccinations         map[string][]string       `json:"vaccinations"`
	VaccinationDurations map[string]map[string]int  `json:"vaccination_durations"`
}

// AddRequest is the JSON body for POST /custom-options
type CustomOptionsAddRequest struct {
	OptionType string `json:"option_type"`
	Value      string `json:"value"`
	Context    string `json:"context"`
}

func (h *CustomOptionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUser(r.Context())
	if u == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	out := CustomOptionsGetResponse{
		Species:              []string{},
		Breeds:               map[string][]string{},
		Vaccinations:         map[string][]string{},
		VaccinationDurations: map[string]map[string]int{},
	}

	var defaults []models.DefaultDropdownOption
	if err := h.GORM.Order("option_type, context, sort_order, value").Find(&defaults).Error; err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	for _, d := range defaults {
		switch d.OptionType {
		case "species":
			out.Species = append(out.Species, d.Value)
		case "breed":
			if out.Breeds[d.Context] == nil {
				out.Breeds[d.Context] = []string{}
			}
			out.Breeds[d.Context] = append(out.Breeds[d.Context], d.Value)
		case "vaccination":
			if out.Vaccinations[d.Context] == nil {
				out.Vaccinations[d.Context] = []string{}
			}
			out.Vaccinations[d.Context] = append(out.Vaccinations[d.Context], d.Value)
			if d.DurationMonths != nil {
				if out.VaccinationDurations[d.Context] == nil {
					out.VaccinationDurations[d.Context] = map[string]int{}
				}
				out.VaccinationDurations[d.Context][d.Value] = *d.DurationMonths
			}
		}
	}

	var customs []models.UserCustomOption
	if err := h.GORM.Where("user_id = ?", u.ID).Order("option_type, context, value").Find(&customs).Error; err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	for _, row := range customs {
		ctx := row.Context
		switch row.OptionType {
		case "species":
			out.Species = appendUnique(out.Species, row.Value)
		case "breed":
			if out.Breeds[ctx] == nil {
				out.Breeds[ctx] = []string{}
			}
			out.Breeds[ctx] = appendUnique(out.Breeds[ctx], row.Value)
		case "vaccination":
			if out.Vaccinations[ctx] == nil {
				out.Vaccinations[ctx] = []string{}
			}
			out.Vaccinations[ctx] = appendUnique(out.Vaccinations[ctx], row.Value)
		}
	}

	sort.Strings(out.Species)
	for k := range out.Breeds {
		sort.Strings(out.Breeds[k])
	}
	for k := range out.Vaccinations {
		sort.Strings(out.Vaccinations[k])
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func appendUnique(slice []string, v string) []string {
	for _, x := range slice {
		if x == v {
			return slice
		}
	}
	return append(slice, v)
}

func (h *CustomOptionsHandler) Add(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	u := middleware.GetUser(r.Context())
	if u == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	var body CustomOptionsAddRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	body.OptionType = strings.TrimSpace(strings.ToLower(body.OptionType))
	body.Value = strings.TrimSpace(body.Value)
	body.Context = strings.TrimSpace(body.Context)
	if body.Value == "" {
		http.Error(w, `{"error":"value required"}`, http.StatusBadRequest)
		return
	}
	if body.OptionType != "species" && body.OptionType != "breed" && body.OptionType != "vaccination" {
		http.Error(w, `{"error":"invalid option_type"}`, http.StatusBadRequest)
		return
	}
	ctxVal := body.Context
	opt := models.UserCustomOption{UserID: u.ID, OptionType: body.OptionType, Value: body.Value, Context: ctxVal}
	h.GORM.Where("user_id = ? AND option_type = ? AND value = ? AND context = ?", u.ID, body.OptionType, body.Value, ctxVal).FirstOrCreate(&opt)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

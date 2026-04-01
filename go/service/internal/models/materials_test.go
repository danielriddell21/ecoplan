package models_test

import (
	"testing"

	"github.com/ecoscan/service/internal/models"
)

func makeDB() models.MaterialsDB {
	return models.MaterialsDB{
		Materials: map[string]models.Material{
			"plastic_bottle": {Label: "Plastic Bottle", Recyclable: "yes", Bin: "green", Tips: "Rinse first"},
			"glass_bottle":   {Label: "Glass Bottle", Recyclable: "yes", Bin: "blue", Tips: "Rinse clean"},
			"plastic_bag":    {Label: "Plastic Bags & Film", Recyclable: "no", Bin: "black", Tips: "Not kerbside"},
		},
		Bins: map[string]models.Bin{
			"green": {Label: "Green Bin"},
			"blue":  {Label: "Blue Bin"},
			"black": {Label: "Black Bin"},
		},
	}
}

func TestNormaliseTag(t *testing.T) {
	db := makeDB()
	cases := []struct {
		input string
		want  string
	}{
		{"en:plastic-bottle", "plastic-bottle"},
		{"fr:verre", "verre"},
		{"de:dose", "dose"},
		{"es:lata", "lata"},
		{"PLASTIC-bottle", "plastic-bottle"},
		{"  cardboard  ", "cardboard"},
	}
	for _, tc := range cases {
		got := db.NormaliseTag(tc.input)
		if got != tc.want {
			t.Errorf("NormaliseTag(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestLookupMaterials_knownTag(t *testing.T) {
	db := makeDB()
	results := db.LookupMaterials([]string{"en:plastic-bottle"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Label != "Plastic Bottle" {
		t.Errorf("unexpected label %q", results[0].Label)
	}
}

func TestLookupMaterials_unknownTag(t *testing.T) {
	db := makeDB()
	results := db.LookupMaterials([]string{"en:unobtanium"})
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestLookupMaterials_deduplicates(t *testing.T) {
	db := makeDB()
	// Both tags map to plastic_bottle — should only get one result.
	results := db.LookupMaterials([]string{"en:plastic-bottle", "en:pet-bottle"})
	if len(results) != 1 {
		t.Errorf("expected 1 deduplicated result, got %d", len(results))
	}
}

func TestSearchMaterials_caseInsensitive(t *testing.T) {
	db := makeDB()
	results := db.SearchMaterials("PLASTIC")
	if len(results) < 2 {
		t.Errorf("expected at least 2 results for 'PLASTIC', got %d", len(results))
	}
}

func TestSearchMaterials_noMatch(t *testing.T) {
	db := makeDB()
	results := db.SearchMaterials("unobtanium")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

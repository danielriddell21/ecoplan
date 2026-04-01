package models

import "strings"

type Material struct {
	Label      string   `json:"label"`
	Examples   []string `json:"examples"`
	Recyclable string   `json:"recyclable"`
	Bin        string   `json:"bin"`
	Tips       string   `json:"tips"`
}

type Bin struct {
	Label       string `json:"label"`
	Description string `json:"description"`
	Collection  string `json:"collection,omitempty"`
	URL         string `json:"url,omitempty"`
}

type MaterialsDB struct {
	Materials map[string]Material `json:"materials"`
	Bins      map[string]Bin      `json:"bins"`
}

// LookupMaterials maps a slice of OpenFoodFacts packaging tags to known materials.
func (db MaterialsDB) LookupMaterials(offTags []string) []Material {
	seen := map[string]bool{}
	var results []Material
	for _, raw := range offTags {
		key, ok := OffTagToMaterial[db.NormaliseTag(raw)]
		if !ok || seen[key] {
			continue
		}
		if m, found := db.Materials[key]; found {
			results = append(results, m)
			seen[key] = true
		}
	}
	return results
}

// SearchMaterials returns all materials whose label contains the query string (case-insensitive).
func (db MaterialsDB) SearchMaterials(query string) []Material {
	results := []Material{}
	query = strings.ToLower(strings.TrimSpace(query))
	for _, material := range db.Materials {
		if strings.Contains(strings.ToLower(material.Label), query) {
			results = append(results, material)
		}
	}
	return results
}

// NormaliseTag strips language prefixes (en:, fr:, etc.) and lowercases the tag.
func (db MaterialsDB) NormaliseTag(raw string) string {
	tag := raw
	for _, prefix := range []string{"en:", "fr:", "de:", "es:"} {
		tag = strings.TrimPrefix(tag, prefix)
	}
	return strings.ToLower(strings.TrimSpace(tag))
}

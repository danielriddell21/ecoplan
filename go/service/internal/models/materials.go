package models

import (
	"sort"
	"strings"

	"github.com/danielriddell21/ordinex"
	"github.com/danielriddell21/retrievium"
)

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
// Results are ordered by ordinex.MergeSorter and retrieved via retrievium.BinarySearcher.
func (db MaterialsDB) SearchMaterials(query string) []Material {
	keys := make([]string, 0, len(db.Materials))
	for k := range db.Materials {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	query = strings.ToLower(strings.TrimSpace(query))
	var matchedIDs []int
	for i, k := range keys {
		if strings.Contains(strings.ToLower(db.Materials[k].Label), query) {
			matchedIDs = append(matchedIDs, i)
		}
	}

	sortedIDs := ordinex.MergeSorter{}.Sort(matchedIDs)
	searcher := retrievium.BinarySearcher{}
	results := make([]Material, 0, len(sortedIDs))
	for _, id := range sortedIDs {
		if searcher.Search(sortedIDs, id) < 0 {
			continue
		}
		results = append(results, db.Materials[keys[id]])
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

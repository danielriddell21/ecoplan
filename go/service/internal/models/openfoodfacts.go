package models

type OFFResponse struct {
	Status  int        `json:"status"`
	Product OFFProduct `json:"product"`
}

type OFFProduct struct {
	ProductName            string   `json:"product_name"`
	Brands                 string   `json:"brands"`
	PackagingTags          []string `json:"packaging_tags"`
	PackagingMaterialsTags []string `json:"packaging_materials_tags"`
	PackagingShapesTags    []string `json:"packaging_shapes_tags"`
	PackagingText          string   `json:"packaging_text"`
}

// OffTagToMaterial maps OpenFoodFacts packaging tags to internal material keys.
var OffTagToMaterial = map[string]string{
	// Plastic bottles
	"plastic-bottle": "plastic_bottle",
	"pet-bottle":     "plastic_bottle",
	"hdpe-bottle":    "plastic_bottle",
	"bottle":         "plastic_bottle",
	"plastic bottle": "plastic_bottle",
	"water bottle":   "plastic_bottle",
	// Plastic tubs/trays
	"plastic-tray": "plastic_tub",
	"plastic-pot":  "plastic_tub",
	"plastic-tub":  "plastic_tub",
	"tray":         "plastic_tub",
	"pot":          "plastic_tub",
	// Black plastic
	"black-plastic": "plastic_black",
	// Bags & film
	"plastic-bag":      "plastic_bag",
	"film":             "plastic_bag",
	"flexible-plastic": "plastic_bag",
	"bag":              "plastic_bag",
	// Polystyrene
	"polystyrene":          "polystyrene",
	"expanded-polystyrene": "polystyrene",
	"foam":                 "polystyrene",
	// Glass
	"glass-bottle": "glass_bottle",
	"glass-jar":    "glass_bottle",
	"jar":          "glass_bottle",
	"glass bottle": "glass_bottle",
	"wine bottle":  "glass_bottle",
	// Cans
	"tin":           "tin_can",
	"steel-can":     "tin_can",
	"aluminium-can": "tin_can",
	"can":           "tin_can",
	"metal-can":     "tin_can",
	"aluminum can":  "tin_can",
	"tin can":       "tin_can",
	"steel can":     "tin_can",
	// Foil
	"aluminium-foil": "foil",
	"foil-tray":      "foil",
	"foil":           "foil",
	// Paper & cardboard
	"cardboard":     "cardboard",
	"cardboard-box": "cardboard",
	"paper-box":     "cardboard",
	"paper":         "paper",
	"paper-bag":     "paper",
	// Cartons
	"tetra-pak":       "tetra_pak",
	"carton":          "tetra_pak",
	"beverage-carton": "tetra_pak",
	// Aerosol
	"aerosol": "aerosol",
	// Crisp packets / wrappers
	"crisp-packet":  "crisp_packet",
	"snack-wrapper": "crisp_packet",
	"wrapper":       "crisp_packet",
}

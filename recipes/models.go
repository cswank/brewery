package recipes

import (
	"encoding/xml"
)

type Fermentable struct {
	Name string  `xml:"NAME"`
        Origin string  `xml:"ORIGIN"`
        Type string `xml:"TYPE"`
        Yield float64 `xml:"YIELD"`
        Amount float64 `xml:"AMOUNT"`
        DisplayAmount string `xml:"DISPLAY_AMOUNT"`
        Potential float64 `xml:"POTENTIAL"`
        Color int `xml:"COLOR"`
        DisplayColor string `xml:"DISPLAY_COLOR"`
        AddAfterBoil string `xml:"ADD_AFTER_BOIL"`
        CoarseFineDiff string `xml:"COARSE_FINE_DIFF"`
        Moisture string `xml:"MOISTURE"`
        DiastaticPower string `xml:"DIASTATIC_POWER"`
        Protein  string `xml:"PROTEIN"`
        MaxInBatch string `xml:"MAX_IN_BATCH"`
        RecommendMash string `xml:"RECOMMEND_MASH"`
        IBUGalPerLb string `xml:"IBU_GAL_PER_LB"`
        Notes string `xml:"NOTES"`
}

type Hop struct {
        Name string `xml:"NAME"`
        Origin string `xml:"ORIGIN"`
        Alpha string `xml:"ALPHA"`
        Beta string `xml:"BETA"`
        Amount float64 `xml:"AMOUNT"`
        DisplayAmount string `xml:"DISPLAY_AMOUNT"`
        Use string `xml:"USE"`
        Form string `xml:"FORM"`
        Time float64 `xml:"TIME"`
        DisplayTime string `xml:"DISPLAY_TIME"`
        Notes string `xml:"NOTES"`
}

type Yeast struct {
	Laboratory string `xml:"LABORATORY"`
        Name string `xml:"NAME"`
	Type string `xml:"TYPE"`
        Form string `xml:"FORM"`
        Attenuation float64 `xml:"ATTENUATION"`
}

type Fermentables struct {
	XMLName xml.Name `xml:"FERMENTABLES"`
	Fermentables []Fermentable `xml:"FERMENTABLE"`
}

type Hops struct {
	XMLName xml.Name `xml:"HOPS"`
	Hops []Hop `xml:"HOP"`
}

type Yeasts struct {
	XMLName xml.Name `xml:"YEASTS"`
	Yeasts []Yeast `xml:"YEAST"`
}

type Recipe struct {
	Name    string `xml:"NAME"`
	Type    string `xml:"TYPE"`
	Brewer  string `xml:"BREWER"`
	BatchSize float64  `xml:"BATCH_SIZE"`
	BoilSize  float64 `xml:"BOIL_SIZE"`
	BoilTime  float64 `xml:"BOIL_TIME"`
	Efficiency float64 `xml:"EFFICIENCY"`
	Fermentables  []Fermentable `xml:"FERMENTABLES>FERMENTABLE"`
	Hops []Hop `xml:"HOPS>HOP"`
	Yeasts []Yeast `xml:"YEASTS>YEAST"`
	Miscs string  `xml:"MISCS"`
}

type Recipes struct {
	XMLName xml.Name `xml:"RECIPES"`
	Recipes []Recipe `xml:"RECIPE"`
}

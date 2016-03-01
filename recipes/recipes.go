package recipes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Fermentable struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Color  int     `json:"color"`
	Unit   string  `json:"unit"`
}

type Hop struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Alpha  float64 `xml:"alpha"`
	Beta   float64 `xml:"beta"`
	Time   float64 `xml:"time"`
}

type Yeast struct {
	Name        string  `json:"name"`
	Attenuation float64 `json:"attenuation"`
}

type MashStep struct {
	Temperature float64 `json:"target_temperature"`
	Metric      bool    `json:"target_temperature_is_metric"`
	Time        float64 `json:"time"`
}

type Recipe struct {
	Name         string        `json:"name"`
	BatchSize    float64       `json:"batch_size"`
	BoilSize     float64       `json:"boil_size"`
	BoilTime     float64       `json:"boil_time"`
	Efficiency   float64       `json:"efficiency"`
	Fermentables []Fermentable `json:"recipe_fermentables"`
	WaterRatio   float64
	Hops         []Hop      `json:"recipe_hops"`
	Yeasts       []Yeast    `json:"recipe_yeasts"`
	MashSteps    []MashStep `json:"recipe_mash_steps"`
}

type Mash struct {
	Volume             float64
	StrikeTemperature  float64
	Time               float64
	SpargeVolume       float64
	SecondSpargeVolume float64
}

func NewRecipe(name string) (r *Recipe, err error) {
	recipeUrl := fmt.Sprintf("http://www.brewtoad.com/recipes/%s.json", name)
	res, err := http.Get(recipeUrl)
	if err != nil {
		return r, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return r, err
	}
	r = &Recipe{}
	err = json.Unmarshal(body, r)
	if err != nil {
		return r, err
	}
	r.WaterRatio = 1.25
	return r, err
}

func (r *Recipe) getMash(grainTemperature float64) *Mash {
	grainWeight := r.getGrainWeight()
	mashVolume := r.getMashVolume(grainWeight)
	return &Mash{
		StrikeTemperature:  r.getStrikeTemperature(grainTemperature),
		Volume:             mashVolume,
		SpargeVolume:       r.getSpargeVolume(mashVolume, grainWeight),
		SecondSpargeVolume: r.BoilSize / 2.0,
		Time:               r.getMashTime(),
	}
}

func (r *Recipe) getGrainWeight() float64 {
	weight := 0.0
	for _, f := range r.Fermentables {
		weight += f.Amount
	}
	return weight
}

func (r *Recipe) getTargetTemperature() (t float64) {
	if len(r.MashSteps) > 0 {
		t = r.MashSteps[0].Temperature
	} else {
		t = 154.0
	}
	return t
}

func (r *Recipe) getMashTime() (t float64) {
	if len(r.MashSteps) > 0 {
		t = r.MashSteps[0].Time
	} else {
		t = 45.0
	}
	return t
}

func (r *Recipe) getStrikeTemperature(grainTemperature float64) float64 {
	targetTemperature := r.getTargetTemperature()
	return (0.2*r.WaterRatio)*(targetTemperature-grainTemperature) + targetTemperature
}

func (r *Recipe) getInfusionVolume(initialTemperature, targetTemperature, volume, grainWeight, waterTemperature, mashTemperature float64) float64 {
	return (targetTemperature - initialTemperature) * (0.2*grainWeight + volume) / (waterTemperature - targetTemperature)
}

func (r *Recipe) getMashVolume(grainWeight float64) float64 {
	return (r.WaterRatio * grainWeight) / 4.0
}

func (r *Recipe) getSpargeVolume(mashVolume, grainWeight float64) float64 {
	absorbtion := grainWeight * 0.1
	targetVolume := r.BoilSize / 2.0
	drainVolume := mashVolume - absorbtion
	volume := targetVolume - drainVolume
	if volume < 0 {
		volume = 0
	}
	return volume + mashVolume
}

func (r *Recipe) GetMethod(grainTemperature float64) []string {
	mash := r.getMash(grainTemperature)
	temperatureUnits := "F"
	volumeUnits := "gallons"
	return []string{
		fmt.Sprintf("fill hlt to 7.0 %s", volumeUnits),
		fmt.Sprintf("heat hlt to %f %s", mash.StrikeTemperature, temperatureUnits),
		fmt.Sprintf("wait for hlt temperature >= %f %s", mash.StrikeTemperature, temperatureUnits),
		fmt.Sprintf("fill tun to %f %s", mash.Volume, volumeUnits),
		fmt.Sprintf("wait for tun volume >= %f %s", mash.Volume, volumeUnits),
		"wait for user to add grains",
		fmt.Sprintf("fill hlt to 7.0 %s", volumeUnits),
		fmt.Sprintf("heat hlt to 185 %s", temperatureUnits),
		fmt.Sprintf("wait for %f minutes", mash.Time),
		fmt.Sprintf("fill tun to %f %s", mash.SpargeVolume, volumeUnits),
		"wait for 10 minutes",
		"wait for user ready to recirculate",
		"fill boiler",
		"wait for user to finish recirculating",
		fmt.Sprintf("fill boiler to %f %s", mash.SpargeVolume, volumeUnits),
		fmt.Sprintf("heat boiler to 190 %s", temperatureUnits),
		fmt.Sprintf("wait for boiler volume >= %f %s", mash.SpargeVolume, volumeUnits),
		fmt.Sprintf("fill tun to %f %s", mash.SecondSpargeVolume, volumeUnits),
		fmt.Sprintf("wait for tun volume >= %f %s", mash.SecondSpargeVolume, volumeUnits),
		"stop heating hlt",
		"wait for 2 minutes",
		"wait for user ready to recirculate",
		"fill boiler",
		"wait for user to finish recirculating",
		fmt.Sprintf("fill boiler to %f %s", mash.SecondSpargeVolume+mash.SpargeVolume, volumeUnits),
		fmt.Sprintf("heat boiler to 204 %s", temperatureUnits),
		"turn on fan",
		fmt.Sprintf("wait for %f minutes", r.BoilTime),
		"stop heating boiler",
		"turn off fan",
		"wait for 5 minutes",
		"cool boiler to 80 F",
		"wait for boiler temperature <= 80 F",
		"wait for user to open ball valve",
		"fill carboy",
		"wait for user to confirm boiler empty",
		"stop filling carboy",
	}
}

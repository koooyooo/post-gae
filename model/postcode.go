package model

type Postcode struct {
	OrgCode        string `json:"org-code"`
	PostcodeOld    string `json:"postcode-old"`
	Postcode       string `json:"postcode"`
	PrefectureRuby string `json:"prefecture-ruby"`
	CityRuby       string `json:"city-ruby"`
	AreaRuby       string `json:"area-ruby"`
	Prefecture     string `json:"prefecture"`
	City           string `json:"city"`
	Area           string `json:"area"`
	Flag1          string `json:"flag1"`
	Flag2          string `json:"flag2"`
	Flag3          string `json:"flag3"`
	Flag4          string `json:"flag4"`
	Flag5          string `json:"flag5"`
	Flag6          string `json:"flag6"`
}

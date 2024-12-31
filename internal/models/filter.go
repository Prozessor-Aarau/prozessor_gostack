package models

type VisitorFilter struct {
	SearchString string
	Abo          bool
	AboValidity  string
	Member       bool
	Badge        bool
	Partner      bool
	Sort         string
	SortDir      string
}

func NewVisitorSearch(search string) VisitorFilter {
	return VisitorFilter{
		SearchString: search,
		Abo:          false,
		AboValidity:  "all",
		Member:       false,
		Badge:        false,
		Partner:      false,
		Sort:         "family_name",
		SortDir:      "ASC",
	}
}

type BadgeFilter struct {
	Number string
	InUse  string
	Type   string
}

func NewBadgeFilter(search string) BadgeFilter {
	return BadgeFilter{
		Number: search,
		InUse:  "all",
		Type:   "all",
	}
}

type AbonnementFilter struct {
	SearchString string
	AboValidity  string
	Member       bool
	Unique       bool
	Sort         string
	SortDir      string
}

func NewAbonnementFilter(search string) AbonnementFilter {
	return AbonnementFilter{
		SearchString: search,
		AboValidity:  "all",
		Member:       false,
		Unique:       false,
		Sort:         "buydate",
		SortDir:      "DESC",
	}
}

type LogisticFilter struct {
	Filter   string
	InUse    string
	Validity string
	Type     int
}

func NewLogisticFilter(search string) LogisticFilter {
	return LogisticFilter{
		Filter:   search,
		InUse:    "all",
		Validity: "all",
		Type:     0,
	}
}

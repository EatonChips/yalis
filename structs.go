package main

// Person is a person result
type Person struct {
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	PublicIdentifier string `json:"publicIdentifier"`
	Occupation       string `json:"occupation"`
	CompanyID        string `json:"companyId"`
}

// SearchResponse is the response from the search request
type SearchResponse struct {
	Data struct {
		Metadata struct {
			TotalResultCount int    `json:"totalResultCount"`
			Origin           string `json:"origin"`
		} `json:"metadata"`
		Paging struct {
			Count int `json:"count"`
			Start int `json:"start"`
			Total int `json:"total"`
		} `json:"paging"`
	} `json:"data"`
	Included []Person `json:"included"`
}

// CompanyLookupResponse is the response from looking up a company
type CompanyLookupResponse struct {
	Data struct {
		Elements []struct {
			Elements []struct {
				TargetURN string `json:"targetUrn"`
				Title     struct {
					Text string `json:"text"`
				} `json:"title"`
			} `json:"elements"`
		} `json:"elements"`
		Included []struct {
			ObjectURN     string `json:"objectUrn"`
			UniversalName string `json:"universalName"`
		}
	} `json:"data"`
}

// StringSlice is a slice of strings
type StringSlice []string

func (i *StringSlice) String() string {
	return "asdf"
}

func (i *StringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

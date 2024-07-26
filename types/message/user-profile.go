package typesMessage

type BusinessPartnerWithDetails struct {
	BusinessPartner          int
	BusinessPartnerType      string
	NickName                 string
	ProfileComment           *string
	PreferableLocalSubRegion string
	PreferableLocalRegion    string
	PreferableCountry        string
	LocalRegionName          *string
	LocalSubRegionName       *string
}

package types

type Request struct {
	BusinessPartner     *int    `json:"BusinessPartner"`
	BusinessPartnerRole *string `json:"BusinessPartnerRole"`
	UserID              *string `json:"UserId"`
	User                *string `json:"User"`
	Language            *string `json:"Language"`
	UserType            *string `json:"UserType"`
	RuntimeSessionID    *string `json:"RuntimeSessionId"`
}

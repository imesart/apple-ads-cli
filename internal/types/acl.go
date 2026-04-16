package types

// UserACL is the response to ACL requests.
type UserACL struct {
	OrgID        int64        `json:"orgId"`
	OrgName      string       `json:"orgName"`
	ParentOrgID  *int64       `json:"parentOrgId,omitempty"`
	Currency     string       `json:"currency"`
	PaymentModel PaymentModel `json:"paymentModel"`
	RoleNames    []string     `json:"roleNames"`
	TimeZone     string       `json:"timeZone"`
}

// MeDetail contains the API caller identifiers.
type MeDetail struct {
	UserID      int64 `json:"userId"`
	ParentOrgID int64 `json:"parentOrgId"`
}

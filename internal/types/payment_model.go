package types

// PaymentModel is the payment model set through the Apple Search Ads UI.
type PaymentModel string

const (
	PaymentModelPAYG PaymentModel = "PAYG"
	PaymentModelLOC  PaymentModel = "LOC"
)

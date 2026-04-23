package types

// CampaignStatus is the user-controlled status to enable or pause the campaign.
type CampaignStatus string

const (
	CampaignStatusEnabled CampaignStatus = "ENABLED"
	CampaignStatusPaused  CampaignStatus = "PAUSED"
)

// CampaignDisplayStatus is the display status of the campaign.
type CampaignDisplayStatus string

const (
	CampaignDisplayStatusRunning CampaignDisplayStatus = "RUNNING"
	CampaignDisplayStatusOnHold  CampaignDisplayStatus = "ON_HOLD"
	CampaignDisplayStatusPaused  CampaignDisplayStatus = "PAUSED"
	CampaignDisplayStatusDeleted CampaignDisplayStatus = "DELETED"
)

// CampaignServingStatus is the serving status of the campaign.
type CampaignServingStatus string

const (
	CampaignServingStatusRunning    CampaignServingStatus = "RUNNING"
	CampaignServingStatusNotRunning CampaignServingStatus = "NOT_RUNNING"
)

// CampaignAdChannelType is the channel type of an ad in a campaign.
type CampaignAdChannelType string

const (
	CampaignAdChannelTypeSearch  CampaignAdChannelType = "SEARCH"
	CampaignAdChannelTypeDisplay CampaignAdChannelType = "DISPLAY"
)

// CampaignBillingEventType is the type of billing event for a campaign.
type CampaignBillingEventType string

const (
	CampaignBillingEventTypeTaps        CampaignBillingEventType = "TAPS"
	CampaignBillingEventTypeImpressions CampaignBillingEventType = "IMPRESSIONS"
)

// CampaignServingStateReason is a reason the system provides when a campaign can't run.
type CampaignServingStateReason string

const (
	CampaignServingStateReasonAdGroupMissing                 CampaignServingStateReason = "AD_GROUP_MISSING"
	CampaignServingStateReasonAppNotCategorized              CampaignServingStateReason = "APP_NOT_CATEGORIZED"
	CampaignServingStateReasonAppNotEligible                 CampaignServingStateReason = "APP_NOT_ELIGIBLE"
	CampaignServingStateReasonAppNotEligibleSearchAds        CampaignServingStateReason = "APP_NOT_ELIGIBLE_SEARCHADS"
	CampaignServingStateReasonAppNotEligibleSupplySource     CampaignServingStateReason = "APP_NOT_ELIGIBLE_SUPPLY_SOURCE"
	CampaignServingStateReasonAppNotLinkedToCampaignGroup    CampaignServingStateReason = "APP_NOT_LINKED_TO_CAMPAIGN_GROUP"
	CampaignServingStateReasonAppNotPublishedYet             CampaignServingStateReason = "APP_NOT_PUBLISHED_YET"
	CampaignServingStateReasonAppSensitiveContent            CampaignServingStateReason = "APP_SENSITIVE_CONTENT"
	CampaignServingStateReasonBOStartDateInFuture            CampaignServingStateReason = "BO_START_DATE_IN_FUTURE"
	CampaignServingStateReasonBOEndDateReached               CampaignServingStateReason = "BO_END_DATE_REACHED"
	CampaignServingStateReasonBOExhausted                    CampaignServingStateReason = "BO_EXHAUSTED"
	CampaignServingStateReasonCampaignEndDateReached         CampaignServingStateReason = "CAMPAIGN_END_DATE_REACHED"
	CampaignServingStateReasonCampaignStartDateInFuture      CampaignServingStateReason = "CAMPAIGN_START_DATE_IN_FUTURE"
	CampaignServingStateReasonContentProviderUnlinked        CampaignServingStateReason = "CONTENT_PROVIDER_UNLINKED"
	CampaignServingStateReasonCreditCardDeclined             CampaignServingStateReason = "CREDIT_CARD_DECLINED"
	CampaignServingStateReasonDailyCapExhausted              CampaignServingStateReason = "DAILY_CAP_EXHAUSTED"
	CampaignServingStateReasonDeletedByUser                  CampaignServingStateReason = "DELETED_BY_USER"
	CampaignServingStateReasonFeatureNoLongerAvailable       CampaignServingStateReason = "FEATURE_NO_LONGER_AVAILABLE"
	CampaignServingStateReasonFeatureNotYetAvailable         CampaignServingStateReason = "FEATURE_NOT_YET_AVAILABLE"
	CampaignServingStateReasonIneligibleBusinessLocation     CampaignServingStateReason = "INELIGIBLE_BUSINESS_LOCATION"
	CampaignServingStateReasonLOCExhausted                   CampaignServingStateReason = "LOC_EXHAUSTED"
	CampaignServingStateReasonMissingBOOrInvoicingFields     CampaignServingStateReason = "MISSING_BO_OR_INVOICING_FIELDS"
	CampaignServingStateReasonNoAvailableAdGroups            CampaignServingStateReason = "NO_AVAILABLE_AD_GROUPS"
	CampaignServingStateReasonNoEligibleCountries            CampaignServingStateReason = "NO_ELIGIBLE_COUNTRIES"
	CampaignServingStateReasonNoPaymentMethodOnFile          CampaignServingStateReason = "NO_PAYMENT_METHOD_ON_FILE"
	CampaignServingStateReasonOrgChargeBackDisputed          CampaignServingStateReason = "ORG_CHARGE_BACK_DISPUTED"
	CampaignServingStateReasonOrgPaymentTypeChanged          CampaignServingStateReason = "ORG_PAYMENT_TYPE_CHANGED"
	CampaignServingStateReasonOrgSuspendedPolicyViolation    CampaignServingStateReason = "ORG_SUSPENDED_POLICY_VIOLATION"
	CampaignServingStateReasonOrgSuspendedFraud              CampaignServingStateReason = "ORG_SUSPENDED_FRAUD"
	CampaignServingStateReasonPausedByUser                   CampaignServingStateReason = "PAUSED_BY_USER"
	CampaignServingStateReasonPausedBySystem                 CampaignServingStateReason = "PAUSED_BY_SYSTEM"
	CampaignServingStateReasonSapinLawAgentUnknown           CampaignServingStateReason = "SAPIN_LAW_AGENT_UNKNOWN"
	CampaignServingStateReasonSapinLawFrenchBizUnknown       CampaignServingStateReason = "SAPIN_LAW_FRENCH_BIZ_UNKNOWN"
	CampaignServingStateReasonSapinLawFrenchBiz              CampaignServingStateReason = "SAPIN_LAW_FRENCH_BIZ"
	CampaignServingStateReasonTaxVerificationPending         CampaignServingStateReason = "TAX_VERIFICATION_PENDING"
	CampaignServingStateReasonTotalBudgetExhausted           CampaignServingStateReason = "TOTAL_BUDGET_EXHAUSTED"
	CampaignServingStateReasonUserRequestedAccountSuspension CampaignServingStateReason = "USER_REQUESTED_ACCOUNT_SUSPENSION"
)

// CampaignCountryOrRegionsServingStateReason is a reason that displays when
// a campaign can't run in a specific country or region.
type CampaignCountryOrRegionsServingStateReason string

const (
	CampaignCountryReasonAppLanguageIncompatible        CampaignCountryOrRegionsServingStateReason = "APP_LANGUAGE_INCOMPATIBLE"
	CampaignCountryReasonAppNotEligible                 CampaignCountryOrRegionsServingStateReason = "APP_NOT_ELIGIBLE"
	CampaignCountryReasonAppNotEligibleSearchAds        CampaignCountryOrRegionsServingStateReason = "APP_NOT_ELIGIBLE_SEARCHADS"
	CampaignCountryReasonAppNotPublishedYet             CampaignCountryOrRegionsServingStateReason = "APP_NOT_PUBLISHED_YET"
	CampaignCountryReasonAppNotEligibleSupplySource     CampaignCountryOrRegionsServingStateReason = "APP_NOT_ELIGIBLE_SUPPLY_SOURCE"
	CampaignCountryReasonAppContentRejected             CampaignCountryOrRegionsServingStateReason = "APP_CONTENT_REJECTED"
	CampaignCountryReasonAppContentReviewPending        CampaignCountryOrRegionsServingStateReason = "APP_CONTENT_REVIEW_PENDING"
	CampaignCountryReasonAppDocApprovalExpired          CampaignCountryOrRegionsServingStateReason = "APP_DOC_APPROVAL_EXPIRED"
	CampaignCountryReasonAppDocApprovalInfected         CampaignCountryOrRegionsServingStateReason = "APP_DOC_APPROVAL_INFECTED"
	CampaignCountryReasonAppDocApprovalNotSubmitted     CampaignCountryOrRegionsServingStateReason = "APP_DOC_APPROVAL_NOT_SUBMITTED"
	CampaignCountryReasonAppDocApprovalPending          CampaignCountryOrRegionsServingStateReason = "APP_DOC_APPROVAL_PENDING"
	CampaignCountryReasonAppDocApprovalRejected         CampaignCountryOrRegionsServingStateReason = "APP_DOC_APPROVAL_REJECTED"
	CampaignCountryReasonAccountDocApprovalExpired      CampaignCountryOrRegionsServingStateReason = "ACCOUNT_DOC_APPROVAL_EXPIRED"
	CampaignCountryReasonAccountDocApprovalInfected     CampaignCountryOrRegionsServingStateReason = "ACCOUNT_DOC_APPROVAL_INFECTED"
	CampaignCountryReasonAccountDocApprovalNotSubmitted CampaignCountryOrRegionsServingStateReason = "ACCOUNT_DOC_APPROVAL_NOT_SUBMITTED"
	CampaignCountryReasonAccountDocApprovalPending      CampaignCountryOrRegionsServingStateReason = "ACCOUNT_DOC_APPROVAL_PENDING"
	CampaignCountryReasonAccountDocApprovalRejected     CampaignCountryOrRegionsServingStateReason = "ACCOUNT_DOC_APPROVAL_REJECTED"
	CampaignCountryReasonFeatureNotAvailableInCountry   CampaignCountryOrRegionsServingStateReason = "FEATURE_NOT_AVAILABLE_IN_COUNTRY_OR_REGION"
	CampaignCountryReasonSapinLawAgentUnknown           CampaignCountryOrRegionsServingStateReason = "SAPIN_LAW_AGENT_UNKNOWN"
	CampaignCountryReasonSapinLawFrenchBizUnknown       CampaignCountryOrRegionsServingStateReason = "SAPIN_LAW_FRENCH_BIZ_UNKNOWN"
	CampaignCountryReasonSapinLawFrenchBiz              CampaignCountryOrRegionsServingStateReason = "SAPIN_LAW_FRENCH_BIZ"
)

// LOCInvoiceDetails contains standard invoice details for a monthly invoicing payment model.
type LOCInvoiceDetails struct {
	BillingContactEmail *string `json:"billingContactEmail,omitempty"`
	BuyerEmail          *string `json:"buyerEmail,omitempty"`
	BuyerName           *string `json:"buyerName,omitempty"`
	ClientName          *string `json:"clientName,omitempty"`
	OrderNumber         *string `json:"orderNumber,omitempty"`
}

// Campaign is the response to a request to create and fetch campaigns.
type Campaign struct {
	ID                                 *int64                                                  `json:"id,omitempty"`
	OrgID                              *int64                                                  `json:"orgId,omitempty"`
	AdamID                             int64                                                   `json:"adamId"`
	Name                               string                                                  `json:"name"`
	AdChannelType                      CampaignAdChannelType                                   `json:"adChannelType"`
	SupplySources                      []SupplySource                                          `json:"supplySources,omitempty"`
	Status                             *CampaignStatus                                         `json:"status,omitempty"`
	DisplayStatus                      *CampaignDisplayStatus                                  `json:"displayStatus,omitempty"`
	ServingStatus                      *CampaignServingStatus                                  `json:"servingStatus,omitempty"`
	ServingStateReasons                []CampaignServingStateReason                            `json:"servingStateReasons,omitempty"`
	CountriesOrRegions                 []string                                                `json:"countriesOrRegions"`
	CountryOrRegionServingStateReasons map[string][]CampaignCountryOrRegionsServingStateReason `json:"countryOrRegionServingStateReasons,omitempty"`
	BillingEvent                       CampaignBillingEventType                                `json:"billingEvent"`
	BudgetAmount                       *Money                                                  `json:"budgetAmount,omitempty"`
	DailyBudgetAmount                  Money                                                   `json:"dailyBudgetAmount"`
	TargetCpa                          *Money                                                  `json:"targetCpa,omitempty"`
	BudgetOrders                       []int64                                                 `json:"budgetOrders,omitempty"`
	PaymentModel                       *PaymentModel                                           `json:"paymentModel,omitempty"`
	LOCInvoiceDetails                  *LOCInvoiceDetails                                      `json:"locInvoiceDetails,omitempty"`
	Deleted                            *bool                                                   `json:"deleted,omitempty"`
	CreationTime                       *string                                                 `json:"creationTime,omitempty"`
	ModificationTime                   *string                                                 `json:"modificationTime,omitempty"`
	StartTime                          *string                                                 `json:"startTime,omitempty"`
	EndTime                            *string                                                 `json:"endTime,omitempty"`
}

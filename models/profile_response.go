package models

type Profile struct {
	Email             string             `json:"email"`
	Username          string             `json:"username"`
	InvestmentProfile *InvestmentProfile `json:"investmentProfile,omitempty"`
}

type ProfileResponse struct {
	Email             string                     `json:"email"`
	Username          string                     `json:"username"`
	InvestmentProfile *InvestmentProfileResponse `json:"investmentProfile,omitempty"`
}

type InvestmentProfileResponse struct {
	Age                                  int     `json:"age"`
	MaxAcceptableShortTermLossPercentage int     `json:"maxAcceptableShortTermLossPercentage"`
	ExpectedAnnualizedRateOfReturn       int     `json:"expectedAnnualizedRateOfReturn"`
	TimeHorizon                          string  `json:"timeHorizon"`
	YearsInvesting                       int     `json:"yearsInvesting"`
	MonthlyCashFlow                      float64 `json:"monthlyCashFlow"`
	DefaultCurrency                      string  `json:"defaultCurrency"`
}

package serializers

import (
	"time"

	"github.com/czConstant/constant-nftylend-api/models"
	"github.com/czConstant/constant-nftylend-api/types/numeric"
)

type LoanResp struct {
	ID                   uint              `json:"id"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
	Network              models.Chain      `json:"network"`
	Owner                string            `json:"owner"`
	Lender               string            `json:"lender"`
	AssetID              uint              `json:"asset_id"`
	Asset                *AssetResp        `json:"asset"`
	Description          string            `json:"description"`
	CurrencyID           uint              `json:"currency_id"`
	Currency             *CurrencyResp     `json:"currency"`
	StartedAt            *time.Time        `json:"started_at"`
	Duration             uint              `json:"duration"`
	ExpiredAt            *time.Time        `json:"expired_at"`
	FinishedAt           *time.Time        `json:"finished_at"`
	PrincipalAmount      numeric.BigFloat  `json:"principal_amount"`
	InterestRate         float64           `json:"interest_rate"`
	InterestAmount       numeric.BigFloat  `json:"interest_amount"`
	FeeRate              float64           `json:"fee_rate"`
	FeeAmount            numeric.BigFloat  `json:"fee_amount"`
	NonceHex             string            `json:"nonce_hex"`
	ImageUrl             string            `json:"image_url"`
	Signature            string            `json:"signature"`
	Status               models.LoanStatus `json:"status"`
	DataLoanAddress      string            `json:"data_loan_address"`
	DataAssetAddress     string            `json:"data_asset_address"`
	Offers               []*LoanOfferResp  `json:"offers"`
	ApprovedOffer        *LoanOfferResp    `json:"approved_offer"`
	OfferStartedAt       *time.Time        `json:"offer_started_at"`
	OfferDuration        uint              `json:"offer_duration"`
	OfferExpiredAt       *time.Time        `json:"offer_expired_at"`
	OfferPrincipalAmount numeric.BigFloat  `json:"offer_principal_amount"`
	OfferInterestRate    float64           `json:"offer_interest_rate"`
	InitTxHash           string            `json:"init_tx_hash"`
	CancelTxHash         string            `json:"cancel_tx_hash"`
	PayTxHash            string            `json:"pay_tx_hash"`
	LiquidateTxHash      string            `json:"liquidate_tx_hash"`
}

func NewLoanResp(m *models.Loan) *LoanResp {
	if m == nil {
		return nil
	}
	resp := &LoanResp{
		ID:                   m.ID,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
		Network:              m.Network,
		Owner:                m.Owner,
		Lender:               m.Lender,
		AssetID:              m.AssetID,
		Asset:                NewAssetResp(m.Asset),
		CurrencyID:           m.CurrencyID,
		Currency:             NewCurrencyResp(m.Currency),
		StartedAt:            m.StartedAt,
		Duration:             m.Duration,
		ExpiredAt:            m.ExpiredAt,
		FinishedAt:           m.FinishedAt,
		PrincipalAmount:      m.PrincipalAmount,
		InterestRate:         m.InterestRate,
		FeeRate:              m.FeeRate,
		FeeAmount:            m.FeeAmount,
		NonceHex:             m.NonceHex,
		ImageUrl:             m.ImageUrl,
		Signature:            m.Signature,
		Status:               m.Status,
		DataLoanAddress:      m.DataLoanAddress,
		DataAssetAddress:     m.DataAssetAddress,
		Offers:               NewLoanOfferRespArr(m.Offers),
		ApprovedOffer:        NewLoanOfferResp(m.ApprovedOffer),
		OfferStartedAt:       m.OfferStartedAt,
		OfferDuration:        m.OfferDuration,
		OfferExpiredAt:       m.OfferExpiredAt,
		OfferPrincipalAmount: m.OfferPrincipalAmount,
		OfferInterestRate:    m.OfferInterestRate,
		InitTxHash:           m.InitTxHash,
		CancelTxHash:         m.CancelTxHash,
		PayTxHash:            m.PayTxHash,
		LiquidateTxHash:      m.LiquidateTxHash,
	}
	return resp
}

func NewLoanRespArr(arr []*models.Loan) []*LoanResp {
	resps := []*LoanResp{}
	for _, m := range arr {
		resps = append(resps, NewLoanResp(m))
	}
	return resps
}

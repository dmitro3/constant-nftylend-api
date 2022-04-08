package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/czConstant/constant-nftylend-api/daos"
	"github.com/czConstant/constant-nftylend-api/errs"
	"github.com/czConstant/constant-nftylend-api/helpers"
	"github.com/czConstant/constant-nftylend-api/models"
	"github.com/czConstant/constant-nftylend-api/serializers"
	"github.com/jinzhu/gorm"
)

func (s *NftLend) GetListingLoans(
	ctx context.Context,
	network models.Network,
	collectionId uint,
	minPrice float64,
	maxPrice float64,
	minDuration uint,
	maxDuration uint,
	minInterestRate float64,
	maxInterestRate float64,
	excludeIds []uint,
	sort []string,
	page int,
	limit int,
) ([]*models.Loan, uint, error) {
	filters := map[string][]interface{}{
		"status in (?)": []interface{}{
			[]models.LoanStatus{
				models.LoanStatusNew,
			}},
	}
	if network != "" {
		filters["network = ?"] = []interface{}{network}
	}
	if collectionId > 0 {
		filters[`
		exists(
			select 1
			from assets
			where asset_id = assets.id
			  and assets.collection_id = ?
		)
		`] = []interface{}{collectionId}
	}
	if minPrice > 0 {
		filters["principal_amount >= ?"] = []interface{}{minPrice}
	}
	if maxPrice > 0 {
		filters["principal_amount <= ?"] = []interface{}{maxPrice}
	}
	if minDuration > 0 {
		filters["duration >= ?"] = []interface{}{minDuration}
	}
	if maxDuration > 0 {
		filters["duration <= ?"] = []interface{}{maxDuration}
	}
	if len(excludeIds) > 0 {
		filters["id not in (?)"] = []interface{}{excludeIds}
	}
	if minInterestRate > 0 {
		filters["interest_rate >= ?"] = []interface{}{minInterestRate}
	}
	if maxInterestRate > 0 {
		filters["interest_rate <= ?"] = []interface{}{maxInterestRate}
	}
	if len(excludeIds) > 0 {
		filters["id not in (?)"] = []interface{}{excludeIds}
	}
	if len(sort) == 0 {
		sort = []string{"id desc"}
	}
	loans, count, err := s.ld.Find4Page(
		daos.GetDBMainCtx(ctx),
		filters,
		map[string][]interface{}{
			"Asset":            []interface{}{},
			"Asset.Collection": []interface{}{},
			"Currency":         []interface{}{},
			"ApprovedOffer": []interface{}{
				"status = ?",
				models.LoanOfferStatusApproved,
			},
		},
		sort,
		page,
		limit,
	)
	if err != nil {
		return nil, 0, errs.NewError(err)
	}
	return loans, count, nil
}

func (s *NftLend) GetLoans(ctx context.Context, network models.Network, owner string, lender string, assetId uint, statues []string, page int, limit int) ([]*models.Loan, uint, error) {
	filters := map[string][]interface{}{}
	if network != "" {
		filters["network = ?"] = []interface{}{network}
	}
	if owner != "" {
		filters["owner = ?"] = []interface{}{owner}
	}
	if lender != "" {
		filters["lender = ?"] = []interface{}{lender}
	}
	if assetId > 0 {
		filters["asset_id = ?"] = []interface{}{assetId}
	}
	if len(statues) > 0 {
		filters["status in (?)"] = []interface{}{statues}
	}
	loans, count, err := s.ld.Find4Page(
		daos.GetDBMainCtx(ctx),
		filters,
		map[string][]interface{}{
			"Asset":            []interface{}{},
			"Asset.Collection": []interface{}{},
			"Currency":         []interface{}{},
			"ApprovedOffer": []interface{}{
				"status = ?",
				models.LoanOfferStatusApproved,
			},
		},
		[]string{"id desc"},
		page,
		limit,
	)
	if err != nil {
		return nil, 0, errs.NewError(err)
	}
	return loans, count, nil
}

func (s *NftLend) GetLoanOffers(ctx context.Context, network models.Network, borrower string, lender string, statues []string, page int, limit int) ([]*models.LoanOffer, uint, error) {
	filters := map[string][]interface{}{}
	if borrower != "" {
		filters[`
		exists(
			select 1
			from loans
			where loan_id = loans.id
			  and loans.owner = ?
		)
		`] = []interface{}{borrower}
	}
	if lender != "" {
		filters["lender = ?"] = []interface{}{lender}
	}
	if network != "" {
		filters["network = ?"] = []interface{}{network}
	}
	if len(statues) > 0 {
		filters["status in (?)"] = []interface{}{statues}
	}
	offers, count, err := s.lod.Find4Page(
		daos.GetDBMainCtx(ctx),
		filters,
		map[string][]interface{}{
			"Loan":                  []interface{}{},
			"Loan.Asset":            []interface{}{},
			"Loan.Asset.Collection": []interface{}{},
			"Loan.Currency":         []interface{}{},
		},
		[]string{"id desc"},
		page,
		limit,
	)
	if err != nil {
		return nil, 0, errs.NewError(err)
	}
	return offers, count, nil
}

func (s *NftLend) GetLastListingLoanByCollection(ctx context.Context, collectionId uint) (*models.Loan, error) {
	filters := map[string][]interface{}{
		"status in (?)": []interface{}{
			[]models.LoanStatus{
				models.LoanStatusNew,
			}},
	}
	filters[`
		exists(
			select 1
			from assets
			where asset_id = assets.id
			  and assets.collection_id = ?
		)
		`] = []interface{}{collectionId}
	loan, err := s.ld.First(
		daos.GetDBMainCtx(ctx),
		filters,
		map[string][]interface{}{
			"Asset": []interface{}{},
			"ApprovedOffer": []interface{}{
				"status = ?",
				models.LoanOfferStatusApproved,
			},
		},
		[]string{"id desc"},
	)
	if err != nil {
		return nil, errs.NewError(err)
	}
	return loan, nil
}

func (s *NftLend) GetRPTCollectionLoan(ctx context.Context, collectionID uint) (*models.NftyRPTCollectionLoan, error) {
	m, err := s.ld.GetRPTCollectionLoan(
		daos.GetDBMainCtx(ctx),
		collectionID,
	)
	if err != nil {
		return nil, errs.NewError(err)
	}
	if m == nil {
		return nil, errs.NewError(errs.ErrBadRequest)
	}
	return m, nil
}

func (s *NftLend) GetLoanTransactions(ctx context.Context, assetId uint, page int, limit int) ([]*models.LoanTransaction, uint, error) {
	filters := map[string][]interface{}{}
	if assetId > 0 {
		filters[`
		exists(
			select 1
			from loans
			where loan_id = loans.id
			  and loans.asset_id = ?
		)
		`] = []interface{}{assetId}
	}
	txns, count, err := s.ltd.Find4Page(
		daos.GetDBMainCtx(ctx),
		filters,
		map[string][]interface{}{
			"Loan":                  []interface{}{},
			"Loan.Asset":            []interface{}{},
			"Loan.Asset.Collection": []interface{}{},
			"Loan.Currency":         []interface{}{},
		},
		[]string{"id desc"},
		page,
		limit,
	)
	if err != nil {
		return nil, 0, errs.NewError(err)
	}
	return txns, count, nil
}

func (s *NftLend) CreateLoan(ctx context.Context, req *serializers.CreateLoanReq) (*models.Loan, error) {
	var loan *models.Loan
	if req.PrincipalAmount.Float.Cmp(big.NewFloat(0)) <= 0 ||
		req.CurrencyID <= 0 ||
		req.Duration <= 0 ||
		req.Borrower == "" ||
		req.ContractAddress == "" ||
		req.TokenID == "" ||
		req.NonceHex == "" ||
		req.Signature == "" {
		return nil, errs.NewError(errs.ErrBadRequest)
	}
	req.Borrower = strings.ToLower(req.Borrower)
	req.NonceHex = strings.ToLower(req.NonceHex)
	req.Signature = strings.ToLower(req.Signature)
	req.ContractAddress = strings.ToLower(req.ContractAddress)
	switch req.Network {
	case models.NetworkMATIC,
		models.NetworkAVAX,
		models.NetworkBSC:
		{
		}
	default:
		{
			return nil, errs.NewError(errs.ErrBadRequest)
		}
	}
	err := daos.WithTransaction(
		daos.GetDBMainCtx(ctx),
		func(tx *gorm.DB) error {
			currency, err := s.GetCurrencyByID(tx, req.CurrencyID, req.Network)
			if err != nil {
				return errs.NewError(err)
			}
			msgHex := helpers.AppendHexStrings(
				helpers.ParseBigInt2Hex(models.Number2BigInt(req.PrincipalAmount.String(), int(currency.Decimals))),
				helpers.ParseBigInt2Hex(models.Number2BigInt(req.TokenID, 0)),
				helpers.ParseBigInt2Hex(big.NewInt(int64(req.Duration))),
				helpers.ParseBigInt2Hex(models.Number2BigInt(fmt.Sprintf("%f", req.InterestRate), 4)),
				helpers.ParseBigInt2Hex(big.NewInt(s.getEvmAdminFee(req.Network))),
				helpers.ParseHex2Hex(req.NonceHex),
				helpers.ParseAddress2Hex(req.ContractAddress),
				helpers.ParseAddress2Hex(currency.ContractAddress),
				helpers.ParseAddress2Hex(req.Borrower),
				helpers.ParseBigInt2Hex(big.NewInt(s.getEvmClientByNetwork(req.Network).ChainID)),
			)
			err = s.getEvmClientByNetwork(req.Network).ValidateSignature(
				msgHex,
				req.Signature,
				req.Borrower,
			)
			if err != nil {
				return errs.NewError(err)
			}
			asset, err := s.ad.First(
				tx,
				map[string][]interface{}{
					"network = ?":          []interface{}{req.Network},
					"contract_address = ?": []interface{}{req.ContractAddress},
					"token_id = ?":         []interface{}{req.TokenID},
				},
				map[string][]interface{}{},
				[]string{},
			)
			if err != nil {
				return errs.NewError(err)
			}
			if asset == nil {
				tokenURL, err := s.getEvmClientByNetwork(req.Network).NftTokenURI(req.ContractAddress, req.TokenID)
				if err != nil {
					return errs.NewError(err)
				}
				meta, err := s.stc.GetEvmNftMetaResp(tokenURL)
				if err != nil {
					return errs.NewError(err)
				}
				collection, err := s.cld.First(
					tx,
					map[string][]interface{}{
						"network = ?":          []interface{}{req.Network},
						"contract_address = ?": []interface{}{req.ContractAddress},
					},
					map[string][]interface{}{},
					[]string{},
				)
				if err != nil {
					return errs.NewError(err)
				}
				if collection == nil {
					collection = &models.Collection{
						Network:         req.Network,
						SeoURL:          helpers.MakeSeoURL(req.ContractAddress),
						ContractAddress: req.ContractAddress,
						Name:            "",
						Description:     meta.Description,
						Enabled:         true,
					}
					err = s.cld.Create(
						tx,
						collection,
					)
					if err != nil {
						return errs.NewError(err)
					}
				}
				attributes, err := json.Marshal(meta.Attributes)
				if err != nil {
					return errs.NewError(err)
				}
				metaJson, err := json.Marshal(meta)
				if err != nil {
					return errs.NewError(err)
				}
				asset = &models.Asset{
					Network:               req.Network,
					CollectionID:          collection.ID,
					SeoURL:                helpers.MakeSeoURL(fmt.Sprintf("%s-%s", req.ContractAddress, req.TokenID)),
					ContractAddress:       collection.ContractAddress,
					TokenID:               req.TokenID,
					Symbol:                "",
					Name:                  meta.Name,
					TokenURL:              meta.Image,
					ExternalUrl:           meta.ExternalUrl,
					SellerFeeRate:         0,
					Attributes:            string(attributes),
					MetaJson:              string(metaJson),
					MetaJsonUrl:           tokenURL,
					OriginNetwork:         "",
					OriginContractAddress: "",
					OriginTokenID:         "",
				}
				err = s.ad.Create(
					tx,
					asset,
				)
				if err != nil {
					return errs.NewError(err)
				}
			}
			loan, err = s.ld.First(
				tx,
				map[string][]interface{}{
					"network = ?":  []interface{}{req.Network},
					"asset_id = ?": []interface{}{asset.ID},
					"status = ?":   []interface{}{models.LoanStatusNew},
				},
				map[string][]interface{}{},
				[]string{},
			)
			if err != nil {
				return errs.NewError(err)
			}
			if loan != nil {
				return errs.NewError(errs.ErrBadRequest)
			}
			loan = &models.Loan{
				Network:         req.Network,
				Owner:           req.Borrower,
				PrincipalAmount: req.PrincipalAmount,
				InterestRate:    req.InterestRate,
				Duration:        req.Duration,
				StartedAt:       helpers.TimeNow(),
				ExpiredAt:       helpers.TimeAdd(time.Now(), time.Duration(req.Duration)*time.Second),
				CurrencyID:      currency.ID,
				AssetID:         asset.ID,
				Status:          models.LoanStatusNew,
				Signature:       req.Signature,
				NonceHex:        req.NonceHex,
			}
			err = s.ld.Create(
				tx,
				loan,
			)
			if err != nil {
				return errs.NewError(err)
			}
			err = s.ltd.Create(
				tx,
				&models.LoanTransaction{
					Network:         req.Network,
					Type:            models.LoanTransactionTypeListed,
					LoanID:          loan.ID,
					Borrower:        loan.Owner,
					PrincipalAmount: loan.PrincipalAmount,
					InterestRate:    loan.InterestRate,
					StartedAt:       loan.StartedAt,
					Duration:        loan.Duration,
					ExpiredAt:       loan.ExpiredAt,
				},
			)
			if err != nil {
				return errs.NewError(err)
			}
			return nil
		},
	)
	if err != nil {
		return nil, errs.NewError(err)
	}
	return loan, nil
}

func (s *NftLend) CreateLoanOffer(ctx context.Context, loanID uint, req *serializers.CreateLoanOfferReq) (*models.LoanOffer, error) {
	var offer *models.LoanOffer
	if req.PrincipalAmount.Float.Cmp(big.NewFloat(0)) <= 0 ||
		req.Duration <= 0 ||
		req.Lender == "" ||
		req.NonceHex == "" ||
		req.Signature == "" {
		return nil, errs.NewError(errs.ErrBadRequest)
	}
	req.Lender = strings.ToLower(req.Lender)
	req.NonceHex = strings.ToLower(req.NonceHex)
	req.Signature = strings.ToLower(req.Signature)
	err := daos.WithTransaction(
		daos.GetDBMainCtx(ctx),
		func(tx *gorm.DB) error {
			loan, err := s.ld.FirstByID(
				tx,
				loanID,
				map[string][]interface{}{},
				false,
			)
			if err != nil {
				return errs.NewError(err)
			}
			if loan == nil {
				return errs.NewError(errs.ErrBadRequest)
			}
			if loan.Status != models.LoanStatusNew {
				return errs.NewError(errs.ErrBadRequest)
			}
			currency, err := s.GetCurrencyByID(tx, loan.CurrencyID, loan.Network)
			if err != nil {
				return errs.NewError(err)
			}
			asset, err := s.ad.FirstByID(tx, loan.AssetID, map[string][]interface{}{}, false)
			if err != nil {
				return errs.NewError(err)
			}
			if asset == nil {
				return errs.NewError(errs.ErrBadRequest)
			}
			msgHex := helpers.AppendHexStrings(
				helpers.ParseBigInt2Hex(models.Number2BigInt(req.PrincipalAmount.String(), int(currency.Decimals))),
				helpers.ParseBigInt2Hex(models.Number2BigInt(asset.TokenID, 0)),
				helpers.ParseBigInt2Hex(big.NewInt(int64(req.Duration))),
				helpers.ParseBigInt2Hex(models.Number2BigInt(fmt.Sprintf("%f", req.InterestRate), 4)),
				helpers.ParseBigInt2Hex(big.NewInt(s.getEvmAdminFee(loan.Network))),
				helpers.ParseHex2Hex(req.NonceHex),
				helpers.ParseAddress2Hex(asset.ContractAddress),
				helpers.ParseAddress2Hex(currency.ContractAddress),
				helpers.ParseAddress2Hex(req.Lender),
				helpers.ParseBigInt2Hex(big.NewInt(s.getEvmClientByNetwork(loan.Network).ChainID)),
			)
			err = s.getEvmClientByNetwork(loan.Network).ValidateSignature(
				msgHex,
				req.Signature,
				req.Lender,
			)
			if err != nil {
				return errs.NewError(err)
			}
			offer, err = s.lod.First(
				tx,
				map[string][]interface{}{
					"lender = ?":    []interface{}{req.Lender},
					"nonce_hex = ?": []interface{}{req.NonceHex},
				},
				map[string][]interface{}{},
				[]string{},
			)
			if err != nil {
				return errs.NewError(err)
			}
			if offer != nil {
				return errs.NewError(errs.ErrBadRequest)
			}
			offer = &models.LoanOffer{
				Network:         loan.Network,
				LoanID:          loan.ID,
				Lender:          req.Lender,
				PrincipalAmount: req.PrincipalAmount,
				InterestRate:    req.InterestRate,
				Duration:        req.Duration,
				Status:          models.LoanOfferStatusNew,
				NonceHex:        req.NonceHex,
				Signature:       req.Signature,
			}
			err = s.lod.Create(
				tx,
				offer,
			)
			if err != nil {
				return errs.NewError(err)
			}
			return nil
		},
	)
	if err != nil {
		return nil, errs.NewError(err)
	}
	return offer, nil
}

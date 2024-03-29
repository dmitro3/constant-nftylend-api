package services

import (
	"context"

	"github.com/czConstant/constant-nftylend-api/daos"
	"github.com/czConstant/constant-nftylend-api/errs"
	"github.com/czConstant/constant-nftylend-api/models"
)

func (s *NftLend) GetListingLoans(
	ctx context.Context,
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

func (s *NftLend) GetLoans(ctx context.Context, owner string, lender string, assetId uint, statues []string, page int, limit int) ([]*models.Loan, uint, error) {
	filters := map[string][]interface{}{}
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

func (s *NftLend) GetLoanOffers(ctx context.Context, borrower string, lender string, statues []string, page int, limit int) ([]*models.LoanOffer, uint, error) {
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

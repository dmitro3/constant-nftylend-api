package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/czConstant/blockchain-api/bcclient"
	"github.com/czConstant/blockchain-api/bcclient/solana"
	"github.com/czConstant/constant-nftylend-api/daos"
	"github.com/czConstant/constant-nftylend-api/errs"
	"github.com/czConstant/constant-nftylend-api/helpers"
	"github.com/czConstant/constant-nftylend-api/models"
	"github.com/czConstant/constant-nftylend-api/services/3rd/saletrack"
	"github.com/czConstant/constant-nftylend-api/types/numeric"
	"github.com/jinzhu/gorm"
)

type NftLend struct {
	bcs  *bcclient.Client
	stc  *saletrack.Client
	cd   *daos.Currency
	cld  *daos.Collection
	clsd *daos.CollectionSubmitted
	ad   *daos.Asset
	atd  *daos.AssetTransaction
	ld   *daos.Loan
	lod  *daos.LoanOffer
	ltd  *daos.LoanTransaction
	id   *daos.Instruction
}

func NewNftLend(
	bcs *bcclient.Client,
	stc *saletrack.Client,
	cd *daos.Currency,
	cld *daos.Collection,
	clsd *daos.CollectionSubmitted,
	ad *daos.Asset,
	atd *daos.AssetTransaction,
	ld *daos.Loan,
	lod *daos.LoanOffer,
	ltd *daos.LoanTransaction,
	id *daos.Instruction,

) *NftLend {
	s := &NftLend{
		bcs:  bcs,
		stc:  stc,
		cd:   cd,
		cld:  cld,
		clsd: clsd,
		ad:   ad,
		atd:  atd,
		ld:   ld,
		lod:  lod,
		ltd:  ltd,
		id:   id,
	}
	go stc.StartWssSolsea(s.solseaMsgReceived)
	return s
}

func (s *NftLend) getLendCurrency(tx *gorm.DB, address string) (*models.Currency, error) {
	c, err := s.cd.First(
		tx,
		map[string][]interface{}{
			"contract_address = ?": []interface{}{address},
		},
		map[string][]interface{}{},
		[]string{},
	)
	if err != nil {
		return nil, errs.NewError(err)
	}
	if c == nil {
		return nil, errs.NewError(errs.ErrBadRequest)
	}
	return c, nil
}

func (s *NftLend) getLendCurrencyBySymbol(tx *gorm.DB, symbol string) (*models.Currency, error) {
	c, err := s.cd.First(
		tx,
		map[string][]interface{}{
			"symbol = ?": []interface{}{symbol},
		},
		map[string][]interface{}{},
		[]string{},
	)
	if err != nil {
		return nil, errs.NewError(err)
	}
	if c == nil {
		return nil, errs.NewError(errs.ErrBadRequest)
	}
	return c, nil
}

func (s *NftLend) GetAssetDetail(ctx context.Context, seoURL string) (*models.Asset, error) {
	m, err := s.ad.First(
		daos.GetDBMainCtx(ctx),
		map[string][]interface{}{
			"seo_url = ?": []interface{}{seoURL},
		},
		map[string][]interface{}{
			"Collection": []interface{}{},
			"NewLoan": []interface{}{
				"status = ?",
				models.LoanStatusNew,
			},
			"NewLoan.Currency": []interface{}{},
			"NewLoan.Offers": []interface{}{
				func(db *gorm.DB) *gorm.DB {
					return db.Order("loan_offers.id DESC")
				},
			},
		},
		[]string{"id desc"},
	)
	if err != nil {
		return nil, errs.NewError(err)
	}
	return m, nil
}

func (s *NftLend) GetCollections(ctx context.Context, page int, limit int) ([]*models.Collection, uint, error) {
	categories, count, err := s.cld.Find4Page(
		daos.GetDBMainCtx(ctx),
		map[string][]interface{}{},
		map[string][]interface{}{
			"ListingAsset": []interface{}{
				`id in (
					select asset_id
					from loans
					where asset_id = assets.id
					  and loans.status in (?)
				)`,
				[]models.LoanOfferStatus{
					models.LoanOfferStatusNew,
				},
				func(db *gorm.DB) *gorm.DB {
					return db.Order(`
					(
						select max(loans.created_at)
						from loans
						where asset_id = assets.id
						  and loans.status in ('new')
					) desc
					`)
				},
			},
		},
		[]string{"id desc"},
		page,
		limit,
	)
	if err != nil {
		return nil, 0, errs.NewError(err)
	}
	return categories, count, nil
}

func (s *NftLend) GetCollectionDetail(ctx context.Context, seoURL string) (*models.Collection, error) {
	m, err := s.cld.First(
		daos.GetDBMainCtx(ctx),
		map[string][]interface{}{
			"seo_url = ?": []interface{}{seoURL},
		},
		map[string][]interface{}{},
		[]string{"id desc"},
	)
	if err != nil {
		return nil, errs.NewError(err)
	}
	return m, nil
}

func (s *NftLend) GetCurrencies(ctx context.Context) ([]*models.Currency, error) {
	currencies, err := s.cd.Find(
		daos.GetDBMainCtx(ctx),
		map[string][]interface{}{
			"enabled = ?": []interface{}{true},
		},
		map[string][]interface{}{},
		[]string{"id desc"},
		0,
		99999999,
	)
	if err != nil {
		return nil, errs.NewError(err)
	}
	return currencies, nil
}

func (s *NftLend) GetRPTListingCollection(ctx context.Context) ([]*models.NftyRPTListingCollection, error) {
	ms, err := s.ad.GetRPTListingCollection(
		daos.GetDBMainCtx(ctx),
	)
	if err != nil {
		return nil, errs.NewError(err)
	}
	return ms, nil
}

func (s *NftLend) GetCollectionVerified(ctx context.Context, mintAddress string) (*models.Collection, error) {
	m, _, err := s.getCollectionVerified(
		daos.GetDBMainCtx(ctx),
		mintAddress,
		nil,
		nil,
	)
	if err != nil {
		return nil, errs.NewError(err)
	}
	return m, nil
}

func (s *NftLend) getCollectionVerified(tx *gorm.DB, mintAddress string, meta *solana.MetadataResp, metaInfo *solana.MetadataInfoResp) (*models.Collection, string, error) {
	vrs, err := s.bcs.SolanaNftVerifier.GetNftVerifier(mintAddress)
	if err != nil {
		return nil, "", errs.NewError(err)
	}
	if vrs.IsWrapped {
		chain := s.bcs.SolanaNftVerifier.ParseChain(vrs.ChainID)
		m, err := s.cld.First(
			tx,
			map[string][]interface{}{
				"origin_network = ?":          []interface{}{chain},
				"origin_contract_address = ?": []interface{}{vrs.AssetAddress},
				"enabled = ?":                 []interface{}{true},
			},
			map[string][]interface{}{},
			[]string{"id desc"},
		)
		if err != nil {
			return nil, "", errs.NewError(err)
		}
		if m != nil {
			return m, vrs.TokenID, nil
		}
	} else {
		if meta == nil {
			meta, err = s.bcs.Solana.GetMetadata(mintAddress)
			if err != nil {
				return nil, "", errs.NewError(err)
			}
		}
		if metaInfo == nil {
			metaInfo, err = s.bcs.Solana.GetMetadataInfo(meta.Data.Uri)
			if err != nil {
				return nil, "", errs.NewError(err)
			}
		}
		collectionName := metaInfo.Collection.Name
		if collectionName == "" {
			collectionName = metaInfo.Collection.Family
			if collectionName == "" {
				names := strings.Split(metaInfo.Name, "#")
				if len(names) >= 2 {
					collectionName = strings.TrimSpace(names[0])
				}
			}
		}
		if collectionName == "" {
			return nil, "", errs.NewError(err)
		}
		for _, creator := range meta.Data.Creators {
			m, err := s.cld.First(
				tx,
				map[string][]interface{}{
					"name = ?":       []interface{}{collectionName},
					"creator like ?": []interface{}{fmt.Sprintf("%%%s%%", creator.Address)},
					"enabled = ?":    []interface{}{true},
				},
				map[string][]interface{}{},
				[]string{},
			)
			if err != nil {
				return nil, "", errs.NewError(err)
			}
			if m != nil {
				return m, "", nil
			}
		}
	}
	return nil, "", nil
}

func (s *NftLend) GetAseetTransactions(ctx context.Context, assetId uint, page int, limit int) ([]*models.AssetTransaction, uint, error) {
	err := s.updateAssetTransactions(ctx, assetId)
	if err != nil {
		return nil, 0, errs.NewError(err)
	}
	filters := map[string][]interface{}{}
	if assetId > 0 {
		filters["asset_id = ?"] = []interface{}{assetId}
	}
	txns, count, err := s.atd.Find4Page(
		daos.GetDBMainCtx(ctx),
		filters,
		map[string][]interface{}{
			"Asset":            []interface{}{},
			"Asset.Collection": []interface{}{},
		},
		[]string{"transaction_at desc"},
		page,
		limit,
	)
	if err != nil {
		return nil, 0, errs.NewError(err)
	}
	return txns, count, nil
}

func (s *NftLend) updateAssetTransactions(ctx context.Context, assetId uint) error {
	asset, err := s.ad.FirstByID(
		daos.GetDBMainCtx(ctx),
		assetId,
		map[string][]interface{}{},
		false,
	)
	if err != nil {
		return errs.NewError(err)
	}
	if asset == nil {
		return errs.NewError(errs.ErrBadRequest)
	}
	if asset.MagicEdenCrawAt == nil ||
		asset.MagicEdenCrawAt.Before(time.Now().Add(-24*time.Hour)) {
		c, err := s.getLendCurrencyBySymbol(daos.GetDBMainCtx(ctx), "SOL")
		if err != nil {
			return errs.NewError(err)
		}
		tokenAddress := asset.ContractAddress
		if asset.TestContractAddress != "" {
			tokenAddress = asset.TestContractAddress
		}
		rs, _ := s.stc.GetMagicEdenSaleHistories(tokenAddress)
		for _, r := range rs {
			if r.TxType == "exchange" {
				txnAt := time.Unix(r.BlockTime, 0)
				_ = s.atd.Create(
					daos.GetDBMainCtx(ctx),
					&models.AssetTransaction{
						Source:        "magiceden.io",
						Network:       models.ChainSOL,
						AssetID:       asset.ID,
						Type:          models.AssetTransactionTypeExchange,
						Seller:        r.SellerAddress,
						Buyer:         r.BuyerAddress,
						TransactionID: r.TransactionID,
						TransactionAt: &txnAt,
						Amount:        numeric.BigFloat{*models.ConvertWeiToBigFloat(big.NewInt(int64(r.ParsedTransaction.TotalAmount)), 9)},
						CurrencyID:    c.ID,
					},
				)
			}
		}
		err = daos.WithTransaction(
			daos.GetDBMainCtx(ctx),
			func(tx *gorm.DB) error {
				asset, err := s.ad.FirstByID(
					tx,
					asset.ID,
					map[string][]interface{}{},
					true,
				)
				if err != nil {
					return errs.NewError(err)
				}
				if asset == nil {
					return errs.NewError(errs.ErrBadRequest)
				}
				asset.MagicEdenCrawAt = helpers.TimeNow()
				err = s.ad.Save(
					tx,
					asset,
				)
				if err != nil {
					return errs.NewError(err)
				}
				return nil
			},
		)
		if err != nil {
			return errs.NewError(err)
		}
	}
	if asset.SolanartCrawAt == nil ||
		asset.SolanartCrawAt.Before(time.Now().Add(-24*time.Hour)) {
		tokenAddress := asset.ContractAddress
		if asset.TestContractAddress != "" {
			tokenAddress = asset.TestContractAddress
		}
		rs, _ := s.stc.GetSolnartSaleHistories(tokenAddress)
		for _, r := range rs {
			c, err := s.getLendCurrencyBySymbol(daos.GetDBMainCtx(ctx), r.Currency)
			if err != nil {
				return errs.NewError(err)
			}
			_ = s.atd.Create(
				daos.GetDBMainCtx(ctx),
				&models.AssetTransaction{
					Source:        "solanart.io",
					Network:       models.ChainSOL,
					AssetID:       asset.ID,
					Type:          models.AssetTransactionTypeExchange,
					Seller:        r.SellerAddress,
					Buyer:         r.BuyerAdd,
					TransactionAt: r.Date,
					Amount:        numeric.BigFloat{*big.NewFloat(r.Price)},
					CurrencyID:    c.ID,
				},
			)
		}
		err = daos.WithTransaction(
			daos.GetDBMainCtx(ctx),
			func(tx *gorm.DB) error {
				asset, err := s.ad.FirstByID(
					tx,
					asset.ID,
					map[string][]interface{}{},
					true,
				)
				if err != nil {
					return errs.NewError(err)
				}
				if asset == nil {
					return errs.NewError(errs.ErrBadRequest)
				}
				asset.SolanartCrawAt = helpers.TimeNow()
				err = s.ad.Save(
					tx,
					asset,
				)
				if err != nil {
					return errs.NewError(err)
				}
				return nil
			},
		)
		if err != nil {
			return errs.NewError(err)
		}
	}
	if asset.SolSeaCrawAt == nil ||
		asset.SolSeaCrawAt.Before(time.Now().Add(-24*time.Hour)) {
		tokenAddress := asset.ContractAddress
		if asset.TestContractAddress != "" {
			tokenAddress = asset.TestContractAddress
		}
		s.stc.PubSolseaMsg(
			fmt.Sprintf(`421["find","listed-archive",{"Mint":"%s","status":"SOLD"}]`, tokenAddress),
		)
		err = daos.WithTransaction(
			daos.GetDBMainCtx(ctx),
			func(tx *gorm.DB) error {
				asset, err := s.ad.FirstByID(
					tx,
					asset.ID,
					map[string][]interface{}{},
					true,
				)
				if err != nil {
					return errs.NewError(err)
				}
				if asset == nil {
					return errs.NewError(errs.ErrBadRequest)
				}
				asset.SolSeaCrawAt = helpers.TimeNow()
				err = s.ad.Save(
					tx,
					asset,
				)
				if err != nil {
					return errs.NewError(err)
				}
				return nil
			},
		)
		if err != nil {
			return errs.NewError(err)
		}
	}
	return nil
}

func (s *NftLend) solseaMsgReceived(msg string) {
	if strings.HasPrefix(msg, "431") {
		msg = strings.TrimLeft(msg, "431")
		resps := []*struct {
			Data []*struct {
				Mint      string     `json:"mint"`
				Price     uint64     `json:"price"`
				SellerKey string     `json:"sellerKey"`
				BuyerKey  string     `json:"buyerKey"`
				Status    string     `json:"status"`
				ListedAt  *time.Time `json:"listedAt"`
			} `json:"data"`
		}{}
		err := json.Unmarshal([]byte(msg), &resps)
		if err != nil {
			return
		}
		c, err := s.getLendCurrencyBySymbol(daos.GetDBMainCtx(context.Background()), "SOL")
		if err != nil {
			return
		}
		for _, resp := range resps {
			if resp != nil {
				for _, d := range resp.Data {
					asset, err := s.ad.First(
						daos.GetDBMainCtx(context.Background()),
						map[string][]interface{}{
							"contract_address = ?": []interface{}{d.Mint},
						},
						map[string][]interface{}{},
						[]string{"id desc"},
					)
					if err != nil {
						return
					}
					if asset == nil {
						asset, err = s.ad.First(
							daos.GetDBMainCtx(context.Background()),
							map[string][]interface{}{
								"test_contract_address = ?": []interface{}{d.Mint},
							},
							map[string][]interface{}{},
							[]string{"id desc"},
						)
						if err != nil {
							return
						}
						if asset == nil {
							return
						}
					}
					_ = s.atd.Create(
						daos.GetDBMainCtx(context.Background()),
						&models.AssetTransaction{
							Source:        "solsea.io",
							Network:       models.ChainSOL,
							AssetID:       asset.ID,
							Type:          models.AssetTransactionTypeExchange,
							Seller:        d.SellerKey,
							Buyer:         d.BuyerKey,
							TransactionAt: d.ListedAt,
							Amount:        numeric.BigFloat{*models.ConvertWeiToBigFloat(big.NewInt(int64(d.Price)), 9)},
							CurrencyID:    c.ID,
						},
					)
				}
			}
		}
	}
}

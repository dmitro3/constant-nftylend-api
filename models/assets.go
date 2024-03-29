package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Asset struct {
	gorm.Model
	Network                   Chain
	CollectionID              uint
	Collection                *Collection
	SeoURL                    string
	ContractAddress           string
	TestContractAddress       string
	TokenURL                  string
	ExternalUrl               string
	Name                      string
	Symbol                    string
	SellerFeeRate             float64 `gorm:"type:decimal(6,4);default:0"`
	Attributes                string  `gorm:"type:text"`
	MetaJson                  string  `gorm:"type:text"`
	MetaJsonUrl               string
	NewLoan                   *Loan
	OriginNetwork             Chain
	OriginContractAddress     string
	OriginTokenID             string
	TestOriginContractAddress string
	TestOriginTokenID         uint
	MagicEdenCrawAt           *time.Time
	SolanartCrawAt            *time.Time
	SolSeaCrawAt              *time.Time
	OpenseaCrawAt             *time.Time
}

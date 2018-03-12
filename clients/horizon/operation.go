package horizon

import (
	"encoding/json"
	"time"

	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/xdr"
)

// Operation represents a single operation within a Transaction.
type Operation interface {
	OperationType() string
}

// UnmarshalOperation deserializes JSON into an Operation.
func UnmarshalOperation(data []byte) (Operation, error) {
	var opType struct {
		TypeI int32 `json:"type_i"`
	}

	err := json.Unmarshal(data, &opType)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshaling operation")
	}

	var op Operation
	switch xdr.OperationType(opType.TypeI) {
	case xdr.OperationTypeCreateAccount:
		o := &CreateAccountOperation{}
		err = json.Unmarshal(data, o)
		op = o
	case xdr.OperationTypePayment:
		o := &PaymentOperation{}
		err = json.Unmarshal(data, o)
		op = o
	case xdr.OperationTypePathPayment:
		o := &PathPaymentOperation{}
		err = json.Unmarshal(data, o)
		op = o
	case xdr.OperationTypeManageOffer:
		o := &ManageOfferOperation{}
		err = json.Unmarshal(data, o)
		op = o
	case xdr.OperationTypeCreatePassiveOffer:
		o := &CreatePassiveOfferOperation{}
		err = json.Unmarshal(data, o)
		op = o
	case xdr.OperationTypeSetOptions:
		o := &SetOptionsOperation{}
		err = json.Unmarshal(data, o)
		op = o
	case xdr.OperationTypeChangeTrust:
		o := &ChangeTrustOperation{}
		err = json.Unmarshal(data, o)
		op = o
	case xdr.OperationTypeAllowTrust:
		o := &AllowTrustOperation{}
		err = json.Unmarshal(data, o)
		op = o
	case xdr.OperationTypeAccountMerge:
		o := &AccountMergeOperation{}
		err = json.Unmarshal(data, o)
		op = o
	case xdr.OperationTypeInflation:
		o := &InflationOperation{}
		err = json.Unmarshal(data, o)
		op = o
	case xdr.OperationTypeManageData:
		o := &ManageDataOperation{}
		err = json.Unmarshal(data, o)
		op = o
	default:
		return nil, errors.Errorf("unknown operation type %d", opType.TypeI)
	}

	return op, errors.Wrap(err, "error unmarshaling operation")
}

type BaseOperation struct {
	Links struct {
		Self        Link `json:"self"`
		Transaction Link `json:"transaction"`
		Effects     Link `json:"effects"`
		Succeeds    Link `json:"succeeds"`
		Precedes    Link `json:"precedes"`
	} `json:"_links"`

	ID              string    `json:"id"`
	PT              string    `json:"paging_token"`
	SourceAccount   string    `json:"source_account"`
	Type            string    `json:"type"`
	TypeI           int32     `json:"type_i"`
	LedgerClosedAt  time.Time `json:"created_at"`
	TransactionHash string    `json:"transaction_hash"`
	Order           int32     `json:"order"`
}

// OperationType returns the type of the operation.
func (o *BaseOperation) OperationType() string { return o.Type }

// CreateAccountOperation is the json resource representing a single operation
// whose type is CreateAccount.
type CreateAccountOperation struct {
	BaseOperation
	StartingBalance string `json:"starting_balance"`
	Funder          string `json:"funder"`
	Account         string `json:"account"`
}

// PaymentOperation is the json resource representing a single operation whose
// type is Payment.
type PaymentOperation struct {
	BaseOperation
	AssetType   string `json:"asset_type"`
	AssetCode   string `json:"asset_code,omitempty"`
	AssetIssuer string `json:"asset_issuer,omitempty"`
	From        string `json:"from"`
	To          string `json:"to"`
	Amount      string `json:"amount"`
}

// PathPaymentOperation is the json resource representing a single operation whose type
// is PathPayment.
type PathPaymentOperation struct {
	PaymentOperation
	Path              []Asset `json:"path"`
	SourceMax         string  `json:"source_max"`
	SourceAssetType   string  `json:"source_asset_type"`
	SourceAssetCode   string  `json:"source_asset_code,omitempty"`
	SourceAssetIssuer string  `json:"source_asset_issuer,omitempty"`
}

// ManageDataOperation represents a ManageData operation as it is serialized into json
// for the horizon API.
type ManageDataOperation struct {
	BaseOperation
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CreatePassiveOfferOperation is the json resource representing a single operation whose
// type is CreatePassiveOffer.
type CreatePassiveOfferOperation struct {
	BaseOperation
	Amount             string `json:"amount"`
	Price              string `json:"price"`
	PriceR             Price  `json:"price_r"`
	BuyingAssetType    string `json:"buying_asset_type"`
	BuyingAssetCode    string `json:"buying_asset_code,omitempty"`
	BuyingAssetIssuer  string `json:"buying_asset_issuer,omitempty"`
	SellingAssetType   string `json:"selling_asset_type"`
	SellingAssetCode   string `json:"selling_asset_code,omitempty"`
	SellingAssetIssuer string `json:"selling_asset_issuer,omitempty"`
}

// ManageOfferOperation is the json resource representing a single operation whose type
// is ManageOffer.
type ManageOfferOperation struct {
	CreatePassiveOfferOperation
	OfferID int64 `json:"offer_id"`
}

// SetOptionsOperation is the json resource representing a single operation whose type is
// SetOptions.
type SetOptionsOperation struct {
	BaseOperation
	HomeDomain    string `json:"home_domain,omitempty"`
	InflationDest string `json:"inflation_dest,omitempty"`

	MasterKeyWeight *int   `json:"master_key_weight,omitempty"`
	SignerKey       string `json:"signer_key,omitempty"`
	SignerWeight    *int   `json:"signer_weight,omitempty"`

	SetFlags    []int    `json:"set_flags,omitempty"`
	SetFlagsS   []string `json:"set_flags_s,omitempty"`
	ClearFlags  []int    `json:"clear_flags,omitempty"`
	ClearFlagsS []string `json:"clear_flags_s,omitempty"`

	LowThreshold  *int `json:"low_threshold,omitempty"`
	MedThreshold  *int `json:"med_threshold,omitempty"`
	HighThreshold *int `json:"high_threshold,omitempty"`
}

// ChangeTrustOperation is the json resource representing a single operation whose type
// is ChangeTrust.
type ChangeTrustOperation struct {
	BaseOperation
	AssetType   string `json:"asset_type"`
	AssetCode   string `json:"asset_code,omitempty"`
	AssetIssuer string `json:"asset_issuer,omitempty"`
	Limit       string `json:"limit"`
	Trustee     string `json:"trustee"`
	Trustor     string `json:"trustor"`
}

// AllowTrustOperation is the json resource representing a single operation whose type is
// AllowTrust.
type AllowTrustOperation struct {
	BaseOperation
	AssetType   string `json:"asset_type"`
	AssetCode   string `json:"asset_code,omitempty"`
	AssetIssuer string `json:"asset_issuer,omitempty"`
	Trustee     string `json:"trustee"`
	Trustor     string `json:"trustor"`
	Authorize   bool   `json:"authorize"`
}

// AccountMergeOperation is the json resource representing a single operation whose type
// is AccountMerge.
type AccountMergeOperation struct {
	BaseOperation
	Account string `json:"account"`
	Into    string `json:"into"`
}

// InflationOperation is the json resource representing a single operation whose type is
// Inflation.
type InflationOperation struct {
	BaseOperation
}

package horizon

import (
	"encoding/json"

	"github.com/stellar/go/support/errors"
)

// Effect represents an effect from an applied operation.
type Effect interface {
	EffectType() string
}

// UnmarshalEffect deserializes JSON into an Effect.
func UnmarshalEffect(data []byte) (Effect, error) {
	var effectType struct {
		Type string `json:"type"`
	}

	err := json.Unmarshal(data, &effectType)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshaling effect")
	}

	var effect Effect
	switch effectType.Type {
	case "account_created":
		e := &AccountCreatedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "account_removed":
		e := &AccountRemovedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "account_credited":
		e := &AccountCreditedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "account_debited":
		e := &AccountDebitedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "account_thresholds_updated":
		e := &AccountCreatedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "account_home_domain_updated":
		e := &AccountHomeDomainUpdatedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "account_flags_updated":
		e := &AccountFlagsUpdatedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "account_inflation_destination_updated":
		e := &AccountInflationDestinationUpdatedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "signer_created":
		e := &SignerCreatedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "signer_removed":
		e := &SignerRemovedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "signer_updated":
		e := &SignerUpdatedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "trustline_created":
		e := &TrustlineCreatedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "trustline_removed":
		e := &TrustlineRemovedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "trustline_updated":
		e := &TrustlineUpdatedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "trustline_authorized":
		e := &TrustlineAuthorizedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "trustline_deauthorized":
		e := &TrustlineDeauthorizedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "offer_created":
		e := &OfferCreatedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "offer_removed":
		e := &OfferRemovedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "offer_updated":
		e := &OfferUpdatedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "trade":
		e := &TradeEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "data_created":
		e := &DataCreatedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "data_removed":
		e := &DataRemovedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	case "data_updated":
		e := &DataUpdatedEffect{}
		err = json.Unmarshal(data, e)
		effect = e
	default:
		return nil, errors.Errorf("unknown effect type %d", effectType.Type)
	}

	return effect, errors.Wrap(err, "error unmarshaling effect")
}

type BaseEffect struct {
	LinksEffect struct {
		Operation Link `json:"operation"`
		Succeeds  Link `json:"succeeds"`
		Precedes  Link `json:"precedes"`
	} `json:"_links"`

	ID          string `json:"id"`
	PT          string `json:"paging_token"`
	Account     string `json:"account"`
	Type        string `json:"type"`
	TypeI       int32  `json:"type_i"`
	OperationID string `json:"operation_id"`
	Order       int32  `json:"order"`
}

// EffectType returns the type of the effect.
func (e *BaseEffect) EffectType() string { return e.Type }

type AccountCreatedEffect struct {
	BaseEffect
	StartingBalance string `json:"starting_balance"`
}

type AccountRemovedEffect struct {
	BaseEffect
}

type AccountCreditedEffect struct {
	BaseEffect
	AssetType   string `json:"asset_type"`
	AssetCode   string `json:"asset_code,omitempty"`
	AssetIssuer string `json:"asset_issuer,omitempty"`
	Amount      string `json:"amount"`
}

type AccountDebitedEffect struct {
	BaseEffect
	AssetType   string `json:"asset_type"`
	AssetCode   string `json:"asset_code,omitempty"`
	AssetIssuer string `json:"asset_issuer,omitempty"`
	Amount      string `json:"amount"`
}

type AccountThresholdsUpdatedEffect struct {
	BaseEffect
	LowThreshold  int32 `json:"low_threshold"`
	MedThreshold  int32 `json:"med_threshold"`
	HighThreshold int32 `json:"high_threshold"`
}

type AccountHomeDomainUpdatedEffect struct {
	BaseEffect
	HomeDomain string `json:"home_domain"`
}

type AccountFlagsUpdatedEffect struct {
	BaseEffect
	AuthRequired  *bool `json:"auth_required_flag,omitempty"`
	AuthRevokable *bool `json:"auth_revokable_flag,omitempty"`
}

type AccountInflationDestinationUpdatedEffect struct {
	BaseEffect
}

type SignerCreatedEffect struct {
	BaseEffect
	Weight    int32  `json:"weight"`
	PublicKey string `json:"public_key"`
	Key       string `json:"key"`
}

type SignerRemovedEffect struct {
	BaseEffect
	Weight    int32  `json:"weight"`
	PublicKey string `json:"public_key"`
	Key       string `json:"key"`
}

type SignerUpdatedEffect struct {
	BaseEffect
	Weight    int32  `json:"weight"`
	PublicKey string `json:"public_key"`
	Key       string `json:"key"`
}

type TrustlineCreatedEffect struct {
	BaseEffect
	AssetType   string `json:"asset_type"`
	AssetCode   string `json:"asset_code,omitempty"`
	AssetIssuer string `json:"asset_issuer,omitempty"`
	Limit       string `json:"limit"`
}

type TrustlineRemovedEffect struct {
	BaseEffect
	AssetType   string `json:"asset_type"`
	AssetCode   string `json:"asset_code,omitempty"`
	AssetIssuer string `json:"asset_issuer,omitempty"`
	Limit       string `json:"limit"`
}

type TrustlineUpdatedEffect struct {
	BaseEffect
	AssetType   string `json:"asset_type"`
	AssetCode   string `json:"asset_code,omitempty"`
	AssetIssuer string `json:"asset_issuer,omitempty"`
	Limit       string `json:"limit"`
}

type TrustlineAuthorizedEffect struct {
	BaseEffect
	Trustor   string `json:"trustor"`
	AssetType string `json:"asset_type"`
	AssetCode string `json:"asset_code,omitempty"`
}

type TrustlineDeauthorizedEffect struct {
	BaseEffect
	Trustor   string `json:"trustor"`
	AssetType string `json:"asset_type"`
	AssetCode string `json:"asset_code,omitempty"`
}

type OfferCreatedEffect struct {
	BaseEffect
}

type OfferRemovedEffect struct {
	BaseEffect
}

type OfferUpdatedEffect struct {
	BaseEffect
}

type TradeEffect struct {
	BaseEffect
	Seller            string `json:"seller"`
	OfferID           int64  `json:"offer_id"`
	SoldAmount        string `json:"sold_amount"`
	SoldAssetType     string `json:"sold_asset_type"`
	SoldAssetCode     string `json:"sold_asset_code,omitempty"`
	SoldAssetIssuer   string `json:"sold_asset_issuer,omitempty"`
	BoughtAmount      string `json:"bought_amount"`
	BoughtAssetType   string `json:"bought_asset_type"`
	BoughtAssetCode   string `json:"bought_asset_code,omitempty"`
	BoughtAssetIssuer string `json:"bought_asset_issuer,omitempty"`
}

type DataCreatedEffect struct {
	BaseEffect
}

type DataRemovedEffect struct {
	BaseEffect
}

type DataUpdatedEffect struct {
	BaseEffect
}

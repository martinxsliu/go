package resource

import (
	"fmt"
	"strings"
	"time"

	"github.com/guregu/null"
	"golang.org/x/net/context"

	"github.com/stellar/go/services/horizon/internal/db2/history"
	"github.com/stellar/go/services/horizon/internal/httpx"
	"github.com/stellar/go/services/horizon/internal/render/hal"
)

// Populate fills out the details
func (t *Transaction) Populate(
	ctx context.Context,
	row history.Transaction,
) (err error) {
	t.ID = row.TransactionHash
	t.PT = row.PagingToken()
	t.Hash = row.TransactionHash
	t.Ledger = row.LedgerSequence
	t.LedgerCloseTime = row.LedgerCloseTime
	t.Account = row.Account
	t.AccountSequence = row.AccountSequence
	t.FeePaid = row.FeePaid
	t.OperationCount = row.OperationCount
	t.EnvelopeXdr = row.TxEnvelope
	t.ResultXdr = row.TxResult
	t.ResultMetaXdr = row.TxMeta
	t.FeeMetaXdr = row.TxFeeMeta
	t.MemoType = row.MemoType
	t.Memo = row.Memo.String
	t.Signatures = strings.Split(row.SignatureString, ",")
	t.ValidBefore = t.timeString(row.ValidBefore)
	t.ValidAfter = t.timeString(row.ValidAfter)
	t.Order = row.ApplicationOrder

	lb := hal.LinkBuilder{Base: httpx.BaseURL(ctx)}
	t.Links.Account = lb.Link("/accounts", t.Account)
	t.Links.Ledger = lb.Link("/ledgers", fmt.Sprintf("%d", t.Ledger))
	t.Links.Operations = lb.PagedLink("/transactions", t.ID, "operations")
	t.Links.Effects = lb.PagedLink("/transactions", t.ID, "effects")
	t.Links.Self = lb.Link("/transactions", t.ID)
	t.Links.Succeeds = lb.Linkf("/transactions?order=desc&cursor=%s", t.PT)
	t.Links.Precedes = lb.Linkf("/transactions?order=asc&cursor=%s", t.PT)
	return
}

// PagingToken implementation for hal.Pageable
func (t Transaction) PagingToken() string {
	return t.PT
}

func (t *Transaction) timeString(in null.Int) string {
	if !in.Valid {
		return ""
	}

	return time.Unix(in.Int64, 0).UTC().Format(time.RFC3339)
}

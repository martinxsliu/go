package operations

import (
	"fmt"
	"strconv"

	"github.com/stellar/go/services/horizon/internal/db2/history"
	"github.com/stellar/go/services/horizon/internal/httpx"
	"github.com/stellar/go/services/horizon/internal/render/hal"
	"golang.org/x/net/context"
)

// PagingToken implements hal.Pageable
func (b Base) PagingToken() string {
	return b.PT
}

// Populate fills out b resource using `row` as the source.
func (b *Base) Populate(
	ctx context.Context,
	row history.Operation,
	ledger history.Ledger,
) {
	b.ID = strconv.FormatInt(row.ID, 10)
	b.PT = row.PagingToken()
	b.SourceAccount = row.SourceAccount
	b.populateType(row)
	b.LedgerCloseTime = ledger.ClosedAt
	b.TransactionHash = row.TransactionHash
	b.Order = row.ApplicationOrder

	lb := hal.LinkBuilder{Base: httpx.BaseURL(ctx)}
	self := fmt.Sprintf("/operations/%d", row.ID)
	b.Links.Self = lb.Link(self)
	b.Links.Succeeds = lb.Linkf("/effects?order=desc&cursor=%s", b.PT)
	b.Links.Precedes = lb.Linkf("/effects?order=asc&cursor=%s", b.PT)
	b.Links.Transaction = lb.Linkf("/transactions/%s", row.TransactionHash)
	b.Links.Effects = lb.Link(self, "effects")
}

func (b *Base) populateType(row history.Operation) {
	var ok bool
	b.TypeI = int32(row.Type)
	b.Type, ok = TypeNames[row.Type]

	if !ok {
		b.Type = "unknown"
	}
}

package effects

import (
	"strconv"

	"golang.org/x/net/context"

	"github.com/stellar/go/services/horizon/internal/db2/history"
	"github.com/stellar/go/services/horizon/internal/httpx"
	"github.com/stellar/go/services/horizon/internal/render/hal"
)

// PagingToken implements `hal.Pageable`
func (b Base) PagingToken() string {
	return b.PT
}

// Populate loads b resource from `row`
func (b *Base) Populate(ctx context.Context, row history.Effect) {
	b.ID = row.ID()
	b.PT = row.PagingToken()
	b.Account = row.Account
	b.populateType(row)
	b.OperationID = strconv.FormatInt(row.HistoryOperationID, 10)
	b.Order = row.Order

	lb := hal.LinkBuilder{Base: httpx.BaseURL(ctx)}
	b.Links.Operation = lb.Linkf("/operations/%d", row.HistoryOperationID)
	b.Links.Succeeds = lb.Linkf("/effects?order=desc&cursor=%s", b.PT)
	b.Links.Precedes = lb.Linkf("/effects?order=asc&cursor=%s", b.PT)
}

func (b *Base) populateType(row history.Effect) {
	var ok bool
	b.TypeI = int32(row.Type)
	b.Type, ok = TypeNames[row.Type]

	if !ok {
		b.Type = "unknown"
	}
}

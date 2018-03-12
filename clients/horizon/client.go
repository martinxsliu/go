package horizon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/xdr"
	"golang.org/x/net/context"
)

// HomeDomainForAccount returns the home domain for the provided strkey-encoded
// account id.
func (c *Client) HomeDomainForAccount(aid string) (string, error) {
	a, err := c.LoadAccount(aid)
	if err != nil {
		return "", errors.Wrap(err, "load account failed")
	}
	return a.HomeDomain, nil
}

// fixURL removes trailing slash from Client.URL. This will prevent situation when
// http.Client does not follow redirects.
func (c *Client) fixURL() {
	c.URL = strings.TrimRight(c.URL, "/")
}

// Root loads the root endpoint of horizon
func (c *Client) Root() (root Root, err error) {
	c.fixURLOnce.Do(c.fixURL)
	resp, err := c.HTTP.Get(c.URL)
	if err != nil {
		return
	}

	err = decodeResponse(resp, &root)
	return
}

// LoadAccount loads the account state from horizon. err can be either error
// object or horizon.Error object.
func (c *Client) LoadAccount(accountID string) (account Account, err error) {
	c.fixURLOnce.Do(c.fixURL)
	resp, err := c.HTTP.Get(c.URL + "/accounts/" + accountID)
	if err != nil {
		return
	}

	err = decodeResponse(resp, &account)
	return
}

// LoadAccountOffers loads the account offers from horizon. err can be either
// error object or horizon.Error object.
func (c *Client) LoadAccountOffers(accountID string, params ...interface{}) (offers OffersPage, err error) {
	c.fixURLOnce.Do(c.fixURL)
	endpoint := ""
	query := url.Values{}

	for _, param := range params {
		switch param := param.(type) {
		case At:
			endpoint = string(param)
		case Limit:
			query.Add("limit", strconv.Itoa(int(param)))
		case Order:
			query.Add("order", string(param))
		case Cursor:
			query.Add("cursor", string(param))
		default:
			err = fmt.Errorf("Undefined parameter (%T): %+v", param, param)
			return
		}
	}

	if endpoint == "" {
		endpoint = fmt.Sprintf(
			"%s/accounts/%s/offers?%s",
			c.URL,
			accountID,
			query.Encode(),
		)
	}

	// ensure our endpoint is a real url
	_, err = url.Parse(endpoint)
	if err != nil {
		err = errors.Wrap(err, "failed to parse endpoint")
		return
	}

	resp, err := c.HTTP.Get(endpoint)
	if err != nil {
		err = errors.Wrap(err, "failed to load endpoint")
		return
	}

	err = decodeResponse(resp, &offers)
	return
}

// LoadMemo loads memo for a transaction in Payment
func (c *Client) LoadMemo(p *Payment) (err error) {
	res, err := c.HTTP.Get(p.Links.Transaction.Href)
	if err != nil {
		return errors.Wrap(err, "load transaction failed")
	}
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(&p.Memo)
}

// SequenceForAccount implements build.SequenceProvider
func (c *Client) SequenceForAccount(
	accountID string,
) (xdr.SequenceNumber, error) {

	a, err := c.LoadAccount(accountID)
	if err != nil {
		return 0, errors.Wrap(err, "load account failed")
	}

	seq, err := strconv.ParseUint(a.Sequence, 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "parse sequence failed")
	}

	return xdr.SequenceNumber(seq), nil
}

// LoadOrderBook loads order book for given selling and buying assets.
func (c *Client) LoadOrderBook(selling Asset, buying Asset, params ...interface{}) (orderBook OrderBookSummary, err error) {
	c.fixURLOnce.Do(c.fixURL)
	query := url.Values{}

	query.Add("selling_asset_type", selling.Type)
	query.Add("selling_asset_code", selling.Code)
	query.Add("selling_asset_issuer", selling.Issuer)

	query.Add("buying_asset_type", buying.Type)
	query.Add("buying_asset_code", buying.Code)
	query.Add("buying_asset_issuer", buying.Issuer)

	for _, param := range params {
		switch param := param.(type) {
		case Limit:
			query.Add("limit", strconv.Itoa(int(param)))
		default:
			err = fmt.Errorf("Undefined parameter (%T): %+v", param, param)
			return
		}
	}

	resp, err := c.HTTP.Get(c.URL + "/order_book?" + query.Encode())
	if err != nil {
		return
	}

	err = decodeResponse(resp, &orderBook)
	return
}

func (c *Client) stream(ctx context.Context, baseURL string, cursor *Cursor, handler func(data []byte) error) error {
	query := url.Values{}
	if cursor != nil {
		query.Set("cursor", string(*cursor))
	}

	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", baseURL, query.Encode()), nil)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "text/event-stream")

		resp, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		scanner.Split(splitSSE)

		var objectBytes []byte

		for scanner.Scan() {
			// Check if ctx is not cancelled
			select {
			case <-ctx.Done():
				return nil
			default:
				// Continue streaming
			}

			if len(scanner.Bytes()) == 0 {
				continue
			}

			ev, err := parseEvent(scanner.Bytes())
			if err != nil {
				return err
			}

			if ev.Event != "message" {
				continue
			}

			switch data := ev.Data.(type) {
			case string:
				err = handler([]byte(data))
				objectBytes = []byte(data)
			case []byte:
				err = handler(data)
				objectBytes = data
			default:
				err = errors.New("Invalid ev.Data type")
			}
			if err != nil {
				return err
			}
		}

		err = scanner.Err()

		// Start streaming from the next object:
		// - if there was no error OR
		// - if connection was lost
		if err == nil || err == io.ErrUnexpectedEOF {
			object := struct {
				PT string `json:"paging_token"`
			}{}

			err := json.Unmarshal(objectBytes, &object)
			if err != nil {
				return errors.Wrap(err, "error unmarshaling objectBytes")
			}

			if object.PT != "" {
				query.Set("cursor", object.PT)
			} else {
				return errors.New("no paging_token in object: cannot continue")
			}

			continue
		}

		if err != nil {
			return err
		}
	}
}

// StreamLedgers streams incoming ledgers. Use context.WithCancel to stop streaming or
// context.Background() if you want to stream indefinitely.
func (c *Client) StreamLedgers(ctx context.Context, cursor *Cursor, handler LedgerHandler) error {
	c.fixURLOnce.Do(c.fixURL)
	url := fmt.Sprintf("%s/ledgers", c.URL)
	return c.stream(ctx, url, cursor, func(data []byte) error {
		var l Ledger
		if err := json.Unmarshal(data, &l); err != nil {
			return errors.Wrap(err, "error unmarshaling data")
		}
		handler(l)
		return nil
	})
}

// StreamPayments streams incoming payments. Use context.WithCancel to stop streaming or
// context.Background() if you want to stream indefinitely.
func (c *Client) StreamPayments(ctx context.Context, accountID string, cursor *Cursor, handler PaymentHandler) error {
	c.fixURLOnce.Do(c.fixURL)
	url := fmt.Sprintf("%s/accounts/%s/payments", c.URL, accountID)
	return c.stream(ctx, url, cursor, func(data []byte) error {
		var p Payment
		if err := json.Unmarshal(data, &p); err != nil {
			return errors.Wrap(err, "error unmarshaling data")
		}
		handler(p)
		return nil
	})
}

// StreamAllTransactions streams all incoming transactions. Use context.WithCancel()
// to stop streaming or context.Background() if you want to stream indefinitely.
func (c *Client) StreamAllTransactions(ctx context.Context, cursor *Cursor, handler TransactionHandler) error {
	c.fixURLOnce.Do(c.fixURL)
	url := fmt.Sprintf("%s/transactions", c.URL)
	return c.stream(ctx, url, cursor, func(data []byte) error {
		var t Transaction
		if err := json.Unmarshal(data, &t); err != nil {
			return errors.Wrap(err, "error unmarshaling data")
		}
		handler(t)
		return nil
	})
}

// StreamTransactions streams incoming transactions for a given account. Use
// context.WithCancel() to stop streaming or context.Background() if you want
// to stream indefinitely.
func (c *Client) StreamTransactions(ctx context.Context, accountID string, cursor *Cursor, handler TransactionHandler) error {
	c.fixURLOnce.Do(c.fixURL)
	url := fmt.Sprintf("%s/accounts/%s/transactions", c.URL, accountID)
	return c.stream(ctx, url, cursor, func(data []byte) error {
		var t Transaction
		if err := json.Unmarshal(data, &t); err != nil {
			return errors.Wrap(err, "error unmarshaling data")
		}
		handler(t)
		return nil
	})
}

// StreamAllOperations streams all incoming operations. Use context.WithCancel()
// to stop streaming or context.Background() if you want to stream indefinitely.
func (c *Client) StreamAllOperations(ctx context.Context, cursor *Cursor, handler OperationHandler) error {
	c.fixURLOnce.Do(c.fixURL)
	url := fmt.Sprintf("%s/operations", c.URL)
	return c.stream(ctx, url, cursor, func(data []byte) error {
		op, err := UnmarshalOperation(data)
		if err != nil {
			return errors.Wrap(err, "error unmarshaling data")
		}
		handler(op)
		return nil
	})
}

// StreamAllEffects streams all incoming effects. Use context.WithCancel()
// to stop streaming or context.Background() if you want to stream indefinitely.
func (c *Client) StreamAllEffects(ctx context.Context, cursor *Cursor, handler EffectHandler) error {
	c.fixURLOnce.Do(c.fixURL)
	url := fmt.Sprintf("%s/effects", c.URL)
	return c.stream(ctx, url, cursor, func(data []byte) error {
		effect, err := UnmarshalEffect(data)
		if err != nil {
			return errors.Wrap(err, "error unmarshaling data")
		}
		handler(effect)
		return nil
	})
}

// SubmitTransaction submits a transaction to the network. err can be either error object or horizon.Error object.
func (c *Client) SubmitTransaction(transactionEnvelopeXdr string) (response TransactionSuccess, err error) {
	c.fixURLOnce.Do(c.fixURL)
	v := url.Values{}
	v.Set("tx", transactionEnvelopeXdr)

	resp, err := c.HTTP.PostForm(c.URL+"/transactions", v)
	if err != nil {
		err = errors.Wrap(err, "http post failed")
		return
	}

	err = decodeResponse(resp, &response)
	if err != nil {
		return
	}

	// WARNING! Do not remove this code. If you include two trailing slashes (`//`) at the end of Client.URL
	// and developers changed Client.HTTP to not follow redirects, this will return empty response and no error!
	if resp.StatusCode != http.StatusOK {
		err = errors.New("Invalid response code")
		return
	}

	return
}

package api

import "github.com/shopspring/decimal"

// init configures package-global behaviour we want at every entry point
// (both cmd/server and cmd/lambda import this package transitively).
//
// `MarshalJSONWithoutQuotes = true` makes shopspring/decimal serialize as
// JSON numbers (e.g. 1500.5) instead of JSON strings ("1500.5"). The Angular
// frontend's `parseBalance` handles both, but raw numbers avoid the brittle
// pt-BR string parsing path on the client.
func init() {
	decimal.MarshalJSONWithoutQuotes = true
}

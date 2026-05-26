package service

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
)

// flowAV is a small DSL so call sites stay readable when building items:
// `{"PK": {S: "..."}, "balance": {N: "100"}, "isSystem": {B: true}}`.
// Exactly one field must be set.
type flowAV struct {
	S string
	N string
	B bool
	// BSet distinguishes "B is meaningful" from the zero value.
	BSet bool
}

// buildItem converts the DSL map into the SDK's AttributeValue map.
func buildItem(in map[string]flowAV) map[string]types.AttributeValue {
	out := make(map[string]types.AttributeValue, len(in))
	for k, v := range in {
		switch {
		case v.BSet:
			out[k] = flowdynamo.Bool(v.B)
		case v.N != "":
			out[k] = flowdynamo.N(v.N)
		default:
			out[k] = flowdynamo.S(v.S)
		}
	}
	return out
}

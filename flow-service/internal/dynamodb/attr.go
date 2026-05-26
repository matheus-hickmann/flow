package dynamodb

import (
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// S wraps a string as an AttributeValue (DynamoDB type S).
func S(v string) types.AttributeValue {
	return &types.AttributeValueMemberS{Value: v}
}

// N wraps a numeric string as an AttributeValue (DynamoDB type N).
func N(v string) types.AttributeValue {
	return &types.AttributeValueMemberN{Value: v}
}

// NInt is N for ints.
func NInt(v int) types.AttributeValue {
	return &types.AttributeValueMemberN{Value: strconv.Itoa(v)}
}

// Bool wraps a bool as an AttributeValue (DynamoDB type BOOL).
func Bool(v bool) types.AttributeValue {
	return &types.AttributeValueMemberBOOL{Value: v}
}

// Str reads a string attribute, returning empty when absent.
func Str(item map[string]types.AttributeValue, key string) string {
	if v, ok := item[key].(*types.AttributeValueMemberS); ok {
		return v.Value
	}
	return ""
}

// Num reads a numeric attribute, returning empty when absent.
func Num(item map[string]types.AttributeValue, key string) string {
	if v, ok := item[key].(*types.AttributeValueMemberN); ok {
		return v.Value
	}
	return ""
}

// ReadBool reads a bool attribute, returning false when absent.
func ReadBool(item map[string]types.AttributeValue, key string) bool {
	if v, ok := item[key].(*types.AttributeValueMemberBOOL); ok {
		return v.Value
	}
	return false
}

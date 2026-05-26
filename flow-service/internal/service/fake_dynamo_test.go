package service

import (
	"context"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// fakeDynamo is a hand-rolled stub for the dynamodb.API interface. Each method
// delegates to an optional field; nil fields return zero values. Tests fill in
// only the methods they care about.
type fakeDynamo struct {
	GetItemFunc            func(ctx context.Context, in *awsdynamodb.GetItemInput, opts ...func(*awsdynamodb.Options)) (*awsdynamodb.GetItemOutput, error)
	PutItemFunc            func(ctx context.Context, in *awsdynamodb.PutItemInput, opts ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error)
	UpdateItemFunc         func(ctx context.Context, in *awsdynamodb.UpdateItemInput, opts ...func(*awsdynamodb.Options)) (*awsdynamodb.UpdateItemOutput, error)
	DeleteItemFunc         func(ctx context.Context, in *awsdynamodb.DeleteItemInput, opts ...func(*awsdynamodb.Options)) (*awsdynamodb.DeleteItemOutput, error)
	QueryFunc              func(ctx context.Context, in *awsdynamodb.QueryInput, opts ...func(*awsdynamodb.Options)) (*awsdynamodb.QueryOutput, error)
	TransactWriteItemsFunc func(ctx context.Context, in *awsdynamodb.TransactWriteItemsInput, opts ...func(*awsdynamodb.Options)) (*awsdynamodb.TransactWriteItemsOutput, error)
}

func (f *fakeDynamo) GetItem(ctx context.Context, in *awsdynamodb.GetItemInput, opts ...func(*awsdynamodb.Options)) (*awsdynamodb.GetItemOutput, error) {
	if f.GetItemFunc != nil {
		return f.GetItemFunc(ctx, in, opts...)
	}
	return &awsdynamodb.GetItemOutput{}, nil
}

func (f *fakeDynamo) PutItem(ctx context.Context, in *awsdynamodb.PutItemInput, opts ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error) {
	if f.PutItemFunc != nil {
		return f.PutItemFunc(ctx, in, opts...)
	}
	return &awsdynamodb.PutItemOutput{}, nil
}

func (f *fakeDynamo) UpdateItem(ctx context.Context, in *awsdynamodb.UpdateItemInput, opts ...func(*awsdynamodb.Options)) (*awsdynamodb.UpdateItemOutput, error) {
	if f.UpdateItemFunc != nil {
		return f.UpdateItemFunc(ctx, in, opts...)
	}
	return &awsdynamodb.UpdateItemOutput{}, nil
}

func (f *fakeDynamo) DeleteItem(ctx context.Context, in *awsdynamodb.DeleteItemInput, opts ...func(*awsdynamodb.Options)) (*awsdynamodb.DeleteItemOutput, error) {
	if f.DeleteItemFunc != nil {
		return f.DeleteItemFunc(ctx, in, opts...)
	}
	return &awsdynamodb.DeleteItemOutput{}, nil
}

func (f *fakeDynamo) Query(ctx context.Context, in *awsdynamodb.QueryInput, opts ...func(*awsdynamodb.Options)) (*awsdynamodb.QueryOutput, error) {
	if f.QueryFunc != nil {
		return f.QueryFunc(ctx, in, opts...)
	}
	return &awsdynamodb.QueryOutput{}, nil
}

func (f *fakeDynamo) TransactWriteItems(ctx context.Context, in *awsdynamodb.TransactWriteItemsInput, opts ...func(*awsdynamodb.Options)) (*awsdynamodb.TransactWriteItemsOutput, error) {
	if f.TransactWriteItemsFunc != nil {
		return f.TransactWriteItemsFunc(ctx, in, opts...)
	}
	return &awsdynamodb.TransactWriteItemsOutput{}, nil
}

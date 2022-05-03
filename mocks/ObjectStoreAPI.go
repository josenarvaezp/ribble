// Code generated by mockery v2.10.4. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	s3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

// ObjectStoreAPI is an autogenerated mock type for the ObjectStoreAPI type
type ObjectStoreAPI struct {
	mock.Mock
}

// CreateBucket provides a mock function with given fields: ctx, params, optFns
func (_m *ObjectStoreAPI) CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
	_va := make([]interface{}, len(optFns))
	for _i := range optFns {
		_va[_i] = optFns[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *s3.CreateBucketOutput
	if rf, ok := ret.Get(0).(func(context.Context, *s3.CreateBucketInput, ...func(*s3.Options)) *s3.CreateBucketOutput); ok {
		r0 = rf(ctx, params, optFns...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*s3.CreateBucketOutput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *s3.CreateBucketInput, ...func(*s3.Options)) error); ok {
		r1 = rf(ctx, params, optFns...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetObject provides a mock function with given fields: ctx, params, optFns
func (_m *ObjectStoreAPI) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	_va := make([]interface{}, len(optFns))
	for _i := range optFns {
		_va[_i] = optFns[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *s3.GetObjectOutput
	if rf, ok := ret.Get(0).(func(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) *s3.GetObjectOutput); ok {
		r0 = rf(ctx, params, optFns...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*s3.GetObjectOutput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) error); ok {
		r1 = rf(ctx, params, optFns...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListObjectsV2 provides a mock function with given fields: ctx, params, optFns
func (_m *ObjectStoreAPI) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	_va := make([]interface{}, len(optFns))
	for _i := range optFns {
		_va[_i] = optFns[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *s3.ListObjectsV2Output
	if rf, ok := ret.Get(0).(func(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) *s3.ListObjectsV2Output); ok {
		r0 = rf(ctx, params, optFns...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*s3.ListObjectsV2Output)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) error); ok {
		r1 = rf(ctx, params, optFns...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
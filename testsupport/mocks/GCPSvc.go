// Code generated by mockery v2.4.0-beta. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// GCPSvc is an autogenerated mock type for the GCPSvc type
type GCPSvc struct {
	mock.Mock
}

// CreateBucket provides a mock function with given fields: ctx, name
func (_m *GCPSvc) CreateBucket(ctx context.Context, name string) error {
	ret := _m.Called(ctx, name)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Code generated by mockery v2.1.0. DO NOT EDIT.

package mocks

import (
	net "net"

	mock "github.com/stretchr/testify/mock"

	ovn "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/ovn"
)

// AddressSetFactory is an autogenerated mock type for the AddressSetFactory type
type AddressSetFactory struct {
	mock.Mock
}

// DestroyAddressSetInBackingStore provides a mock function with given fields: name
func (_m *AddressSetFactory) DestroyAddressSetInBackingStore(name string) error {
	ret := _m.Called(name)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ForEachAddressSet provides a mock function with given fields: iteratorFn
func (_m *AddressSetFactory) ForEachAddressSet(iteratorFn ovn.AddressSetIterFunc) error {
	ret := _m.Called(iteratorFn)

	var r0 error
	if rf, ok := ret.Get(0).(func(ovn.AddressSetIterFunc) error); ok {
		r0 = rf(iteratorFn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewAddressSet provides a mock function with given fields: name, ips
func (_m *AddressSetFactory) NewAddressSet(name string, ips []net.IP) (ovn.AddressSet, error) {
	ret := _m.Called(name, ips)

	var r0 ovn.AddressSet
	if rf, ok := ret.Get(0).(func(string, []net.IP) ovn.AddressSet); ok {
		r0 = rf(name, ips)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(ovn.AddressSet)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, []net.IP) error); ok {
		r1 = rf(name, ips)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
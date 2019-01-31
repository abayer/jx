// Code generated by pegomock. DO NOT EDIT.
package matchers

import (
	"reflect"
	"github.com/petergtz/pegomock"
	versioned3 "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
)

func AnyVersioned3Interface() versioned3.Interface {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(versioned3.Interface))(nil)).Elem()))
	var nullValue versioned3.Interface
	return nullValue
}

func EqVersioned3Interface(value versioned3.Interface) versioned3.Interface {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue versioned3.Interface
	return nullValue
}

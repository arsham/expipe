package datatype_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDatatype(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Datatype Suite")
}

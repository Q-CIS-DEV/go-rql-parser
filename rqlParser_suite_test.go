package rqlParser_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRqlParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RqlParser Suite")
}

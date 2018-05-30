package GoRqlParser__test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/url"
	"strings"
	"rqlParser"
	"fmt"
)

var _ = Describe("GoRqlParser", func() {

	It("Can parse expressions containing encoded special characters", func() {
		Describe("having filter containing special characters", func() {
			dateString := "2018-05-29T15:29:58.627755+05:00"
			filter := fmt.Sprintf(
				"in(date,(%s))",
				url.QueryEscape(dateString))

			parser := rqlParser.NewParser()
			rqlNode, err := parser.Parse(strings.NewReader(filter))
			Expect(err).To(BeNil())
			Expect(rqlNode.Node.Op).To(BeEquivalentTo("in"))
			Expect(rqlNode.Node.Args[1].(string)).To(BeEquivalentTo(dateString))
		})
	})
})

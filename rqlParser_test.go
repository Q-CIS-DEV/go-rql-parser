package rqlParser_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rqlParser "go-rql-parser"

	"fmt"
	"net/url"
)

var _ = Describe("GoRqlParser", func() {

	It("Can parse plain expressiond", func() {
		dateString := "2018-05-29T15:29:58.627755Z"
		parser := rqlParser.NewParser()
		rqlNode, err := parser.Parse(fmt.Sprintf("in(date,%s)", dateString))

		Expect(err).To(BeNil())
		Expect(rqlNode.Node.Op).To(BeEquivalentTo("in"))
		Expect(rqlNode.Node.Args[1].(string)).To(BeEquivalentTo(dateString))
	})

	It("Can parse expressions containing encoded special characters", func() {
		Describe("filter value containing special characters", func() {
			dateString := "2018-05-29T15:29:58.627755Z"
			filter := fmt.Sprintf(
				"in(date,(%s))",
				url.QueryEscape(dateString))

			parser := rqlParser.NewParser()
			rqlNode, err := parser.Parse(filter)
			Expect(err).To(BeNil())
			Expect(rqlNode.Node.Op).To(BeEquivalentTo("in"))
			Expect(rqlNode.Node.Args[1].(string)).To(BeEquivalentTo(dateString))
		})

		Describe("Full expression encoded", func() {
			dateString := "2018-05-29T15:29:58.627755Z"
			filter := url.QueryEscape(fmt.Sprintf("in(date,%s)", dateString))
			parser := rqlParser.NewParser()
			rqlNode, err := parser.Parse(filter)

			Expect(err).To(BeNil())
			Expect(rqlNode.Node.Op).To(BeEquivalentTo("in"))
			Expect(rqlNode.Node.Args[1].(string)).To(BeEquivalentTo(dateString))
		})

		Describe("Sort expression encoded", func() {
			filter := url.QueryEscape("sort(date)")
			parser := rqlParser.NewParser()
			rqlNode, err := parser.Parse(filter)

			Expect(err).To(BeNil())
			res := rqlNode.Sort()[0]
			Expect(res.By).To(BeEquivalentTo("date"))
			Expect(res.Desc).To(BeFalse())
		})
	})
})

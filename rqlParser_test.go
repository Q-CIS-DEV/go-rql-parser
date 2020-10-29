package rqlParser_test

import (
	rqlParser "github.com/Q-CIS-DEV/go-rql-parser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"net/url"
)

var _ = Describe("GoRqlParser", func() {

	It("Must parse whitespace", func() {
		testString := "Some cool name"
		parser := rqlParser.NewParser()
		rqlNode, err := parser.Parse(fmt.Sprintf("eq(name,%s)", testString))

		Expect(err).To(BeNil())
		Expect(rqlNode.Node.Op).To(BeEquivalentTo("eq"))
		Expect(rqlNode.Node.Args[1].(string)).To(BeEquivalentTo(testString))
	})

	It("Must parse utf8", func() {
		testString := "Русский текст"
		parser := rqlParser.NewParser()
		rqlNode, err := parser.Parse(fmt.Sprintf("like(name,*%s*)", testString))
		Expect(err).To(BeNil())
		Expect(rqlNode.Node.Op).To(BeEquivalentTo("like"))
		Expect(rqlNode.Node.Args[1].(string)).To(BeEquivalentTo(fmt.Sprintf("*%s*", testString)))
	})

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
			filter := url.QueryEscape("sort(+date)")
			parser := rqlParser.NewParser()
			rqlNode, err := parser.Parse(filter)

			Expect(err).To(BeNil())
			res := rqlNode.Sort()[0]
			Expect(res.By).To(BeEquivalentTo("date"))
			Expect(res.Desc).To(BeFalse())
		})
	})
	It("Can parse special symbol inside string", func() {
		sheldingString := "H\\&M"
		valueStirng := "H&M"
		parser := rqlParser.NewParser()
		rqlNode, err := parser.Parse(fmt.Sprintf("like(name,*%s*)", sheldingString))

		Expect(err).To(BeNil())
		Expect(rqlNode.Node.Op).To(BeEquivalentTo("like"))
		Expect(rqlNode.Node.Args[1].(string)).To(BeEquivalentTo(fmt.Sprintf("*%s*", valueStirng)))
	})
	It("Can parse special symbol in the end of the string", func() {
		sheldingString := "H\\&"
		valueStirng := "H&"
		parser := rqlParser.NewParser()
		rqlNode, err := parser.Parse(fmt.Sprintf("like(name,*%s)", sheldingString))

		Expect(err).To(BeNil())
		Expect(rqlNode.Node.Op).To(BeEquivalentTo("like"))
		Expect(rqlNode.Node.Args[1].(string)).To(BeEquivalentTo(fmt.Sprintf("*%s", valueStirng)))
	})
	It("Can escape backslash", func() {
		sheldingString := "9\\\\4"
		valueStirng := "9\\4"
		parser := rqlParser.NewParser()
		rqlNode, err := parser.Parse(fmt.Sprintf("like(name,*%s*)", sheldingString))

		Expect(err).To(BeNil())
		Expect(rqlNode.Node.Op).To(BeEquivalentTo("like"))
		Expect(rqlNode.Node.Args[1].(string)).To(BeEquivalentTo(fmt.Sprintf("*%s*", valueStirng)))
	})
})

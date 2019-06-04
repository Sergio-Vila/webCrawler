package htmlscanner

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
	"webCrawler/crawler"
)

type docscanTest struct {
	desc string	// A short description of the test case.
	html string	// The HTML to crawl.
	expectedTitle string // Expected title
	expectedLinks []string // Expected links
}

type docscanTestSuite struct {
	desc string
	tests []docscanTest
}

var titleTests = docscanTestSuite{"Tests for titles", []docscanTest{
	{
		"valid title",
		"<head><title>This is my title</title></head>",
		"This is my title",
		nil,
	},
	{
		"valid title with special chars",
		"<head><title>This is my title with some special chars %^&#</title></head>",
		"This is my title with some special chars %^&#",
		nil,
	},
	{
		"title inside a comment",
		"<head><!-- <title>This is not the title</title> --></head>",
		"",
		nil,
	},
	{
		"should stop scanning after end of head",
		"<head></head><title>This is not the title</title>",
		"",
		nil,
	},
	{
		"valid title in multiline multiline",
		`<head>
					<title>This is my title</title>
				</head>`,
		"This is my title",
		nil,
	},
	{
		"empty title",
		"<head><title></title></head>",
		"",
		nil,
	},
}}

var linksTests = docscanTestSuite{"Tests for links", []docscanTest{
	{
		"valid links",
		`
			<head></head>
			<body>
				<a href="/resource1">Link to resource 1</a>
				<div>
					<a href="/resource2?query#pos">Link to resource 2</a>
				</div>
			</body>`,
		"",
		[]string{"/resource1", "/resource2?query#pos"},
	},
	{
		"href in different tag than anchor",
		`<head></head>
               <body><div href="/resource1"></div></body>`,
		"",
		nil,
	},
	{
		"link after body",
		`<head></head>
               <body></body>
               <a href="/resource1></a>"`,
		"",
		nil,
	},
	{
		"link inside comment",
		`<head></head>
               <body><!-- <a href="/resource1">Link to resource 1</a> --></body>`,
		"",
		nil,
	},
}}

var linksAndTitle = docscanTestSuite{"Tests for links and title", []docscanTest{
	{
		"empty",
		"",
		"",
		nil,
	},
	{
		"not html",
		"lorem ipsu",
		"",
		nil,
	},
	{
		"valid links and title",
		`
			<head>
				<title>This is my title</title>
			</head>
			<body>
				<a href="/resource1">Link to resource 1</a>
				<div>
					<a href="/resource2?query#pos">Link to resource 2</a>
				</div>
			</body>`,
		"This is my title",
		[]string{"/resource1", "/resource2?query#pos"},
	},
	{
		"not valid title, valid links",
		`
			<head>
			</head>
			<title>This is my title</title>

			<body>
				<a href="/resource1">Link to resource 1</a>
				<div>
					<a href="/resource2?query#pos">Link to resource 2</a>
				</div>
			</body>`,
		"",
		[]string{"/resource1", "/resource2?query#pos"},
	},
}}

func TestHtmlScanner_Scan(t *testing.T) {
	assert := assert.New(t)

	testSuites := []docscanTestSuite{titleTests, linksTests, linksAndTitle}

	numTestsRan := 0
	numTestSuitesRan := 0

	scanner := New()

	for _, testSuite := range testSuites {

		for i, test := range testSuite.tests {

			docId := crawler.Id(fmt.Sprintf("DOC_ID_%d", i))

			scanOutputCh := make(chan crawler.Message)

			docReader := crawler.DocReader{docId, ioutil.NopCloser(strings.NewReader(test.html)) }

			go scanner.Scan(docReader, scanOutputCh)

			var actualLinks []string
			actualTitle := ""

		loopOverMessages:
			for {
				msg := <- scanOutputCh

				assert.Equal(docId, msg.DocId,
					"Test '%s' of test suite '%s' failed. Expected DocId '%s' but got '%s'",
					test.desc, testSuite.desc, docId, msg.DocId)

				switch msg.Type {
					case crawler.Title:
						actualTitle = msg.Content

					case crawler.Link:
						actualLinks = append(actualLinks, msg.Content)

					case crawler.EndOfStream:
						break loopOverMessages
				}
			}

			assert.Equal(actualTitle, test.expectedTitle,
				"Test '%s' of test suite '%s' failed. Expected title '%s' but got '%s'",
				test.desc, testSuite.desc, test.expectedTitle, actualTitle)

			assert.True(reflect.DeepEqual(test.expectedLinks, actualLinks),
				"Test %s of test suite '%s' failed. Expected to find links '%v' but found '%v'",
				test.desc, testSuite.desc, test.expectedLinks, actualLinks)

			close(scanOutputCh)

			numTestsRan++
		}

		numTestSuitesRan++
	}

	fmt.Printf("Ran %d tests from %d test suites\n", numTestsRan, numTestSuitesRan)
}
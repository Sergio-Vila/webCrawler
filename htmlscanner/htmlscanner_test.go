package htmlscanner

import (
    "bytes"
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
        "multiple valid links",
        `
            <head></head>
            <body>
                <a href="/r1">r1</a>
                <a href="/r2">r2</a>
                <a href="/r3">r3</a>
                <a href="/r4">r4</a>
                <a href="/r5">r5</a>
                <a href="/r6">r6</a>
                <a href="/r7">r7</a>
                <a href="/r8">r8</a>
                <a href="/r9">r9</a>
                <a href="/r10">r10</a>
                <a href="/r11">r11</a>
                <a href="/r12">r12</a>
                <a href="/r13">r13</a>
                <a href="/r14">r14</a>
                <a href="/r15">r15</a>
                <a href="/r16">r16</a>
                <a href="/r17">r17</a>
                <a href="/r18">r18</a>
                <a href="/r19">r19</a>
                <a href="/r20">r20</a>
                <a href="/r21">r21</a>
                <a href="/r22">r22</a>
                <a href="/r23">r23</a>
                <a href="/r24">r24</a>
                <a href="/r25">r25</a>
                <a href="/r26">r26</a>
                <a href="/r27">r27</a>
                <a href="/r28">r28</a>
                <a href="/r29">r29</a>
                <a href="/r30">r30</a>
                <a href="/r31">r31</a>
                <a href="/r32">r32</a>
                <a href="/r33">r33</a>
                <a href="/r34">r34</a>
                <a href="/r35">r35</a>
                <a href="/r36">r36</a>
                <a href="/r37">r37</a>
            </body>`,
        "",
        []string{"/r1", "/r2", "/r3", "/r4", "/r5", "/r6", "/r7", "/r8", "/r9", "/r10",
            "/r11", "/r12", "/r13", "/r14", "/r15", "/r16", "/r17", "/r18", "/r19", "/r20",
            "/r21", "/r22", "/r23", "/r24", "/r25", "/r26", "/r27", "/r28", "/r29", "/r30",
            "/r31", "/r32", "/r33", "/r34", "/r35", "/r36", "/r37"},
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

            docId := crawler.DocId(fmt.Sprintf("DOC_ID_%d", i))

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
                        actualTitle = msg.Content[0]

                    case crawler.Link:
                        actualLinks = append(actualLinks, msg.Content...)

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


func benchmarkHtmlScanner_Scan(fileName string, b *testing.B) {

    var numLinks = 0
    var title string

    scanner := New()

    scanOutputCh := make(chan crawler.Message)

    fileStr, err := ioutil.ReadFile(fileName)
    if err != nil {
        b.Fatalf("Error opening file '%s': %s", fileName, err.Error());
    }

    for n := 0; n < b.N; n++ {

        b.StopTimer()

        numLinks = 0

        docReader := crawler.DocReader {
            DocId: crawler.DocId(fileName),
            Reader: ioutil.NopCloser(bytes.NewReader(fileStr)),
        }

        b.StartTimer()

        go scanner.Scan(docReader, scanOutputCh)

    loopOverMessages:
        for {
            msg := <-scanOutputCh

            switch msg.Type {
            case crawler.Title:
                title = msg.Content[0]

            case crawler.Link:
                numLinks = numLinks + len(msg.Content)

            case crawler.EndOfStream:
                break loopOverMessages
            }
        }

    }

    close(scanOutputCh)

    _ = title
    _ = numLinks
}

func BenchmarkHtmlScanner_Scan_GoReleaseNotes(b *testing.B) {
    benchmarkHtmlScanner_Scan("./testdata/go1.html", b)
}

// The Spring Championship Of Online Poker Wikipedia page is
// currently the longest page of Wikipedia
func BenchmarkHtmlScanner_Scan_WikipediaScoop(b *testing.B) {
    benchmarkHtmlScanner_Scan("./testdata/wikipediascoop.html", b)
}
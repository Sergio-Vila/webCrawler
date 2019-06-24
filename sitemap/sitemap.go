package sitemap

import (
    "errors"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
    "webCrawler/crawler"
    "webCrawler/htmlscanner"
    "webCrawler/threadpool"
)

const numberOfConcurrentConnections = 6
const docOutputChSize = 1024

type SiteMap struct {
    crawler crawler.Crawler
    docs map [crawler.DocId] *crawler.DocInfo
    root crawler.DocId
}

func NewSiteMap() *SiteMap {

    pool, err := threadpool.NewFixed(numberOfConcurrentConnections)
    if err != nil {
        return nil
    }

    return &SiteMap{
        crawler.New(
            htmlscanner.New(),
            crawler.RequesterFunc(addHttpRequest),
            crawler.ResolverFunc(idFromLocator),
            pool,
        ),
        make(map [crawler.DocId] *crawler.DocInfo),
        "",
    }
}

func addHttpRequest(requestedUrlStr crawler.DocId) (io.ReadCloser, error) {
    var emptyReadCloser io.ReadCloser

    requestedUrl, err := url.Parse(string(requestedUrlStr))
    if err != nil {
        return emptyReadCloser, err
    }

    if !requestedUrl.IsAbs() {
        return emptyReadCloser, errors.New("URL to request should be absolute")
    }

    resp, err := http.Get(requestedUrl.String())
    if err != nil {
        return emptyReadCloser, err
    }

    return resp.Body, nil
}

func (sm *SiteMap) ProduceFrom(startingPoint string) error {

    docInfoCh := make(chan crawler.DocInfo, docOutputChSize)

    startingPointUrl, err := url.ParseRequestURI(startingPoint)
    if err != nil {
        return errors.New("Starting point URL " + startingPoint + " is not valid")
    }
    sm.root = idFromAbsUrl(startingPointUrl)

    go sm.crawler.Crawl(sm.root, docInfoCh)

loopOverCompletedPages:
    for {
        select {
            case completedPage, isOpen := <- docInfoCh:
                if !isOpen {
                    break loopOverCompletedPages
                }

                sm.docs[completedPage.DocId] = &completedPage
        }
    }

    return nil
}

func (sm *SiteMap) print(
    fromPageId crawler.DocId,
    visited map [crawler.DocId] bool,
    spacing string,
    level int) {

    page := sm.docs[fromPageId]

    visited[page.DocId] = true

    fmt.Printf("%s- %s\n", spacing, strings.Trim(page.Title, " \t\n"))

    for _, link := range page.Links {
        if _, wasVisited := visited[link]; !wasVisited {
            sm.print(link, visited, spacing+" ", level+1)
        } else {
            fmt.Printf("%s* %s\n", spacing, strings.Trim(sm.docs[link].Title, " \t\n"))
        }
    }
}

func (sm *SiteMap) Print() {
    fmt.Printf("SITE MAP\n" +
        " A line starting with a - character indicates the links of the page are just below it.\n" +
        " A line starting with a * character indicates the page is already in the output and links\n" +
        " won't be printed again.\n\n\n")
    sm.print(sm.root, make(map[crawler.DocId]bool), " ", 0)
}
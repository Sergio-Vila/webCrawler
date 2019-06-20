package sitemap

import (
    "net/url"
    "webCrawler/crawler"
)

func idFromAbsUrl(url *url.URL) crawler.Id {

    // Remove the elements of the URL that could be different for
    // the same page
    url.RawQuery = ""
    url.Fragment = ""
    url.User = nil

    // Do not show '?' character on empty queries
    url.ForceQuery = false

    // Remove trailing '/' characters on URL
    for last := len(url.Path)-1; last >= 0 && url.Path[last] == '/'; last = len(url.Path)-1 {
        url.Path = url.Path[:len(url.Path)-1]
    }

    return crawler.Id(url.String())
}

func idFromLocator(locator crawler.Loc, from crawler.Id) (id crawler.Id, idIsValid bool) {
    parentUrl, parentUrlError := url.ParseRequestURI(string(from))
    locatorAbsUrl, locatorUrlError := parentUrl.Parse(string(locator))

    if parentUrlError != nil || locatorUrlError != nil {
        return "", false
    }

    // Ignore different domains or same domain from a different protocol
    if locatorAbsUrl.Hostname() != parentUrl.Hostname() ||
        locatorAbsUrl.Scheme != parentUrl.Scheme {
        return "", false
    }

    return idFromAbsUrl(locatorAbsUrl), true
}
package htmlscanner

import (
    "go.uber.org/zap"
    "golang.org/x/net/html"
    "io/ioutil"
    "webCrawler/crawler"
)

const linksPerMsg = 20

type HtmlScanner struct {}

func New() crawler.DocScanner {
    return &HtmlScanner{}
}

func (*HtmlScanner) Scan(r crawler.DocReader, outCh chan crawler.Message) {
    tokenizer := html.NewTokenizer(r.Reader)
    logger, _ := zap.NewProduction()

    findTitle(tokenizer, r.DocId, outCh, logger)
    findLinks(tokenizer, r.DocId, outCh, logger)

    outCh <- crawler.EndOfStreamMsg(r.DocId)

    logger.Debug("End Of Stream", zap.String("Doc", string(r.DocId)))
    logger.Sync()

    // Read until end of file and close
    _, _ = ioutil.ReadAll(r.Reader)
    r.Reader.Close()
}

func findTitle(token *html.Tokenizer, docId crawler.Id, outCh chan crawler.Message, logger *zap.Logger) {

loopOverTokens:
    for {
        switch token.Next() {
            case html.StartTagToken:
                if tagNameEquals("title", token) {
                    if tokenType := token.Next(); tokenType == html.TextToken {
                        title := string(token.Text())

                        logger.Debug("Found title",
                            zap.String("Doc", string(docId)),
                            zap.String("Title", title))

                        outCh <- crawler.Message{
                            Content: []string{title},
                            DocId: docId,
                            Type: crawler.Title,
                        }

                        break loopOverTokens
                    }
                }

            case html.ErrorToken:
                break loopOverTokens

            case html.EndTagToken:
                if tagNameEquals("head", token) {
                    break loopOverTokens
                }
        }
    }

}

func findLinks(token *html.Tokenizer, docId crawler.Id, linksCh chan crawler.Message, logger *zap.Logger) {

    unsentLinks := make([]string, 0, linksPerMsg)

loopOverTokens:
    for {
        switch token.Next() {
            case html.StartTagToken:
                if tagName, hasAttributes := token.TagName();
                    areEqual(tagName,"a") && hasAttributes {

                loopOverAttributes:
                    for hasMoreAttr := true; hasMoreAttr; {
                        var key, val []byte
                        key, val, hasMoreAttr = token.TagAttr()

                        if areEqual(key, "href") {
                            link := string(val)

                            logger.Debug("Found link",
                                zap.String("Doc", string(docId)),
                                zap.String("Link", link))

                            if len(unsentLinks) < linksPerMsg {
                                unsentLinks = append(unsentLinks, link)
                            }

                            if len(unsentLinks) == linksPerMsg {
                                linksCh <- crawler.Message{
                                    Content: unsentLinks,
                                    DocId:   docId,
                                    Type:    crawler.Link,
                                }

                                unsentLinks = make([]string, 0, linksPerMsg)
                            }

                            break loopOverAttributes
                        }
                    }
                }

            case html.ErrorToken:
                break loopOverTokens

            case html.EndTagToken:
                if tagNameEquals("body", token) {
                    break loopOverTokens
                }
        }

    }

    if len(unsentLinks) != 0 {
        linksCh <- crawler.Message{
            Content: unsentLinks,
            DocId:   docId,
            Type:    crawler.Link,
        }
    }
}

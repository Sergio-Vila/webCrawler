package htmlscanner

import (
    "go.uber.org/zap"
    "golang.org/x/net/html"
    "io/ioutil"
    "webCrawler/crawler"
)

const linksPerMsg = 20

type HtmlScanner struct {}

func New() crawler.Scanner {
    return &HtmlScanner{}
}

func (*HtmlScanner) Scan(r crawler.DocReader, outCh chan crawler.Message) {
    tokenizer := html.NewTokenizer(r.Reader)
    logger, _ := zap.NewProduction(
        zap.Fields(zap.String("DocId", string(r.DocId))),
    )

    findTitle(tokenizer, r.DocId, outCh, logger)
    findLinks(tokenizer, r.DocId, outCh, logger)

    eos := crawler.EndOfStreamMsg(r.DocId)

    logger.Debug("Send EoS", zap.Object("Msg", eos))
    outCh <- eos

    logger.Sync()

    // Read until end of file and close
    _, _ = ioutil.ReadAll(r.Reader)
    r.Reader.Close()
}

func findTitle(token *html.Tokenizer, docId crawler.DocId, outCh chan crawler.Message, logger *zap.Logger) {

loopOverTokens:
    for {
        switch token.Next() {
            case html.StartTagToken:
                if tagNameEquals("title", token) {
                    if tokenType := token.Next(); tokenType == html.TextToken {
                        title := string(token.Text())

                        logger.Debug("Found title", zap.String("Title", title))

                        msg := crawler.Message{
                            Content: []string{title},
                            DocId: docId,
                            Type: crawler.Title,
                        }

                        logger.Debug("Send title", zap.Object("Msg", msg))

                        outCh <- msg

                        break loopOverTokens
                    }
                }

            case html.ErrorToken:
                break loopOverTokens

            case html.EndTagToken:
                if tagNameEquals("head", token) {
                    logger.Debug("Reached </head>")
                    break loopOverTokens
                }
        }
    }

}

func findLinks(token *html.Tokenizer, docId crawler.DocId, linksCh chan crawler.Message, logger *zap.Logger) {

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

                            logger.Debug("Found link", zap.String("Link", link))

                            if len(unsentLinks) < linksPerMsg {
                                unsentLinks = append(unsentLinks, link)
                            }

                            if len(unsentLinks) == linksPerMsg {
                                msg := crawler.Message{
                                    Content: unsentLinks,
                                    DocId:   docId,
                                    Type:    crawler.Link,
                                }

                                logger.Debug("Send links", zap.Object("Msg", msg))
                                linksCh <- msg

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
                    logger.Debug("Reached </body>")
                    break loopOverTokens
                }
        }

    }

    if len(unsentLinks) != 0 {

        msg := crawler.Message{
            Content: unsentLinks,
            DocId:   docId,
            Type:    crawler.Link,
        }

        logger.Debug("Send links", zap.Object("Msg", msg))

        linksCh <- msg
    }
}

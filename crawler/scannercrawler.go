package crawler

import (
    "go.uber.org/zap"
    "webCrawler/threadpool"
)

const docRequestsBufferSize = 1024

type ScannerCrawler struct {
    pool       threadpool.Pool
    docScanner Scanner
    requester  Requester
    resolver   Resolver
    crawled    map [DocId] *DocInfo
    logger     *zap.Logger
}

func New(
    docScanner Scanner,
    requester Requester,
    resolver Resolver,
    pool threadpool.Pool) Crawler {

    logger, _ := zap.NewProduction()

    return &ScannerCrawler{
        pool,
        docScanner,
        requester,
        resolver,
        make(map [DocId] *DocInfo),
        logger,
    }
}

func (c ScannerCrawler) Crawl(
    startId DocId,
    outCh chan DocInfo) {

    scanResCh := make(chan Message, docRequestsBufferSize)
    docIdsCh := make(chan DocId, docRequestsBufferSize)

    c.logger.Info("Crawl started", zap.String("Root page", string(startId)))

    go c.produceDocs(docIdsCh, scanResCh)

    docIdsCh <- DocId(startId)

    c.consumeDocs(scanResCh, outCh, docIdsCh)

    c.logger.Debug("Crawl stopping", zap.String("Root page", string(startId)))

    close(outCh)
    close(docIdsCh)
    close(scanResCh)
    c.pool.Stop()

    c.logger.Info("Crawl stopped", zap.String("Root page", string(startId)))
}

func (c ScannerCrawler) produceDocs(
    docIdsInCh chan DocId,
    scanResCh chan Message) {

loopOverNewDocIds:
    for {
        nextDocId, docsChIsOpen := <- docIdsInCh
        if !docsChIsOpen {
            break loopOverNewDocIds
        }

        c.pool.Run("Scanner for " + string(nextDocId), func () {
            c.logger.Debug("Running task for document scan",
                zap.String("DocId", string(nextDocId)))

            nextDoc, err := c.requester.Request(nextDocId)
            if err != nil {
                c.logger.Error("Error while requesting doc",
                    zap.String("DocId", string(nextDocId)),
                    zap.Error(err))
            } else {
                nextDocReader := DocReader{
                    DocId: nextDocId,
                    Reader: nextDoc,
                }

                c.docScanner.Scan(nextDocReader, scanResCh)
            }

            c.logger.Info("Finished task for document scan",
                zap.String("DocId", string(nextDocId)))

            c.logger.Sync()
        })
    }
}

func (c ScannerCrawler) consumeDocs(
    scanResInCh chan Message,
    outCh chan DocInfo,
    docIdsOutCh chan DocId) {

    pendingDocs := 1

loopOverDocScannerMessages:
    for {
        msg := <- scanResInCh

        doc, exists := c.crawled[msg.DocId]
        if !exists {
            c.crawled[msg.DocId] = DefaultDocInfo(msg.DocId)
            doc = c.crawled[msg.DocId]
        }

        switch msg.Type {
            case Title:
                doc.Title = msg.Content[0]
                c.logger.Debug("Got title from Scanner",
                    zap.String("DocId", string(doc.DocId)),
                    zap.String("Title", doc.Title))

            case Link:
            loopOverLinks:
                for _, link := range msg.Content {
                    linkedId, linkHasId := c.resolver.Resolve(Loc(link), msg.DocId)
                    if !linkHasId {
                        c.logger.Debug("Got link - Ignored by 'idFromLoc' function",
                            zap.String("DocId", string(doc.DocId)),
                            zap.String("Link location", string(link)))
                        continue loopOverLinks
                    }

                    doc.Links = append(doc.Links, linkedId)

                    if _, alreadyScanned := c.crawled[linkedId]; alreadyScanned {
                        c.logger.Debug("Got link - Already scanned",
                            zap.String("DocId", string(doc.DocId)),
                            zap.String("Link location", string(link)),
                            zap.String("Linked DocId", string(linkedId)))

                        continue loopOverLinks
                    }

                    c.crawled[linkedId] = DefaultDocInfo(linkedId)
                    docIdsOutCh <- linkedId

                    c.logger.Debug("Got link - Requested",
                        zap.String("DocId", string(doc.DocId)),
                        zap.String("Link location", string(link)))

                    pendingDocs++
                }

            case EndOfStream:
                doc.completed = true
                outCh <- *c.crawled[msg.DocId]

                c.logger.Sync()

                pendingDocs--
                if pendingDocs == 0 {
                    break loopOverDocScannerMessages
                }
        }
    }
}

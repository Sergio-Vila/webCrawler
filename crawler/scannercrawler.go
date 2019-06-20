package crawler

import (
    "go.uber.org/zap"
    "webCrawler/threadpool"
)

const docRequestsBufferSize = 1024

type ScannerCrawler struct {
    pool threadpool.Pool
    docScanner  DocScanner
    crawled     map [Id] *DocInfo
    logger		*zap.Logger
}

func New(docScanner DocScanner, pool threadpool.Pool) Crawler {
    logger, _ := zap.NewProduction()

    return &ScannerCrawler{pool, docScanner, make(map [Id] *DocInfo), logger}
}

func (c ScannerCrawler) Crawl(
    startId string,
    outCh chan DocInfo,
    getDocReader func(docId Id) (DocReader, error),
    idFromLoc func (locator Loc, fromId Id) (id Id, hasId bool)) {

    scanResCh := make(chan Message, docRequestsBufferSize)
    docIdsCh := make(chan Id, docRequestsBufferSize)

    c.logger.Info("Crawl started", zap.String("Root page", startId))

    go c.produceDocs(docIdsCh, scanResCh, getDocReader)

    docIdsCh <- Id(startId)

    c.consumeDocs(scanResCh, outCh, idFromLoc, docIdsCh)

    c.logger.Debug("Crawl stopping", zap.String("Root page", startId))

    close(outCh)
    close(docIdsCh)
    close(scanResCh)
    c.pool.Stop()

    c.logger.Info("Crawl stopped", zap.String("Root page", startId))
}

func (c ScannerCrawler) produceDocs(
    docIdsInCh chan Id,
    scanResCh chan Message,
    getDocReader func(docId Id) (DocReader, error)) {

loopOverNewDocIds:
    for {
        nextDocId, docsChIsOpen := <- docIdsInCh
        if !docsChIsOpen {
            break loopOverNewDocIds
        }

        c.pool.Run("DocScanner for " + string(nextDocId), func () {
            c.logger.Debug("Running task for document scan",
                zap.String("DocId", string(nextDocId)))

            nextDoc, err := getDocReader(nextDocId)
            if err != nil {
                c.logger.Error("Error on getting DocReader",
                    zap.String("DocId", string(nextDocId)),
                    zap.Error(err))
            } else {
                c.docScanner.Scan(nextDoc, scanResCh)
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
    idFromLoc func (locator Loc, fromId Id) (id Id, hasId bool),
    docIdsOutCh chan Id) {

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
            doc.Title = msg.Content
            c.logger.Debug("Got title from DocScanner",
                zap.String("DocId", string(doc.DocId)),
                zap.String("Title", doc.Title))

        case Link:
            linkedId, linkHasId := idFromLoc(Loc(msg.Content), msg.DocId)
            if !linkHasId {
                c.logger.Debug("Got link - Ignored by 'idFromLoc' function",
                    zap.String("DocId", string(doc.DocId)),
                    zap.String("Link location", msg.Content))
                continue loopOverDocScannerMessages
            }

            doc.Links = append(doc.Links, linkedId)

            if _, alreadyScanned := c.crawled[linkedId]; alreadyScanned {
                c.logger.Debug("Got link - Already scanned",
                    zap.String("DocId", string(doc.DocId)),
                    zap.String("Link location", msg.Content),
                    zap.String("Linked DocId", string(linkedId)))

                continue loopOverDocScannerMessages
            }

            c.crawled[linkedId] = DefaultDocInfo(linkedId)
            docIdsOutCh <- linkedId

            c.logger.Debug("Got link - Requested",
                zap.String("DocId", string(doc.DocId)),
                zap.String("Link location", msg.Content))

            pendingDocs++

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

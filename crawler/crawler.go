package crawler

type Loc string

type DocInfo struct {
    DocId     Id
    Title     string
    Links     []Id
    completed bool
}

func DefaultDocInfo(DocId Id) *DocInfo {
    return &DocInfo {
        DocId: DocId,
        Title: "Untitled document",
        Links: nil,
        completed: false,
    }
}

type Crawler interface {
    // Navigates through a website, starting from the document with 'startId'
    // and sending the found document information through 'outCh'.
    // 'getDocReader' function should have the logic to get a document from its id.
    // 'idFromLoc' function should have the logic to get a document id from the link value.
    Crawl(
        startId string,
        outCh chan DocInfo,
        getDocReader func(docId Id) (DocReader, error),
        idFromLoc func (locator Loc, fromId Id) (id Id, hasId bool))
}

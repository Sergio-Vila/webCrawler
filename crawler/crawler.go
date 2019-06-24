package crawler

type Loc string

type DocInfo struct {
    DocId     DocId
    Title     string
    Links     []DocId
    completed bool
}

func DefaultDocInfo(DocId DocId) *DocInfo {
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
        startId DocId,
        outCh chan DocInfo)
}

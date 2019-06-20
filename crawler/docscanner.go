package crawler

import "io"

type Id string

type MessageType int
const (
    Title MessageType = iota
    Link
    EndOfStream
)

type Message struct {
    Content string
    DocId   Id
    Type	MessageType
}

type DocReader struct {
    DocId 	Id
    Reader 	io.ReadCloser
}

type DocScanner interface {
    // Scans a document and looks for its title, and for
    // links to other documents.
    //
    // The title and links are sent via outCh channel,
    // and once the scan is done an 'EndOfStream' message with empty
    // content is sent.
    Scan(docReader DocReader, outCh chan Message)
}

func EndOfStreamMsg(docId Id) Message {
    return Message {
        Content: "",
        DocId: docId,
        Type: EndOfStream,
    }
}


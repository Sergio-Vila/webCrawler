package crawler

import (
    "fmt"
    "go.uber.org/zap/zapcore"
    "io"
)

type DocId string

type MessageType int
const (
    Title MessageType = iota
    Link
    EndOfStream
)

func (mt MessageType) String() string {
    switch (mt) {
    case Title:
        return "Title"
    case Link:
        return "Link"
    case EndOfStream:
        return "EndOfStream"
    default:
        return fmt.Sprintf("%d", int(mt))
    }
}

type Message struct {
    Content []string
    DocId   DocId
    Type    MessageType
}

type DocReader struct {
    DocId  DocId
    Reader io.ReadCloser
}

type Scanner interface {
    // Scans a document and looks for its title, and for
    // links to other documents.
    //
    // The title and links are sent via outCh channel,
    // and once the scan is done an 'EndOfStream' message with empty
    // content is sent.
    Scan(docReader DocReader, outCh chan Message)
}

func EndOfStreamMsg(docId DocId) Message {
    return Message {
        Content: nil,
        DocId: docId,
        Type: EndOfStream,
    }
}

func (msg Message) MarshalLogObject(enc zapcore.ObjectEncoder) error {
    _ = enc.AddArray("Content", zapcore.ArrayMarshalerFunc(func(arr zapcore.ArrayEncoder) error {
        for i := range msg.Content {
            arr.AppendString(msg.Content[i])
        }
        return nil
    }))

    enc.AddString("DocId", string(msg.DocId))
    enc.AddString("Type", msg.Type.String())

    return nil
}
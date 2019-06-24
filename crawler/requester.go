package crawler

import "io"

type Requester interface {
	// Function called to request a reader to the document with 'docId'. This function
	// may be called simultaneously from multiple threads.
	Request(docId DocId) (io.ReadCloser, error)
}

type requesterFuncImpl struct {
	request func(docId DocId) (io.ReadCloser, error)
}

func (rfi requesterFuncImpl) Request(docId DocId) (io.ReadCloser, error) {
	return rfi.request(docId)
}

func RequesterFunc(request func(docId DocId) (io.ReadCloser, error)) Requester {
	return requesterFuncImpl{request}
}
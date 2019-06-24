package crawler

type Resolver interface {
	// Gets the id of the document refered to when using 'locator'
	// inside the document with id 'fromId'. This function may
	// be called simultaneously from multiple threads.
	Resolve(locator Loc, fromId DocId) (id DocId, hasId bool)
}

type ResolverFuncImpl struct {
	resolve func(locator Loc, fromId DocId) (id DocId, hasId bool)
}

func (rfi ResolverFuncImpl) Resolve(locator Loc, fromId DocId) (id DocId, hasId bool) {
	return rfi.resolve(locator, fromId)
}

func ResolverFunc(resolve func(locator Loc, fromId DocId) (id DocId, hasId bool)) Resolver {
	return ResolverFuncImpl{resolve}
}
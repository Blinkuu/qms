package alloclimit

type Strategy interface {
	Alloc(tokens int64) (remainingTokens int64, ok bool, err error)
	Free(tokens int64) (remainingTokens int64, ok bool, err error)
}

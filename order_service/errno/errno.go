package errno

import "errors"

var (
	ErrQueryFailed           = errors.New("query db failed")
	ErrUnderRepertory        = errors.New("lack of repertory")
	ErrQueryEmpty            = errors.New("query empty")
	ErrReduceRepertoryFailed = errors.New("reduce repertory failed for too many request")
)

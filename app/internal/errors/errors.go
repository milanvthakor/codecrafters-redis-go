package errors

import "fmt"

var (
	ErrXaddIdIsZero      = fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")
	ErrXaddIdIsEqOrSmall = fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	ErrNotANumericValue  = fmt.Errorf("ERR value is not an integer or out of range")
	ErrExecWoMulti       = fmt.Errorf("ERR EXEC without MULTI")
	ErrDiscardWoMulti    = fmt.Errorf("ERR DISCARD without MULTI")
	ErrInvalidCmd        = fmt.Errorf("invalid command")
)

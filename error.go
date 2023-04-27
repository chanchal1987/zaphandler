package zaphandler

import "strings"

type Error struct {
	strings.Builder
}

func (e *Error) Sync() error   { return nil }
func (e *Error) Error() string { return e.String() }

func (e *Error) Err() error {
	if e.Len() < 1 {
		return nil
	}

	return e
}

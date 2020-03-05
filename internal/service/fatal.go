package service

import "fmt"

type fatalError struct {
	err error
}

func (f fatalError) Error() string {
	return fmt.Sprintf("fatal error: %s", f.err)
}

func (f fatalError) Cause() error {
	return f.err
}

func Fatal(err error) error {
	if err == nil {
		return nil
	}

	return fatalError{err: err}
}

func IsFatal(err error) bool {
	type causer interface {
		Cause() error
	}

	for {
		if err == nil {
			return false
		}

		_, ok := err.(fatalError)
		if ok {
			return true
		}

		causeErr, ok := err.(causer)
		if !ok {
			return false
		}

		err = causeErr.Cause()
	}
}

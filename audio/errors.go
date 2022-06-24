package audio

import "errors"

var (
	ErrNameCannotBeOmitted = errors.New("name must not be omitted")
	ErrWriterNotFound      = errors.New("writer was not found in the index map")
)

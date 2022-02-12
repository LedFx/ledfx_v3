package audio

import "errors"

var (
	NameCannotBeOmitted = errors.New("name must not be omitted")
	WriterNotFound      = errors.New("writer was not found in the index map")
)

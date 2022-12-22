package storage

type NotFoundError struct {
	errorMessage string
}

func (n NotFoundError) Error() string {
	return n.errorMessage
}

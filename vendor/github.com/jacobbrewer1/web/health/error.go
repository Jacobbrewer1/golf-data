package health

type StatusError struct {
	error
	Status Status
}

func NewStatusError(err error, status Status) *StatusError {
	return &StatusError{
		error:  err,
		Status: status,
	}
}

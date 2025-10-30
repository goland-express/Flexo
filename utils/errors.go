package utils

type UserError struct {
	Message string
}

func (e *UserError) Error() string {
	return e.Message
}

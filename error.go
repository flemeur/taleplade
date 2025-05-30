package taleplade

// Error is used for constant errors.
type Error string

func (e Error) Error() string { return string(e) }

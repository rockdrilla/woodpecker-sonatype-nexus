package main

type ErrEmpty struct {
}

func (e *ErrEmpty) Error() string {
	return "empty"
}

type ErrMalformed struct {
}

func (e *ErrMalformed) Error() string {
	return "malformed"
}

type ErrMissing struct {
}

func (e *ErrMissing) Error() string {
	return "missing"
}

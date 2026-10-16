package errors

import "net/http"

type Error struct {
	Status int
	Text   string
}

func (e *Error) Error() string { return e.Text }

func NotFound() *Error {
	return &Error{Status: http.StatusNotFound, Text: "not found"}
}

func Unavailable() *Error {
	return &Error{Status: http.StatusServiceUnavailable, Text: "unavailable"}
}

func Internal() *Error {
	return &Error{Status: http.StatusInternalServerError, Text: "internal"}
}

func Write(w http.ResponseWriter, err *Error) {
	http.Error(w, err.Text, err.Status)
}

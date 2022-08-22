package error

import "github.com/pkg/errors"

//ClientError represents a problem with client request
type ClientError struct {
	Err error
}

//Error - returns description of the error
func (e *ClientError) Error() string {
	return e.Err.Error()
}

//NewClientError - creates an instance of ClientError
func NewClientError(description string) *ClientError {
	return &ClientError{
		Err: errors.New(description),
	}
}

//NewClientErrorWrap - creates an instance of ClientError, which wrapps around given error
func NewClientErrorWrap(err error, description string) *ClientError {
	return &ClientError{
		Err: errors.Wrap(err, description),
	}
}

//ItemNotFoundError - primary used for db queries, when a particular resource doesnt exist
type ItemNotFoundError struct {
	Err error
}

//Error - returns description of the error
func (e *ItemNotFoundError) Error() string {
	return e.Err.Error()
}

//NewItemNotFoundError - creates an instance of ItemNotFoundError
func NewItemNotFoundError(description string) *ItemNotFoundError {
	return &ItemNotFoundError{
		Err: errors.New(description),
	}
}

//ServerError represents a problem with server
type ServerError struct {
	Err error
}

//Error - returns description of the error
func (e *ServerError) Error() string {
	return e.Err.Error()
}

//NewServerError - creates an instance of ServerError
func NewServerError(description string) *ServerError {
	return &ServerError{
		Err: errors.New(description),
	}
}

//NewServerErrorWrap - creates an instance of ServerError, wrapping around a given error
func NewServerErrorWrap(err error, description string) *ServerError {
	return &ServerError{
		Err: errors.Wrapf(err, description),
	}
}

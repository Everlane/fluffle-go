package fluffle

type Error interface {
	Code() int
	Message() string
	Data() interface{}
}

type CustomError struct {
	code int
	message string
	data interface{}
}

func (err *CustomError) Code() int {
	return err.code
}

func (err *CustomError) Message() string {
	return err.message
}

func (err *CustomError) Error() string {
	return err.Message()
}

func (err *CustomError) Data() interface{} {
	return err.data
}

type MethodNotFoundError struct {
	message string
}

func (err *MethodNotFoundError) Code() int {
	return -32601
}

func (err *MethodNotFoundError) Message() string {
	if err.message == "" {
		return "Method not found"
	} else {
		return err.message
	}
}

func (err *MethodNotFoundError) Error() string {
	return err.Message()
}

type ParseError struct{}

func (err *ParseError) Code() int {
	return -32700
}

func (err *ParseError) Message() string {
	return "Parse error"
}

func (err *ParseError) Error() string {
	return err.Message()
}

func (err *ParseError) Data() interface{} {
	return nil
}

// Wraps a normal Go error to provide methods fulfilling the Error interface
// for use in crafting error objects in a response.
type WrappedError struct {
	error
}

func (err *WrappedError) Code() int {
	return 0
}

func (err *WrappedError) Message() string {
	return err.Error()
}

func (err *WrappedError) Data() interface{} {
	return nil
}

// Wrap a normal Go error in a WrappedError struct. Skips wrapping errors that
// are already fluffle-go error types.
func WrapError(err error) Error {
	switch err := err.(type) {
	case Error:
		return err
	default:
		return &WrappedError{err}
	}
}

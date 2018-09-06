package openapierr

const (
	// NotFound const defines the code value for openapi internal NotFound errors
	NotFound = "NotFound"
)

// Error defines the interface that OpenAPI internal errors must be compliant with
type Error interface {
	// Inherit from go error builtin interface
	error

	// Code that briefly describes the type of error
	Code() string
}

// NotFoundError represent a NotFound error and implements the openapi Error interface
type NotFoundError struct {
	OriginalError error
}

// Error returns a string containing the original error; or an empty string otherwise
func (e *NotFoundError) Error() string {
	if e.OriginalError != nil {
		return e.OriginalError.Error()
	}
	return ""
}

// Code returns the code that represents the NotFound error
func (e *NotFoundError) Code() string {
	return NotFound
}

package errs

import "fmt"

type ErrorType string

const (
	// Used in KeyNotFoundError
	NotFoundInDB  ErrorType = "NotFoundInDB"
	NotFoundInAVC ErrorType = "NotFoundInAVC"
	Deleted       ErrorType = "Deleted"

	// Used in SerializerError
	DeserializeFailed ErrorType = "DeserializeFailed"
	SerializeFailed   ErrorType = "SerializeFailed"

	// Used in ConnectorError
	ConnectFailed ErrorType = "ConnectFailed"
	KeyExists     ErrorType = "KeyExists"
)

type KeyNotFoundError struct {
	Key       string
	ErrorType ErrorType
}

func (e *KeyNotFoundError) Error() string {
	var msg string
	if e.ErrorType == NotFoundInDB {
		msg = fmt.Sprintf("Key[%s] not found in DB", e.Key)
	}
	if e.ErrorType == NotFoundInAVC {
		msg = fmt.Sprintf("Key[%s] not found in AVC", e.Key)
	}
	if e.ErrorType == Deleted {
		msg = fmt.Sprintf("Key[%s] is deleted", e.Key)
	}
	return msg
}

func NewKeyNotFoundError(key string, errType ErrorType) error {
	return &KeyNotFoundError{Key: key, ErrorType: errType}
}

type TypeAssertionError struct {
	VariableName string
	ExpectedType string
	ActualType   string
}

func (e *TypeAssertionError) Error() string {
	return fmt.Sprintf("TypeAssertionError: Expected[%s], Actual[%s], VariableName[%s]", e.ExpectedType, e.ActualType, e.VariableName)
}

func NewTypeAssertionError(variableName, expectedType, actualType string) error {
	return &TypeAssertionError{VariableName: variableName, ExpectedType: expectedType, ActualType: actualType}
}

type SerializerError struct {
	ErrorType ErrorType
	Err       error
}

func (e *SerializerError) Error() string {
	var msg string
	if e.ErrorType == DeserializeFailed {
		msg = fmt.Sprintf("SerializerError: Op[Deserialize]: %s", e.Err)
	}
	if e.ErrorType == SerializeFailed {
		msg = fmt.Sprintf("SerializerError: Op[Serialize]: %s", e.Err)
	}
	return msg
}

func (e *SerializerError) Unwrap() error {
	return e.Err
}

func NewSerializerError(errType ErrorType, err error) error {
	return &SerializerError{ErrorType: errType, Err: err}
}

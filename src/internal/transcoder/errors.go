package transcoder

import "fmt"

type TranscoderError struct {
	Type    ErrorType
	Message string
	Cause   error
}

type ErrorType string

const (
	ErrorTypeFFmpegNotFound  ErrorType = "ffmpeg_not_found"
	ErrorTypeGPUNotAvailable ErrorType = "gpu_not_available"
	ErrorTypeEncoderNotFound ErrorType = "encoder_not_found"
	ErrorTypeInvalidPreset   ErrorType = "invalid_preset"
	ErrorTypeInvalidFilePath ErrorType = "invalid_file_path"
	ErrorTypeEncodingFailed  ErrorType = "encoding_failed"
	ErrorTypeFileSystemError ErrorType = "file_system_error"
)

func (e *TranscoderError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *TranscoderError) Unwrap() error {
	return e.Cause
}

func NewTranscoderError(errorType ErrorType, message string, cause error) *TranscoderError {
	return &TranscoderError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}

func IsTranscoderError(err error, errorType ErrorType) bool {
	if te, ok := err.(*TranscoderError); ok {
		return te.Type == errorType
	}
	return false
}

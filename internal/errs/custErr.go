package errs

import "errors"

var (
	ErrInvalidInput        = errors.New("invalid input parameters")
	ErrUnsupportedLanguage = errors.New("unsupported programming language")
	ErrCodeTooLarge        = errors.New("code size exceeds maximum limit")
	ErrTaskNotFound        = errors.New("task not found")
	ErrSubmissionNotFound  = errors.New("submission not found")
	ErrUserNotFound        = errors.New("user not found")
	ErrNotFound            = errors.New("resource not found")
	ErrLanguageNotFound    = errors.New("language not found")
	ErrTestNotFound        = errors.New("test_case not found")
	ErrEmptySlice          = errors.New("empty slice")
	ErrUserDiscrepancy     = errors.New("You cannot get submission that was sent by another user")
	ErrWrongUserIDFormat   = errors.New("wrong user id format")
	ErrDataBase            = errors.New("db error")
	ErrUndefinedLanguage   = errors.New("Undefined language")
)

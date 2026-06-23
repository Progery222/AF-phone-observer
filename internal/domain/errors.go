package domain

import "errors"

var (
	ErrInvalidSerial      = errors.New("serial телефона не указан")
	ErrScreenshotFailed   = errors.New("не удалось сделать скриншот")
	ErrUIDumpFailed       = errors.New("не удалось получить UI dump")
	ErrStorageUnavailable = errors.New("хранилище недоступно")
)

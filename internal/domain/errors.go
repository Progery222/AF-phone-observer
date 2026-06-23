package domain

import "errors"

var (
	ErrInvalidSerial       = errors.New("serial телефона не указан")
	ErrScreenshotFailed    = errors.New("не удалось сделать скриншот")
	ErrUIDumpFailed        = errors.New("не удалось получить UI dump")
	ErrStorageUnavailable  = errors.New("хранилище недоступно")
	ErrInvalidElementQuery = errors.New("некорректный запрос поиска элемента")
	ErrElementNotFound     = errors.New("элемент не найден")
	ErrElementWaitTimeout  = errors.New("элемент не появился за заданное время")
	ErrInvalidDetectMode   = errors.New("неизвестный mode detect-state")
	ErrVLMUnavailable      = errors.New("VLM backend недоступен")
	ErrVLMFailed           = errors.New("ошибка VLM анализа")
)

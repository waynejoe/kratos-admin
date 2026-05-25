package errorx

// I18nError 国际化错误
type I18nError struct {
	cause error
	lang  string
}

func NewI18nError(lang string, cause error) error {
	if cause == nil {
		return nil
	}

	return &I18nError{cause: cause, lang: lang}
}

func (e *I18nError) Error() string {
	return e.cause.Error()
}

func (e *I18nError) Unwrap() error {
	return e.cause
}

func (e *I18nError) Lang() string {
	return e.lang
}

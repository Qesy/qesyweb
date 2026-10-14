package apperror

import "errors"

type AppError struct {
	Code   string                 `json:"code"`   // 错误代号/i18n Key，如 "auth.password_incorrect"
	Params map[string]interface{} `json:"params"` // 动态参数，如 {"max_retry": 3}
	Err    error                  `json:"-"`      // 底层错误，可通过 Unwrap 获取
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Code
}

func (e *AppError) Unwrap() error { return e.Err }

// GetError returns the translation key and interpolation parameters.
func GetError(err error) (string, map[string]interface{}) {
	if err == nil {
		return "", nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.Code != "" {
			return appErr.Code, appErr.Params
		}
		if appErr.Err != nil {
			return appErr.Err.Error(), appErr.Params
		}
	}
	return "error.INTERNAL_ERROR", nil
}

func New(code string, err error, params ...map[string]interface{}) error {
	if code == "" && err == nil {
		return nil
	}

	var p map[string]interface{}
	if len(params) > 0 {
		p = params[0]
	}
	return &AppError{Code: code, Err: err, Params: p}
}

package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/op/go-logging"
)

var logPanic = logging.MustGetLogger("middleware_panic")

func PanicMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		defer func() {
			err := recover()
			if err != nil {
				logPanic.Critical(err)
			}
		}()
		return next(ctx)
	}
}

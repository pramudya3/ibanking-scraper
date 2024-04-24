package middleware

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func WithAccessLogger(log *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			stop := time.Now()

			fields := []zapcore.Field{}
			if err != nil {
				fields = append(fields, zap.Error(err))
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			fields = append(fields,
				zap.String("remote_ip", c.RealIP()),
				zap.String("request_id", req.Header.Get(echo.HeaderXRequestID)),
				zap.String("host", req.Host),
				zap.String("user_agent", req.UserAgent()),
				zap.String("protocol", req.Proto),
				zap.String("referer", req.Referer()),
				zap.String("latency", stop.Sub(start).String()),
				zap.String("request", fmt.Sprintf("%s %s", req.Method, req.RequestURI)),
				zap.Int("status", res.Status),
				zap.Int64("size", res.Size),
			)

			switch sc := res.Status; {
			case sc >= 500:
				log.Warn("server error", fields...)
			case sc >= 400:
				log.Info("client error", fields...)
			case sc >= 300:
				log.Info("redirection", fields...)
			default:
				//log.Info("success", fields...)
			}

			return nil
		}
	}
}

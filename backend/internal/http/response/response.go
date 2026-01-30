package response

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"mini-banking-platform/internal/errorsx"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func WithError(c *gin.Context, message string, statusCode int) {
	c.JSON(statusCode, gin.H{"error": message})
}

func WithJSON(c *gin.Context, statusCode int, payload interface{}) {
	c.JSON(statusCode, payload)
}

func WithServiceError(c *gin.Context, err error) {
	var pub *errorsx.PublicError
	if errors.As(err, &pub) && pub != nil {
		slog.Default().Warn(
			"request failed (public error)",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"error", err,
			"cause", errorsx.RootCause(err),
		)
		WithError(c, pub.Message, pub.Status)
		return
	}

	cause := errorsx.RootCause(err)

	isClientError :=
		errors.Is(cause, errorsx.ErrUserExists) ||
			errors.Is(cause, errorsx.ErrInvalidCredentials) ||
			errors.Is(cause, errorsx.ErrInvalidToken) ||
			errors.Is(cause, errorsx.ErrUnauthorized) ||
			errors.Is(cause, errorsx.ErrAccountNotFound) ||
			errors.Is(cause, errorsx.ErrTransactionNotFound) ||
			errors.Is(cause, errorsx.ErrUserNotFound) ||
			errors.Is(cause, errorsx.ErrInsufficientFunds) ||
			errors.Is(cause, errorsx.ErrInvalidAmount) ||
			errors.Is(cause, errorsx.ErrInvalidCurrency) ||
			errors.Is(cause, errorsx.ErrCurrenciesMustDiffer) ||
			errors.Is(cause, errorsx.ErrCannotTransferToSelf)

	if isClientError {
		slog.Default().Warn(
			"request failed",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"error", err,
			"cause", cause,
		)
	} else {
		slog.Default().Error(
			"request failed",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"error", err,
			"cause", cause,
		)
	}

	switch {
	case errors.Is(cause, errorsx.ErrUserExists):
		WithError(c, errorsx.ErrUserExists.Error(), http.StatusConflict)
	case errors.Is(cause, errorsx.ErrInvalidCredentials):
		WithError(c, errorsx.ErrInvalidCredentials.Error(), http.StatusUnauthorized)
	case errors.Is(cause, errorsx.ErrInvalidToken):
		WithError(c, errorsx.ErrInvalidToken.Error(), http.StatusUnauthorized)
	case errors.Is(cause, errorsx.ErrUnauthorized):
		WithError(c, errorsx.ErrUnauthorized.Error(), http.StatusForbidden)
	case errors.Is(cause, errorsx.ErrAccountNotFound):
		WithError(c, errorsx.ErrAccountNotFound.Error(), http.StatusNotFound)
	case errors.Is(cause, errorsx.ErrTransactionNotFound):
		WithError(c, errorsx.ErrTransactionNotFound.Error(), http.StatusNotFound)
	case errors.Is(cause, errorsx.ErrUserNotFound):
		if c.FullPath() == "/api/v1/transactions/transfer" {
			WithError(c, "recipient not found", http.StatusBadRequest)
			return
		}
		WithError(c, errorsx.ErrUserNotFound.Error(), http.StatusNotFound)
	case errors.Is(cause, errorsx.ErrInsufficientFunds):
		WithError(c, errorsx.ErrInsufficientFunds.Error(), http.StatusBadRequest)
	case errors.Is(cause, errorsx.ErrInvalidAmount):
		WithError(c, errorsx.ErrInvalidAmount.Error(), http.StatusBadRequest)
	case errors.Is(cause, errorsx.ErrInvalidCurrency):
		WithError(c, errorsx.ErrInvalidCurrency.Error(), http.StatusBadRequest)
	case errors.Is(cause, errorsx.ErrCurrenciesMustDiffer):
		WithError(c, errorsx.ErrCurrenciesMustDiffer.Error(), http.StatusBadRequest)
	case errors.Is(cause, errorsx.ErrCannotTransferToSelf):
		WithError(c, errorsx.ErrCannotTransferToSelf.Error(), http.StatusBadRequest)
	default:
		WithError(c, "internal_error", http.StatusInternalServerError)
	}
}

type validationFieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func WithBindError(c *gin.Context, err error) {
	if err == nil {
		WithError(c, "invalid request", http.StatusBadRequest)
		return
	}

	if errors.Is(err, io.EOF) {
		WithError(c, "request body is required", http.StatusBadRequest)
		return
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		WithError(c, "invalid json", http.StatusBadRequest)
		return
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "invalid character") || strings.Contains(msg, "unexpected eof") || strings.Contains(msg, "unexpected end of json") {
		WithError(c, "invalid json", http.StatusBadRequest)
		return
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		WithError(c, "invalid field type: "+toSnakeCase(typeErr.Field), http.StatusBadRequest)
		return
	}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make([]validationFieldError, 0, len(ve))
		for _, fe := range ve {
			field := toSnakeCase(fe.Field())
			out = append(out, validationFieldError{
				Field:   field,
				Message: validationMessage(fe),
			})
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "fields": out})
		return
	}

	WithError(c, "invalid request body", http.StatusBadRequest)
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "oneof":
		return "must be one of: " + fe.Param()
	case "gt":
		return "must be greater than " + fe.Param()
	default:
		return "is invalid"
	}
}

func toSnakeCase(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 4)
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteByte(byte(r + ('a' - 'A')))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}


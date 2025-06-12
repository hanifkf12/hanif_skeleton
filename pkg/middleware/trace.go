package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hanifkf12/hanif_skeleton/pkg/telemetry"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel"
)

// TraceMiddleware extracts trace context from incoming requests and creates spans
func TraceMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract trace context from headers
		propagator := otel.GetTextMapPropagator()
		ctx := propagator.Extract(c.Context(), propagation.HeaderCarrier(c.GetReqHeaders()))

		// Create a new span for this request
		ctx, span := telemetry.StartSpan(ctx, "http."+c.Method()+"."+c.Path())
		defer span.End()

		// Store the context in fiber
		c.SetUserContext(ctx)

		// Call the next handler
		return c.Next()
	}
}

package config

import (
	"github.com/igurianova/logistic_optimizer/internal/middleware"
	"github.com/igurianova/logistic_optimizer/internal/route"
	"github.com/igurianova/logistic_optimizer/internal/service"
	"github.com/igurianova/logistic_optimizer/internal/transport"
	"net/http"
)

// RouteConfig Конфигурация маршрутов
type RouteConfig struct {
}

func (r *RouteConfig) Routes() map[string]http.Handler {
	builder := route.Builder()

	vrpCalculatorService := service.VrpCalculatorService{}

	builder.AddHandler(
		&transport.VrpCalculatorHandler{
			BaseUrl:              "/vrp-calculator",
			VrpCalculatorService: vrpCalculatorService,
		})
	builder.AddMiddleware(middleware.RequestLoggerMiddleware)
	return builder.Build()
}

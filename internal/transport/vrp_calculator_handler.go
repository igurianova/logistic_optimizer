package transport

import (
	"encoding/json"
	"github.com/igurianova/logistic_optimizer/internal/service"
	"github.com/igurianova/logistic_optimizer/pkg"
	"net/http"
	"regexp"
)

// VrpCalculatorHandler Обработчик запросов выполянющий подбор оптимальных маршрутов
type VrpCalculatorHandler struct {
	BaseUrl              string
	VrpCalculatorService service.VrpCalculatorService
}

func (h *VrpCalculatorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	PathRe := regexp.MustCompile(`^` + h.BaseUrl + `/*$`)
	switch {
	case r.Method == http.MethodPost && PathRe.MatchString(r.URL.Path):
		routes, err := h.VrpCalculatorService.Calculate(r)
		if err != nil {
			return
		}

		jsonBytes, err := json.MarshalIndent(routes, "", "	")
		if err != nil {
			_ = pkg.WriteHttpStatusResponse(w, http.StatusInternalServerError)
			return
		}

		_ = pkg.WriteHttpByteResponse(w, jsonBytes, http.StatusOK)
	default:
		_ = pkg.WriteHttpMessageResponse(w, "request not implemented", http.StatusBadRequest)
	}
}

func (h *VrpCalculatorHandler) GetUrls() []string {
	return []string{h.BaseUrl, h.BaseUrl + "/"}
}

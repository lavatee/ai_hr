package endpoint

import "github.com/lavatee/ai_hr/internal/service"

type Endpoint struct {
	services *service.Service
}

func NewEndpoint(services *service.Service) *Endpoint {
	return &Endpoint{services: services}
}

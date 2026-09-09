package health

import (
	"net/http"

	httptransport "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http"
)

type Controller struct{}

func NewController() *Controller { return &Controller{} }

func (c *Controller) Live(w http.ResponseWriter, _ *http.Request) {
	httptransport.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

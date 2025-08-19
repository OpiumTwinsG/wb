package httphandlers
import (
	// "checkdown/apiService/internal/pkg/logger"
	// "checkdown/apiService/internal/usecase"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	// "strconv"
	"awesomeproject7/orderservice/internal/model"
	"context"
	"awesomeproject7/orderservice/internal/cache"
)


type GetOrderRequest struct {
    OrderUID string `json:"order_uid"` // Идентификатор заказа
	
}

type OrderService interface{
	SaveOrder(ctx context.Context, order *model.Order) error
	GetOrders(ctx context.Context) ([]model.Order,error)
	GetOrderByID(ctx context.Context, id string) (model.Order,error)
}

// func atoi(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) }

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

type Handler struct{
	svc OrderService
	cache *cache.Cache
}
func New(svc OrderService, cache *cache.Cache) *Handler { return &Handler{
	svc: svc,
	cache: cache,
	} }

func (h *Handler) NewRouter() http.Handler{
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
           w.WriteHeader(http.StatusOK)         
		   _, _ = w.Write([]byte("ok")) })
	r.Route("/orders", func(r chi.Router){
		r.Post("/", h.saveOrder)
		r.Get("/", h.getOrders)
		r.Get("/{id}", h.getOrderByID)
		
		// r.Delete("/{id}", )
	})
	return r
}
func (h *Handler) saveOrder(w http.ResponseWriter, r *http.Request) {
	var order model.Order
	if err:=json.NewDecoder(r.Body).Decode(&order);err!=nil{
		writeJSON(w, http.StatusBadRequest, err) 
		return
	}
	err := h.svc.SaveOrder(r.Context(), &order)
	if err !=nil{
		writeJSON(w, http.StatusInternalServerError, err)
		return
	}
	h.cache.Set(order.OrderUID, order)
	   writeJSON(w, http.StatusCreated, map[string]any{       "status":    "ok",       "order_uid": order.OrderUID,
  })
}
func (h *Handler) getOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.svc.GetOrders(r.Context())
	if err !=nil{
		writeJSON(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, orders)
}
func (h *Handler) getOrderByID(w http.ResponseWriter, r *http.Request)  {
	id := chi.URLParam(r, "id")
	order, fl := h.cache.Get(id)
	if !fl{
		order, err := h.svc.GetOrderByID(r.Context(), id)
		if err !=nil{
			writeJSON(w, http.StatusBadRequest, err)
			return
		}
		h.cache.Set(id, order)
		writeJSON(w, http.StatusOK, order)
		return
	}
	writeJSON(w, http.StatusOK, order)	
}
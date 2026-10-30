package orderhttp_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	orderdomain "github.com/dujiao-next/internal/modules/order/domain"
	orderhttp "github.com/dujiao-next/internal/modules/order/transport/http"
	"github.com/gin-gonic/gin"
)

type singlePurchaseService struct {
	calls int
	items []orderhttp.CreateOrderItem
}

func (s *singlePurchaseService) CreateOrder(input orderhttp.CreateOrderInput) (*orderdomain.Order, error) {
	s.calls++
	s.items = input.Items
	return &orderdomain.Order{OrderNo: "SINGLE-TEST"}, nil
}
func (s *singlePurchaseService) CreateGuestOrder(input orderhttp.CreateGuestOrderInput) (*orderdomain.Order, error) {
	s.calls++
	s.items = input.Items
	return &orderdomain.Order{OrderNo: "SINGLE-TEST"}, nil
}
func (s *singlePurchaseService) PreviewOrder(input orderhttp.CreateOrderInput) (*orderhttp.OrderPreview, error) {
	s.calls++
	s.items = input.Items
	return &orderhttp.OrderPreview{}, nil
}
func (s *singlePurchaseService) PreviewGuestOrder(input orderhttp.CreateGuestOrderInput) (*orderhttp.OrderPreview, error) {
	s.calls++
	s.items = input.Items
	return &orderhttp.OrderPreview{}, nil
}

func TestStorefrontPurchaseAcceptsOneItemAndRejectsMultipleBeforeService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, endpoint := range []string{"user-preview", "guest-preview", "user-create", "guest-create", "user-pay", "guest-pay"} {
		for _, count := range []int{0, 1, 2} {
			t.Run(endpoint+string(rune('0'+count)), func(t *testing.T) {
				service := &singlePurchaseService{}
				preview := orderhttp.NewPreviewHandler(service)
				create := orderhttp.NewCreateHandler(service, nil, nil, nil)
				handler := map[string]gin.HandlerFunc{"user-preview": preview.PreviewOrder, "guest-preview": preview.PreviewGuestOrder, "user-create": create.CreateOrder, "guest-create": create.CreateGuestOrder, "user-pay": create.CreateOrderAndPay, "guest-pay": create.CreateGuestOrderAndPay}[endpoint]
				items := []orderhttp.OrderItemRequest{}
				for n := 0; n < count; n++ {
					items = append(items, orderhttp.OrderItemRequest{ProductID: uint(n + 1), SKUID: uint(n + 10), Quantity: 3})
				}
				body, err := json.Marshal(map[string]any{"items": items, "email": "guest@example.test", "order_password": "valid-password-123"})
				if err != nil {
					t.Fatal(err)
				}
				router := gin.New()
				router.POST("/purchase", func(c *gin.Context) { c.Set("user_id", uint(7)); handler(c) })
				recorder := httptest.NewRecorder()
				request := httptest.NewRequest(http.MethodPost, "/purchase", bytes.NewReader(body))
				request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(recorder, request)
				var envelope struct {
					StatusCode int `json:"status_code"`
				}
				if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
					t.Fatal(err)
				}
				if count == 1 {
					if recorder.Code != http.StatusOK || envelope.StatusCode != 0 || service.calls != 1 {
						t.Fatalf("single item failed: status=%d calls=%d body=%s", recorder.Code, service.calls, recorder.Body.String())
					}
					if len(service.items) != 1 || service.items[0].Quantity != 3 {
						t.Fatalf("single item quantity not preserved: %+v", service.items)
					}
				} else if envelope.StatusCode != http.StatusBadRequest || service.calls != 0 {
					t.Fatalf("invalid item count reached order service: status=%d calls=%d body=%s", recorder.Code, service.calls, recorder.Body.String())
				}
			})
		}
	}
}

package application_test

import (
	"errors"
	"testing"

	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
	. "github.com/dujiao-next/internal/modules/order/application"
	orderdomain "github.com/dujiao-next/internal/modules/order/domain"
)

func TestGuestPurchaseIgnoresRetiredProductIdentityRestriction(t *testing.T) {
	for _, legacyColumn := range []bool{false, true} {
		name := "new_product_schema"
		if legacyColumn {
			name = "existing_member_only_product"
		}
		t.Run(name, func(t *testing.T) {
			f := newOrderResellerSnapshotFixture(t)
			if legacyColumn {
				// Existing databases may retain this retired column and its member default.
				if err := f.db.Exec("ALTER TABLE products ADD COLUMN purchase_type varchar(20) NOT NULL DEFAULT 'member'").Error; err != nil {
					t.Fatalf("add legacy purchase column: %v", err)
				}
				if err := f.db.AutoMigrate(&productdomain.Product{}); err != nil {
					t.Fatalf("migrate product without dropping legacy data: %v", err)
				}
				var storedType string
				if err := f.db.Table("products").Select("purchase_type").Where("id = ?", f.product.ID).Scan(&storedType).Error; err != nil || storedType != "member" {
					t.Fatalf("legacy column must remain untouched: type=%q error=%v", storedType, err)
				}
			}

			input := CreateGuestOrderInput{
				Email:         "guest-purchase@example.test",
				OrderPassword: "local-guest-password",
				Items:         []CreateOrderItem{{ProductID: f.product.ID, SKUID: f.sku.ID, Quantity: 1}},
			}
			preview, err := f.svc.PreviewGuestOrder(input)
			if err != nil {
				t.Fatalf("guest preview must not require an account: %v", err)
			}
			order, err := f.svc.CreateGuestOrder(input)
			if err != nil {
				t.Fatalf("guest order must not require an account: %v", err)
			}
			if order.UserID != 0 || order.GuestEmail != input.Email {
				t.Fatalf("guest ownership changed: user=%d email=%q", order.UserID, order.GuestEmail)
			}
			if !order.TotalAmount.Decimal.Equal(preview.TotalAmount.Decimal) {
				t.Fatalf("guest preview and order totals differ: preview=%s order=%s", preview.TotalAmount, order.TotalAmount)
			}

			memberOrder, err := f.svc.CreateOrder(CreateOrderInput{UserID: f.buyer.ID, Items: input.Items})
			if err != nil {
				t.Fatalf("signed-in purchase must remain available: %v", err)
			}
			if memberOrder.UserID != f.buyer.ID || memberOrder.GuestEmail != "" {
				t.Fatalf("signed-in order lost account ownership: %+v", memberOrder)
			}
		})
	}
}

func TestOpenGuestPurchasePreservesCredentialAndStockValidation(t *testing.T) {
	f := newOrderResellerSnapshotFixture(t)
	input := CreateGuestOrderInput{
		Email:         "guest-purchase@example.test",
		OrderPassword: "local-guest-password",
		Items:         []CreateOrderItem{{ProductID: f.product.ID, SKUID: f.sku.ID, Quantity: 1}},
	}
	for _, tc := range []struct {
		name     string
		email    string
		password string
		want     error
	}{
		{"missing_email", "", input.OrderPassword, ErrGuestEmailRequired},
		{"missing_password", input.Email, "", ErrGuestPasswordRequired},
		{"weak_password", input.Email, "short", ErrGuestPasswordTooWeak},
	} {
		t.Run(tc.name, func(t *testing.T) {
			invalid := input
			invalid.Email, invalid.OrderPassword = tc.email, tc.password
			if _, err := f.svc.CreateGuestOrder(invalid); !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
	if err := f.db.Model(&productdomain.ProductSKU{}).Where("id = ?", f.sku.ID).Update("manual_stock_total", 0).Error; err != nil {
		t.Fatalf("set exhausted stock: %v", err)
	}
	if _, err := f.svc.CreateGuestOrder(input); !errors.Is(err, ErrManualStockInsufficient) {
		t.Fatalf("guest purchase must still reject exhausted stock: %v", err)
	}
	var count int64
	if err := f.db.Model(&orderdomain.Order{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("rejected guest requests persisted orders: count=%d error=%v", count, err)
	}
}

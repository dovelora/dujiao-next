package upstream

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	siteconnectiondomain "github.com/dujiao-next/internal/modules/siteconnection/domain"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

func TestSharedStockSignMatchesPHPVector(t *testing.T) {
	values := url.Values{
		"app_id":           {"42"},
		"app_key":          {"s3cr3t"},
		"name":             {"中文"},
		"sku[color]":       {"blue"},
		"empty_is_omitted": {""},
	}
	if got, want := sharedStockSign(values, "s3cr3t"), "1adf7319c16c8133378afdca42c2e209"; got != want {
		t.Fatalf("sharedStockSign() = %s, want %s", got, want)
	}
}

func TestSharedStockPreferredPriceSkipsZeroFactoryPrice(t *testing.T) {
	if got := sharedStockPreferredPrice("700", "700", "0"); got != "700" {
		t.Fatalf("preferred price = %q, want 700", got)
	}
	if got := sharedStockPreferredPrice("1.3", "1.3", "0.8"); got != "1.3" {
		t.Fatalf("merchant price = %q, want 1.3 instead of factory cost", got)
	}
	if got := sharedStockPreferredPrice("0", "0.00", "0"); got != "0" {
		t.Fatalf("all-zero preferred price = %q, want first zero", got)
	}
}

func TestReferenceKeepsNonCanonicalNumericStringsOpaque(t *testing.T) {
	encoded, err := json.Marshal(Reference("00123"))
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `"00123"` {
		t.Fatalf("encoded reference = %s", encoded)
	}
	var decoded Reference
	if err := json.Unmarshal(encoded, &decoded); err != nil || decoded != "00123" {
		t.Fatalf("decoded reference = %q, %v", decoded, err)
	}
}

func TestReferenceKeepsLargeNumericStringsExactInJSON(t *testing.T) {
	encoded, err := json.Marshal(Reference("9007199254740993"))
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `"9007199254740993"` {
		t.Fatalf("encoded reference = %s", encoded)
	}
}

func TestDujiaoNextCreateOrderKeepsNumericWireContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), `"sku_id":201`) || strings.Contains(string(body), `"sku_id":"201"`) {
			t.Fatalf("unexpected Dujiao request body: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true,"order_id":301,"order_no":"UP-301","status":"accepted"}`)
	}))
	defer server.Close()
	adapter := NewDujiaoNextAdapter(&siteconnectiondomain.Connection{
		BaseURL: server.URL, ApiKey: "key", ApiSecret: "secret",
	}, t.TempDir())
	result, err := adapter.CreateOrder(context.Background(), CreateUpstreamOrderReq{SKUID: "201", Quantity: 1})
	if err != nil || result.OrderID != "301" {
		t.Fatalf("CreateOrder = %+v, %v", result, err)
	}
}

func TestSharedStockConfigBuildsOpaqueRaceAndSKUReferences(t *testing.T) {
	config := parseSharedStockConfig(strings.Join([]string{
		"[category]",
		"基础=11.00",
		"[category_factory]",
		"基础=9.50",
		"[sku]",
		"地区.美国=1.00",
		"地区.日本=2.00",
		"时长.月付=0",
		"时长.年付=5.00",
	}, "\n"))
	skus, err := buildSharedStockSKUs("PROD-001", "8.00", 20, config, nil)
	if err != nil {
		t.Fatalf("buildSharedStockSKUs: %v", err)
	}
	if len(skus) != 4 {
		t.Fatalf("expected 4 SKU combinations, got %d", len(skus))
	}
	for _, sku := range skus {
		if !strings.HasPrefix(sku.ID.String(), sharedStockRefPrefix) {
			t.Fatalf("SKU reference is not opaque: %q", sku.ID)
		}
		ref, err := decodeSharedStockSKURef(sku.ID)
		if err != nil {
			t.Fatalf("decodeSharedStockSKURef: %v", err)
		}
		if ref.Code != "PROD-001" || ref.Race != "基础" || len(ref.SKU) != 2 {
			t.Fatalf("unexpected decoded ref: %+v", ref)
		}
	}
}

func TestSharedStockWidgetDictUsesSubmittedOptionValues(t *testing.T) {
	schema := sharedStockManualFormSchema(json.RawMessage(`[{"cn":"地区","name":"region","type":"select","dict":"美国=us,日本=jp"}]`))
	fields, ok := schema["fields"].([]interface{})
	if !ok || len(fields) != 1 {
		t.Fatalf("unexpected fields: %#v", schema)
	}
	field := fields[0].(jsonmap.JSON)
	options, ok := field["options"].([]string)
	if !ok || len(options) != 2 || options[0] != "us" || options[1] != "jp" {
		t.Fatalf("unexpected options: %#v", field["options"])
	}
}

func TestSharedStockWidgetFieldMapPreservesRemoteCase(t *testing.T) {
	raw := json.RawMessage(`[{"cn":"账号","name":"AccountID","type":"text"}]`)
	fields := sharedStockWidgetFieldMap(raw)
	if fields["accountid"] != "AccountID" {
		t.Fatalf("unexpected field mapping: %#v", fields)
	}
}

func TestSharedStockCheckboxValuesKeepArrayFormShape(t *testing.T) {
	value, err := sharedStockFormValue([]interface{}{"us", "jp"})
	if err != nil {
		t.Fatal(err)
	}
	form, err := flattenSharedStockForm(map[string]interface{}{"Region": value})
	if err != nil {
		t.Fatal(err)
	}
	if form.Get("Region[0]") != "us" || form.Get("Region[1]") != "jp" || form.Get("Region") != "" {
		t.Fatalf("unexpected checkbox form: %#v", form)
	}
}

func TestSharedStockCategorySlugIsNamespacedByMerchant(t *testing.T) {
	first := NewSharedStockAdapter(&siteconnectiondomain.Connection{BaseURL: "https://supplier.example", ApiKey: "merchant-a"}, t.TempDir())
	second := NewSharedStockAdapter(&siteconnectiondomain.Connection{BaseURL: "https://supplier.example", ApiKey: "merchant-b"}, t.TempDir())
	if first.categorySlug(1) == second.categorySlug(1) {
		t.Fatalf("merchant category slugs collided: %s", first.categorySlug(1))
	}
}

func TestSharedStockAdapterEndToEnd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		assertSharedStockAuth(t, r.PostForm, "merchant-1", "secret-1")
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case sharedStockConnectPath:
			_, _ = io.WriteString(w, `{"code":200,"msg":"success","data":{"shopName":"Web3Chirou","balance":"88.50"}}`)
		case sharedStockItemsPath:
			_, _ = io.WriteString(w, `{"code":200,"msg":"success","data":[{"id":7,"name":"账号","children":[{"id":9,"name":"测试商品","code":"PROD-001","factory_price":"9.50","stock":"20","delivery_way":0,"config":"[category]\n基础=11.00\n[category_factory]\n基础=9.50\n[sku]\n地区.美国=1.00","widget":"[{\"cn\":\"登录邮箱\",\"name\":\"Email\",\"type\":\"input\",\"required\":1}]"}]}]}`)
		case sharedStockItemPath:
			if r.PostForm.Get("code") != "PROD-001" {
				t.Fatalf("unexpected item code: %q", r.PostForm.Get("code"))
			}
			_, _ = io.WriteString(w, `{"code":200,"msg":"success","data":{"id":9,"name":"测试商品","code":"PROD-001","category_id":7,"factory_price":"9.50","stock":"20","delivery_way":0,"config":{"category":{"基础":"11.00"},"category_factory":{"基础":"9.50"},"sku":{"地区":{"美国":"1.00"}}},"widget":"[{\"cn\":\"登录邮箱\",\"name\":\"Email\",\"type\":\"input\",\"required\":1}]"}}`)
		case sharedStockStockPath:
			if r.PostForm.Get("sku[地区]") != "美国" || r.PostForm.Get("race") != "基础" {
				t.Fatalf("stock did not receive nested SKU/race: %#v", r.PostForm)
			}
			_, _ = io.WriteString(w, `{"code":200,"msg":"success","data":{"stock":"17"}}`)
		case sharedStockValuationPath:
			_, _ = io.WriteString(w, `{"code":200,"msg":"success","data":{"price":"10.50"}}`)
		case sharedStockInventoryStatePath:
			_, _ = io.WriteString(w, `{"code":200,"msg":"success","data":[]}`)
		case sharedStockTradePath:
			if r.PostForm.Get("request_no") != sharedStockRequestNo("LOCAL-001") || len(r.PostForm.Get("request_no")) != 19 || r.PostForm.Get("Email") != "buyer@example.com" {
				t.Fatalf("unexpected trade form: %#v", r.PostForm)
			}
			_, _ = io.WriteString(w, `{"code":200,"msg":"success","data":{"amount":"10.50","tradeNo":"ORDER-X","secret":"账号:a 密码:b"}}`)
		case sharedStockQueryPath:
			if r.PostForm.Get("tradeNo") != "ORDER-X" {
				t.Fatalf("unexpected query order: %q", r.PostForm.Get("tradeNo"))
			}
			_, _ = io.WriteString(w, `{"code":200,"msg":"success","data":{"status":1,"secret":"账号:a 密码:b","widget":null}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	adapter := NewSharedStockAdapter(&siteconnectiondomain.Connection{
		BaseURL: server.URL, ApiKey: "merchant-1", ApiSecret: "secret-1",
	}, t.TempDir())
	ctx := context.Background()
	ping, err := adapter.Ping(ctx)
	if err != nil || ping.SiteName != "Web3Chirou" || ping.Balance != "88.50" {
		t.Fatalf("Ping = %+v, %v", ping, err)
	}
	products, err := adapter.ListProducts(ctx, ListProductsOpts{Page: 1, PageSize: 20})
	if err != nil || products.Total != 1 || products.Items[0].ID != "PROD-001" {
		t.Fatalf("ListProducts = %+v, %v", products, err)
	}
	if fields, ok := products.Items[0].ManualFormSchema["fields"].([]interface{}); !ok || len(fields) != 1 {
		t.Fatalf("manual form schema not converted: %#v", products.Items[0].ManualFormSchema)
	}
	product, err := adapter.GetProduct(ctx, "PROD-001")
	if err != nil || product.CategoryID != 7 || len(product.SKUs) != 1 || product.SKUs[0].StockQuantity != 17 || product.SKUs[0].PriceAmount != "10.50" {
		t.Fatalf("GetProduct = %+v, %v", product, err)
	}
	created, err := adapter.CreateOrder(ctx, CreateUpstreamOrderReq{
		SKUID: product.SKUs[0].ID, Quantity: 1, DownstreamOrderNo: "LOCAL-001",
		ManualFormData: jsonmap.JSON{"email": "buyer@example.com"},
	})
	if err != nil || !created.OK || created.OrderID != "ORDER-X" || created.Fulfillment == nil {
		t.Fatalf("CreateOrder = %+v, %v", created, err)
	}
	queried, err := adapter.GetOrder(ctx, created.OrderID)
	if err != nil || queried.Status != "delivered" || queried.Fulfillment == nil {
		t.Fatalf("GetOrder = %+v, %v", queried, err)
	}
}

func TestSharedStockTradeTransportFailureIsAmbiguous(t *testing.T) {
	ref, err := encodeSharedStockSKURef(sharedStockSKURef{Version: 1, Code: "PROD-1", SKU: map[string]string{}})
	if err != nil {
		t.Fatal(err)
	}
	adapter := NewSharedStockAdapter(&siteconnectiondomain.Connection{
		BaseURL: "https://upstream.example", ApiKey: "merchant", ApiSecret: "secret",
	}, t.TempDir())
	adapter.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == sharedStockInventoryStatePath {
			return sharedStockHTTPResponse(`{"code":200,"msg":"success","data":[]}`), nil
		}
		return nil, errors.New("connection reset after write")
	})
	_, err = adapter.CreateOrder(context.Background(), CreateUpstreamOrderReq{
		SKUID: ref, Quantity: 1, DownstreamOrderNo: "LOCAL-AMBIGUOUS",
	})
	if !errors.Is(err, ErrOrderOutcomeUnknown) {
		t.Fatalf("CreateOrder error = %v, want ErrOrderOutcomeUnknown", err)
	}
}

func TestSharedStockAdapterDoesNotForwardCredentialsAcrossRedirects(t *testing.T) {
	forwarded := false
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		forwarded = true
	}))
	defer target.Close()
	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", target.URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))
	defer redirector.Close()

	adapter := NewSharedStockAdapter(&siteconnectiondomain.Connection{
		BaseURL: redirector.URL, ApiKey: "merchant", ApiSecret: "secret",
	}, t.TempDir())
	if _, err := adapter.Ping(context.Background()); err == nil || !strings.Contains(err.Error(), "HTTP status 307") {
		t.Fatalf("Ping error = %v, want redirect rejection", err)
	}
	if forwarded {
		t.Fatal("merchant credentials were forwarded to the redirect target")
	}
}

func assertSharedStockAuth(t *testing.T, form url.Values, merchantID, secret string) {
	t.Helper()
	if form.Get("app_id") != merchantID || form.Get("app_key") != secret {
		t.Fatalf("missing merchant credentials: %#v", form)
	}
	signature := form.Get("sign")
	unsigned := make(url.Values, len(form))
	for key, list := range form {
		if key == "sign" {
			continue
		}
		unsigned[key] = append([]string(nil), list...)
	}
	if expected := sharedStockSign(unsigned, secret); signature != expected {
		t.Fatalf("signature = %q, want %q", signature, expected)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func sharedStockHTTPResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

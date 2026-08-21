package upstream

import (
	"bufio"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
	siteconnectiondomain "github.com/dujiao-next/internal/modules/siteconnection/domain"
	"github.com/dujiao-next/internal/shared/jsonmap"

	"github.com/shopspring/decimal"
)

const (
	sharedStockConnectPath        = "/shared/authentication/connect"
	sharedStockItemsPath          = "/shared/commodity/items"
	sharedStockItemPath           = "/shared/commodity/item"
	sharedStockStockPath          = "/shared/commodity/stock"
	sharedStockValuationPath      = "/shared/commodity/valuation"
	sharedStockInventoryStatePath = "/shared/commodity/inventoryState"
	sharedStockTradePath          = "/shared/commodity/trade"
	sharedStockQueryPath          = "/shared/commodity/query"
	sharedStockMaxResponseBytes   = 16 << 20
	sharedStockMaxSKUCombinations = 512
	sharedStockRefPrefix          = "ss1_"
)

var sharedStockFormKeyPattern = regexp.MustCompile(`^[a-z0-9_]{1,64}$`)

// SharedStockAdapter 对接异次元店铺共享标准接口。每个连接独立保存商户 ID 与密钥，
// 因而同一上游站点可以配置多个商户连接。
type SharedStockAdapter struct {
	baseURL    string
	merchantID string
	secret     string
	uploadsDir string
	client     *http.Client
}

func NewSharedStockAdapter(conn *siteconnectiondomain.Connection, uploadsDir string) *SharedStockAdapter {
	return &SharedStockAdapter{
		baseURL:    strings.TrimRight(strings.TrimSpace(conn.BaseURL), "/"),
		merchantID: strings.TrimSpace(conn.ApiKey),
		secret:     conn.ApiSecret,
		uploadsDir: uploadsDir,
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

type sharedStockEnvelope struct {
	Code json.RawMessage `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type sharedStockAPIError struct {
	Code    int
	Message string
}

func (e *sharedStockAPIError) Error() string {
	if strings.TrimSpace(e.Message) == "" {
		return fmt.Sprintf("SharedStock 请求失败（业务码 %d）", e.Code)
	}
	return e.Message
}

type sharedStockString string

func (s *sharedStockString) UnmarshalJSON(data []byte) error {
	data = []byte(strings.TrimSpace(string(data)))
	if string(data) == "null" {
		*s = ""
		return nil
	}
	if len(data) > 0 && data[0] == '"' {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		*s = sharedStockString(value)
		return nil
	}
	var number json.Number
	if err := json.Unmarshal(data, &number); err == nil {
		*s = sharedStockString(number.String())
		return nil
	}
	var boolean bool
	if err := json.Unmarshal(data, &boolean); err == nil {
		*s = sharedStockString(strconv.FormatBool(boolean))
		return nil
	}
	return errors.New("SharedStock 字段不是标量")
}

type sharedStockCategoryPayload struct {
	ID       sharedStockString `json:"id"`
	Name     string            `json:"name"`
	Children []sharedStockItem `json:"children"`
}

type sharedStockItem struct {
	ID           sharedStockString        `json:"id"`
	CategoryID   sharedStockString        `json:"category_id"`
	Name         string                   `json:"name"`
	Code         sharedStockString        `json:"code"`
	Price        sharedStockString        `json:"price"`
	UserPrice    sharedStockString        `json:"user_price"`
	FactoryPrice sharedStockString        `json:"factory_price"`
	Stock        sharedStockString        `json:"stock"`
	DeliveryWay  sharedStockString        `json:"delivery_way"`
	Config       sharedStockConfigPayload `json:"config"`
	Cover        string                   `json:"cover"`
	PictureURL   string                   `json:"picture_url"`
	Description  string                   `json:"description"`
	Content      string                   `json:"content"`
	Widget       json.RawMessage          `json:"widget"`
	UpdatedAt    sharedStockString        `json:"updated_at"`
}

type sharedStockConfigPayload struct {
	value interface{}
}

func (p *sharedStockConfigPayload) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	return decoder.Decode(&p.value)
}

type sharedStockSKURef struct {
	Version int               `json:"v"`
	Code    string            `json:"code"`
	Race    string            `json:"race,omitempty"`
	SKU     map[string]string `json:"sku,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type sharedStockConfig struct {
	races  []sharedStockVariant
	groups []sharedStockGroup
}

type sharedStockVariant struct {
	name  string
	price string
}

type sharedStockGroup struct {
	name    string
	options []sharedStockVariant
}

func (a *SharedStockAdapter) Ping(ctx context.Context) (*PingResult, error) {
	var data struct {
		ShopName string            `json:"shopName"`
		Balance  sharedStockString `json:"balance"`
	}
	if err := a.postForm(ctx, sharedStockConnectPath, nil, &data); err != nil {
		return nil, err
	}
	userID, _ := strconv.ParseUint(a.merchantID, 10, 64)
	return &PingResult{
		SiteName: strings.TrimSpace(data.ShopName), ProtocolVersion: "shared-stock",
		UserID: uint(userID), Balance: string(data.Balance), Currency: "CNY",
	}, nil
}

func (a *SharedStockAdapter) ListCategories(ctx context.Context) (*CategoryListResult, error) {
	categories, err := a.fetchItems(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]UpstreamCategory, 0, len(categories))
	for index, category := range categories {
		id := parseSharedStockUint(category.ID)
		if id == 0 {
			id = uint(index + 1)
		}
		result = append(result, UpstreamCategory{
			ID: id, Slug: a.categorySlug(id),
			Name: localizedSharedStockText(category.Name), SortOrder: index,
		})
	}
	return &CategoryListResult{Supported: true, Categories: result}, nil
}

func (a *SharedStockAdapter) categorySlug(categoryID uint) string {
	sum := sha256.Sum256([]byte(a.baseURL + "\x00" + a.merchantID))
	return fmt.Sprintf("shared-stock-%x-%d", sum[:6], categoryID)
}

func (a *SharedStockAdapter) ListProducts(ctx context.Context, opts ListProductsOpts) (*ProductListResult, error) {
	categories, err := a.fetchItems(ctx)
	if err != nil {
		return nil, err
	}
	products := make([]UpstreamProduct, 0)
	seen := make(map[string]struct{})
	for categoryIndex, category := range categories {
		categoryID := parseSharedStockUint(category.ID)
		if categoryID == 0 {
			categoryID = uint(categoryIndex + 1)
		}
		for _, item := range category.Children {
			product, buildErr := buildSharedStockProduct(item, categoryID)
			if buildErr != nil {
				return nil, buildErr
			}
			if _, exists := seen[product.ID.String()]; exists {
				return nil, fmt.Errorf("SharedStock 返回重复商品对接码 %q", product.ID)
			}
			seen[product.ID.String()] = struct{}{}
			products = append(products, product)
		}
	}

	page, pageSize := opts.Page, opts.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start > len(products) {
		start = len(products)
	}
	end := start + pageSize
	if end > len(products) {
		end = len(products)
	}
	return &ProductListResult{
		Total: len(products), Items: append([]UpstreamProduct(nil), products[start:end]...),
		IncludesInactive: false,
	}, nil
}

func (a *SharedStockAdapter) GetProduct(ctx context.Context, productID Reference) (*UpstreamProduct, error) {
	code := strings.TrimSpace(productID.String())
	if code == "" {
		return nil, ErrUpstreamProductDeleted
	}
	var item sharedStockItem
	if err := a.postForm(ctx, sharedStockItemPath, map[string]interface{}{"code": code}, &item); err != nil {
		var apiErr *sharedStockAPIError
		if errors.As(err, &apiErr) && classifySharedStockError(apiErr.Message) == "product_unavailable" {
			return nil, ErrUpstreamProductDeleted
		}
		return nil, err
	}
	if strings.TrimSpace(string(item.Code)) == "" {
		item.Code = sharedStockString(code)
	}
	product, err := buildSharedStockProduct(item, 0)
	if err != nil {
		return nil, err
	}
	for index := range product.SKUs {
		ref, decodeErr := decodeSharedStockSKURef(product.SKUs[index].ID)
		if decodeErr != nil {
			return nil, decodeErr
		}
		stock, stockErr := a.getStock(ctx, ref)
		if stockErr != nil {
			return nil, stockErr
		}
		product.SKUs[index].StockQuantity = stock
		product.SKUs[index].StockStatus = sharedStockStatus(stock)
		price, priceErr := a.getValuation(ctx, ref, 1)
		if priceErr == nil && price != "" && price != "0" {
			product.SKUs[index].PriceAmount = price
		}
	}
	product.PriceAmount = lowestSharedStockPrice(product.SKUs, product.PriceAmount)
	return &product, nil
}

func (a *SharedStockAdapter) CreateOrder(ctx context.Context, req CreateUpstreamOrderReq) (*CreateUpstreamOrderResp, error) {
	ref, err := decodeSharedStockSKURef(req.SKUID)
	if err != nil {
		return &CreateUpstreamOrderResp{OK: false, ErrorCode: "sku_unavailable", ErrorMessage: err.Error()}, nil
	}
	check := map[string]interface{}{
		"shared_code": ref.Code, "num": req.Quantity, "card_id": 0,
		"race": ref.Race, "sku": ref.SKU,
	}
	if err := a.postForm(ctx, sharedStockInventoryStatePath, check, nil); err != nil {
		return sharedStockOrderFailure(err), nil
	}

	requestNo := strings.TrimSpace(req.DownstreamOrderNo)
	if requestNo == "" {
		requestNo = strings.TrimSpace(req.TraceID)
	}
	requestNo = sharedStockRequestNo(requestNo)
	if requestNo == "" {
		return &CreateUpstreamOrderResp{OK: false, ErrorCode: "invalid_request", ErrorMessage: "SharedStock 下单缺少幂等请求号"}, nil
	}
	trade := map[string]interface{}{
		"shared_code": ref.Code, "contact": "-", "num": req.Quantity,
		"card_id": 0, "device": 0, "password": "", "race": ref.Race,
		"request_no": requestNo, "sku": ref.SKU,
	}
	for key, raw := range req.ManualFormData {
		remoteKey := key
		if mapped := strings.TrimSpace(ref.Fields[key]); mapped != "" {
			remoteKey = mapped
		}
		if isSharedStockReservedTradeKey(remoteKey) {
			continue
		}
		value, valueErr := sharedStockFormValue(raw)
		if valueErr != nil {
			return &CreateUpstreamOrderResp{OK: false, ErrorCode: "invalid_request", ErrorMessage: valueErr.Error()}, nil
		}
		trade[remoteKey] = value
	}

	var data struct {
		Amount  sharedStockString `json:"amount"`
		TradeNo sharedStockString `json:"tradeNo"`
		Secret  string            `json:"secret"`
	}
	if err := a.postForm(ctx, sharedStockTradePath, trade, &data); err != nil {
		var apiErr *sharedStockAPIError
		if errors.As(err, &apiErr) {
			return sharedStockOrderFailure(err), nil
		}
		return nil, fmt.Errorf("%w: request_no=%s: %v", ErrOrderOutcomeUnknown, requestNo, err)
	}
	tradeNo := strings.TrimSpace(string(data.TradeNo))
	if tradeNo == "" {
		return nil, fmt.Errorf("%w: request_no=%s: SharedStock 下单响应缺少订单号", ErrOrderOutcomeUnknown, requestNo)
	}
	response := &CreateUpstreamOrderResp{
		OK: true, OrderID: Reference(tradeNo), OrderNo: tradeNo,
		Status: "accepted", Amount: string(data.Amount), Currency: "CNY",
	}
	if sharedStockSecretDelivered(data.Secret) {
		now := time.Now()
		response.Status = "delivered"
		response.Fulfillment = &UpstreamFulfillment{
			Type: "shared-stock", Status: "delivered", Payload: data.Secret,
			DeliveryData: jsonmap.JSON{"secret": data.Secret}, DeliveredAt: &now,
		}
	}
	return response, nil
}

func (a *SharedStockAdapter) GetOrder(ctx context.Context, orderID Reference) (*UpstreamOrderDetail, error) {
	tradeNo := strings.TrimSpace(orderID.String())
	var data struct {
		Secret string            `json:"secret"`
		Widget interface{}       `json:"widget"`
		Status sharedStockString `json:"status"`
		Amount sharedStockString `json:"amount"`
	}
	if err := a.postForm(ctx, sharedStockQueryPath, map[string]interface{}{"tradeNo": tradeNo}, &data); err != nil {
		return nil, err
	}
	status := "processing"
	var fulfillment *UpstreamFulfillment
	if sharedStockSecretDelivered(data.Secret) {
		status = "delivered"
		now := time.Now()
		deliveryData := jsonmap.JSON{"secret": data.Secret}
		if data.Widget != nil {
			deliveryData["widget"] = data.Widget
		}
		fulfillment = &UpstreamFulfillment{
			Type: "shared-stock", Status: "delivered", Payload: data.Secret,
			DeliveryData: deliveryData, DeliveredAt: &now,
		}
	} else if parseSharedStockInt(data.Status) == 0 {
		status = "pending"
	}
	return &UpstreamOrderDetail{
		OrderID: orderID, OrderNo: tradeNo, Status: status,
		Amount: string(data.Amount), Currency: "CNY", Fulfillment: fulfillment,
	}, nil
}

func (a *SharedStockAdapter) CancelOrder(context.Context, Reference) error {
	return errors.New("SharedStock 协议不支持取消已提交的采购单")
}

func (a *SharedStockAdapter) DownloadImage(ctx context.Context, imageURL string) (string, error) {
	return downloadUpstreamImage(ctx, a.baseURL, imageURL, a.uploadsDir)
}

func (a *SharedStockAdapter) fetchItems(ctx context.Context) ([]sharedStockCategoryPayload, error) {
	var raw json.RawMessage
	if err := a.postForm(ctx, sharedStockItemsPath, nil, &raw); err != nil {
		return nil, err
	}
	var categories []sharedStockCategoryPayload
	if err := json.Unmarshal(raw, &categories); err == nil {
		return categories, nil
	}
	var wrapper struct {
		Items []sharedStockCategoryPayload `json:"items"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, fmt.Errorf("decode SharedStock items: %w", err)
	}
	return wrapper.Items, nil
}

func (a *SharedStockAdapter) getStock(ctx context.Context, ref sharedStockSKURef) (int, error) {
	var data struct {
		Stock sharedStockString `json:"stock"`
	}
	params := map[string]interface{}{"code": ref.Code, "race": ref.Race, "sku": ref.SKU}
	if err := a.postForm(ctx, sharedStockStockPath, params, &data); err != nil {
		return 0, err
	}
	return parseSharedStockInt(data.Stock), nil
}

func (a *SharedStockAdapter) getValuation(ctx context.Context, ref sharedStockSKURef, quantity int) (string, error) {
	var data struct {
		Price sharedStockString `json:"price"`
	}
	params := map[string]interface{}{
		"code": ref.Code, "num": quantity, "race": ref.Race, "sku": ref.SKU, "card_id": 0,
	}
	if err := a.postForm(ctx, sharedStockValuationPath, params, &data); err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data.Price)), nil
}

func (a *SharedStockAdapter) postForm(ctx context.Context, path string, params map[string]interface{}, output interface{}) error {
	values, err := flattenSharedStockForm(params)
	if err != nil {
		return err
	}
	values.Set("app_id", a.merchantID)
	values.Set("app_key", a.secret)
	values.Set("sign", sharedStockSign(values, a.secret))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+path, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("create SharedStock request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("send SharedStock request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, sharedStockMaxResponseBytes+1))
	if err != nil {
		return fmt.Errorf("read SharedStock response: %w", err)
	}
	if len(body) > sharedStockMaxResponseBytes {
		return errors.New("SharedStock response exceeds size limit")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("SharedStock HTTP status %d", resp.StatusCode)
	}
	var envelope sharedStockEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("decode SharedStock response: %w", err)
	}
	code := parseSharedStockCode(envelope.Code)
	if code != http.StatusOK {
		return &sharedStockAPIError{Code: code, Message: strings.TrimSpace(envelope.Msg)}
	}
	if output == nil {
		return nil
	}
	if raw, ok := output.(*json.RawMessage); ok {
		*raw = append((*raw)[:0], envelope.Data...)
		return nil
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return errors.New("SharedStock response is missing data")
	}
	if err := json.Unmarshal(envelope.Data, output); err != nil {
		return fmt.Errorf("decode SharedStock data: %w", err)
	}
	return nil
}

func flattenSharedStockForm(params map[string]interface{}) (url.Values, error) {
	values := make(url.Values)
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := appendSharedStockFormValue(values, key, params[key]); err != nil {
			return nil, err
		}
	}
	return values, nil
}

func appendSharedStockFormValue(values url.Values, key string, raw interface{}) error {
	switch value := raw.(type) {
	case nil:
		return nil
	case map[string]string:
		keys := make([]string, 0, len(value))
		for child := range value {
			keys = append(keys, child)
		}
		sort.Strings(keys)
		for _, child := range keys {
			if err := appendSharedStockFormValue(values, key+"["+child+"]", value[child]); err != nil {
				return err
			}
		}
		return nil
	case map[string]interface{}:
		keys := make([]string, 0, len(value))
		for child := range value {
			keys = append(keys, child)
		}
		sort.Strings(keys)
		for _, child := range keys {
			if err := appendSharedStockFormValue(values, key+"["+child+"]", value[child]); err != nil {
				return err
			}
		}
		return nil
	case []interface{}:
		for index, item := range value {
			if err := appendSharedStockFormValue(values, fmt.Sprintf("%s[%d]", key, index), item); err != nil {
				return err
			}
		}
		return nil
	case []string:
		for index, item := range value {
			if err := appendSharedStockFormValue(values, fmt.Sprintf("%s[%d]", key, index), item); err != nil {
				return err
			}
		}
		return nil
	case string:
		values.Set(key, value)
		return nil
	case bool:
		values.Set(key, strconv.FormatBool(value))
		return nil
	case int:
		values.Set(key, strconv.Itoa(value))
		return nil
	case int64:
		values.Set(key, strconv.FormatInt(value, 10))
		return nil
	case uint:
		values.Set(key, strconv.FormatUint(uint64(value), 10))
		return nil
	case float64:
		values.Set(key, strconv.FormatFloat(value, 'f', -1, 64))
		return nil
	default:
		return fmt.Errorf("SharedStock 参数 %s 的类型不受支持", key)
	}
}

func sharedStockSign(values url.Values, secret string) string {
	unsigned := make(url.Values, len(values))
	for key, list := range values {
		if key == "sign" {
			continue
		}
		for _, value := range list {
			if value != "" {
				unsigned.Add(key, value)
			}
		}
	}
	query := unsigned.Encode()
	decoded, err := url.QueryUnescape(query)
	if err != nil {
		decoded = query
	}
	sum := md5.Sum([]byte(decoded + "&key=" + secret))
	return hex.EncodeToString(sum[:])
}

func buildSharedStockProduct(item sharedStockItem, categoryID uint) (UpstreamProduct, error) {
	code := strings.TrimSpace(string(item.Code))
	if code == "" {
		return UpstreamProduct{}, errors.New("SharedStock 商品缺少对接码")
	}
	// SharedStock 的实时 valuation 与 user_price/price 一致；部分站点的
	// factory_price 是内部供货成本，并不是当前商户实际下单价。
	basePrice := sharedStockPreferredPrice(item.UserPrice, item.Price, item.FactoryPrice)
	if basePrice == "" {
		basePrice = "0"
	}
	config := parseSharedStockConfigPayload(item.Config)
	if categoryID == 0 {
		categoryID = parseSharedStockUint(item.CategoryID)
	}
	skus, err := buildSharedStockSKUs(code, basePrice, parseSharedStockInt(item.Stock), config, sharedStockWidgetFieldMap(item.Widget))
	if err != nil {
		return UpstreamProduct{}, err
	}
	images := []string{}
	if cover := firstNonEmptyString(item.Cover, item.PictureURL); cover != "" {
		images = append(images, cover)
	}
	updatedAt := time.Now()
	if parsed, parseErr := parseSharedStockTime(string(item.UpdatedAt)); parseErr == nil {
		updatedAt = parsed
	}
	fulfillmentType := "manual"
	if parseSharedStockInt(item.DeliveryWay) == 0 {
		fulfillmentType = "auto"
	}
	return UpstreamProduct{
		ID: Reference(code), Title: localizedSharedStockText(item.Name),
		Description: localizedSharedStockText(item.Description), Content: localizedSharedStockText(item.Content),
		Images: images, PriceAmount: lowestSharedStockPrice(skus, basePrice), Currency: "CNY",
		FulfillmentType: fulfillmentType, ManualFormSchema: sharedStockManualFormSchema(item.Widget),
		IsActive: true, CategoryID: categoryID, SKUs: skus, UpdatedAt: updatedAt,
		WholesalePrices: productdomain.WholesalePriceTiers{},
	}, nil
}

func buildSharedStockSKUs(code, basePrice string, stock int, config sharedStockConfig, widgetFields map[string]string) ([]UpstreamSKU, error) {
	traces := config.races
	if len(traces) == 0 {
		traces = []sharedStockVariant{{}}
	}
	selections := []map[string]string{{}}
	for _, group := range config.groups {
		if len(selections)*len(group.options) > sharedStockMaxSKUCombinations {
			return nil, fmt.Errorf("SharedStock 商品 %s 的 SKU 组合超过 %d 个", code, sharedStockMaxSKUCombinations)
		}
		next := make([]map[string]string, 0, len(selections)*len(group.options))
		for _, selection := range selections {
			for _, option := range group.options {
				copySelection := make(map[string]string, len(selection)+1)
				for key, value := range selection {
					copySelection[key] = value
				}
				copySelection[group.name] = option.name
				next = append(next, copySelection)
			}
		}
		selections = next
	}
	if len(traces)*len(selections) > sharedStockMaxSKUCombinations {
		return nil, fmt.Errorf("SharedStock 商品 %s 的 SKU 组合超过 %d 个", code, sharedStockMaxSKUCombinations)
	}

	skus := make([]UpstreamSKU, 0, len(traces)*len(selections))
	for _, race := range traces {
		for _, selection := range selections {
			ref := sharedStockSKURef{Version: 1, Code: code, Race: race.name, SKU: selection, Fields: widgetFields}
			encoded, err := encodeSharedStockSKURef(ref)
			if err != nil {
				return nil, err
			}
			price := race.price
			if price == "" {
				price = basePrice
			}
			price = addSharedStockSKUPremiums(price, selection, config.groups)
			specValues := jsonmap.JSON{}
			if race.name != "" {
				specValues["race"] = race.name
			}
			for key, value := range selection {
				specValues[key] = value
			}
			skus = append(skus, UpstreamSKU{
				ID: encoded, SKUCode: sharedStockSKUCode(code, encoded), SpecValues: specValues,
				PriceAmount: price, StockStatus: sharedStockStatus(stock), StockQuantity: stock, IsActive: true,
			})
		}
	}
	return skus, nil
}

func encodeSharedStockSKURef(ref sharedStockSKURef) (Reference, error) {
	if strings.TrimSpace(ref.Code) == "" {
		return "", errors.New("SharedStock SKU 引用缺少商品对接码")
	}
	if ref.Version == 0 {
		ref.Version = 1
	}
	encoded, err := json.Marshal(ref)
	if err != nil {
		return "", err
	}
	return Reference(sharedStockRefPrefix + base64.RawURLEncoding.EncodeToString(encoded)), nil
}

func decodeSharedStockSKURef(raw Reference) (sharedStockSKURef, error) {
	value := strings.TrimSpace(raw.String())
	if !strings.HasPrefix(value, sharedStockRefPrefix) {
		return sharedStockSKURef{}, errors.New("SharedStock SKU 引用格式不正确")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(value, sharedStockRefPrefix))
	if err != nil {
		return sharedStockSKURef{}, errors.New("SharedStock SKU 引用无法解码")
	}
	var ref sharedStockSKURef
	if err := json.Unmarshal(decoded, &ref); err != nil || ref.Version != 1 || strings.TrimSpace(ref.Code) == "" {
		return sharedStockSKURef{}, errors.New("SharedStock SKU 引用内容不正确")
	}
	if ref.SKU == nil {
		ref.SKU = map[string]string{}
	}
	if ref.Fields == nil {
		ref.Fields = map[string]string{}
	}
	return ref, nil
}

func parseSharedStockConfig(raw string) sharedStockConfig {
	type raceValue struct{ retail, factory string }
	traces := make(map[string]*raceValue)
	raceOrder := make([]string, 0)
	groups := make(map[string]map[string]string)
	groupOrder := make([]string, 0)
	optionOrder := make(map[string][]string)

	scanner := bufio.NewScanner(strings.NewReader(raw))
	section := ""
	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\ufeff"))
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		left := strings.TrimSpace(parts[0])
		base, keys := parseSharedStockINIKey(left)
		if len(keys) == 0 && section != "" {
			base = section
			if section == "sku" {
				keys = strings.SplitN(left, ".", 2)
			} else {
				keys = []string{left}
			}
		}
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		switch {
		case (base == "category" || base == "category_factory") && len(keys) == 1:
			name := keys[0]
			entry, exists := traces[name]
			if !exists {
				entry = &raceValue{}
				traces[name] = entry
				raceOrder = append(raceOrder, name)
			}
			if base == "category_factory" {
				entry.factory = value
			} else {
				entry.retail = value
			}
		case base == "sku" && len(keys) == 2:
			group, option := keys[0], keys[1]
			if _, exists := groups[group]; !exists {
				groups[group] = make(map[string]string)
				groupOrder = append(groupOrder, group)
			}
			if _, exists := groups[group][option]; !exists {
				optionOrder[group] = append(optionOrder[group], option)
			}
			groups[group][option] = value
		}
	}

	result := sharedStockConfig{}
	for _, name := range raceOrder {
		price := traces[name].factory
		if price == "" {
			price = traces[name].retail
		}
		result.races = append(result.races, sharedStockVariant{name: name, price: price})
	}
	for _, group := range groupOrder {
		entry := sharedStockGroup{name: group}
		for _, option := range optionOrder[group] {
			entry.options = append(entry.options, sharedStockVariant{name: option, price: groups[group][option]})
		}
		result.groups = append(result.groups, entry)
	}
	return result
}

func parseSharedStockConfigPayload(payload sharedStockConfigPayload) sharedStockConfig {
	switch value := payload.value.(type) {
	case string:
		return parseSharedStockConfig(value)
	case map[string]interface{}:
		result := sharedStockConfig{}
		categories := sharedStockConfigMap(value["category"])
		factory := sharedStockConfigMap(value["category_factory"])
		categoryNames := sortedSharedStockKeys(categories)
		for _, name := range categoryNames {
			price := sharedStockAnyString(factory[name])
			if price == "" {
				price = sharedStockAnyString(categories[name])
			}
			result.races = append(result.races, sharedStockVariant{name: name, price: price})
		}
		groups := sharedStockConfigMap(value["sku"])
		for _, groupName := range sortedSharedStockKeys(groups) {
			options := sharedStockConfigMap(groups[groupName])
			group := sharedStockGroup{name: groupName}
			for _, optionName := range sortedSharedStockKeys(options) {
				group.options = append(group.options, sharedStockVariant{
					name: optionName, price: sharedStockAnyString(options[optionName]),
				})
			}
			if len(group.options) > 0 {
				result.groups = append(result.groups, group)
			}
		}
		return result
	default:
		return sharedStockConfig{}
	}
}

func sharedStockConfigMap(value interface{}) map[string]interface{} {
	result, _ := value.(map[string]interface{})
	return result
}

func sortedSharedStockKeys(values map[string]interface{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func parseSharedStockINIKey(raw string) (string, []string) {
	open := strings.IndexByte(raw, '[')
	if open < 0 {
		return strings.TrimSpace(raw), nil
	}
	base := strings.TrimSpace(raw[:open])
	keys := make([]string, 0, 2)
	rest := raw[open:]
	for len(rest) > 0 {
		start := strings.IndexByte(rest, '[')
		end := strings.IndexByte(rest, ']')
		if start < 0 || end <= start {
			break
		}
		key := strings.Trim(strings.TrimSpace(rest[start+1:end]), "\"'")
		if key != "" {
			keys = append(keys, key)
		}
		rest = rest[end+1:]
	}
	return base, keys
}

func addSharedStockSKUPremiums(base string, selection map[string]string, groups []sharedStockGroup) string {
	amount, err := decimal.NewFromString(strings.TrimSpace(base))
	if err != nil {
		return base
	}
	for _, group := range groups {
		selected := selection[group.name]
		for _, option := range group.options {
			if option.name != selected || strings.TrimSpace(option.price) == "" {
				continue
			}
			premium, parseErr := decimal.NewFromString(strings.TrimSpace(option.price))
			if parseErr == nil {
				amount = amount.Add(premium)
			}
		}
	}
	return amount.StringFixed(2)
}

func sharedStockManualFormSchema(raw json.RawMessage) jsonmap.JSON {
	if len(raw) == 0 || string(raw) == "null" || string(raw) == `""` {
		return jsonmap.JSON{"fields": []interface{}{}}
	}
	if raw[0] == '"' {
		var encoded string
		if json.Unmarshal(raw, &encoded) != nil || strings.TrimSpace(encoded) == "" {
			return jsonmap.JSON{"fields": []interface{}{}}
		}
		raw = json.RawMessage(encoded)
	}
	var widgets []map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	if decoder.Decode(&widgets) != nil {
		return jsonmap.JSON{"fields": []interface{}{}}
	}
	fields := make([]interface{}, 0, len(widgets))
	seen := make(map[string]struct{})
	for _, widget := range widgets {
		key := strings.ToLower(strings.TrimSpace(sharedStockAnyString(widget["name"])))
		if !sharedStockFormKeyPattern.MatchString(key) {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		fieldType := sharedStockManualFieldType(sharedStockAnyString(widget["type"]))
		label := firstNonEmptyString(sharedStockAnyString(widget["cn"]), sharedStockAnyString(widget["title"]), key)
		field := jsonmap.JSON{
			"key": key, "type": fieldType, "required": sharedStockAnyBool(widget["required"]),
			"label": localizedSharedStockText(label),
		}
		if placeholder := sharedStockAnyString(widget["placeholder"]); placeholder != "" {
			field["placeholder"] = localizedSharedStockText(placeholder)
		}
		if pattern := sharedStockAnyString(widget["regex"]); pattern != "" {
			if _, err := regexp.Compile(pattern); err == nil {
				field["regex"] = pattern
			}
		}
		if fieldType == "select" || fieldType == "radio" || fieldType == "checkbox" {
			options := sharedStockOptions(widget["options"])
			if len(options) == 0 {
				options = sharedStockOptions(widget["dict"])
			}
			if len(options) == 0 {
				continue
			}
			field["options"] = options
		}
		fields = append(fields, field)
	}
	return jsonmap.JSON{"fields": fields}
}

func sharedStockWidgetFieldMap(raw json.RawMessage) map[string]string {
	if len(raw) == 0 || string(raw) == "null" || string(raw) == `""` {
		return nil
	}
	if raw[0] == '"' {
		var encoded string
		if json.Unmarshal(raw, &encoded) != nil || strings.TrimSpace(encoded) == "" {
			return nil
		}
		raw = json.RawMessage(encoded)
	}
	var widgets []map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	if decoder.Decode(&widgets) != nil {
		return nil
	}
	result := make(map[string]string)
	for _, widget := range widgets {
		remoteKey := strings.TrimSpace(sharedStockAnyString(widget["name"]))
		localKey := strings.ToLower(remoteKey)
		if !sharedStockFormKeyPattern.MatchString(localKey) {
			continue
		}
		if _, exists := result[localKey]; !exists {
			result[localKey] = remoteKey
		}
	}
	return result
}

func sharedStockSKUCode(code string, ref Reference) string {
	sum := sha256.Sum256([]byte(ref.String()))
	prefix := strings.ToLower(strings.TrimSpace(code))
	prefix = regexp.MustCompile(`[^a-z0-9_-]+`).ReplaceAllString(prefix, "-")
	prefix = strings.Trim(prefix, "-")
	if len(prefix) > 32 {
		prefix = prefix[:32]
	}
	if prefix == "" {
		prefix = "shared-stock"
	}
	return fmt.Sprintf("%s-%x", prefix, sum[:8])
}

func sharedStockRequestNo(source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(source))
	return hex.EncodeToString(sum[:])[:19]
}

func parseSharedStockCode(raw json.RawMessage) int {
	var number int
	if json.Unmarshal(raw, &number) == nil {
		return number
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		number, _ = strconv.Atoi(text)
	}
	return number
}

func parseSharedStockUint(raw sharedStockString) uint {
	value, _ := strconv.ParseUint(strings.TrimSpace(string(raw)), 10, 64)
	return uint(value)
}

func parseSharedStockInt(raw sharedStockString) int {
	value := strings.TrimSpace(string(raw))
	if integer, err := strconv.Atoi(value); err == nil {
		return integer
	}
	if decimalValue, err := strconv.ParseFloat(value, 64); err == nil {
		return int(decimalValue)
	}
	return 0
}

func parseSharedStockTime(raw string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
		if value, err := time.Parse(layout, strings.TrimSpace(raw)); err == nil {
			return value, nil
		}
	}
	return time.Time{}, errors.New("invalid SharedStock time")
}

func localizedSharedStockText(value string) jsonmap.JSON {
	value = strings.TrimSpace(value)
	if value == "" {
		return jsonmap.JSON{}
	}
	return jsonmap.JSON{"zh-CN": value, "zh-TW": value, "en-US": value}
}

func sharedStockStatus(stock int) string {
	if stock == 0 {
		return "out_of_stock"
	}
	return "in_stock"
}

func lowestSharedStockPrice(skus []UpstreamSKU, fallback string) string {
	lowest := decimal.Zero
	lowestRaw := ""
	for _, sku := range skus {
		value, err := decimal.NewFromString(strings.TrimSpace(sku.PriceAmount))
		if err == nil && !value.IsNegative() && (lowestRaw == "" || value.LessThan(lowest)) {
			lowest, lowestRaw = value, sku.PriceAmount
		}
	}
	if lowestRaw == "" {
		return fallback
	}
	return lowestRaw
}

func sharedStockPreferredPrice(values ...sharedStockString) string {
	fallback := ""
	for _, value := range values {
		trimmed := strings.TrimSpace(string(value))
		if trimmed == "" {
			continue
		}
		if fallback == "" {
			fallback = trimmed
		}
		amount, err := decimal.NewFromString(trimmed)
		if err == nil && amount.GreaterThan(decimal.Zero) {
			return trimmed
		}
	}
	return fallback
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func sharedStockOrderFailure(err error) *CreateUpstreamOrderResp {
	return &CreateUpstreamOrderResp{OK: false, ErrorCode: classifySharedStockError(err.Error()), ErrorMessage: err.Error()}
}

func classifySharedStockError(message string) string {
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(message, "余额") || strings.Contains(lower, "balance"):
		return "insufficient_balance"
	case strings.Contains(message, "库存") || strings.Contains(message, "无货") || strings.Contains(lower, "stock"):
		return "product_out_of_stock"
	case strings.Contains(message, "商品不存在") || strings.Contains(message, "停售") || strings.Contains(message, "下架"):
		return "product_unavailable"
	case strings.Contains(message, "商户") || strings.Contains(message, "密钥") || strings.Contains(message, "签名") || strings.Contains(lower, "sign"):
		return "unauthorized"
	case strings.Contains(message, "重复") || strings.Contains(message, "已存在") || strings.Contains(lower, "duplicate"):
		return "duplicate_order"
	default:
		return "upstream_error"
	}
}

func isSharedStockReservedTradeKey(key string) bool {
	switch key {
	case "app_id", "app_key", "sign", "shared_code", "num", "request_no", "contact", "card_id", "device", "password", "race", "sku":
		return true
	default:
		return false
	}
}

func sharedStockFormValue(raw interface{}) (interface{}, error) {
	switch value := raw.(type) {
	case nil:
		return "", nil
	case string:
		return strings.TrimSpace(value), nil
	case bool:
		return strconv.FormatBool(value), nil
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64), nil
	case int:
		return strconv.Itoa(value), nil
	case int64:
		return strconv.FormatInt(value, 10), nil
	case []interface{}:
		items := make([]interface{}, 0, len(value))
		for _, item := range value {
			normalized, err := sharedStockFormValue(item)
			if err != nil {
				return "", err
			}
			items = append(items, normalized)
		}
		return items, nil
	case []string:
		items := make([]interface{}, len(value))
		for index, item := range value {
			items[index] = strings.TrimSpace(item)
		}
		return items, nil
	default:
		return "", errors.New("SharedStock 自定义表单包含不支持的字段值")
	}
}

func sharedStockSecretDelivered(secret string) bool {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return false
	}
	lower := strings.ToLower(secret)
	return !strings.Contains(secret, "正在发货") && !strings.Contains(secret, "没有发货信息") && !strings.Contains(lower, "processing")
}

func sharedStockManualFieldType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "textarea":
		return "textarea"
	case "select":
		return "select"
	case "radio":
		return "radio"
	case "checkbox":
		return "checkbox"
	case "email":
		return "email"
	case "phone", "tel":
		return "phone"
	case "number":
		return "number"
	default:
		return "text"
	}
}

func sharedStockAnyString(raw interface{}) string {
	switch value := raw.(type) {
	case string:
		return strings.TrimSpace(value)
	case json.Number:
		return value.String()
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	default:
		return ""
	}
}

func sharedStockAnyBool(raw interface{}) bool {
	switch value := raw.(type) {
	case bool:
		return value
	case json.Number:
		return value.String() != "0"
	case float64:
		return value != 0
	case string:
		value = strings.ToLower(strings.TrimSpace(value))
		return value == "1" || value == "true" || value == "yes" || value == "required"
	default:
		return false
	}
}

func sharedStockOptions(raw interface{}) []string {
	result := make([]string, 0)
	switch value := raw.(type) {
	case []interface{}:
		for _, item := range value {
			if option := sharedStockAnyString(item); option != "" {
				result = append(result, option)
			}
		}
	case string:
		for _, item := range strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == '|' || r == '\n' }) {
			option := strings.TrimSpace(item)
			if _, submitted, found := strings.Cut(option, "="); found {
				option = strings.TrimSpace(submitted)
			}
			if option != "" {
				result = append(result, option)
			}
		}
	}
	return result
}

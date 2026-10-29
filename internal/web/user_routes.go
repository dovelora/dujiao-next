package web

import "strings"

// Keep the history fallback limited to actual storefront routes. A cross-source
// regression test checks these against frontend/user/src/router/index.ts.
func isUserRoute(requestPath string) bool {
	p := strings.ToLower(strings.TrimSuffix(requestPath, "/"))
	switch p {
	case "", "/products", "/cart", "/checkout", "/pay",
		"/me", "/me/profile", "/me/security", "/me/orders", "/me/wallet", "/me/gift-cards", "/me/api", "/me/affiliate", "/me/reseller",
		"/reseller", "/reseller/apply", "/reseller/domains", "/reseller/site", "/reseller/products", "/reseller/orders", "/reseller/finance", "/reseller/ledger", "/reseller/withdraws",
		"/guest/orders", "/blog", "/notice", "/about", "/terms", "/privacy",
		"/auth/login", "/auth/register", "/auth/forgot", "/auth/telegram/callback", "/auth/google/callback":
		return true
	}
	for _, prefix := range []string{"/products/", "/categories/", "/blog/", "/orders/", "/recharge-orders/", "/guest/orders/", "/reseller/orders/"} {
		if strings.HasPrefix(p, prefix) {
			param := strings.TrimPrefix(p, prefix)
			return param != "" && !strings.Contains(param, "/")
		}
	}
	return false
}

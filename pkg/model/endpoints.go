package model

// CLOB API endpoint paths.
const (
	// Health
	EndpointOk         = "/"
	EndpointServerTime = "/time"

	// Auth (L1)
	EndpointCreateAPIKey = "/auth/api-key"
	EndpointDeriveAPIKey = "/auth/derive-api-key"

	// Auth (L2)
	EndpointGetAPIKeys    = "/auth/api-keys"
	EndpointDeleteAPIKey  = "/auth/api-key"
	EndpointClosedOnly    = "/auth/ban-status/closed-only"

	// Market data (L0)
	EndpointMidpoint        = "/midpoint"
	EndpointMidpoints       = "/midpoints"
	EndpointPrice           = "/price"
	EndpointPrices          = "/prices"
	EndpointSpread          = "/spread"
	EndpointSpreads         = "/spreads"
	EndpointTickSize        = "/tick-size"
	EndpointNegRisk         = "/neg-risk"
	EndpointFeeRate         = "/fee-rate"
	EndpointBook            = "/book"
	EndpointBooks           = "/books"
	EndpointLastTradePrice  = "/last-trade-price"
	EndpointLastTradesPrices = "/last-trades-prices"

	// Markets (L0)
	EndpointMarkets                   = "/markets"
	EndpointSimplifiedMarkets         = "/simplified-markets"
	EndpointSamplingMarkets           = "/sampling-markets"
	EndpointSamplingSimplifiedMarkets = "/sampling-simplified-markets"

	// Orders (L2)
	EndpointOrder             = "/order"
	EndpointOrders            = "/orders"
	EndpointCancelAll         = "/cancel-all"
	EndpointCancelMarketOrders = "/cancel-market-orders"

	// Order data (L2)
	EndpointDataOrders = "/data/orders"
	EndpointDataOrder  = "/data/order"
	EndpointDataTrades = "/data/trades"

	// Balance (L2)
	EndpointBalanceAllowance       = "/balance-allowance"
	EndpointBalanceAllowanceUpdate = "/balance-allowance/update"

	// Notifications (L2)
	EndpointNotifications = "/notifications"

	// Scoring (L2)
	EndpointOrderScoring  = "/order-scoring"
	EndpointOrdersScoring = "/orders-scoring"

	// Heartbeat (L2)
	EndpointHeartbeat = "/v1/heartbeats"

	// Pagination
	EndCursor     = "LTE="
	InitialCursor = "MA=="
)

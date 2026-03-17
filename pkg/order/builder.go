package order

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/jaxxjj/go-poly-clob-client/pkg/model"
	gomodel "github.com/polymarket/go-order-utils/pkg/model"

	"github.com/polymarket/go-order-utils/pkg/builder"
)

// Builder creates and signs Polymarket orders using go-order-utils.
type Builder struct {
	privateKey   *ecdsa.PrivateKey
	address      common.Address
	funder       common.Address
	chainID      int64
	sigType      int
	orderBuilder builder.ExchangeOrderBuilder
}

// NewBuilder creates an order builder.
//
// funder is the address holding funds (defaults to signer address if empty).
// sigType is the signature type (0=EOA, 1=PolyProxy, 2=GnosisSafe).
func NewBuilder(privateKey *ecdsa.PrivateKey, chainID int64, sigType int, funder string) *Builder {
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	funderAddr := address
	if funder != "" {
		funderAddr = common.HexToAddress(funder)
	}

	return &Builder{
		privateKey:   privateKey,
		address:      address,
		funder:       funderAddr,
		chainID:      chainID,
		sigType:      sigType,
		orderBuilder: builder.NewExchangeOrderBuilderImpl(big.NewInt(chainID), nil),
	}
}

// CreateOrder creates and signs a limit order.
func (b *Builder) CreateOrder(args model.OrderArgs, opts model.CreateOrderOptions) (*gomodel.SignedOrder, error) {
	rc := model.RoundConfigForTickSize(opts.TickSize)

	side, makerAmount, takerAmount, err := GetOrderAmounts(args.Side, args.Size, args.Price, rc)
	if err != nil {
		return nil, err
	}

	return b.buildAndSign(args.TokenID, side, makerAmount, takerAmount,
		args.FeeRateBps, args.Nonce, args.Expiration, args.Taker, opts.NegRisk)
}

// CreateMarketOrder creates and signs a market order.
func (b *Builder) CreateMarketOrder(args model.MarketOrderArgs, opts model.CreateOrderOptions) (*gomodel.SignedOrder, error) {
	rc := model.RoundConfigForTickSize(opts.TickSize)

	side, makerAmount, takerAmount, err := GetMarketOrderAmounts(args.Side, args.Amount, args.Price, rc)
	if err != nil {
		return nil, err
	}

	taker := args.Taker
	if taker == "" {
		taker = model.ZeroAddress.Hex()
	}

	return b.buildAndSign(args.TokenID, side, makerAmount, takerAmount,
		args.FeeRateBps, args.Nonce, 0, taker, opts.NegRisk)
}

func (b *Builder) buildAndSign(
	tokenID string,
	side, makerAmount, takerAmount int,
	feeRateBps, nonce, expiration int,
	taker string,
	negRisk bool,
) (*gomodel.SignedOrder, error) {
	if taker == "" {
		taker = model.ZeroAddress.Hex()
	}

	data := &gomodel.OrderData{
		Maker:         b.funder.Hex(),
		Taker:         taker,
		TokenId:       tokenID,
		MakerAmount:   strconv.Itoa(makerAmount),
		TakerAmount:   strconv.Itoa(takerAmount),
		Side:          side,
		FeeRateBps:    strconv.Itoa(feeRateBps),
		Nonce:         strconv.Itoa(nonce),
		Signer:        b.address.Hex(),
		Expiration:    strconv.Itoa(expiration),
		SignatureType: b.sigType,
	}

	contract := gomodel.CTFExchange
	if negRisk {
		contract = gomodel.NegRiskCTFExchange
	}

	signed, err := b.orderBuilder.BuildSignedOrder(b.privateKey, data, contract)
	if err != nil {
		return nil, fmt.Errorf("build signed order: %w", err)
	}

	return signed, nil
}

// CalculateBuyMarketPrice walks the ask side of the order book to find
// the price at which the given USDC amount can be filled.
func CalculateBuyMarketPrice(asks []model.OrderLevel, amount float64, orderType model.OrderType) (float64, error) {
	if len(asks) == 0 {
		return 0, fmt.Errorf("no asks available")
	}

	cumulative := 0.0
	for i := len(asks) - 1; i >= 0; i-- {
		price, err := strconv.ParseFloat(asks[i].Price, 64)
		if err != nil {
			return 0, fmt.Errorf("parse ask price %q: %w", asks[i].Price, err)
		}
		size, err := strconv.ParseFloat(asks[i].Size, 64)
		if err != nil {
			return 0, fmt.Errorf("parse ask size %q: %w", asks[i].Size, err)
		}
		cumulative += size * price
		if cumulative >= amount {
			return price, nil
		}
	}

	if orderType == model.OrderTypeFOK {
		return 0, fmt.Errorf("insufficient liquidity for FOK order")
	}

	price, _ := strconv.ParseFloat(asks[0].Price, 64)
	return price, nil
}

// CalculateSellMarketPrice walks the bid side of the order book to find
// the price at which the given share amount can be filled.
func CalculateSellMarketPrice(bids []model.OrderLevel, amount float64, orderType model.OrderType) (float64, error) {
	if len(bids) == 0 {
		return 0, fmt.Errorf("no bids available")
	}

	cumulative := 0.0
	for i := len(bids) - 1; i >= 0; i-- {
		size, err := strconv.ParseFloat(bids[i].Size, 64)
		if err != nil {
			return 0, fmt.Errorf("parse bid size %q: %w", bids[i].Size, err)
		}
		cumulative += size
		if cumulative >= amount {
			price, err := strconv.ParseFloat(bids[i].Price, 64)
			if err != nil {
				return 0, fmt.Errorf("parse bid price %q: %w", bids[i].Price, err)
			}
			return price, nil
		}
	}

	if orderType == model.OrderTypeFOK {
		return 0, fmt.Errorf("insufficient liquidity for FOK order")
	}

	price, _ := strconv.ParseFloat(bids[0].Price, 64)
	return price, nil
}

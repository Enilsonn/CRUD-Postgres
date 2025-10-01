package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Enilsonn/CRUD-Postgres/internal/model"
	"github.com/Enilsonn/CRUD-Postgres/internal/repository"
)

var allowedPaymentMethods = map[string]struct{}{
	"CARD":    {},
	"BOLETO":  {},
	"PIX":     {},
	"BERRIES": {},
}

type OrderService struct {
	orders  *repository.OrderRepository
	clients *repository.ClientRepository
	sellers *repository.SellerRepository
	plans   *repository.ProductRepository
	wallets *repository.WalletRepository
}

type OrderItemRequest struct {
	PlanID   int64
	Quantity int
}

type CreateOrderRequest struct {
	ClientID      int64
	SellerID      int64
	PaymentMethod string
	Items         []OrderItemRequest
}

type FinalizeOrderResponse struct {
	Order         *model.Order
	WalletBalance int64
}

func NewOrderService(
	orders *repository.OrderRepository,
	clients *repository.ClientRepository,
	sellers *repository.SellerRepository,
	plans *repository.ProductRepository,
	wallets *repository.WalletRepository,
) *OrderService {
	return &OrderService{
		orders:  orders,
		clients: clients,
		sellers: sellers,
		plans:   plans,
		wallets: wallets,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*model.Order, error) {
	if req.ClientID <= 0 {
		return nil, fmt.Errorf("client_id must be positive")
	}
	if req.SellerID <= 0 {
		return nil, fmt.Errorf("seller_id must be positive")
	}
	if len(req.Items) == 0 {
		return nil, repository.ErrOrderWithoutItems
	}

	paymentMethod := strings.ToUpper(strings.TrimSpace(req.PaymentMethod))
	if _, ok := allowedPaymentMethods[paymentMethod]; !ok {
		return nil, fmt.Errorf("invalid payment method: %s", req.PaymentMethod)
	}

	if _, err := s.clients.GetClientByID(ctx, req.ClientID); err != nil {
		return nil, fmt.Errorf("client lookup failed: %w", err)
	}

	if _, err := s.sellers.GetByID(ctx, req.SellerID); err != nil {
		return nil, fmt.Errorf("seller lookup failed: %w", err)
	}

	order := &model.Order{
		ClientID:      req.ClientID,
		SellerID:      req.SellerID,
		PaymentMethod: paymentMethod,
		Items:         make([]model.OrderItem, len(req.Items)),
	}

	for idx, item := range req.Items {
		if item.PlanID <= 0 {
			return nil, fmt.Errorf("plan_id must be positive")
		}
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("quantity must be positive for plan %d", item.PlanID)
		}

		plan, err := s.plans.GetProductByID(ctx, item.PlanID)
		if err != nil {
			return nil, fmt.Errorf("plan %d retrieval failed: %w", item.PlanID, err)
		}
		if plan.Stock < item.Quantity {
			return nil, fmt.Errorf("plan %d has insufficient stock", item.PlanID)
		}

		order.Items[idx] = model.OrderItem{
			PlanID:   item.PlanID,
			Quantity: item.Quantity,
		}
	}

	if _, err := s.orders.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) FinalizeOrder(ctx context.Context, orderID int64) (*FinalizeOrderResponse, error) {
	if orderID <= 0 {
		return nil, fmt.Errorf("order_id must be positive")
	}

	if err := s.orders.FinalizeOrder(ctx, orderID); err != nil {
		return nil, err
	}

	order, err := s.orders.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	wallet, err := s.wallets.GetWalletByClientID(ctx, order.ClientID)
	if err != nil {
		return nil, fmt.Errorf("fetch wallet after finalize: %w", err)
	}

	return &FinalizeOrderResponse{
		Order:         order,
		WalletBalance: wallet.BalanceCredits,
	}, nil
}

func (s *OrderService) ListOrdersByClient(ctx context.Context, clientID int64) ([]model.Order, error) {
	if clientID <= 0 {
		return nil, fmt.Errorf("client_id must be positive")
	}
	return s.orders.ListOrdersByClient(ctx, clientID)
}

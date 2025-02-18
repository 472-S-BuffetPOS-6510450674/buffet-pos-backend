package usecases

import (
	"context"
	"strconv"

	"github.com/cs471-buffetpos/buffet-pos-backend/configs"
	"github.com/cs471-buffetpos/buffet-pos-backend/domain/exceptions"
	"github.com/cs471-buffetpos/buffet-pos-backend/domain/repositories"
	"github.com/cs471-buffetpos/buffet-pos-backend/domain/requests"
	"github.com/cs471-buffetpos/buffet-pos-backend/domain/responses"
	"golang.org/x/crypto/bcrypt"
)

type CustomerUseCase interface {
	Register(ctx context.Context, req *requests.CustomerRegisterRequest) error
	FindAll(ctx context.Context) ([]responses.BaseCustomer, error)
	AddPoint(ctx context.Context, req *requests.CustomerAddPointRequest) (*responses.BaseCustomer, error)
	RedeemPoint(ctx context.Context, req *requests.CustomerRedeemRequest) (*responses.BaseCustomer, error)
	DeleteCustomer(ctx context.Context, customerID string) error
}

type customerService struct {
	customerRepo   repositories.CustomerRepository
	settingRepo    repositories.SettingRepository
	config         *configs.Config
}

func NewCustomerService(customerRepo repositories.CustomerRepository, settingRepo repositories.SettingRepository, config *configs.Config) CustomerUseCase {
	return &customerService{
		customerRepo:   customerRepo,
		settingRepo: 	settingRepo,
		config:         config,
	}
}

func (c *customerService) Register(ctx context.Context, req *requests.CustomerRegisterRequest) error {
	customer, _ := c.customerRepo.FindByPhone(ctx, req.Phone)

	// Check if customer already exist
	if customer != nil {
		return exceptions.ErrDuplicatedPhone
	}

	// Hash PIN
	hashedPIN, err := bcrypt.GenerateFromPassword([]byte(req.PIN), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Asign PIN
	req.PIN = string(hashedPIN)

	return c.customerRepo.Create(ctx, req)
}

func (c *customerService) FindAll(ctx context.Context) ([]responses.BaseCustomer, error) {
	return c.customerRepo.FindAll(ctx)
}

func (c *customerService) AddPoint(ctx context.Context, req *requests.CustomerAddPointRequest) (*responses.BaseCustomer, error) {
	customer, _ := c.customerRepo.FindByPhone(ctx, req.Phone)

	if customer == nil {
		return nil, exceptions.ErrCustomerNotFound
	}

	// Compare PIN
	if bcrypt.CompareHashAndPassword([]byte(customer.PIN), []byte(req.PIN)) != nil {
		return nil, exceptions.ErrIncorrectPIN
	}

	
	settingPoint, _ := c.settingRepo.GetSetting(ctx, "limitPoint")
	limitPoint, _ := strconv.Atoi(settingPoint.Value)

	// Check point is limit
	if customer.Point >= limitPoint {
		return nil, exceptions.ErrPointLimit
	}
	// Check point is positive number
	if req.Point < 1 {
		return nil, exceptions.ErrInvalidPoint
	}

	return c.customerRepo.AddPoint(ctx, req)
}

func (c *customerService) RedeemPoint(ctx context.Context, req *requests.CustomerRedeemRequest) (*responses.BaseCustomer, error) {
	customer, _ := c.customerRepo.FindByPhone(ctx, req.Phone)

	if customer == nil {
		return nil, exceptions.ErrCustomerNotFound
	}

	// Compare PIN
	if bcrypt.CompareHashAndPassword([]byte(customer.PIN), []byte(req.PIN)) != nil {
		return nil, exceptions.ErrIncorrectPIN
	}
	

	// Check point is enough to redeem
	settingPoint, _ := c.settingRepo.GetSetting(ctx, "usePointPerPerson")
	usePointPerPerson, _ := strconv.Atoi(settingPoint.Value)
	if customer.Point < usePointPerPerson {
		return nil, exceptions.ErrNotEnoughPoints
	}

	return c.customerRepo.RedeemPoint(ctx, req, usePointPerPerson)	
}

func (c *customerService) DeleteCustomer(ctx context.Context, customerID string) error {
	customer, err := c.customerRepo.FindByID(ctx, customerID)
	if err != nil {
		return err
	}

	if customer == nil {
		return exceptions.ErrCustomerNotFound
	}
	return c.customerRepo.Delete(ctx, customerID)
}
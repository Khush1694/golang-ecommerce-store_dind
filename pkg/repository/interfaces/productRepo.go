package interfaces

import (
	"context"

	"github.com/Noush-012/Project-eCommerce-smart_gads/pkg/domain"
	"github.com/Noush-012/Project-eCommerce-smart_gads/pkg/utils/request"
	"github.com/Noush-012/Project-eCommerce-smart_gads/pkg/utils/response"
)

type ProductRepository interface {
	// Product CRUD section
	GetAllProducts(ctx context.Context, page request.ReqPagination) (products []response.ResponseProduct, err error)
	FindProduct(ctx context.Context, product domain.Product) (domain.Product, error)
	SaveProduct(ctx context.Context, product domain.Product) error

	// Brand CRUD section
	FindBrand(ctx context.Context, brand domain.Brand) (domain.Brand, error)
	SaveBrand(ctx context.Context, brand domain.Brand) (err error)
	GetAllBrand(ctx context.Context) ([]response.Brand, error)
}

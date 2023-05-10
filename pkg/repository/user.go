package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Noush-012/Project-eCommerce-smart_gads/pkg/domain"
	repo "github.com/Noush-012/Project-eCommerce-smart_gads/pkg/repository/interfaces"
	"github.com/Noush-012/Project-eCommerce-smart_gads/pkg/utils/request"
	"github.com/Noush-012/Project-eCommerce-smart_gads/pkg/utils/response"
	"gorm.io/gorm"
)

type userDatabase struct {
	DB *gorm.DB
}

func NewUserRepository(DB *gorm.DB) repo.UserRepository {
	return &userDatabase{DB: DB}
}

func (i *userDatabase) FindUser(ctx context.Context, user domain.Users) (domain.Users, error) {
	// Check any of the user details matching with db user list
	query := `SELECT * FROM users WHERE id = ? OR email = ? OR phone = ? OR user_name = ?`
	if err := i.DB.Raw(query, user.ID, user.Email, user.Phone, user.UserName).Scan(&user).Error; err != nil {
		return user, errors.New("failed to get user")
	}
	return user, nil
}

func (i *userDatabase) GetUserbyID(ctx context.Context, userId uint) (domain.Users, error) {
	var user domain.Users
	query := `SELECT * FROM users WHERE id = ?`
	if err := i.DB.Raw(query, userId).Scan(&user).Error; err != nil {
		return user, err
	}
	return user, nil
}

func (i *userDatabase) SaveUser(ctx context.Context, user domain.Users) error {
	query := `INSERT INTO users (user_name, first_name, last_name, age, email, phone, password,created_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	createdAt := time.Now()
	err := i.DB.Exec(query, user.UserName, user.FirstName, user.LastName, user.Age,
		user.Email, user.Phone, user.Password, createdAt).Error
	if err != nil {
		return fmt.Errorf("failed to save user %s", user.UserName)
	}
	return nil
}

func (i *userDatabase) SaveAddress(ctx context.Context, userAddress domain.Address) error {
	var defaultAddressID uint
	query := `INSERT INTO addresses (user_id ,house,address_line1,address_line2,city,state,zip_code,country) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	if err := i.DB.Raw(query, userAddress.UserID, userAddress.House, userAddress.AddressLine1,
		userAddress.AddressLine2, userAddress.City, userAddress.State, userAddress.ZipCode, userAddress.Country).Scan(&defaultAddressID).Error; err != nil {
		return err
	}

	// set default if no existing default address
	query = `INSERT INTO user_addresses (user_id,address_id,is_default)
	SELECT $1, $2, true
	WHERE NOT EXISTS (
  	SELECT 1 FROM user_addresses WHERE user_id = $1 AND is_default = true)`
	if err := i.DB.Exec(query, userAddress.UserID, defaultAddressID).Error; err != nil {
		return err
	}
	return nil
}

func (u *userDatabase) GetAllAddress(ctx context.Context, userId uint) (address []response.Address, err error) {
	query := `SELECT * FROM addresses WHERE user_id = ?`
	if err := u.DB.Raw(query, userId).Scan(&address).Error; err != nil {
		return address, err
	}
	return address, nil
}

func (i *userDatabase) SavetoCart(ctx context.Context, addToCart request.AddToCartReq) error {
	// get product item details
	query := `SELECT discount_price FROM product_items WHERE id = $1`
	if err := i.DB.Raw(query, addToCart.ProductItemID).Scan(&addToCart.Discount_price).Error; err != nil {
		return err
	}

	// get cart id with user id
	query = `SELECT id FROM carts WHERE user_id = $1`
	var cartID int
	if err := i.DB.Raw(query, addToCart.UserID).Scan(&cartID).Error; err != nil {
		return err
	}
	if cartID == 0 {
		// create a cart for user with userID if not exist
		query = `INSERT INTO carts (user_id) VALUES ($1) RETURNING id`
		if err := i.DB.Raw(query, addToCart.UserID).Scan(&cartID).Error; err != nil {
			return err
		}
	}
	// Check if the product item already exist in cart
	query = `SELECT id FROM cart_items WHERE product_item_id = $1 AND cart_id = $2`
	var cartItemID int
	if err := i.DB.Raw(query, addToCart.ProductItemID, cartID).Scan(&cartItemID).Error; err != nil {
		return err
	}
	if cartItemID != 0 {
		query = `UPDATE cart_items SET quantity = quantity + $1, updated_at = $2 WHERE id = $3`
		UpdatedAt := time.Now()
		if err := i.DB.Exec(query, addToCart.Quantity, UpdatedAt, cartItemID).Error; err != nil {
			return fmt.Errorf("failed to save cart item %v", addToCart.ProductItemID)
		}
	} else {
		// insert product items to cart items
		query = `INSERT INTO cart_items (cart_id,product_item_id,quantity,price,created_at)
	VALUES ($1,$2, $3, $4, $5)`
		CreatedAt := time.Now()
		if err := i.DB.Exec(query, cartID, addToCart.ProductItemID, addToCart.Quantity, addToCart.Discount_price, CreatedAt).Error; err != nil {
			return fmt.Errorf("failed to save cart item %v", addToCart.ProductItemID)
		}
	}
	var cartItems []domain.CartItem
	if err := i.DB.Where("cart_id = ?", cartID).Find(&cartItems).Error; err != nil {
		return err
	}
	// Calculate the new total based on the updated cart items
	var total float64
	for _, item := range cartItems {
		total += float64(item.Quantity) * item.Price
	}
	if err := i.DB.Exec("UPDATE carts SET total = $1 WHERE user_id = $2", total, addToCart.UserID).Error; err != nil {
		return err
	}
	return nil
}

func (i *userDatabase) GetCartIdByUserId(ctx context.Context, userId uint) (cartId uint, err error) {
	query := `SELECT id FROM carts WHERE user_id = $1`
	if err := i.DB.Raw(query, userId).Scan(&cartId).Error; err != nil {
		return cartId, err
	}
	return cartId, nil
}

func (i *userDatabase) GetCartItemsbyUserId(ctx context.Context, page request.ReqPagination, userID uint) (CartItems []response.CartItemResp, err error) {

	limit := page.Count
	offset := (page.PageNumber - 1) * limit
	// get cartID by user id
	cartID, err := i.GetCartIdByUserId(ctx, userID)
	if err != nil {
		return CartItems, err
	}
	// get cartItems with cartID
	query := `SELECT ci.product_item_id, p.name,p.price,ci.price AS discount_price, 
	ci.quantity,pi.qty_in_stock AS qty_left, ci.price * ci.quantity AS sub_total
	FROM cart_items ci
	JOIN product_items pi ON ci.product_item_id = pi.id
	JOIN products p ON pi.product_id = p.id
	WHERE cart_id = $1
	ORDER BY ci.created_at DESC LIMIT $2 OFFSET $3`
	if err := i.DB.Raw(query, cartID, limit, offset).Scan(&CartItems).Error; err != nil {
		return CartItems, err
	}
	return CartItems, nil
}

func (i *userDatabase) UpdateCart(ctx context.Context, cartUpadates request.UpdateCartReq) error {

	// get cartID by user id
	cartID, err := i.GetCartIdByUserId(ctx, cartUpadates.UserID)
	if err != nil {
		return err
	}
	// update cart
	query := `UPDATE carts SET
    product_item_id = COALESCE($1, product_item_id),
    quantity = COALESCE($2, quantity)
	WHERE id = $3`
	if err := i.DB.Exec(query, cartUpadates.ProductItemID, cartUpadates.Quantity, cartID).Error; err != nil {
		return err
	}
	return nil
}

func (i *userDatabase) RemoveCartItem(ctx context.Context, DelCartItem request.DeleteCartItemReq) error {
	// get cartID by user id
	cartID, err := i.GetCartIdByUserId(ctx, DelCartItem.UserID)
	if err != nil {
		return err
	}
	// delete cartItems
	query := `DELETE FROM cart_items WHERE cart_id = $1 AND product_item_id = $2`
	if err := i.DB.Exec(query, cartID, DelCartItem.ProductItemID).Error; err != nil {
		return err
	}
	return nil

}
func (i *userDatabase) GetEmailPhoneByUserId(ctx context.Context, userID uint) (contact response.UserContact, err error) {
	// find data
	query := `SELECT email, phone FROM users WHERE id = ?`
	if err := i.DB.Raw(query, userID).Scan(&contact).Error; err != nil {
		return contact, err
	}
	return contact, nil
}

func (i *userDatabase) GetDefaultAddress(ctx context.Context, userId uint) (address response.Address, err error) {
	query := `SELECT a.house, a.address_line1, a.address_line2, a.city, a.state, a.zip_code, a.country
FROM addresses a
JOIN user_addresses ua ON ua.address_id  = a.id
WHERE ua.user_id = ? AND ua.is_default = true`
	if err := i.DB.Raw(query, userId).Scan(&address).Error; err != nil {
		return address, err
	}
	return address, nil
}

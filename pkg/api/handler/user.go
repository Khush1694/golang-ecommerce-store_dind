package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Noush-012/Project-eCommerce-smart_gads/pkg/api/auth"
	"github.com/Noush-012/Project-eCommerce-smart_gads/pkg/domain"
	"github.com/Noush-012/Project-eCommerce-smart_gads/pkg/useCase/interfaces"
	"github.com/Noush-012/Project-eCommerce-smart_gads/pkg/utils"
	request "github.com/Noush-012/Project-eCommerce-smart_gads/pkg/utils/request"

	"github.com/Noush-012/Project-eCommerce-smart_gads/pkg/utils/response"
	"github.com/Noush-012/Project-eCommerce-smart_gads/pkg/verify"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

type UserHandler struct {
	userService interfaces.UserService
}

func NewUserHandler(userUsecase interfaces.UserService) *UserHandler {
	return &UserHandler{userService: userUsecase}
}

// UserSignUp godoc
// @summary api for register user
// @security ApiKeyAuth
// @id UserSignUp
// @tags User Signup
// @Param input body request.SignupUserData{} true "Input Fields"
// @Router /signup [post]
// @Success 200 "Account created successfuly"
// @Failure 400 "invalid entry"
func (u *UserHandler) UserSignup(c *gin.Context) {
	var body request.SignupUserData
	if err := c.ShouldBindJSON(&body); err != nil {
		response := "invalid input"
		c.JSON(http.StatusBadRequest, response)
		return
	}
	//
	var user domain.Users
	// var user domain.Users
	if err := copier.Copy(&user, body); err != nil {
		fmt.Println("Copy failed")
	}

	// Check the user already exist in DB and save user if not
	if err := u.userService.SignUp(c, user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// success response
	c.JSON(http.StatusOK, gin.H{"success": "Account created successfuly"})

}

// LoginSubmit godoc
// @summary api for user login
// @security ApiKeyAuth
// @id UserLogin
// @tags User Login
// @Param input body request.LoginData{} true "Input Fields"
// @Router /login [post]
// @Success 200 {object} response.Response{} "Login successful"
// @Failure 400  {object} response.Response{} "Missing or invalid entry"
// @Failure 500 {object} response.Response{}  "Something went wrong !"
func (u *UserHandler) LoginSubmit(c *gin.Context) {
	var body request.LoginData
	if err := c.ShouldBindJSON(&body); err != nil {
		response := response.ErrorResponse(400, "Missing or invalid entry", err.Error(), body)
		c.JSON(http.StatusBadRequest, response)
		return
	}
	if body.Email == "" && body.Password == "" && body.UserName == "" {
		_ = errors.New("please enter user_name and password")
		response := "Field should not be empty"
		c.JSON(http.StatusBadRequest, response)
		return
	}
	// Copying
	var user domain.Users
	copier.Copy(&user, body)

	dbUser, err := u.userService.Login(c, user)
	if err != nil {
		response := response.ErrorResponse(500, "Something went wrong !", err.Error(), nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := gin.H{"Successfuly send OTP to registered mobile number": dbUser.ID}
	c.JSON(http.StatusOK, response)
}

// OTPVerification godoc
// @summary api for user otp verification
// @security ApiKeyAuth
// @id UserOtpVerify
// @tags User OTP verification
// @Param input body request.OTPVerify{} true "Input Fields"
// @Router /otp-verify [post]
// @Success 200 {object} response.Response{} "Login successful"
// @Failure 400  {object} response.Response{} "Missing or invalid entry"
// @Failure 500 {object} response.Response{}  "failed to send OTP"
func (u *UserHandler) UserOTPVerify(c *gin.Context) {

	var body request.OTPVerify
	if err := c.ShouldBindJSON(&body); err != nil {
		response := response.ErrorResponse(400, "Missing or invalid entry", err.Error(), body)
		c.JSON(http.StatusBadRequest, response)
		return
	}
	var user = domain.Users{
		ID: body.UserID,
	}

	usr, err := u.userService.OTPLogin(c, user)
	if err != nil {
		response := response.ErrorResponse(500, "user not registered", err.Error(), user)
		c.JSON(http.StatusBadRequest, response)
		return
	}
	fmt.Println(body.OTP)
	// Verify otp
	err = verify.TwilioVerifyOTP("+91"+usr.Phone, body.OTP)
	if err != nil {
		response := gin.H{"error": err.Error()}
		c.JSON(http.StatusBadRequest, response)
		return
	}
	// setup JWT
	ok := auth.JwtCookieSetup(c, "user-auth", usr.ID)
	if !ok {
		response := response.ErrorResponse(500, "failed to login", "", nil)
		c.JSON(http.StatusInternalServerError, response)
		return

	}
	response := response.SuccessResponse(200, "Successfuly logged in!", nil)
	c.JSON(http.StatusOK, response)
}

// Home godoc
// @summary api for user home page
// @description after user login user will seen this page with user informations
// @security ApiKeyAuth
// @id User Home
// @tags Home
// @Router / [get]
// @Success 200 "Welcome to home !"
func (u *UserHandler) Home(c *gin.Context) {

	response := response.SuccessResponse(200, "Welcome to home !", nil)
	c.JSON(http.StatusOK, response)
}

// Logout godoc
// @summary api for user to logout
// @description user can logout
// @security ApiKeyAuth
// @id UserLogout
// @tags User Logout
// @Router /logout [post]
// @Success 200 "Log out successful"
func (u *UserHandler) LogoutUser(c *gin.Context) {
	c.SetCookie("user-auth", "", -1, "", "", false, true)
	response := response.SuccessResponse(http.StatusOK, "Log out successful", nil)
	c.JSON(http.StatusOK, response)
}

func (u *UserHandler) GetcartItems(c *gin.Context) {
	var page request.ReqPagination
	count, err0 := utils.StringToUint(c.Query("count"))
	page_number, err1 := utils.StringToUint(c.Query("page_number"))
	err0 = errors.Join(err0, err1)
	if err0 != nil {
		response := response.ErrorResponse(400, "Missing or invalid inputs", err0.Error(), nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}
	page.PageNumber = page_number
	page.Count = count

	userId := utils.GetUserIdFromContext(c)
	cartItems, err := u.userService.GetCartItemsbyCartId(c, page, userId)
	if err != nil {
		response := response.ErrorResponse(500, "Something went wrong!", err.Error(), nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}
	response := response.SuccessResponse(200, "Get Cart Items successful", cartItems)
	c.JSON(http.StatusOK, response)
}

// AddToCart godoc
// @summary api for add product item to user cart
// @description user can add a stock in product to cart
// @security ApiKeyAuth
// @id AddToCart
// @tags User Cart
// @Param input body request.AddToCartReq{} true "Input Field"
// @Router /cart [post]
// @Success 200 "Successfuly added product item to cart "
// @Failure 400 "Failed to add product item in cart"
func (u *UserHandler) AddToCart(c *gin.Context) {
	var body request.AddToCartReq

	if err := c.ShouldBindJSON(&body); err != nil {
		response := response.ErrorResponse(400, "invalid input", err.Error(), body.ProductItemID)
		c.JSON(http.StatusBadRequest, response)
		return
	}
	// get userId from context
	body.UserID = utils.GetUserIdFromContext(c)
	if body.UserID == 0 {
		c.JSON(400, "No user id on context")
		return
	}
	if err := u.userService.SaveCartItem(c, body); err != nil {
		response := response.ErrorResponse(500, "Failed to add product item in cart", err.Error(), body)
		c.JSON(500, response)
		return
	}
	response := response.SuccessResponse(200, "Successfuly added product item to cart ", body)
	c.JSON(200, response)

}

func (u *UserHandler) UpdateCart(c *gin.Context) {
	var body request.UpdateCartReq

	if err := c.ShouldBindJSON(&body); err != nil {
		response := response.ErrorResponse(400, "invalid input", err.Error(), body)
		c.JSON(400, response)
		return
	}
	// get userId from context
	body.UserID = utils.GetUserIdFromContext(c)
	if body.UserID == 0 {
		response := response.ErrorResponse(400, "No user id on context", "", nil)
		c.JSON(400, response)
		return
	}

}

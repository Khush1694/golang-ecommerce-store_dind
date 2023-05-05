package handler

import (
	service "github.com/Noush-012/Project-eCommerce-smart_gads/pkg/useCase/interfaces"
	"github.com/Noush-012/Project-eCommerce-smart_gads/pkg/utils/response"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	PaymentService service.PaymentService
}

func NewPaymentHandler(payUseCase service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		PaymentService: payUseCase,
	}
}

// GetAllPaymentOptions godoc
// @summary api for user to get all options for payment
// @security ApiKeyAuth
// @tags User Payment
// @id AllPayment
// @Success 200 {object} response.Response{} "Payment option successfull"
// @Failure 500 {object} response.Response{}  "Something went wrong!"
func (p *PaymentHandler) GetAllPaymentOptions(c *gin.Context) {
	payOptions, err := p.PaymentService.GetAllPaymentOptions(c)
	if err != nil {
		response := response.ErrorResponse(500, "Something went wrong! ", err.Error(), nil)
		c.JSON(500, response)
	}
	response := response.SuccessResponse(200, "Payment option successfull", payOptions)
	c.JSON(200, response)
}

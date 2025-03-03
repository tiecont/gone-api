package controllers

import (
	"github.com/gin-gonic/gin"
	Users "gone-be/src/modules/user/models"
	verificationservices "gone-be/src/modules/user/services"
	"net/http"
)

// VerifyCode handles the verification code process and token generation
func VerifyCode(c *gin.Context) {
	var code Users.VerificationCode

	// Parse JSON input
	if err := c.ShouldBindJSON(&code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": gin.H{
				"code":    http.StatusBadRequest,
				"message": "Invalid input",
			},
			"error": err.Error(),
		})
		return
	}

	// Call the service to verify the code and generate tokens
	token, err := verificationservices.VerifyCode(code.Code, code.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": gin.H{
				"code":    http.StatusUnauthorized,
				"message": err.Error(),
			},
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Verification successful",
		},
		"data": token,
	})
}

// VerifyCodeAndGenerateToken handles the verification code process and token generation
func VerifyCodeAndGenerateToken(c *gin.Context) {
	var code Users.VerificationCode

	// Parse JSON input
	if err := c.ShouldBindJSON(&code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": gin.H{
				"code":    http.StatusBadRequest,
				"message": "Invalid input",
			},
			"error": err.Error(),
		})
		return
	}

	// Call the service to verify the code and generate tokens
	response, status, err := verificationservices.VerifyCodeAndGenerateTokens(code)
	if err != nil {
		c.JSON(status, gin.H{
			"status": gin.H{
				"code":    status,
				"message": err.Error(),
			},
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Verification successful",
		},
		"data": response,
	})
}

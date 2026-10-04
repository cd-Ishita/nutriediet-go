package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cd-Ishita/nutriediet-go/database"
	"github.com/cd-Ishita/nutriediet-go/model"
	"gorm.io/gorm"

	"github.com/cd-Ishita/nutriediet-go/helpers"
	"github.com/gin-gonic/gin"
)

func Authenticate(c *gin.Context) {
	clientToken := ""
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			fmt.Println("invalid authorization header format")
			c.JSON(http.StatusBadRequest, gin.H{"err": "invalid token format"})
			c.Abort()
			return
		}
		clientToken = parts[1]
	} else if queryToken := c.Query("token"); queryToken != "" {
		// EventSource cannot set Authorization headers; allow token query for SSE.
		clientToken = queryToken
	} else {
		fmt.Println("no authorization header received")
		c.JSON(http.StatusBadRequest, gin.H{"err": "no token received"})
		c.Abort()
		return
	}

	claims, err := helpers.ValidateToken(clientToken)
	if err != nil {
		fmt.Println("error validating the token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": err})
		c.Abort()
		return
	}

	// setting context with this information
	c.Set("email", claims.Email)
	c.Set("first_name", claims.FirstName)
	c.Set("last_name", claims.LastName)
	c.Set("user_type", claims.UserType)
	c.Set("user_id", claims.UserID)

	if claims.UserType == "CLIENT" && !IsClientActive(claims.Email) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "account is inactive", "isActive": false})
		return
	}

	c.Next()
}

func ClientAuthentication(emailFromContext string, clientIDFromReq string) (bool, bool) {
	// To authenticate, fetch the client_id associated with this email id
	db := database.DB
	client := model.Client{}

	err := db.Table("clients").Where("email = ?", emailFromContext).First(&client).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Errorf("error: client with email %s does not exist", emailFromContext)
		return false, false
	} else if err != nil {
		fmt.Errorf("error: could not fetch client with email %s | err: %v", emailFromContext, err)
		return false, false
	}

	if strconv.FormatUint(client.ID, 10) != clientIDFromReq {
		fmt.Errorf("error: client with ID: %d trying to access another clients information", client.ID)
		return false, false
	}

	return true, client.IsActive
}

func IsClientActive(emailFromContext string) bool {
	db := database.DB
	client := model.Client{}
	err := db.Table("clients").Where("email = ?", emailFromContext).First(&client).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Errorf("error: client with email %s does not exist", emailFromContext)
		return false
	} else if err != nil {
		fmt.Errorf("error: could not fetch client with email %s | err: %v", emailFromContext, err)
		return false
	}

	return client.IsActive
}

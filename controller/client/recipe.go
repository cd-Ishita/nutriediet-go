package client

import (
	"net/http"

	"github.com/cd-Ishita/nutriediet-go/database"
	"github.com/cd-Ishita/nutriediet-go/middleware"
	"github.com/cd-Ishita/nutriediet-go/model"
	"github.com/gin-gonic/gin"
)

func GetRecipesForClient(c *gin.Context) {
	clientEmail := c.GetString("email")
	clientID := c.Param("client_id")

	// Authentication check
	isAllowed, isActive := middleware.ClientAuthentication(clientEmail, clientID)
	if !isAllowed {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized access",
			"details": gin.H{
				"clientEmail":     clientEmail,
				"requestClientID": clientID,
			},
		})
		return
	}

	if !isActive {
		c.JSON(http.StatusOK, gin.H{"isActive": false})
		return
	}

	db := database.DB
	var rows []model.Recipe
	if err := db.Where("deleted_at IS NULL").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load recipes"})
		return
	}

	recipeCards := make([]model.RecipeCard, len(rows))
	for i, row := range rows {
		recipeCards[i] = model.ToRecipeCard(row)
	}

	c.JSON(http.StatusOK, gin.H{
		"isActive": true,
		"recipes":  recipeCards,
	})
}

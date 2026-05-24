package admin

import (
	"errors"
	"fmt"
	"github.com/cd-Ishita/nutriediet-go/database"
	"github.com/cd-Ishita/nutriediet-go/helpers"
	"github.com/cd-Ishita/nutriediet-go/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func GetDietTemplatesList(c *gin.Context) {
	if !helpers.CheckUserType(c, "ADMIN") {
		fmt.Errorf("error: client user not allowed to access")
		c.JSON(http.StatusUnauthorized, gin.H{"err": "unauthorized access by client"})
		return
	}

	db := database.DB

	dietTemplates := []model.DietTemplate{}
	err := db.Where("deleted_at IS NULL").Select("id", "name").Order("name ASC").Find(&dietTemplates).Error
	if err != nil {
		fmt.Errorf("error: could not fetch diet templates for GetDietTemplatesList API | err: %v", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}

	fmt.Println(dietTemplates)

	// transform the results into usable format
	var res []struct {
		ID   uint
		Name string
	}

	for _, dietTemplate := range dietTemplates {
		res = append(res, struct {
			ID   uint
			Name string
		}{
			ID:   dietTemplate.ID,
			Name: dietTemplate.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{"list": res})
	return
}

func GetDietTemplateByID(c *gin.Context) {
	if !helpers.CheckUserType(c, "ADMIN") {
		fmt.Errorf("error: client user not allowed to access")
		c.JSON(http.StatusUnauthorized, gin.H{"err": "unauthorized access by client"})
		return
	}

	db := database.DB

	var dietTemplate model.DietTemplate
	err := db.Model(&model.DietTemplate{}).Where("id = ? and deleted_at IS NULL", c.Param("diet_template_id")).Select("diet_string", "name").Find(&dietTemplate).Error
	if err != nil {
		fmt.Errorf("error: could not fetch dietTemplate with id: %s for GetDietTemplateByID | err: %v", c.Param("diet_template_id"), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}

	if dietTemplate.DietString == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no diet"})
	}

	c.JSON(http.StatusOK, gin.H{"name": dietTemplate.Name, "template": dietTemplate.DietString})
}

func CreateDietTemplate(c *gin.Context) {
	if !helpers.CheckUserType(c, "ADMIN") {
		fmt.Errorf("error: client user not allowed to access")
		c.JSON(http.StatusUnauthorized, gin.H{"err": "unauthorized access by client"})
		return
	}

	var template model.CreateDietTemplateRequest
	if err := c.BindJSON(&template); err != nil {
		fmt.Errorf("error: could not extract request from context for CreateDietTemplate | err: %v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.DB
	err := db.Table("diet_templates").Where("deleted_at IS NULL and name = ?", template.Name).First(&model.DietTemplate{}).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// continue
	} else if err != nil {
		fmt.Errorf("error: CreateDietTemplate | could not check for existing diet template with name | err: %v", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		fmt.Errorf("error: CreateDietTemplate | already exists diet template with name: %s", template.Name)
		c.JSON(http.StatusConflict, gin.H{"error": "diet template already exists"})
		return
	}

	dietTemplate := model.DietTemplate{
		Name:       template.Name,
		DietString: &template.Diet,
	}
	err = db.Table("diet_templates").Save(&dietTemplate).Error
	if err != nil {
		fmt.Errorf("error: could not save dietTemplate %v for CreateDietTemplate | err: %v", template, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
	return
}

func UpdateDietTemplate(c *gin.Context) {
	if !helpers.CheckUserType(c, "ADMIN") {
		fmt.Errorf("error: client user not allowed to access")
		c.JSON(http.StatusUnauthorized, gin.H{"err": "unauthorized access by client"})
		return
	}

	var template model.UpdateDietTemplateRequest
	if err := c.BindJSON(&template); err != nil {
		fmt.Errorf("error: could not extract request from context for UpdateDietTemplateByID | err: %v", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dietTemplate := model.DietTemplate{
		Name:       template.Name,
		DietString: &template.Diet,
		ID:         template.ID,
	}
	db := database.DB
	err := db.Table("diet_templates").Where("id = ? and deleted_at IS NULL", c.Param("diet_template_id")).Select("name", "diet_string").Updates(&dietTemplate).Error
	if err != nil {
		fmt.Errorf("error: could not update dietTemplate %v for UpdateDietTemplateByID | err: %v", template, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
	return
}

func DeleteDietTemplateByID(c *gin.Context) {
	if !helpers.CheckUserType(c, "ADMIN") {
		fmt.Errorf("error: client user not allowed to access")
		c.JSON(http.StatusUnauthorized, gin.H{"err": "unauthorized access by client"})
		return
	}

	db := database.DB

	err := db.Table("diet_templates").Where("id = ?", c.Param("diet_template_id")).Update("deleted_at", time.Now()).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Errorf("error: diet template with id %s does not exist in DeleteDietTemplateByID", c.Param("diet_template_id"))
		c.JSON(http.StatusInternalServerError, gin.H{"err": "record not found"})
		return
	} else if err != nil {
		fmt.Errorf("error: diet template with id %s could not be marked deleted in DeleteDietTemplateByID", c.Param("diet_template_id"))
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

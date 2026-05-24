package admin

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cd-Ishita/nutriediet-go/constants"
	"github.com/cd-Ishita/nutriediet-go/database"
	"github.com/cd-Ishita/nutriediet-go/helpers"
	"github.com/cd-Ishita/nutriediet-go/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetDietHistoryForClient(c *gin.Context) {
	if !helpers.CheckUserType(c, "ADMIN") {
		fmt.Errorf("error: client user not allowed to access")
		c.JSON(http.StatusUnauthorized, gin.H{"err": "unauthorized access by client"})
		return
	}

	db := database.DB

	var dietHistory []model.DietHistoryResponse
	err := db.Model(&model.DietHistory{}).
		Joins("left outer join diet_templates on diet_template_id = diet_templates.id").
		Select("diet_histories.*, diet_templates.name as diet_template_name").
		Where("client_id = ? and week_number > 0 and diet_type = ? and diet_histories.deleted_at IS NULL", c.Param("client_id"), constants.RegularDiet.Uint32()).
		Find(&dietHistory).
		Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Errorf("error: diet does not exist for client_id %d", c.Param("client_id"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if err != nil {
		fmt.Errorf("error: could not fetch diet for client_id %s", c.Param("client_id"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// transform the results into given format
	var resRegularDiet []model.DietHistoryResponse
	for _, diet := range dietHistory {
		if diet.DietType == constants.RegularDiet.Uint32() {
			// regular diet
			resRegularDiet = append(resRegularDiet, diet)
		}
	}

	c.JSON(http.StatusOK, gin.H{"diet_history_regular": resRegularDiet})
	return
}

// SaveDietForClient used to store regular diets for clients from the client profiles page for ADMIN
func SaveDietForClient(c *gin.Context) {

	if !helpers.CheckUserType(c, "ADMIN") {
		fmt.Errorf("error: client user not allowed to access")
		c.JSON(http.StatusUnauthorized, gin.H{"err": "unauthorized access by client"})
		return
	}

	// Parse the request body to extract the diet information
	var schedule model.SaveDietForClientRequest
	if err := c.BindJSON(&schedule); err != nil {
		fmt.Errorf("error: could not bind request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if schedule.Diet == "" || schedule.DietType == 0 {
		fmt.Errorf("SaveDietForClient | error: request sent without diet or diet type or diet template id: %v", schedule)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("diet or Diet Type or diet template id not given")})
		return
	}

	db := database.DB

	// only regular diets allowed to use this client specific route
	if schedule.DietType != constants.RegularDiet.Uint32() {
		fmt.Errorf("SaveDietForClient | error: request sent with wrong diet type, only regular diets allowed on this route %v", schedule)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("diet type given is not regular diet")})
		return
	}

	clientID, _ := strconv.ParseUint(c.Param("client_id"), 10, 64)

	// fetch the week_number of the last diet sent
	var weekNumber int
	err := db.Model(&model.DietHistory{}).
		Where("client_id = ? and diet_type = ? and deleted_at IS NULL", clientID, schedule.DietType).
		Select("week_number").
		Order("date DESC").
		Limit(1).
		Find(&weekNumber).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		weekNumber = 0
	} else if err != nil {
		fmt.Errorf("err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dietRecord := model.DietHistory{
		WeekNumber:     weekNumber + 1,
		ClientID:       clientID,
		Date:           time.Now(),
		Weight:         nil,
		DietType:       schedule.DietType,
		DietString:     &schedule.Diet,
		DietTemplateID: schedule.DietTemplateID,
	}
	if err := db.Create(&dietRecord).Error; err != nil {
		fmt.Errorf("error: SaveDietForClient | could not create empty diet_history_id %d for client_id %s | err: %v", schedule.Diet, clientID, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return a success response
	c.JSON(http.StatusCreated, gin.H{"message": "Diet information saved successfully"})
	return
}

func EditDietForClient(c *gin.Context) {

	if !helpers.CheckUserType(c, "ADMIN") {
		fmt.Errorf("error: client user not allowed to access")
		c.JSON(http.StatusUnauthorized, gin.H{"err": "unauthorized access by client"})
		return
	}

	// Parse the request body to extract the diet information
	var schedule model.EditDietForClientRequest
	if err := c.BindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.DB
	clientID, _ := strconv.ParseUint(c.Param("client_id"), 10, 64)

	timeNow := time.Now()
	if err := db.Table("diet_histories").Where("id = ? and diet_type = ? and client_id = ?", schedule.DietID, schedule.DietType, clientID).Updates(map[string]interface{}{
		"diet_string": schedule.Diet,
		"date":        timeNow,
		"updated_at":  timeNow,
	}).Error; err != nil {
		fmt.Errorf("error: SaveDietForClient | could not save diet for diet_history_id %d for client_id %s | err: %v", schedule.Diet, c.Param("client_id"), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return a success response
	c.JSON(http.StatusOK, gin.H{"message": "Diet information saved successfully"})
	return
}



func DeleteDietForClientByAdmin(c *gin.Context) {
	if !helpers.CheckUserType(c, "ADMIN") {
		fmt.Errorf("error: client user not allowed to access")
		c.JSON(http.StatusUnauthorized, gin.H{"err": "unauthorized access by client"})
		return
	}

	clientID := c.Param("client_id")
	if clientID == "" || clientID == "0" {
		fmt.Errorf("error: client_id cannot be empty string")
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id cannot be empty string"})
		return
	}

	// request contains the diet id to be deleted
	req := uint(0)
	if err := c.BindJSON(&req); err != nil {
		fmt.Println("Wrong request, cannot be extracted. For client_id: " + c.Param("client_id"))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.DB
	// verify that the given id exists and is the latest diet of that type
	diet := model.DietHistory{}
	err := db.Where("id = ? and client_id = ? and deleted_at IS NULL", req, clientID).Find(&diet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Errorf("error: could not find diet_history_id %d for client_id %s", req, clientID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if err != nil {
		fmt.Println("Could not retrieve diet record for client_id: " + c.Param("client_id"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// try to see if a more recent diet of that type exists
	var latestDiet model.DietHistory
	err = db.Model(&model.DietHistory{}).
		Where("client_id = ? and diet_type = ? and deleted_at IS NULL", clientID, diet.DietType).
		Order("date DESC, created_at DESC").
		Limit(1).
		First(&latestDiet).
		Error
	if err != nil {
		fmt.Errorf("error: could not find diet_history_id for client_id %s | err: %v", clientID, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if latestDiet.ID != diet.ID {
		fmt.Errorf("error: trying to delete older diet, not allowed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad delete request"})
		return
	}
	timeNow := time.Now()
	diet.DeletedAt = &timeNow

	err = db.Model(&model.DietHistory{}).Where("id = ?", diet.ID).Update("deleted_at", timeNow).Error
	if err != nil {
		fmt.Errorf("error: could not delete diet_history_id for client_id %s | err: %v", clientID, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func SaveCommonDietForClients(c *gin.Context) {

	if !helpers.CheckUserType(c, "ADMIN") {
		fmt.Errorf("error: client user not allowed to access")
		c.JSON(http.StatusUnauthorized, gin.H{"err": "unauthorized access by client"})
		return
	}

	// Parse the request body to extract the diet information
	var req model.SaveCommonDietForClientsRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Diet == "" || req.DietType == 0 {
		fmt.Errorf("SaveCommonDietForClients | error: request sent without diet or diet type or diet template id: %v", req)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("diet or Diet Type or diet template id not given")})
		return
	}

	if req.DietType == constants.RegularDiet.Uint32() {
		fmt.Errorf("SaveCommonDietForClients | error: request sent of type regular diet not allowed on common route: %v", req)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("regular diet type sent on common diets route")})
		return
	}

	db := database.DB
	var createDietReq []model.DietHistory
	// find the week number of this common diet for each group number
	for _, group := range req.Groups {
		// fetch the week number of last common diet of this type sent to this group
		var weekNumber int
		err := db.Model(&model.DietHistory{}).
			Where("group_id = ? and diet_type = ? and deleted_at IS NULL", group, req.DietType).
			Select("week_number").
			Order("date DESC").
			Limit(1).
			Find(&weekNumber).
			Error
		if err != nil && errors.Is(gorm.ErrRecordNotFound, err) {
			weekNumber = 0
		} else if err != nil {
			fmt.Errorf("SaveCommonDietForClients | error : fetching the week number for group %d and diet type %v | err: %v", group, req.DietType, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		createDietReq = append(createDietReq, model.DietHistory{
			WeekNumber:     weekNumber + 1,
			GroupID:        group,
			Date:           time.Now(),
			Weight:         nil,
			DietType:       req.DietType,
			DietString:     &req.Diet,
			DietTemplateID: req.DietTemplateID,
		})
	}

	if err := db.Create(&createDietReq).Error; err != nil {
		fmt.Errorf("error: SaveCommonDietForClients | could not create diets %v for group %v | err: %v", createDietReq, req.Groups, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return a success response
	c.JSON(http.StatusCreated, gin.H{"message": "Diet information saved successfully"})
	return
}

func GetCommonDietsHistory(c *gin.Context) {
	if !helpers.CheckUserType(c, "ADMIN") {
		fmt.Errorf("error: client user not allowed to access")
		c.JSON(http.StatusUnauthorized, gin.H{"err": "unauthorized access by client"})
		return
	}

	db := database.DB

	var dietHistory []model.DietHistoryResponse
	err := db.Model(&model.DietHistory{}).
		Joins("left outer join diet_templates on diet_template_id = diet_templates.id").
		Select("diet_histories.*, diet_templates.name as diet_template_name").
		Where("group_id = ? and week_number > 0 and diet_type != ? and diet_histories.deleted_at IS NULL", c.Param("group_id"), constants.RegularDiet.Uint32()).
		Find(&dietHistory).
		Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Errorf("error: no diets exist does not exist for group_id %s", c.Param("group_id"))
		c.JSON(http.StatusOK, gin.H{"diet_history_detox_diet": dietHistory, "diet_history_detox_water": dietHistory})
		return
	} else if err != nil {
		fmt.Errorf("error: could not fetch diet for client_id %s", c.Param("client_id"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// transform the results into given format
	var resDetoxDiet []model.DietHistoryResponse
	var resDetoxWater []model.DietHistoryResponse
	for _, diet := range dietHistory {
		if diet.DietType == constants.DetoxDiet.Uint32() {
			resDetoxDiet = append(resDetoxDiet, diet)
		} else if diet.DietType == constants.DetoxWater.Uint32() {
			resDetoxWater = append(resDetoxWater, diet)
		}
	}

	c.JSON(http.StatusOK, gin.H{"diet_history_detox_diet": resDetoxDiet, "diet_history_detox_water": resDetoxWater})
	return
}

func EditCommonDiet(c *gin.Context) {

	if !helpers.CheckUserType(c, "ADMIN") {
		fmt.Errorf("error: client user not allowed to access")
		c.JSON(http.StatusUnauthorized, gin.H{"err": "unauthorized access by client"})
		return
	}

	// Parse the request body to extract the diet information
	var schedule model.EditDietForClientRequest
	if err := c.BindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.DB
	groupID, _ := strconv.ParseUint(c.Param("group_id"), 10, 64)

	if schedule.DietType == constants.RegularDiet.Uint32() {
		fmt.Errorf("error: wrong diet type %v | group_id: %d", schedule, groupID)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("wrong diet type")})
		return
	}

	timeNow := time.Now()
	if err := db.Table("diet_histories").Where("id = ? and diet_type = ? and group_id = ?", schedule.DietID, schedule.DietType, groupID).Updates(map[string]interface{}{
		"diet_string": schedule.Diet,
		"date":        timeNow,
		"updated_at":  timeNow,
	}).Error; err != nil {
		fmt.Errorf("error: SaveDietForClient | could not save diet for diet_history_id %d for client_id %s | err: %v", schedule.Diet, c.Param("client_id"), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return a success response
	c.JSON(http.StatusOK, gin.H{"message": "Diet information saved successfully"})
	return
}

func DeleteCommonDiet(c *gin.Context) {
	if !helpers.CheckUserType(c, "ADMIN") {
		fmt.Errorf("error: client user not allowed to access")
		c.JSON(http.StatusUnauthorized, gin.H{"err": "unauthorized access by client"})
		return
	}

	groupID := c.Param("group_id")
	if groupID == "" || groupID == "0" {
		fmt.Errorf("error: client_id cannot be empty string")
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id cannot be empty string"})
		return
	}

	// request contains the diet id to be deleted
	req := uint(0)
	if err := c.BindJSON(&req); err != nil {
		fmt.Println("Wrong request, cannot be extracted. For client_id: " + c.Param("client_id"))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.DB
	// verify that the given id exists and is the latest diet of that type
	diet := model.DietHistory{}
	err := db.Where("id = ? and group_id = ? and deleted_at IS NULL", req, groupID).Find(&diet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Errorf("error: could not find diet_history_id %d for groupID %s", req, groupID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if err != nil {
		fmt.Println("Could not retrieve diet record for groupID: " + c.Param("group_id"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// try to see if a more recent diet of that type exists
	var latestDiet model.DietHistory
	err = db.Model(&model.DietHistory{}).
		Where("group_id = ? and diet_type = ? and deleted_at IS NULL", groupID, diet.DietType).
		Order("date DESC, created_at DESC").
		Limit(1).
		First(&latestDiet).
		Error
	if err != nil {
		fmt.Errorf("error: could not find diet_history_id for group_id %s | err: %v", groupID, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if latestDiet.ID != diet.ID {
		fmt.Errorf("error: trying to delete older diet, not allowed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad delete request"})
		return
	}
	timeNow := time.Now()
	diet.DeletedAt = &timeNow

	err = db.Model(&model.DietHistory{}).Where("id = ?", diet.ID).Update("deleted_at", timeNow).Error
	if err != nil {
		fmt.Errorf("error: could not delete diet_history_id for group_id %s | err: %v", groupID, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

//dietHistoryRecord := model.DietHistory{}
//err := db.Where("client_id = ?", c.Param("client_id")).Order("date DESC").First(&dietHistoryRecord).Error
//if errors.Is(gorm.ErrRecordNotFound, err) {
//	c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
//	return
//} else if err != nil {
//	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//	return
//}

// a new diet always creates a new record in the diet history table
//dietJSON, err := json.Marshal(schedule.Diet)
//if err != nil {
//	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal diet to JSON"})
//	return
//}

//dietHistory := model.DietHistory{
//	ClientID:   clientID,
//	WeekNumber: schedule.WeekNumber,
//	Date:       time.Now(),
//}
//
//// Save the diet history record to the database
//if err := db.Save(&dietHistory).Error; err != nil {
//	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//	return
//}

//if err = db.Table("diet_histories").Where("id = ?", emptyDietRecord.ID).Update("diet", dietJSON).Error; err != nil {
//	fmt.Errorf("error: SaveDietForClient | could not save diet for diet_history_id %d for client_id %s | err: %v", schedule.Diet, clientID, err.Error())
//	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//	return
//}

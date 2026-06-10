package model

import "time"

type DietTemplate struct {
	ID   uint   `gorm:"primaryKey;autoIncrement:true" json:"id"`
	Name string `gorm:"column:name" json:"name,omitempty"`
	//Diet *DietSchedule `gorm:"column:diet;type:json" json:"diet,omitempty"`
	DietString *string `gorm:"column:diet_string;size:2500" json:"diet_string,omitempty"`

	// POSTGRES fields
	//CreatedAt *time.Time   `gorm:"column:created_at;type:timestamp not null;default:CURRENT_TIMESTAMP;" json:"created_at"`
	//UpdatedAt *time.Time   `gorm:"column:updated_at;type:timestamp not null;default:CURRENT_TIMESTAMP;" json:"updated_at"`
	//DeletedAt *time.Time   `gorm:"column:deleted_at;type:timestamp;default:NULL;" json:"deleted_at,omitempty"`

	CreatedAt *time.Time `gorm:"column:created_at;type:datetime not null;default:CURRENT_TIMESTAMP;" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:datetime not null;default:CURRENT_TIMESTAMP;" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:datetime;default:NULL;omitempty;" json:"deleted_at,omitempty"`
}

type DietHistory struct {
	ID             uint      `gorm:"primaryKey;autoIncrement:true" json:"id"`
	ClientID       uint64    `gorm:"column:client_id;foreignKey:ClientID" json:"client_id,omitempty"`
	GroupID        int       `gorm:"column:group_id" json:"group_id,omitempty"`
	WeekNumber     int       `gorm:"column:week_number" json:"week_number,omitempty"`
	Date           time.Time `gorm:"column:date" json:"date,omitempty"`
	Weight         *float32  `gorm:"column:weight" json:"weight,omitempty"`
	DietString     *string   `gorm:"column:diet_string;size:2500" json:"diet_string,omitempty"`
	Feedback       string    `gorm:"column:feedback" json:"feedback,omitempty"`
	Tags           string    `gorm:"column:tags" json:"tags,omitempty"`
	DietType       uint32    `gorm:"column:diet_type" json:"diet_type,omitempty"`
	DietTemplateID uint      `gorm:"column:diet_template_id" json:"diet_template_id,omitempty"`

	CreatedAt *time.Time `gorm:"column:created_at;type:datetime not null;default:CURRENT_TIMESTAMP;" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:datetime not null;default:CURRENT_TIMESTAMP;" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:datetime;default:NULL;omitempty;" json:"deleted_at,omitempty"`
}

type Item struct {
	ID          int    `json:"ID,omitempty"`
	Name        string `json:"Name,omitempty"`
	Quantity    string `json:"Quantity,omitempty"`
	Preparation string `json:"Preparation,omitempty"`
	Consumption string `json:"Consumption,omitempty"`
	HasRecipe   bool   `json:"HasRecipe,omitempty"`
}

type Meal struct {
	Timing      string `json:"Timing"`
	Primary     []Item `json:"Primary"`
	Alternative []Item `json:"Alternative"`
}

type DietSchedule struct {
	OnWakingUp Meal `json:"On Waking Up"`
	Breakfast  Meal `json:"Breakfast"`
	MidMorning Meal `json:"MidMorning"`
	Lunch      Meal `json:"Lunch"`
	Evening    Meal `json:"Dinner"`
	Night      Meal `json:"Night"`
}

type SaveDietForClientRequest struct {
	//Diet       DietSchedule
	Diet           string `json:"diet,omitempty"`
	WeekNumber     int    `json:"week_number,omitempty"`
	DietType       uint32 `json:"diet_type,omitempty"`
	DietTemplateID uint   `json:"diet_template_id,omitempty"`
}

type EditDietForClientRequest struct {
	DietID   uint   `json:"diet_id,omitempty"`
	Diet     string `json:"diet,omitempty"`
	DietType uint32 `json:"diet_type,omitempty"`
}

type CreateDietTemplateRequest struct {
	//Diet DietSchedule `json:"diet,omitempty"`
	Diet string `json:"diet,omitempty""`
	Name string `json:"name,omitempty"`
}

type UpdateDietTemplateRequest struct {
	ID   uint   `json:"id,omitempty"`
	Diet string `json:"diet,omitempty"`
	Name string `json:"name,omitempty"`
}

type GetDietHistoryForClientResponse struct {
	DietID     uint   `json:"diet_id"`
	WeekNumber int    `json:"week_number,omitempty"`
	Diet       string `json:"diet,omitempty"`
}

type SaveCommonDietForClientsRequest struct {
	Diet           string `json:"diet,omitempty"`
	DietType       uint32 `json:"diet_type,omitempty"`
	Groups         []int  `json:"groups,omitempty"`
	DietTemplateID uint   `json:"diet_template_id,omitempty"`
}

type DietHistoryResponse struct {
	ID         uint      `gorm:"column:id" json:"id"`
	ClientID   uint64    `gorm:"column:client_id" json:"client_id,omitempty"`
	GroupID    int       `gorm:"column:group_id" json:"group_id,omitempty"`
	WeekNumber int       `gorm:"column:week_number" json:"week_number,omitempty"`
	Date       time.Time `gorm:"column:date" json:"date,omitempty"`
	Weight     *float32  `gorm:"column:weight" json:"weight,omitempty"`

	DietString       *string `gorm:"column:diet_string;size:2500" json:"diet_string,omitempty"`
	Feedback         string  `gorm:"column:feedback" json:"feedback,omitempty"`
	Tags             string  `gorm:"column:tags" json:"tags,omitempty"`
	DietType         uint32  `gorm:"column:diet_type" json:"diet_type,omitempty"`
	DietTemplateID   uint    `gorm:"column:diet_template_id" json:"diet_template_id,omitempty"`
	DietTemplateName string  `gorm:"column:diet_template_name" json:"name,omitempty"`

	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at,omitempty"`
}

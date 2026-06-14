package routes

import (
	"github.com/cd-Ishita/nutriediet-go/controller"
	clientController "github.com/cd-Ishita/nutriediet-go/controller/client"
	"github.com/cd-Ishita/nutriediet-go/middleware"
	"github.com/gin-gonic/gin"

	adminController "github.com/cd-Ishita/nutriediet-go/controller/admin"
)

func UserRoutes(incomingRoutes gin.IRouter) {
	// Rate limiting for authenticated API endpoints (100 requests/minute)
	apiRateLimit := middleware.RateLimitAPI()

	// Scoped protected group — middleware ONLY applies to routes inside this group,
	// not to public routes (login, signup, etc.) registered on the engine elsewhere.
	protected := incomingRoutes.Group("/")
	protected.Use(apiRateLimit, middleware.Authenticate)

	// USER ROUTES
	protected.GET("/users", controller.GetUsers) // DEPRECATED
	protected.GET("/user/:user_id", controller.GetUser) // DEPRECATED

	// <<<<<<<<===============================================================================>>>>>>
	// CLIENT ROUTES (Prefix with `/clients` for all client-related routes)

	// CLIENT - WEIGHT UPDATE
	protected.POST("/clients/:client_id/weight_update", clientController.UpdateWeightForClient)
	protected.GET("/clients/:client_id/weight_update", clientController.WeightUpdationStatus)
	protected.GET("/clients/:client_id/weight-history", clientController.GetWeightHistoryForClient)

	// CLIENT - DIET
	protected.GET("/clients/:client_id/diet", clientController.GetDietsForClient)

	// CLIENT - EXERCISE
	protected.GET("/clients/:client_id/exercise", clientController.GetExercisesForClient)
	protected.POST("/clients/:client_id/exercise/favorite", clientController.ToggleFavoriteExercise)

	// CLIENT - PROFILE
	protected.POST("/clients/:client_id/my_profile", clientController.UpdateProfileByClient)
	protected.GET("/clients/:client_id/my_profile", clientController.GetProfileForClient)
	protected.GET("/clients/:client_id/profile_created", clientController.HasClientCreatedProfile)

	// CLIENT - RECIPE
	protected.GET("/clients/:client_id/recipe", clientController.GetRecipeImageForClients)

	// CLIENT - MOTIVATION
	protected.GET("/clients/:client_id/motivation", clientController.GetActiveMotivationsForClients)

	// EMAIL-BASED PROFILE CREATION (Separate from client routes to avoid conflicts)
	protected.POST("/users/:email/create_profile", clientController.CreateProfileByClient)

	// <<<<<<<<===============================================================================>>>>>>
	// ADMIN ROUTES (Prefix with `/admin` for all admin-related routes)

	protected.GET("/admin/clients", adminController.GetAllClients)
	protected.GET("/admin/client/:client_id", adminController.GetClientInfo)
	protected.POST("/admin/client/:client_id", adminController.UpdateClientInfo)
	protected.POST("/admin/client/:client_id/activation", adminController.ActivateOrDeactivateClientAccount)
	protected.GET("/admin/client/:client_id/weight_history", adminController.GetWeightHistoryForClient)
	protected.GET("/admin/client/:client_id/diet_history", adminController.GetDietHistoryForClient)

	// ADMIN - DIET
	protected.GET("/admin/meal_list", adminController.GetMealList) // DEPRECATED
	protected.GET("/admin/quantity_list", adminController.GetQuantityList) // DEPRECATED
	protected.POST("/admin/:client_id/diet", adminController.SaveDietForClient)
	protected.POST("/admin/:client_id/edit_diet", adminController.EditDietForClient)
	protected.POST("/admin/:client_id/weight_update", adminController.UpdateWeightForClientByAdmin)
	protected.POST("/admin/:client_id/delete_diet", adminController.DeleteDietForClientByAdmin)
	protected.POST("/admin/common_diet", adminController.SaveCommonDietForClients)
	protected.GET("/admin/common_diet/:group_id", adminController.GetCommonDietsHistory)
	protected.POST("/admin/common_diet/:group_id/update", adminController.EditCommonDiet)
	protected.POST("/admin/common_diet/:group_id/delete_diet", adminController.DeleteCommonDiet)

	// <<<<<<<<===============================================================================>>>>>>

	// ADMIN - DIET TEMPLATES
	protected.GET("/admin/diet_templates", adminController.GetDietTemplatesList)
	protected.GET("/admin/diet_templates/:diet_template_id", adminController.GetDietTemplateByID)
	protected.POST("/admin/diet_templates/new", adminController.CreateDietTemplate)
	protected.POST("/admin/diet_templates/:diet_template_id", adminController.UpdateDietTemplate)
	protected.POST("/admin/diet_templates/:diet_template_id/delete", adminController.DeleteDietTemplateByID)

	// ADMIN - RECIPES
	//protected.GET("/admin/recipe/:id", adminController.GetRecipeByID)
	//protected.POST("/admin/recipe/:id", adminController.UpdateRecipeByID)
	//protected.POST("/admin/recipe/new", adminController.CreateRecipe)
	//protected.POST("/admin/recipe/:id/delete", adminController.DeleteRecipeByID)
	protected.GET("/admin/recipes", adminController.GetListOfRecipes)
	protected.POST("/admin/recipes/upload", adminController.UploadRecipeImage)
	protected.GET("/admin/recipes/:recipe_id", adminController.GetRecipeImageForAdmin)
	protected.POST("/admin/recipes/:recipe_id/update", adminController.UpdateRecipeImageByAdmin)
	protected.POST("/admin/recipes/:recipe_id/delete", adminController.DeleteRecipeImageByAdmin)

	// ADMIN - EXERCISES
	protected.GET("/admin/exercises", adminController.GetListOfExercises)
	protected.GET("/admin/exercise/:exercise_id", adminController.GetExerciseByID)
	protected.POST("/admin/exercise/new", adminController.CreateExercise)
	protected.POST("/admin/exercise/:exercise_id", adminController.UpdateExerciseByID)
	protected.POST("/admin/exercise/:exercise_id/delete", adminController.DeleteExerciseByID)

	// ADMIN - MOTIVATION
	protected.POST("/admin/motivations/new", adminController.CreateNewMotivation)
	protected.POST("/admin/motivation/:motivation_id/unpost", adminController.UnpostMotivation)
	protected.POST("/admin/motivation/:motivation_id/post", adminController.PostMotivation)
	protected.GET("/admin/motivation", adminController.GetAllMotivations)

	// <<<<<<<<===============================================================================>>>>>>
	// ADMIN - USER MANAGEMENT (Protected routes)

	protected.GET("/admin/users", controller.GetUsers)

	// ADMIN - EXERCISE MANAGEMENT (Previously unprotected in main.go - now secured)
	protected.GET("/admin/exercises/all", controller.GetExercisesForAdmin)
	protected.GET("/admin/exercises/detail/:exercise_id", controller.GetExercise)
	protected.POST("/admin/exercises/:exercise_id/delete", controller.RemoveExerciseFromList)
	protected.POST("/admin/exercises/:exercise_id/update", controller.UpdateExerciseFromList)
	protected.POST("/admin/exercises/submit", controller.AddExerciseFromList)
}

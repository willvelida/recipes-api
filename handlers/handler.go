package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"willvelida/recipes-api/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecipesHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection) *RecipesHandler {
	return &RecipesHandler{
		collection: collection,
		ctx:        ctx,
	}
}

func (handler *RecipesHandler) NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := handler.collection.InsertOne(handler.ctx, recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new recipe"})
		return
	}
	c.JSON(http.StatusOK, recipe)
}

// swagger:operation GET /recipes recipes listRecipes
// Returns list of recipes
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//		description: Successful operation
func (handler *RecipesHandler) ListRecipesHandler(c *gin.Context) {
	cur, err := handler.collection.Find(handler.ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(handler.ctx)
	recipes := make([]models.Recipe, 0)
	for cur.Next(handler.ctx) {
		var recipe models.Recipe
		cur.Decode(&recipe)
		recipes = append(recipes, recipe)
	}
	c.JSON(http.StatusOK, recipes)
}

// swagger:operation PUT /recipes/{id} recipes updateRecipe
// Update an existing recipe
// ---
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// produces:
// - application/json
// responses:
//
//	'200':
//		description: Successful operation
//	'400':
//		description: Invalid input
//	'404':
//		description: Invalid recipe ID
func (handler *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.UpdateOne(handler.ctx, bson.M{
		"_id": objectId,
	}, bson.D{{"$set", bson.D{
		{"name", recipe.Name},
		{"instructions", recipe.Instructions},
		{"ingredients", recipe.Ingredients},
		{"tags", recipe.Tags},
	}}})
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been updated"})
}

func (handler *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.DeleteOne(handler.ctx, bson.M{
		"_id": objectId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been deleted",
	})
}

func (handler *RecipesHandler) GetOneRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	cur := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": objectId,
	})
	var recipe models.Recipe
	err := cur.Decode(&recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

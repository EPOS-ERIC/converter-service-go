package routes

import (
	"errors"
	"net/http"

	"github.com/epos-eu/converter-service/connection"
	"github.com/epos-eu/converter-service/dao/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetAllPluginRelations retrieves all plugin relations from the database
//
//	@Summary		Get all plugin relations
//	@Description	Retrieve all plugin relations from the database
//	@Tags			plugin-relations
//	@Produce		json
//	@Success		200	{array}		model.PluginRelation
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugin-relations [get]
func GetAllPluginRelations(c *gin.Context) {
	plugins, err := connection.GetPluginRelation()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(plugins) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No plugin relation found"})
		return
	}

	c.JSON(http.StatusOK, plugins)
}

// GetPluginRelation retrieves a plugin relation from the database
//
//	@Summary		Get a plugin relation
//	@Description	Retrieve a plugin relation from the database
//	@Tags			plugin-relations
//	@Produce		json
//	@Param			id	path		string	true	"Plugin Relation ID"
//	@Success		200	{object}	model.PluginRelation
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugin-relations/{id} [get]
func GetPluginRelation(c *gin.Context) {
	id := c.Param("id")
	plugin, err := connection.GetPluginById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No plugin found with id: " + id})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plugin)
}

// UpdatePluginRelation updates a plugin relation in the database
//
//	@Summary		Update a plugin relation
//	@Description	Update an existing plugin relation in the database. Even if explicitly passed in the body, the Id of the plugin relation will not be changed
//	@Tags			plugin-relations
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string			true	"Plugin Relation ID"
//	@Param			plugin	body		model.PluginRelation	true	"PluginRelation object"
//	@Success		200		{object}	model.PluginRelation
//	@Failure		400		{object}	HTTPError
//	@Failure		500		{object}	HTTPError
//	@Router			/plugin-relations/{id} [put]
func UpdatePluginRelation(c *gin.Context) {
	id := c.Param("id")

	var relation model.PluginRelation
	if err := c.ShouldBindJSON(&relation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure the id of the relation is correct (prevent clients from changing it)
	relation.ID = id

	// Validate the relation
	if err := relation.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := connection.UpdatePluginRelation(id, relation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, relation)
}

// DeletePluginRelation deletes a plugin relation from the database
//
//	@Summary		Delete a plugin relation
//	@Description	Delete a plugin relation from the database
//	@Tags			plugin-relations
//	@Produce		json
//	@Param			id	path		string	true	"Plugin Relation ID"
//	@Success		200	{object}	model.PluginRelation
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugin-relations/{id} [delete]
func DeletePluginRelation(c *gin.Context) {
	id := c.Param("id")

	deletedRelation, err := connection.DeletePluginRelation(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plugin relation not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deletedRelation)
}

// CreatePluginRelation creates a new plugin relation in the database
//
//	@Summary		Create a new plugin relation
//	@Description	Create a new plugin relation in the database. The plugin relation ID will be assigned upon creation.
//	@Tags			plugin-relations
//	@Accept			json
//	@Produce		json
//	@Param			plugin	body		model.PluginRelation	true	"PluginRelation object"
//	@Success		201		{object}	model.PluginRelation
//	@Failure		400		{object}	HTTPError
//	@Failure		500		{object}	HTTPError
//	@Router			/plugin-relations [post]
func CreatePluginRelation(c *gin.Context) {
	var relation model.PluginRelation
	if err := c.ShouldBindJSON(&relation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate an ID for the plugin relation
	relation.ID = uuid.New().String()

	// Validate the relation
	if err := relation.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdPlugin, err := connection.CreatePluginRelation(relation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdPlugin)
}

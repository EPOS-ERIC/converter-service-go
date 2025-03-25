package routes

import (
	"errors"
	"net/http"

	"github.com/epos-eu/converter-service/connection"
	"github.com/epos-eu/converter-service/dao/model"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetAllPluginRelations Retrieve all plugins from the database
//
//	@Summary		Get all plugin relations
//	@Description	Retrieve all plugin relations from the database
//	@Tags			plugin-relations
//	@Produce		json
//	@Success		200	{array}		model.PluginRelation
//	@Failure		204	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugin-relations [get]
func GetAllPluginRelations(c *gin.Context) {
	plugins, err := connection.GetPluginRelation()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if len(plugins) == 0 {
		c.String(http.StatusNoContent, "No plugin relation found")
		return
	}

	c.JSON(http.StatusOK, plugins)
}

// GetPluginRelations Retrieve a plugin relation from the database
//
//	@Summary		Get a plugin relation
//	@Description	Retrieve a plugin relation from the database
//	@Tags			plugin-relations
//	@Produce		json
//	@Param			id	path		string	true	"Plugin Relation ID"
//	@Success		200	{object}	model.PluginRelation
//	@Failure		204	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugin-relations/{id} [get]
func GetPluginRelations(c *gin.Context) {
	id := c.Param("id")
	plugin, err := connection.GetPluginById(id)
	if err != nil {
		if err == pg.ErrNoRows {
			c.String(http.StatusNoContent, "No plugin found with id: %s", id)
			return
		} else {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}
	c.JSON(http.StatusOK, plugin)
}

// UpdatePluginRelation Update a plugin relation in the database
//
//	@Summary		Update a plugin relation
//	@Description	Update an existing plugin relation in the database
//	@Tags			plugin-relations
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string			true	"Plugin ID"
//	@Param			plugin	body		model.PluginRelation	true	"PluginRelation object"
//	@Success		200		{object}	model.PluginRelation
//	@Failure		400		{object}	HTTPError
//	@Failure		500		{object}	HTTPError
//	@Router			/plugins-relations/{id} [put]
func UpdatePluginRelation(c *gin.Context) {
	id := c.Param("id")

	var relation model.PluginRelation
	if err := c.ShouldBindJSON(&relation); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// make sure the id of the relation is correct (do not allow users to change it)
	relation.ID = id

	// to update just one attribute of the pluginRelation, the whole relation (as json) must be passed in the body of the request, otherwise the validation will not work
	err := relation.Validate()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	err = connection.UpdatePluginRelation(id, relation)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, relation)
}

// DeletePluginRelation Deletes a plugin relation from the database
//
//	@Summary		Delete a plugin relation
//	@Description	Delete a plugin relation from the database
//	@Tags			plugin-relations
//	@Produce		json
//	@Param			id	path		string	true	"Plugin ID"
//	@Success		200	{object}	model.PluginRelation
//	@Failure		204	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugin-relations/{id} [delete]
func DeletePluginRelation(c *gin.Context) {
	id := c.Param("id")

	deletedRelation, err := connection.DeletePluginRelation(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, deletedRelation)
}

// CreatePluginRelation Create a new plugin relation in the database
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
//	@Router			/plugin-relation [post]
func CreatePluginRelation(c *gin.Context) {
	var relation model.PluginRelation
	if err := c.ShouldBindJSON(&relation); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// Generate an id for the plugin
	relation.ID = uuid.New().String()

	// if the relation type is empty assume we are talking about operations
	if relation.RelationType == "" {
		relation.RelationType = "operation"
	}

	err := relation.Validate()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	createdPlugin, err := connection.CreatePluginRelation(relation)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, createdPlugin)
}

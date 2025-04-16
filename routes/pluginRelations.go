package routes

import (
	"errors"
	"net/http"

	"github.com/epos-eu/converter-service/connection"
	"github.com/epos-eu/converter-service/dao/model"
	"github.com/epos-eu/converter-service/loggers"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetAllPluginRelations retrieves all plugin relations from the database
//
//	@Summary		Get all plugin relations
//	@Description	Retrieve all plugin relations from the database
//	@Tags			Converter Service
//	@Produce		json
//	@Success		200	{array}		model.PluginRelation
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugin-relations [get]
func GetAllPluginRelations(c *gin.Context) {
	loggers.API_LOGGER.Debug("GetAllPluginRelations request received")

	plugins, err := connection.GetPluginRelation()
	if err != nil {
		loggers.API_LOGGER.Error("Failed to get plugin relations from DB", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve plugin relations"})
		return
	}

	if len(plugins) == 0 {
		loggers.API_LOGGER.Warn("No plugin relations found in DB")
		c.JSON(http.StatusNotFound, gin.H{"error": "No plugin relation found"})
		return
	}

	loggers.API_LOGGER.Debug("GetAllPluginRelations request successful", "count", len(plugins))
	c.JSON(http.StatusOK, plugins)
}

// GetPluginRelation retrieves a plugin relation from the database
//
//	@Summary		Get a plugin relation
//	@Description	Retrieve a plugin relation from the database
//	@Tags			Converter Service
//	@Produce		json
//	@Param			relation_id	path		string	true	"Plugin Relation ID"
//	@Success		200			{object}	model.PluginRelation
//	@Failure		404			{object}	HTTPError
//	@Failure		500			{object}	HTTPError
//	@Router			/plugin-relations/{relation_id} [get]
func GetPluginRelation(c *gin.Context) {
	id := c.Param("relation_id")
	loggers.API_LOGGER.Debug("GetPluginRelation request received", "relation_id", id)

	plugin, err := connection.GetPluginRelationById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			loggers.API_LOGGER.Warn("Plugin relation not found in DB", "relation_id", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "No plugin relation found with relation_id: " + id})
			return
		}
		loggers.API_LOGGER.Error("Failed to get plugin relation from DB", "relation_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve plugin relation"})
		return
	}

	loggers.API_LOGGER.Debug("GetPluginRelation request successful", "relation_id", id)
	c.JSON(http.StatusOK, plugin)
}

type PluginRelationUpdate struct {
	PluginID     *string `json:"plugin_id"`
	RelationID   *string `json:"relation_id"`
	InputFormat  *string `json:"input_format"`
	OutputFormat *string `json:"output_format"`
}

// UpdatePluginRelation updates a plugin relation in the database
//
//	@Summary		Update a plugin relation
//	@Description	Update an existing plugin relation in the database. Even if explicitly passed in the body, the Id of the plugin relation will not be changed
//	@Tags			Converter Service
//	@Accept			json
//	@Produce		json
//	@Param			relation_id	path		string					true	"Plugin Relation ID"
//	@Param			plugin		body		PluginRelationUpdate	true	"PluginRelation object"
//	@Success		200			{object}	model.PluginRelation
//	@Failure		400			{object}	HTTPError
//	@Failure		404			{object}	HTTPError
//	@Failure		500			{object}	HTTPError
//	@Router			/plugin-relations/{relation_id} [put]
func UpdatePluginRelation(c *gin.Context) {
	id := c.Param("relation_id")
	loggers.API_LOGGER.Debug("UpdatePluginRelation request received", "relation_id", id)

	var relationUpdate PluginRelationUpdate
	if err := c.ShouldBindJSON(&relationUpdate); err != nil {
		loggers.API_LOGGER.Warn("Failed to bind JSON for plugin relation update", "relation_id", id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format: " + err.Error()})
		return
	}

	// get current relation
	relation, err := connection.GetPluginRelationById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			loggers.API_LOGGER.Warn("Plugin relation to update not found in DB", "relation_id", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "No plugin relation found with relation_id: " + id})
			return
		}
		loggers.API_LOGGER.Error("Failed to get plugin relation for update", "relation_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve existing plugin relation"})
		return
	}

	// merge and validate
	newRelation := mergePluginRelationUpdate(relationUpdate, relation)
	if err := newRelation.Validate(); err != nil {
		loggers.API_LOGGER.Warn("Plugin relation validation failed on update", "relation_id", id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
		return
	}

	// update (using the merged and validated 'newRelation')
	err = connection.UpdatePluginRelation(id, newRelation)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// This case might be redundant if GetPluginRelationById succeeded earlier, but keep for safety
			loggers.API_LOGGER.Warn("Plugin relation vanished before update completed", "relation_id", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "No plugin relation found with relation_id: " + id})
			return
		}
		loggers.API_LOGGER.Error("Failed to update plugin relation in DB", "relation_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save plugin relation update"})
		return
	}

	loggers.API_LOGGER.Info("Plugin relation updated successfully", "relation_id", id)
	c.JSON(http.StatusOK, relation)
}

// DeletePluginRelation deletes a plugin relation from the database
//
//	@Summary		Delete a plugin relation
//	@Description	Delete a plugin relation from the database
//	@Tags			Converter Service
//	@Produce		json
//	@Param			relation_id	path		string	true	"Plugin Relation ID"
//	@Success		200			{object}	model.PluginRelation
//	@Failure		404			{object}	HTTPError
//	@Failure		500			{object}	HTTPError
//	@Router			/plugin-relations/{relation_id} [delete]
func DeletePluginRelation(c *gin.Context) {
	id := c.Param("relation_id")
	loggers.API_LOGGER.Debug("DeletePluginRelation request received", "relation_id", id)

	deletedRelation, err := connection.DeletePluginRelation(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			loggers.API_LOGGER.Warn("Plugin relation to delete not found in DB", "relation_id", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "Plugin relation not found"})
			return
		}
		loggers.API_LOGGER.Error("Failed to delete plugin relation from DB", "relation_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete plugin relation"})
		return
	}

	loggers.API_LOGGER.Info("Plugin relation deleted successfully", "relation_id", deletedRelation.ID)
	c.JSON(http.StatusOK, deletedRelation)
}

// CreatePluginRelation creates a new plugin relation in the database
//
//	@Summary		Create a new plugin relation
//	@Description	Create a new plugin relation in the database. The plugin relation ID will be assigned upon creation.
//	@Tags			Converter Service
//	@Accept			json
//	@Produce		json
//	@Param			plugin_relation	body		PluginRelationUpdate	true	"PluginRelation object"
//	@Success		201				{object}	model.PluginRelation
//	@Failure		400				{object}	HTTPError
//	@Failure		500				{object}	HTTPError
//	@Router			/plugin-relations [post]
func CreatePluginRelation(c *gin.Context) {
	loggers.API_LOGGER.Debug("CreatePluginRelation request received")

	var newRelationData PluginRelationUpdate
	if err := c.ShouldBindJSON(&newRelationData); err != nil {
		loggers.API_LOGGER.Warn("Failed to bind JSON for plugin relation creation", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format: " + err.Error()})
		return
	}

	// Prepare the model.PluginRelation for DB
	relationToCreate := mergePluginRelationUpdate(newRelationData, model.PluginRelation{
		ID: uuid.NewString(),
	})

	// Validate the relation
	if err := relationToCreate.Validate(); err != nil {
		loggers.API_LOGGER.Warn("Plugin relation validation failed on create", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
		return
	}

	// Create in DB
	createdRelation, err := connection.CreatePluginRelation(relationToCreate)
	if err != nil {
		// Add duplicate check if necessary based on DB constraints
		loggers.API_LOGGER.Error("Failed to create plugin relation in DB", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save new plugin relation"})
		return
	}

	loggers.API_LOGGER.Info("Plugin relation created successfully", "relation_id", createdRelation.ID)
	c.JSON(http.StatusCreated, createdRelation)
}

// mergePluginRelationUpdate takes the update payload and the existing relation data,
// returning a new Plugin struct representing the merged state.
// Fields are updated only if the corresponding pointer in 'update' is not nil.
func mergePluginRelationUpdate(update PluginRelationUpdate, old model.PluginRelation) model.PluginRelation {
	merged := old

	// explicitly ignoring the id, using the old one

	if update.InputFormat != nil {
		merged.InputFormat = *update.InputFormat
	}
	if update.OutputFormat != nil {
		merged.OutputFormat = *update.OutputFormat
	}
	if update.PluginID != nil {
		merged.PluginID = *update.PluginID
	}
	if update.RelationID != nil {
		merged.RelationID = *update.RelationID
	}

	return merged
}

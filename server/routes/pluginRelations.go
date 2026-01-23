package routes

import (
	"errors"
	"net/http"

	"github.com/epos-eu/converter-service/dao/model"
	"github.com/epos-eu/converter-service/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DistributionInfo struct {
	InstanceID string                      `json:"instance_id"`
	Relations  []PluginWithRelationDetails `json:"relations"`
}

type PluginWithRelationDetails struct {
	Plugin   model.Plugin         `json:"plugin"`
	Relation model.PluginRelation `json:"relation"`
}

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
	log.Debug("GetAllPluginRelations request received")

	plugins, err := db.GetAllPluginRelations()
	if err != nil {
		log.Error("Failed to get plugin relations from DB", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve plugin relations"})
		return
	}

	if len(plugins) == 0 {
		log.Warn("No plugin relations found in DB")
		c.JSON(http.StatusNotFound, gin.H{"error": "No plugin relation found"})
		return
	}

	log.Debug("GetAllPluginRelations request successful", "count", len(plugins))
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
	log.Debug("GetPluginRelation request received", "relation_id", id)

	plugin, err := db.GetPluginRelationByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("Plugin relation not found in DB", "relation_id", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "No plugin relation found with relation_id: " + id})
			return
		}
		log.Error("Failed to get plugin relation from DB", "relation_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve plugin relation"})
		return
	}

	log.Debug("GetPluginRelation request successful", "relation_id", id)
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
//	@Param			relation_id		path		string					true	"Plugin Relation ID"
//	@Param			relation_update	body		PluginRelationUpdate	true	"PluginRelation object"
//	@Success		200				{object}	model.PluginRelation
//	@Failure		400				{object}	HTTPError
//	@Failure		404				{object}	HTTPError
//	@Failure		500				{object}	HTTPError
//	@Router			/plugin-relations/{relation_id} [put]
func UpdatePluginRelation(c *gin.Context) {
	id := c.Param("relation_id")
	log.Debug("UpdatePluginRelation request received", "relation_id", id)

	var relationUpdate PluginRelationUpdate
	if err := c.ShouldBindJSON(&relationUpdate); err != nil {
		log.Warn("Failed to bind JSON for plugin relation update", "relation_id", id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format: " + err.Error()})
		return
	}

	// get current relation
	relation, err := db.GetPluginRelationByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("Plugin relation to update not found in DB", "relation_id", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "No plugin relation found with relation_id: " + id})
			return
		}
		log.Error("Failed to get plugin relation for update", "relation_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve existing plugin relation"})
		return
	}

	// merge and validate
	newRelation := mergePluginRelationUpdate(relationUpdate, relation)
	if err := newRelation.Validate(); err != nil {
		log.Warn("Plugin relation validation failed on update", "relation_id", id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
		return
	}

	// update (using the merged and validated 'newRelation')
	err = db.UpdatePluginRelation(newRelation)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// This case might be redundant if GetPluginRelationById succeeded earlier, but keep for safety
			log.Warn("Plugin relation vanished before update completed", "relation_id", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "No plugin relation found with relation_id: " + id})
			return
		}
		log.Error("Failed to update plugin relation in DB", "relation_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save plugin relation update"})
		return
	}

	log.Info("Plugin relation updated successfully", "relation_id", id)
	c.JSON(http.StatusOK, newRelation)
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
	log.Debug("DeletePluginRelation request received", "relation_id", id)

	deletedRelation, err := db.DeletePluginRelation(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("Plugin relation to delete not found in DB", "relation_id", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "Plugin relation not found"})
			return
		}
		log.Error("Failed to delete plugin relation from DB", "relation_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete plugin relation"})
		return
	}

	log.Info("Plugin relation deleted successfully", "relation_id", deletedRelation.ID)
	c.JSON(http.StatusOK, deletedRelation)
}

// DeleteRelationsByDistributionID deletes all plugin relations for a given distribution instanceID
//
//	@Summary		Delete all relations for a distribution
//	@Description	Delete all plugin relations associated with a specific distribution instanceID
//	@Tags			Converter Service
//	@Produce		json
//	@Param			relation_id	path		string	true	"Instance ID of the Distribution"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		500			{object}	HTTPError
//	@Router			/plugin-relations/distribution/{relation_id} [delete]
func DeleteRelationsByDistributionID(c *gin.Context) {
	distributionID := c.Param("relation_id")
	log.Debug("DeleteRelationsByDistributionID request received", "distribution_id", distributionID)

	deletedCount, err := db.DeletePluginRelationsByRelationID(distributionID)
	if err != nil {
		log.Error("Failed to delete plugin relations from DB", "distribution_id", distributionID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete plugin relations"})
		return
	}

	log.Info("Plugin relations deleted successfully", "distribution_id", distributionID, "count", deletedCount)
	c.JSON(http.StatusOK, gin.H{"deleted_count": deletedCount})
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
	log.Debug("CreatePluginRelation request received")

	var newRelationData PluginRelationUpdate
	if err := c.ShouldBindJSON(&newRelationData); err != nil {
		log.Warn("Failed to bind JSON for plugin relation creation", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format: " + err.Error()})
		return
	}

	// Prepare the model.PluginRelation for DB
	relationToCreate := mergePluginRelationUpdate(newRelationData, model.PluginRelation{
		ID: uuid.NewString(),
	})

	// Validate the relation
	if err := relationToCreate.Validate(); err != nil {
		log.Warn("Plugin relation validation failed on create", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
		return
	}

	// Create in DB
	createdRelation, err := db.CreatePluginRelation(relationToCreate)
	if err != nil {
		log.Error("Failed to create plugin relation in DB", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save new plugin relation"})
		return
	}

	log.Info("Plugin relation created successfully", "relation_id", createdRelation.ID)
	c.JSON(http.StatusCreated, createdRelation)
}

// GetDistributionByInstanceID retrieves all plugins and their relations for a given distribution instance ID
//
//	@Summary		Get distribution by instance ID
//	@Description	Retrieve all plugins and their relations for a specific distribution instance
//	@Tags			Converter Service
//	@Produce		json
//	@Param			instance_id	path		string	true	"Distribution Instance ID"
//	@Success		200			{object}	DistributionInfo
//	@Failure		500			{object}	HTTPError
//	@Router			/distributions/{instance_id} [get]
func GetDistributionByInstanceID(c *gin.Context) {
	instanceID := c.Param("instance_id")
	log.Debug("GetDistributionByInstanceID request received", "instance_id", instanceID)

	relations, err := db.GetPluginRelationsByRelationID(instanceID)
	if err != nil {
		log.Error("Failed to get plugin relations from DB", "instance_id", instanceID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve distribution"})
		return
	}

	distributionInfo := DistributionInfo{
		InstanceID: instanceID,
		Relations:  make([]PluginWithRelationDetails, 0),
	}

	for _, rel := range relations {
		plugin, err := db.GetPluginByID(rel.PluginID)
		if err != nil {
			log.Warn("Plugin not found for relation", "plugin_id", rel.PluginID, "relation_id", rel.ID)
			continue
		}

		distributionInfo.Relations = append(distributionInfo.Relations, PluginWithRelationDetails{
			Plugin:   plugin,
			Relation: rel,
		})
	}

	log.Debug("GetDistributionByInstanceID request successful", "instance_id", instanceID, "count", len(distributionInfo.Relations))
	c.JSON(http.StatusOK, distributionInfo)
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

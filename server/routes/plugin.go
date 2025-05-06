package routes

import (
	"errors"
	"net/http"

	"github.com/epos-eu/converter-service/dao/model"
	"github.com/epos-eu/converter-service/db"
	"github.com/epos-eu/converter-service/routine"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// HTTPError is used just by swag
type HTTPError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"status bad request"`
}

// GetAllPlugins retrieves all plugins from the database
//
//	@Summary		Get all plugins
//	@Description	Retrieve all plugins from the database
//	@Tags			Converter Service
//	@Produce		json
//	@Success		200	{array}		model.Plugin
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugins [get]
func GetAllPlugins(c *gin.Context) {
	logger.Debug("GetAllPlugins request received")

	plugins, err := db.GetPlugins()
	if err != nil {
		logger.Error("Failed to get plugins from DB", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve plugins"})
		return
	}

	if len(plugins) == 0 {
		logger.Warn("No plugins found in DB")
		c.JSON(http.StatusNotFound, gin.H{"error": "No plugins found"})
		return
	}

	logger.Debug("GetAllPlugins request successful", "count", len(plugins))
	c.JSON(http.StatusOK, plugins)
}

// GetPlugin retrieves a plugin from the database
//
//	@Summary		Get a plugin
//	@Description	Retrieve a plugin from the database
//	@Tags			Converter Service
//	@Produce		json
//	@Param			plugin_id	path		string	true	"Plugin ID"
//	@Success		200			{object}	model.Plugin
//	@Failure		404			{object}	HTTPError
//	@Failure		500			{object}	HTTPError
//	@Router			/plugins/{plugin_id} [get]
func GetPlugin(c *gin.Context) {
	id := c.Param("plugin_id")
	logger.Debug("GetPlugin request received", "plugin_id", id)

	plugin, err := db.GetPluginById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("Plugin not found in DB", "plugin_id", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "No plugin found with plugin_id: " + id})
			return
		}
		logger.Error("Failed to get plugin from DB", "plugin_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve plugin"})
		return
	}

	logger.Debug("GetPlugin request successful", "plugin_id", id)
	c.JSON(http.StatusOK, plugin)
}

type Plugin struct {
	Name        *string                  `json:"name"`
	Description *string                  `json:"description"`
	Version     *string                  `json:"version"`
	VersionType *model.VersionType       `json:"version_type"`
	Repository  *string                  `json:"repository"`
	Runtime     *model.SupportedRuntimes `json:"runtime"`
	Executable  *string                  `json:"executable"`
	Arguments   *string                  `json:"arguments"`
	Enabled     *bool                    `json:"enabled"`
}

// UpdatePlugin updates a plugin in the database
//
//	@Summary		Update a plugin
//	@Description	Update an existing plugin in the database. Even if explicitly passed in the body, the Id of the plugin will not be changed
//	@Tags			Converter Service
//	@Accept			json
//	@Produce		json
//	@Param			plugin_id	path		string	true	"Plugin ID"
//	@Param			plugin		body		Plugin	true	"Plugin object"
//	@Success		200			{object}	model.Plugin
//	@Success		202			{object}	model.Plugin "Plugin created in DB. Initial sync failed, will be retried by background task."
//	@Failure		400			{object}	HTTPError
//	@Failure		404			{object}	HTTPError
//	@Failure		500			{object}	HTTPError
//	@Router			/plugins/{plugin_id} [put]
func UpdatePlugin(c *gin.Context) {
	id := c.Param("plugin_id")
	logger.Debug("UpdatePlugin request received", "plugin_id", id)

	var pluginUpdate Plugin
	if err := c.ShouldBindJSON(&pluginUpdate); err != nil {
		logger.Warn("Failed to bind JSON for plugin update", "plugin_id", id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format: " + err.Error()})
		return
	}

	// get the current version of this plugin
	plugin, err := db.GetPluginById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("Plugin to update not found in DB", "plugin_id", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "No plugin found with plugin_id: " + id})
			return
		}
		logger.Error("Failed to get plugin for update", "plugin_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve existing plugin"})
		return
	}

	// merge the two to make a new complete plugin with the new updates
	updatedPlugin := mergePluginUpdate(pluginUpdate, plugin)
	if err = updatedPlugin.Validate(); err != nil {
		logger.Warn("Plugin validation failed on update", "plugin_id", id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
		return
	}

	// update the plugin in the db
	if err := db.UpdatePlugin(updatedPlugin); err != nil {
		logger.Error("Failed to update plugin in DB", "plugin_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save plugin update"})
		return
	}

	logger.Debug("Plugin DB record updated successfully", "plugin_id", updatedPlugin.ID)

	// clean and sync if necessary
	needsSync := pluginUpdate.Version != nil || pluginUpdate.VersionType != nil || pluginUpdate.Repository != nil
	if needsSync && updatedPlugin.Installed {
		logger.Debug("Plugin update requires clean/sync", "plugin_id", updatedPlugin.ID)
		err := routine.Clean(updatedPlugin.ID)
		if err != nil {
			// Plugin updated successfully, but failed during post-update clean step
			logger.Error("Post-update clean step failed", "plugin_id", updatedPlugin.ID, "error", err)
			c.JSON(http.StatusAccepted, updatedPlugin)
			return
		}

		err = routine.SyncPlugin(updatedPlugin.ID)
		if err != nil {
			// Plugin updated successfully, but failed during post-update sync step
			logger.Warn("Post-update sync step failed", "plugin_id", updatedPlugin.ID, "error", err)
			c.JSON(http.StatusAccepted, updatedPlugin)
			return
		}
		logger.Debug("Post-update clean and sync successful", "plugin_id", updatedPlugin.ID)
	}

	logger.Info("Plugin updated successfully", "plugin_id", updatedPlugin.ID, "sync_required", needsSync)
	c.JSON(http.StatusOK, updatedPlugin)
}

// DeletePlugin deletes a plugin from the database
//
//	@Summary		Delete a plugin
//	@Description	Delete a plugin from the database
//	@Tags			Converter Service
//	@Produce		json
//	@Param			plugin_id	path		string	true	"Plugin ID"
//	@Success		200			{object}	model.Plugin
//	@Failure		404			{object}	HTTPError
//	@Failure		500			{object}	HTTPError
//	@Router			/plugins/{plugin_id} [delete]
func DeletePlugin(c *gin.Context) {
	id := c.Param("plugin_id")
	logger.Debug("DeletePlugin request received", "plugin_id", id)

	// Delete the plugin from the database
	deletedPlugin, err := db.DeletePlugin(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("Plugin to delete not found in DB", "plugin_id", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
			return
		}
		logger.Error("Failed to delete plugin from DB", "plugin_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete plugin from database"})
		return
	}

	// NOTE: see if this is ok or if it is better to clean manually
	// the plugin dir will be deleted by the cron task automatically

	logger.Info("Plugin deleted successfully", "plugin_id", deletedPlugin.ID)
	c.JSON(http.StatusOK, deletedPlugin)
}

// CreatePlugin creates a new plugin in the database
//
//	@Summary		Create a new plugin
//	@Description	Create a new plugin in the database. The plugin ID will be assigned upon creation.
//	@Tags			Converter Service
//	@Accept			json
//	@Produce		json
//	@Param			plugin	body		Plugin			true	"Plugin object for creation"
//	@Success		201		{object}	model.Plugin	"Plugin created in DB. Sync succeded."
//	@Success		202		{object}	model.Plugin	"Plugin created in DB. Initial sync failed, will be retried by background task."
//	@Failure		400		{object}	HTTPError
//	@Failure		500		{object}	HTTPError
//	@Router			/plugins [post]
func CreatePlugin(c *gin.Context) {
	logger.Debug("CreatePlugin request received")

	var newPlugin Plugin
	if err := c.ShouldBindJSON(&newPlugin); err != nil {
		logger.Warn("Failed to bind JSON for plugin creation", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format: " + err.Error()})
		return
	}

	// Prepare the model.Plugin for DB, generating ID etc.
	pluginToCreate := mergePluginUpdate(newPlugin, model.Plugin{
		ID:        uuid.NewString(),
		Installed: false,
	})

	if err := pluginToCreate.Validate(); err != nil {
		logger.Warn("Plugin validation failed on create", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
		return
	}

	// Create in DB
	createdPlugin, err := db.CreatePlugin(pluginToCreate)
	if err != nil {
		logger.Error("Failed to create plugin in DB", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save new plugin"})
		return
	}

	// try sync
	err = routine.SyncPlugin(createdPlugin.ID)
	if err != nil {
		// handle the sync error
		logger.Warn("Initial SyncPlugin failed after DB creation. Will rely on cron task.", "plugin_id", createdPlugin.ID, "error", err)
		// sending accepted to say that the request was accepted but the installation is not complete
		c.JSON(http.StatusAccepted, createdPlugin)
		return
	}

	logger.Info("Plugin created successfully", "plugin_id", createdPlugin.ID)
	c.JSON(http.StatusCreated, createdPlugin)
}

// mergePluginUpdate takes the update payload (routes.Plugin) and the existing plugin data (model.Plugin),
// returning a new model.Plugin struct representing the merged state.
// Fields are updated only if the corresponding pointer in 'update' is not nil.
func mergePluginUpdate(update Plugin, old model.Plugin) model.Plugin {
	merged := old

	// explicitly ignoring id and installed, we will use the old values

	// Apply updates field by field if the pointer in 'update' is not nil
	if update.Name != nil {
		merged.Name = *update.Name
	}
	if update.Description != nil {
		merged.Description = *update.Description
	}
	if update.Version != nil {
		merged.Version = *update.Version
	}
	if update.VersionType != nil {
		merged.VersionType = *update.VersionType
	}
	if update.Repository != nil {
		merged.Repository = *update.Repository
	}
	if update.Runtime != nil {
		merged.Runtime = *update.Runtime
	}
	if update.Executable != nil {
		merged.Executable = *update.Executable
	}
	if update.Arguments != nil {
		merged.Arguments = *update.Arguments
	}
	if update.Enabled != nil {
		merged.Enabled = *update.Enabled
	}

	return merged
}

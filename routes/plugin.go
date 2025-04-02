package routes

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/epos-eu/converter-service/connection"
	"github.com/epos-eu/converter-service/dao/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const PluginsPath = "./plugins/"

// HTTPError is used just by swag
type HTTPError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"status bad request"`
}

// GetAllPlugins retrieves all plugins from the database
//
//	@Summary		Get all plugins
//	@Description	Retrieve all plugins from the database
//	@Tags			plugins
//	@Produce		json
//	@Success		200	{array}		model.Plugin
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugins [get]
func GetAllPlugins(c *gin.Context) {
	plugins, err := connection.GetPlugins()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(plugins) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No plugins found"})
		return
	}

	c.JSON(http.StatusOK, plugins)
}

// GetPlugin retrieves a plugin from the database
//
//	@Summary		Get a plugin
//	@Description	Retrieve a plugin from the database
//	@Tags			plugins
//	@Produce		json
//	@Param			id	path		string	true	"Plugin ID"
//	@Success		200	{object}	model.Plugin
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugins/{id} [get]
func GetPlugin(c *gin.Context) {
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

// UpdatePlugin updates a plugin in the database
//
//	@Summary		Update a plugin
//	@Description	Update an existing plugin in the database. Even if explicitly passed in the body, the Id of the plugin will not be changed
//	@Tags			plugins
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"Plugin ID"
//	@Param			plugin	body		model.Plugin	true	"Plugin object"
//	@Success		200	{object}	model.Plugin
//	@Failure		400	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugins/{id} [put]
func UpdatePlugin(c *gin.Context) {
	id := c.Param("id")

	var plugin model.Plugin
	if err := c.ShouldBindJSON(&plugin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := connection.UpdatePlugin(id, plugin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plugin)
}

// DeletePlugin deletes a plugin from the database
//
//	@Summary		Delete a plugin
//	@Description	Delete a plugin from the database
//	@Tags			plugins
//	@Produce		json
//	@Param			id	path		string	true	"Plugin ID"
//	@Success		200	{object}	model.Plugin
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugins/{id} [delete]
func DeletePlugin(c *gin.Context) {
	id := c.Param("id")

	// Delete the plugin from the database
	deletedPlugin, err := connection.DeletePlugin(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Delete the plugin directory
	err = os.RemoveAll(path.Join(PluginsPath, deletedPlugin.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error cleaning plugin directory %s: %v", deletedPlugin.ID, err)})
		return
	}

	// Delete the relations related to this plugin
	err = connection.DeletePluginRelationsForPlugin(deletedPlugin.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error deleting plugin relations related to plugin: %v", err)})
		return
	}

	c.JSON(http.StatusOK, deletedPlugin)
}

// CreatePlugin creates a new plugin in the database
//
//	@Summary		Create a new plugin
//	@Description	Create a new plugin in the database. The plugin ID will be assigned upon creation.
//	@Tags			plugins
//	@Accept			json
//	@Produce		json
//	@Param			plugin	body		model.Plugin	true	"Plugin object"
//	@Success		201	{object}	model.Plugin
//	@Failure		400	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugins [post]
func CreatePlugin(c *gin.Context) {
	var plugin model.Plugin
	if err := c.ShouldBindJSON(&plugin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate an ID for the plugin
	plugin.ID = uuid.New().String()
	// By default, the plugin is not installed (it will be when the routine installs it)
	plugin.Installed = false

	// Validate the plugin
	if err := plugin.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdPlugin, err := connection.CreatePlugin(plugin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdPlugin)
}

package routes

import (
	"errors"
	"net/http"
	"os"
	"path"

	"github.com/epos-eu/converter-service/connection"
	"github.com/epos-eu/converter-service/dao/model"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const PluginsPath = "./plugins/"

// HTTPError Used just by swag
type HTTPError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"status bad request"`
}

// GetAllPlugins Retrieve all plugins from the database
//
//	@Summary		Get all plugins
//	@Description	Retrieve all plugins from the database
//	@Tags			plugins
//	@Produce		json
//	@Success		200	{array}		model.Plugin
//	@Failure		204	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugins [get]
func GetAllPlugins(c *gin.Context) {
	plugins, err := connection.GetPlugins()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if len(plugins) == 0 {
		c.AbortWithError(http.StatusNoContent, err)
		return
	}

	c.JSON(http.StatusOK, plugins)
}

// GetPlugin Retrieve a plugin from the database
//
//	@Summary		Get a plugin
//	@Description	Retrieve a plugin from the database
//	@Tags			plugins
//	@Produce		json
//	@Param			id	path		string	true	"Plugin ID"
//	@Success		200	{object}	model.Plugin
//	@Failure		204	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugins/{id} [get]
func GetPlugin(c *gin.Context) {
	id := c.Param("id")
	plugin, err := connection.GetPluginById(id)
	if err != nil {
		if err == pg.ErrNoRows {
			c.AbortWithError(http.StatusNoContent, err)
			return
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}
	c.JSON(http.StatusOK, plugin)
}

// UpdatePlugin Update a plugin in the database
//
//	@Summary		Update a plugin
//	@Description	Update an existing plugin in the database
//	@Tags			plugins
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string			true	"Plugin ID"
//	@Param			plugin	body		model.Plugin	true	"Plugin object"
//	@Success		200		{object}	model.Plugin
//	@Failure		400		{object}	HTTPError
//	@Failure		500		{object}	HTTPError
//	@Router			/plugins/{id} [put]
func UpdatePlugin(c *gin.Context) {
	id := c.Param("id")

	var plugin model.Plugin
	if err := c.ShouldBindJSON(&plugin); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err := connection.UpdatePlugin(id, plugin)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, plugin)
}

// DeletePlugin Delete a plugin from the database
//
//	@Summary		Delete a plugin
//	@Description	Delete a plugin from the database
//	@Tags			plugins
//	@Produce		json
//	@Param			id	path		string	true	"Plugin ID"
//	@Success		200	{object}	model.Plugin
//	@Failure		204	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugins/{id} [delete]
func DeletePlugin(c *gin.Context) {
	id := c.Param("id")

	// delete the plugin from the db
	deletedPlugin, err := connection.DeletePlugin(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// delete the plugin directory
	err = os.RemoveAll(path.Join(PluginsPath, deletedPlugin.ID))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error cleaning plugin directory %s: %w", deletedPlugin.ID, err)
		return
	}

	// delete the relations related to this plugin
	err = connection.DeletePluginRelationsForPlugin(deletedPlugin.ID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error deleting plugin relations related to plugin %w", err)
		return
	}

	c.JSON(http.StatusOK, deletedPlugin)
}

// CreatePlugin Create a new plugin in the database
//
//	@Summary		Create a new plugin
//	@Description	Create a new plugin in the database. The plugin ID will be assigned upon creation.
//	@Tags			plugins
//	@Accept			json
//	@Produce		json
//	@Param			plugin	body		model.Plugin	true	"Plugin object"
//	@Success		201		{object}	model.Plugin
//	@Failure		400		{object}	HTTPError
//	@Failure		500		{object}	HTTPError
//	@Router			/plugin [post]
func CreatePlugin(c *gin.Context) {
	var plugin model.Plugin
	if err := c.ShouldBindJSON(&plugin); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// Generate an id for the plugin
	plugin.ID = uuid.New().String()
	// By default the plugin is not installed (it will be when the routine installs it)
	plugin.Installed = false
	// plugin.Runtime = strings.ToLower(plugin.Runtime)

	// Make sure that the plugin given in the post is valid
	err := plugin.Validate()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	createdPlugin, err := connection.CreatePlugin(plugin)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, createdPlugin)
}

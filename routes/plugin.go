package routes

import (
	"github.com/epos-eu/converter-service/connection"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"net/http"
)

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
//	@Success		200	{array}		orms.Plugin
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
//	@Success		200	{object}	orms.Plugin
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

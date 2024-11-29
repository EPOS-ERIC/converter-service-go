package routes

import (
	"github.com/epos-eu/converter-service/connection"
	"github.com/gin-gonic/gin"
	"net/http"
)

// EnablePlugin Enable a plugin
//
//	@Summary		Enable a plugin
//	@Description	Enable a plugin by its ID
//	@Tags			plugins
//	@Produce		json
//	@Param			id	path		string	true	"Plugin ID"
//	@Success		200	{string}	string	"Plugin {id} enabled correctly"
//	@Failure		500	{object}	HTTPError
//	@Router			/plugins/{id}/enable [post]
func EnablePlugin(c *gin.Context) {
	id := c.Param("id")

	err := connection.EnablePlugin(id, true)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusOK, "Plugin "+id+" enabled correctly")
}

// DisablePlugin Disable a plugin
//
//	@Summary		Disable a plugin
//	@Description	Disable a plugin by its ID
//	@Tags			plugins
//	@Produce		json
//	@Param			id	path		string	true	"Plugin ID"
//	@Success		200	{string}	string	"Plugin {id} disabled correctly"
//	@Failure		500	{object}	HTTPError
//	@Router			/plugins/{id}/disable [post]
func DisablePlugin(c *gin.Context) {
	id := c.Param("id")

	err := connection.EnablePlugin(id, false)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusOK, "Plugin "+id+" disabled correctly")
}

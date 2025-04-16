package routes

import (
	"net/http"

	"github.com/epos-eu/converter-service/connection"
	"github.com/gin-gonic/gin"
)

// EnablePlugin enables a plugin by its ID.
//
//	@Summary		Enable a plugin
//	@Description	Enables a plugin, making it available for use, by setting its enabled state to true.
//	@Tags			Converter Service
//	@Accept			json
//	@Produce		json
//	@Param			plugin_id	path		string		true	"Plugin ID"
//	@Success		200			{string}	string		"Plugin {plugin_id} enabled correctly"
//	@Failure		500			{object}	HTTPError	"Internal Server Error"
//	@Router			/plugins/{plugin_id}/enable [post]
func EnablePlugin(c *gin.Context) {
	id := c.Param("plugin_id")

	err := connection.EnablePlugin(id, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "Plugin "+id+" enabled correctly")
}

// DisablePlugin disables a plugin by its ID.
//
//	@Summary		Disable a plugin
//	@Description	Disables a plugin, preventing its usage, by setting its enabled state to false.
//	@Tags			Converter Service
//	@Accept			json
//	@Produce		json
//	@Param			plugin_id	path		string		true	"Plugin ID"
//	@Success		200			{string}	string		"Plugin {plugin_id} disabled correctly"
//	@Failure		500			{object}	HTTPError	"Internal Server Error"
//	@Router			/plugins/{plugin_id}/disable [post]
func DisablePlugin(c *gin.Context) {
	id := c.Param("plugin_id")

	err := connection.EnablePlugin(id, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "Plugin "+id+" disabled correctly")
}

package routes

import (
	"github.com/epos-eu/converter-service/connection"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"net/http"
)

// GetAllPluginRelations Retrieve all plugins from the database
//
//	@Summary		Get all plugin relations
//	@Description	Retrieve all plugin relations from the database
//	@Tags			plugin-relations
//	@Produce		json
//	@Success		200	{array}		orms.PluginRelations
//	@Failure		204	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugin-relations [get]
func GetAllPluginRelations(c *gin.Context) {
	plugins, err := connection.GetPluginRelations()
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

// GetPluginRelations Retrieve a plugin relation from the database
//
//	@Summary		Get a plugin relation
//	@Description	Retrieve a plugin relation from the database
//	@Tags			plugin-relations
//	@Produce		json
//	@Param			id	path		string	true	"Plugin Relation ID"
//	@Success		200	{object}	orms.PluginRelations
//	@Failure		204	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/plugin-relations/{id} [get]
func GetPluginRelations(c *gin.Context) {
	id := c.Param("id")
	plugin, err := connection.GetPluginRelationsById(id)
	if err != nil {
		if err == pg.ErrNoRows {
			c.AbortWithError(http.StatusNoContent, err)
			return
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		return
	}
	c.JSON(http.StatusOK, plugin)
}

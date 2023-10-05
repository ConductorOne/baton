package explorer

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) GetResourceHandler(c *gin.Context) {
	resource, err := ctrl.baton.GetResourceById(c.Request.Context(), c.Param("resourceType"), c.Param("resourceId"))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resource})
}

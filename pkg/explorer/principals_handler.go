package explorer

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) GetResourcesWithPrincipalCountHandler(c *gin.Context) {
	resourceType := c.Param("resourceType")
	resources, err := ctrl.baton.GetResourcesWithPrincipalCount(c.Request.Context(), resourceType)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resources})
}

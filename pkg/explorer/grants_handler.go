package explorer

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) GetGrantsForResourceHandler(c *gin.Context) {
	grants, err := ctrl.baton.GetAccessForResource(c.Request.Context(), c.Param("resourceType"), c.Param("resourceId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"data": grants})
}

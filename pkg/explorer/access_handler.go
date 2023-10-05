package explorer

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) GetAccessHandler(c *gin.Context) {
	access, err := ctrl.baton.GetAccess(c.Request.Context(), c.Param("resourceType"), c.Param("resourceId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": access})
}

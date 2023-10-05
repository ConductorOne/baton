package explorer

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) GetResourceTypesHandler(c *gin.Context) {
	resourceTypes, err := ctrl.baton.GetResourceTypes(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resourceTypes})
}

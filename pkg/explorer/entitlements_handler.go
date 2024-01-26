package explorer

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) GetEntitlementsHandler(c *gin.Context) {
	entitlements, err := ctrl.baton.GetEntitlements(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entitlements})
}

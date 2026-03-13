package explorer

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) GetEntitlementsHandler(c *gin.Context) {
	pageToken := c.Query("page_token")

	entitlements, nextToken, err := ctrl.baton.GetEntitlements(c.Request.Context(), pageToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := gin.H{"data": entitlements}
	if nextToken != "" {
		resp["next_page_token"] = nextToken
	}

	c.JSON(http.StatusOK, resp)
}

package explorer

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) GetResourceTypesHandler(c *gin.Context) {
	pageToken := c.Query("page_token")

	resourceTypes, nextToken, err := ctrl.baton.GetResourceTypes(c.Request.Context(), pageToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := gin.H{"data": resourceTypes}
	if nextToken != "" {
		resp["next_page_token"] = nextToken
	}

	c.JSON(http.StatusOK, resp)
}

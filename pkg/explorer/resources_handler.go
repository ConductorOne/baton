package explorer

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) GetResourcesHandler(c *gin.Context) {
	resourceTypeID := c.Query("resource_type_id")
	pageToken := c.Query("page_token")

	resources, nextToken, err := ctrl.baton.GetResources(c.Request.Context(), resourceTypeID, pageToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := gin.H{"data": resources}
	if nextToken != "" {
		resp["next_page_token"] = nextToken
	}

	if pageToken == "" && resourceTypeID != "" {
		if count, countErr := ctrl.baton.countResources(c.Request.Context(), resourceTypeID); countErr == nil {
			resp["total_count"] = count
		}
	}

	c.JSON(http.StatusOK, resp)
}

package explorer

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) GetGrantsForResourceHandler(c *gin.Context) {
	pageToken := c.Query("page_token")
	pageSize := 100
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
		}
	}

	ctx := c.Request.Context()
	resourceType := c.Param("resourceType")
	resourceID := c.Param("resourceId")

	grants, nextToken, counts, err := ctrl.baton.GetAccessForResource(ctx, resourceType, resourceID, pageSize, pageToken, pageToken == "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := gin.H{"data": grants}
	if nextToken != "" {
		resp["next_page_token"] = nextToken
	}

	if counts != nil {
		resp["total_count"] = counts.TotalPrincipals
		resp["counts_by_type"] = counts.CountsByType
	}

	c.JSON(http.StatusOK, resp)
}

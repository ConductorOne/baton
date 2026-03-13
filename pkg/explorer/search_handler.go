package explorer

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) SearchHandler(c *gin.Context) {
	celExpr := c.Query("cel")
	scope := c.Query("scope")

	if celExpr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cel parameter is required"})
		return
	}
	if scope != "resources" && scope != "grants" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scope must be 'resources' or 'grants'"})
		return
	}

	pageSize := 100
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
		}
	}
	pageToken := c.Query("page_token")

	ctx := c.Request.Context()

	if scope == "resources" {
		resourceTypeID := c.Query("resource_type_id")
		results, nextToken, err := ctrl.baton.SearchResources(ctx, celExpr, resourceTypeID, pageSize, pageToken)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp := gin.H{"data": gin.H{"resources": results}}
		if nextToken != "" {
			resp["next_page_token"] = nextToken
		}
		c.JSON(http.StatusOK, resp)
		return
	}

	// scope == "grants"
	resourceType := c.Query("resource_type")
	resourceID := c.Query("resource_id")

	if resourceType == "" || resourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource_type and resource_id are required for grants scope"})
		return
	}

	results, nextToken, err := ctrl.baton.SearchGrants(ctx, celExpr, resourceType, resourceID, pageSize, pageToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := gin.H{"data": gin.H{"access": results}}
	if nextToken != "" {
		resp["next_page_token"] = nextToken
	}
	c.JSON(http.StatusOK, resp)
}

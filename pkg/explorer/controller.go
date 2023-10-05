package explorer

import (
	"github.com/gin-gonic/gin"
)

type Controller struct {
	baton BatonService
}

func NewController(filePath, syncID, resourceType string) Controller {
	return Controller{BatonService{
		filePath:     filePath,
		syncID:       syncID,
		resourceType: resourceType,
	}}
}

func (ctrl *Controller) Run(addr string) error {
	return ctrl.router().Run(addr)
}

func (ctrl *Controller) router() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	{
		api.GET("/entitlements", ctrl.GetEntitlementsHandler)
		api.GET("/resources", ctrl.GetResourcesHandler)
		api.GET("/resourceTypes", ctrl.GetResourceTypesHandler)
		api.GET("/grants/:resourceType/:resourceId", ctrl.GetGrantsForResourceHandler)
		api.GET("/access/:resourceType/:resourceId", ctrl.GetAccessHandler)
		api.GET("/:resourceType/:resourceId", ctrl.GetResourceHandler)
	}
	return router
}

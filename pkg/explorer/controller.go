package explorer

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"github.com/conductorone/baton-sdk/pkg/dotc1z"
	"github.com/conductorone/baton/pkg/storecache"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	baton *BatonService
}

func NewController(ctx context.Context, store *dotc1z.C1File, syncID, resourceType string) Controller {
	return Controller{&BatonService{
		storeCache:   storecache.NewStoreCache(ctx, store),
		store:        store,
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
	router.Use(static.Serve("/", static.LocalFile("frontend/build", true)))
	// todo: make this configurable
	err := openBrowser("http://localhost:8080")
	if err != nil {
		log.Default().Print("error opening browser: ", err)
	}

	// on reload it throws 404, so we need to redirect to index.html.
	router.NoRoute(func(ctx *gin.Context) {
		ctx.File("frontend/build/index.html")
	})

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

func openBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()

	default:
		err = fmt.Errorf("platform not supported")
	}
	if err != nil {
		return fmt.Errorf("error opening browser: %w", err)
	}

	return nil
}

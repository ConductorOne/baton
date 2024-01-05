package explorer

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/conductorone/baton-sdk/pkg/dotc1z"
	"github.com/conductorone/baton/pkg/storecache"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

//go:embed frontend/*
var frontend embed.FS

type EmbededFS struct {
	http.FileSystem
}

func (efs EmbededFS) Exists(prefix string, path string) bool {
	_, err := efs.Open(path)
	return err == nil
}

func newEmbeddedFS(efs embed.FS) EmbededFS {
	httpfs, err := fs.Sub(efs, "frontend")
	if err != nil {
		panic(err)
	}
	return EmbededFS{
		FileSystem: http.FS(httpfs),
	}
}

type Controller struct {
	baton *BatonService
}

func NewController(ctx context.Context, store *dotc1z.C1File, syncID, resourceType string, devMode bool) Controller {
	return Controller{&BatonService{
		storeCache:   storecache.NewStoreCache(ctx, store),
		store:        store,
		syncID:       syncID,
		resourceType: resourceType,
		devMode:      devMode,
	}}
}

func (ctrl *Controller) Run(addr string) error {
	return ctrl.router().Run(addr)
}

// TODO - this is a hack to get the frontend to work. Should be rewritten.
func runNpmInstallAndBuild(projectPath string) error {
	nodeModulesPath := filepath.Join(projectPath, "node_modules")
	if _, err := os.Stat(nodeModulesPath); os.IsNotExist(err) {
		log.Default().Print("node_modules folder not found. Running npm install...")
		cmd := exec.Command("npm", "install")
		cmd.Dir = projectPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Default().Print("Error running npm install:", err)
			log.Default().Print(string(output))
			return fmt.Errorf("error running 'npm install': %w", err)
		}

		log.Default().Print("npm install completed successfully.")
	}

	buildPath := filepath.Join(projectPath, "build")
	if _, err := os.Stat(buildPath); os.IsNotExist(err) {
		log.Default().Print("Build folder not found. Running npm build...")

		cmd := exec.Command("npm", "run", "build")
		cmd.Dir = projectPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Default().Print("Error running npm build:", err)
			log.Default().Print(string(output))
			return fmt.Errorf("error running 'npm run build': %w", err)
		}

		log.Default().Print("npm build completed successfully.")
	}
	return nil
}

func (ctrl *Controller) router() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	if !ctrl.baton.devMode {
		err := runNpmInstallAndBuild("frontend")
		if err != nil {
			log.Default().Println("error setting up frontend: ", err)
		}
	}

	// router.Use(static.Serve("/", static.LocalFile("frontend/build", true)))
	router.Use(static.Serve("/", newEmbeddedFS(frontend)))

	// todo: make this configurable
	if !ctrl.baton.devMode {
		err := openBrowser("http://localhost:8080")
		if err != nil {
			log.Default().Print("error opening browser: ", err)
		}
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
		api.GET("/principals/:resourceType", ctrl.GetResourcesWithPrincipalCountHandler)
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

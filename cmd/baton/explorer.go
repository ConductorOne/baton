package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton/pkg/explorer"
	"github.com/spf13/cobra"
)

func explorerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explorer",
		Short: "Run explorer UI in local browser",
		RunE:  runExplorer,
	}

	addResourceTypeFlag(cmd)
	addSyncIDFlag(cmd)

	cmd.Flags().Bool("dev", false, "Runs the frontend in development mode")
	err := cmd.Flags().MarkHidden("dev")
	if err != nil {
		log.Default().Println("error marking dev flag hidden", err)
	}

	return cmd
}

func runNpmInstallAndStart(projectPath string) error {
	installCmd := exec.Command("npm", "install")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	installCmd.Dir = projectPath
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("error running 'npm install': %w", err)
	}

	startCmd := exec.Command("npm", "start")
	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr
	startCmd.Dir = projectPath
	if err := startCmd.Run(); err != nil {
		return fmt.Errorf("error running 'npm start': %w", err)
	}

	return nil
}

func startFrontendServer() error {
	err := runNpmInstallAndStart("frontend")
	if err != nil {
		return fmt.Errorf("error running npm start: %w", err)
	}

	return nil
}

func startExplorerAPI(cmd *cobra.Command, devMode bool) {
	ctx := cmd.Context()

	filePath, err := cmd.Flags().GetString("file")
	if err != nil {
		log.Fatal("error fetching file path", err)
	}

	syncID, err := cmd.Flags().GetString("sync-id")
	if err != nil {
		log.Fatal("error fetching syncID", err)
	}

	resourceType, err := cmd.Flags().GetString(resourceTypeFlag)
	if err != nil {
		log.Fatal("error fetching resourceType", err)
	}

	m, err := manager.New(ctx, filePath)
	if err != nil {
		log.Fatal("error creating c1z manager", err)
	}
	defer m.Close(ctx)

	store, err := m.LoadC1Z(ctx)
	if err != nil {
		log.Fatal("error loading c1z", err) //nolint:gocritic // reason: in this case store is nil
	}
	defer store.Close()

	ctrl := explorer.NewController(ctx, store, syncID, resourceType, devMode)
	e := ctrl.Run(":8080")
	if e != nil {
		log.Fatal("error running explorer", err)
	}
}

func runExplorer(cmd *cobra.Command, args []string) error {
	isDevMode, err := cmd.Flags().GetBool("dev")
	if err != nil {
		return fmt.Errorf("error getting dev flag: %w", err)
	}

	if isDevMode {
		go startExplorerAPI(cmd, isDevMode)
		err = startFrontendServer()
		if err != nil {
			log.Fatal(err)
		}
	}
	startExplorerAPI(cmd, isDevMode)

	return nil
}

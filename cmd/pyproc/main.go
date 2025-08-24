package main

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

//go:embed templates/*
var templates embed.FS

var rootCmd = &cobra.Command{
	Use:   "pyproc",
	Short: "PyProc - Call Python from Go without CGO",
	Long: `PyProc is a high-performance IPC library for Go and Python integration.
It uses Unix domain sockets for fast, secure communication between Go and Python processes.`,
	Version: "0.1.0",
}

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new PyProc project",
	Long:  `Creates a new PyProc project with Go and Python scaffolding.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runInit,
}

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold [type]",
	Short: "Generate scaffold code",
	Long:  `Generate scaffold code for Go or Python workers.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runScaffold,
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(scaffoldCmd)

	initCmd.Flags().String("go-module", "", "Go module name (e.g., github.com/user/project)")
	initCmd.Flags().Bool("with-docker", false, "Include Docker Compose configuration")
	initCmd.Flags().Bool("with-k8s", false, "Include Kubernetes manifests")

	scaffoldCmd.Flags().String("name", "worker", "Name of the worker")
	scaffoldCmd.Flags().String("output", ".", "Output directory")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	projectName := "pyproc-app"
	if len(args) > 0 {
		projectName = args[0]
	}

	goModule, _ := cmd.Flags().GetString("go-module")
	withDocker, _ := cmd.Flags().GetBool("with-docker")
	withK8s, _ := cmd.Flags().GetBool("with-k8s")

	if goModule == "" {
		goModule = fmt.Sprintf("github.com/example/%s", projectName)
	}

	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Generate project structure
	dirs := []string{
		filepath.Join(projectName, "cmd", "app"),
		filepath.Join(projectName, "worker", "python"),
		filepath.Join(projectName, "api"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Template data
	data := struct {
		ProjectName string
		GoModule    string
	}{
		ProjectName: projectName,
		GoModule:    goModule,
	}

	// Generate files from templates
	files := map[string]string{
		"templates/go.mod.tmpl":           filepath.Join(projectName, "go.mod"),
		"templates/main.go.tmpl":          filepath.Join(projectName, "cmd", "app", "main.go"),
		"templates/worker.py.tmpl":        filepath.Join(projectName, "worker", "python", "worker.py"),
		"templates/requirements.txt.tmpl": filepath.Join(projectName, "worker", "python", "requirements.txt"),
		"templates/README.md.tmpl":        filepath.Join(projectName, "README.md"),
		"templates/Makefile.tmpl":         filepath.Join(projectName, "Makefile"),
	}

	if withDocker {
		files["templates/docker-compose.yml.tmpl"] = filepath.Join(projectName, "docker-compose.yml")
		files["templates/Dockerfile.go.tmpl"] = filepath.Join(projectName, "Dockerfile.go")
		files["templates/Dockerfile.python.tmpl"] = filepath.Join(projectName, "Dockerfile.python")
	}

	if withK8s {
		k8sDir := filepath.Join(projectName, "k8s")
		if err := os.MkdirAll(k8sDir, 0755); err != nil {
			return fmt.Errorf("failed to create k8s directory: %w", err)
		}
		files["templates/k8s-deployment.yaml.tmpl"] = filepath.Join(k8sDir, "deployment.yaml")
		files["templates/k8s-service.yaml.tmpl"] = filepath.Join(k8sDir, "service.yaml")
	}

	for tmplPath, outPath := range files {
		if err := generateFromTemplate(tmplPath, outPath, data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", outPath, err)
		}
	}

	fmt.Printf("✅ Created PyProc project: %s\n", projectName)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Printf("  go mod tidy\n")
	fmt.Printf("  pip install -r worker/python/requirements.txt\n")
	fmt.Printf("  make run\n")

	return nil
}

func runScaffold(cmd *cobra.Command, args []string) error {
	scaffoldType := args[0]
	name, _ := cmd.Flags().GetString("name")
	output, _ := cmd.Flags().GetString("output")

	data := struct {
		Name string
	}{
		Name: name,
	}

	switch scaffoldType {
	case "go", "golang":
		outPath := filepath.Join(output, fmt.Sprintf("%s_client.go", name))
		return generateFromTemplate("templates/scaffold_go.tmpl", outPath, data)

	case "python", "py":
		outPath := filepath.Join(output, fmt.Sprintf("%s_worker.py", name))
		return generateFromTemplate("templates/scaffold_python.tmpl", outPath, data)

	default:
		return fmt.Errorf("unknown scaffold type: %s (use 'go' or 'python')", scaffoldType)
	}
}

func generateFromTemplate(tmplPath, outPath string, data interface{}) error {
	// Read template from embedded files
	tmplContent, err := templates.ReadFile(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Parse and execute template
	tmpl, err := template.New(filepath.Base(tmplPath)).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create output file
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	// Execute template
	if err := tmpl.Execute(outFile, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	fmt.Printf("Generated: %s\n", outPath)
	return nil
}

// Helper function to copy embedded file
func copyEmbeddedFile(src, dst string) error {
	srcFile, err := templates.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// Helper function to sanitize project names
func sanitizeName(name string) string {
	// Replace non-alphanumeric characters with underscores
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return '_'
	}, name)
}

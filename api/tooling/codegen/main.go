package main

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"text/template"

	"github.com/ardanlabs/conf/v3"

	"github.com/rmsj/service/foundation/logger"
)

//go:embed templates
var templates embed.FS

var build = "develop"

type config struct {
	conf.Version
	Args conf.Args
}

func main() {
	log := logger.New(io.Discard, logger.LevelInfo, "ADMIN", func(context.Context) string { return "00000000-0000-0000-0000-000000000000" })

	if err := run(log); err != nil {
		log.Error(context.Background(), "error generating code", err)
	}
}

func run(log *logger.Logger) error {
	cfg := config{
		Version: conf.Version{
			Build: build,
			Desc:  "copyright information here",
		},
	}

	const prefix = "ENGINE"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}

		out, err := conf.String(&cfg)
		if err != nil {
			return fmt.Errorf("generating config for output: %w", err)
		}
		log.Info(context.Background(), "startup", "config", out)

		return fmt.Errorf("parsing config: %w", err)
	}

	if err := genCode(cfg.Args, log, cfg); err != nil {
		log.Error(context.Background(), "error generating code", err)
		os.Exit(1)

		return err
	}

	return nil
}

// genCode adds a new domain to the system with all the basic CRUD functionality.
func genCode(args conf.Args, log *logger.Logger, cfg config) error {
	name := args.Num(0)

	if err := addAPILayer(name); err != nil {
		return fmt.Errorf("adding api layer files: %w", err)
	}

	if err := addAPITestLayer(name); err != nil {
		return fmt.Errorf("adding api test layer files: %w", err)
	}

	if err := addAppLayer(name); err != nil {
		return fmt.Errorf("adding app layer files: %w", err)
	}

	if err := addBusinessLayer(name); err != nil {
		return fmt.Errorf("adding business layer files: %w", err)
	}

	if err := addStorageLayer(name); err != nil {
		return fmt.Errorf("adding storage layer files: %w", err)
	}

	if err := addAppMiddlewareLayer(name); err != nil {
		return fmt.Errorf("adding api auth middleware layer files: %w", err)
	}

	return nil
}

func addAPILayer(domain string) error {
	const basePath = "api/domain/http"

	app, err := fs.Sub(templates, "templates/api")
	if err != nil {
		return fmt.Errorf("switching to template/api folder: %w", err)
	}

	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		return walkWork(domain, basePath, app, fileName, dirEntry, err)
	}

	fmt.Println("=======================================================")
	fmt.Println("API LAYER CODE")

	if err := fs.WalkDir(app, ".", fn); err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	return nil
}

func addAPITestLayer(domain string) error {
	const basePath = "api/services/engine/tests"

	app, err := fs.Sub(templates, "templates/apitests")
	if err != nil {
		return fmt.Errorf("switching to template/apitests folder: %w", err)
	}

	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		return walkWork(domain, basePath, app, fileName, dirEntry, err)
	}

	fmt.Println("=======================================================")
	fmt.Println("API TEST LAYER CODE")

	if err := fs.WalkDir(app, ".", fn); err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	return nil
}

func addAppLayer(domain string) error {
	const basePath = "app/domain"

	app, err := fs.Sub(templates, "templates/app")
	if err != nil {
		return fmt.Errorf("switching to template/app folder: %w", err)
	}

	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		return walkWork(domain, basePath, app, fileName, dirEntry, err)
	}

	fmt.Println("=======================================================")
	fmt.Println("APP LAYER CODE")

	if err := fs.WalkDir(app, ".", fn); err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	return nil
}

func addBusinessLayer(domain string) error {
	const basePath = "business/domain"

	app, err := fs.Sub(templates, "templates/business")
	if err != nil {
		return fmt.Errorf("switching to template/business folder: %w", err)
	}

	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		return walkWork(domain, basePath, app, fileName, dirEntry, err)
	}

	fmt.Println("=======================================================")
	fmt.Println("BUSINESS LAYER CODE")

	if err := fs.WalkDir(app, ".", fn); err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	return nil
}

func addStorageLayer(domain string) error {
	basePath := fmt.Sprintf("business/domain/%sbus/stores", domain)

	app, err := fs.Sub(templates, "templates/storage")
	if err != nil {
		return fmt.Errorf("switching to template/storage folder: %w", err)
	}

	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		return walkWork(domain, basePath, app, fileName, dirEntry, err)
	}

	fmt.Println("=======================================================")
	fmt.Println("STORAGE LAYER CODE")

	if err := fs.WalkDir(app, ".", fn); err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	return nil
}

func addAppMiddlewareLayer(domain string) error {
	basePath := "app/sdk/mid"

	app, err := fs.Sub(templates, "templates/appmid")
	if err != nil {
		return fmt.Errorf("switching to template/appmid folder: %w", err)
	}

	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		return walkWork(domain, basePath, app, fileName, dirEntry, err)
	}

	fmt.Println("=======================================================")
	fmt.Println("MIDDLEWARE LAYER CODE")

	if err := fs.WalkDir(app, ".", fn); err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	return nil
}

func walkWork(domain string, basePath string, app fs.FS, fileName string, dirEntry fs.DirEntry, err error) error {
	if err != nil {
		return fmt.Errorf("walkdir failure: %w", err)
	}

	if dirEntry.IsDir() {
		return nil
	}

	f, err := app.Open(fileName)
	if err != nil {
		return fmt.Errorf("opening %s: %w", fileName, err)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("reading %s: %w", fileName, err)
	}

	tmpl := template.Must(template.New("code").Parse(string(data)))

	domainVar := domain
	if len(domainVar) > 3 {
		vowels := []string{"a", "e", "i", "o", "u"}
		var startVowel string
		for _, c := range vowels {
			if strings.HasPrefix(domainVar, c) {
				startVowel = c
			}
		}
		// remove vowels
		for _, c := range []string{"a", "e", "i", "o", "u"} {
			domainVar = strings.ReplaceAll(domainVar, c, "")
		}
		// put back initial vowel if any
		domainVar = startVowel + domainVar
		if len(domainVar) > 3 {
			domainVar = domain[:3]
		}
	}

	d := struct {
		DomainLower  string
		DomainUpper  string
		DomainVar    string
		DomainVarU   string
		DomainNewVar string
		DomainUpdVar string
	}{
		DomainLower:  strings.ToLower(domain),
		DomainUpper:  strings.ToUpper(domain[0:1]) + domain[1:],
		DomainVar:    domainVar,
		DomainVarU:   strings.ToUpper(domainVar[0:1]) + domainVar[1:],
		DomainNewVar: "n" + domain[0:1],
		DomainUpdVar: "u" + domain[0:1],
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, d); err != nil {
		return err
	}

	if err := writeFile(basePath, domain, fileName, b); err != nil {
		return fmt.Errorf("writing %s: %w", fileName, err)
	}

	return nil
}

func writeFile(basePath string, domain string, fileName string, b bytes.Buffer) error {
	path := basePath
	switch {
	case basePath == "api/domain/http":
		path = fmt.Sprintf("%s/%sapi", basePath, domain)
	case basePath == "api/services/engine/tests":
		path = fmt.Sprintf("%s/%sapi", basePath, domain)
	case basePath == "app/domain":
		path = fmt.Sprintf("%s/%sapp", basePath, domain)
	case basePath == "business/domain":
		path = fmt.Sprintf("%s/%sbus", basePath, domain)
	case strings.Contains(basePath, "stores"):
		path = fmt.Sprintf("%s/%sdb", basePath, domain)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Creating directory:", path)

		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return fmt.Errorf("write app directory: %w", err)
		}
	}

	path = fmt.Sprintf("%s/%s", path, fileName[:len(fileName)-1])
	path = strings.Replace(path, "new", domain, 1)

	fmt.Println("Add file:", path)
	if fileExists(path) {
		fmt.Println("file already exists: ", path)
		return nil
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	fmt.Println("Writing code:", path)
	if _, err := f.Write(b.Bytes()); err != nil {
		return fmt.Errorf("writing bytes: %w", err)
	}

	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

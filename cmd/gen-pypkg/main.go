package main

import (
	"embed"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/util/gconv"
	"github.com/jessevdk/go-flags"
)

type ProjectConfig struct {
	ProjectPath    string `required:"true" short:"p" long:"project-path" description:"Project path." json:"projectPath"`
	ProjectName    string `required:"false" description:"Base name of project path, replace DASH(-) with UNDERLINE(_)." json:"projectName"`
	BaseImage      string `required:"true" short:"i" long:"base-image" description:"Base image for create project image." json:"dockerBaseImage"`
	Registry       string `required:"true" short:"r" long:"docker-registry" description:"Docker registry for create project image." json:"dockerRegistry"`
	NexusUrl       string `required:"true" long:"nexus-url" description:"Nexus url for uploading python package." json:"nexusUrl"`
	NexusUsername  string `required:"true" long:"nexus-username" description:"Nexus username for uploading python package." json:"nexusUsername"`
	NexusPassword  string `required:"true" long:"nexus-password" description:"Nexus password for uploading python package." json:"nexusPassword"`
	NexusPypiPath  string `required:"true" long:"nexus-pypi-path" description:"Nexus path for uploading python package." json:"nexusPypiPath"`
	PdmSourceName  string `required:"true" long:"pdm-source-name" description:"Pdm source name." json:"pdmSourceName"`
	PdmSourceUrl   string `required:"true" long:"pdm-source-url" description:"Pdm source url." json:"pdmSourceUrl"`
	PdmAuthorName  string `required:"true" long:"pdm-author-name" description:"Pdm author name." json:"pdmAuthorName"`
	PdmAuthorEmail string `required:"true" long:"pdm-author-email" description:"Pdm author email." json:"pdmAuthorEmail"`
	PdmLicense     string `required:"true" long:"pdm-license" description:"Pdm license." json:"pdmLicense"`
}

var (
	// Example: params in example are invalid.
	// go run "github.com/artistml/toolkits/cmd/gen-pypkg" -c config
	// go run "github.com/artistml/toolkits/cmd/gen-pypkg" -c $HOME/{project}/config
	opts struct {
		ConfigPath string `required:"true" short:"c" long:"config-path" description:"Config releactive path" json:"ConfigPath"`
	}

	//go:embed templates/*.tmpl
	fs embed.FS
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	rootPath, err := os.Getwd()
	checkErr(err)
	_, err = flags.Parse(&opts)
	checkErr(err)
	if !strings.HasPrefix(opts.ConfigPath, "/") {
		// use rootPath to join absolute path.
		opts.ConfigPath = path.Join(rootPath, opts.ConfigPath)
	}
	err = g.Cfg().SetPath(opts.ConfigPath)
	checkErr(err)
	projectConfig := ProjectConfig{}
	gconv.Struct(g.Cfg().Get("project"), &projectConfig)

	err = os.MkdirAll(projectConfig.ProjectPath, os.ModePerm)
	checkErr(err)
	projectConfig.ProjectName = strings.ReplaceAll(path.Base(projectConfig.ProjectPath), "-", "_")
	err = os.MkdirAll(path.Join(projectConfig.ProjectPath, projectConfig.ProjectName), os.ModePerm)
	checkErr(err)
	initPythonFile := path.Join(projectConfig.ProjectPath, projectConfig.ProjectName, "__init__.py")
	if _, err = os.Stat(initPythonFile); os.IsNotExist(err) {
		_, err = os.Create(initPythonFile)
		checkErr(err)
	}
	templates, err := template.ParseFS(fs, "templates/*.tmpl")
	checkErr(err)
	for _, file := range templates.Templates() {
		var (
			output = path.Join(projectConfig.ProjectPath, fmt.Sprintf("%s", file.Name()[:len(file.Name())-5]))
		)
		fmt.Printf("Generate file by tmp %s: %s\n", file.Name(), output)
		func() {
			var f *os.File
			f, err = os.Create(output)
			checkErr(err)
			if strings.HasSuffix(output, ".sh") {
				err = os.Chmod(output, os.ModePerm)
				checkErr(err)
			}

			defer func() {
				_ = f.Sync()
				_ = f.Close()
			}()

			err = file.Execute(f, &projectConfig)
			checkErr(err)
		}()

	}
}

package main

import (
	"embed"
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
	"path"
	"strings"
	"text/template"
)


var (
	// Example: params in example are invalid.
	// go run "github.com/artistml/toolkits/cmd/gen-pypkg" -p /tmp/toolkits/toolkits-python -i github.com/artistml/base-python:3.8.13 -r github.com/artistml --nexus-url=http://github.com/artistml --nexus-username=artistml --nexus-password=artistml-pw --nexus-pypi-path=pypi-hosted --pdm-source-name=pypi --pdm-source-url=http://github.com/artistml/repository/pypi-group/simple --pdm-author-name=artistml --pdm-author-email=artistml@github.com --pdm-license=ArtistML
	opts struct {
		ProjectPath    string `required:"true" short:"p" long:"project-path" description:"Project path." json:"ProjectPath"`
		ProjectName    string `required:"false" description:"Base name of project path, replace DASH(-) with UNDERLINE(_)." json:"ProjectName"`
		BaseImage      string `required:"true" short:"i" long:"base-image" description:"Base image for create project image." json:"BaseImage"`
		Registry       string `required:"true" short:"r" long:"docker-registry" description:"Docker registry for create project image." json:"Registry"`
		NexusUrl       string `required:"true" long:"nexus-url" description:"Nexus url for uploading python package." json:"NexusUrl"`
		NexusUsername  string `required:"true" long:"nexus-username" description:"Nexus username for uploading python package." json:"NexusUsername"`
		NexusPassword  string `required:"true" long:"nexus-password" description:"Nexus password for uploading python package." json:"NexusPassword"`
		NexusPypiPath  string `required:"true" long:"nexus-pypi-path" description:"Nexus path for uploading python package." json:"NexusPypiPath"`
		PdmSourceName  string `required:"true" long:"pdm-source-name" description:"Pdm source name." json:"PdmSourceName"`
		PdmSourceUrl   string `required:"true" long:"pdm-source-url" description:"Pdm source url." json:"PdmSourceUrl"`
		PdmAuthorName  string `required:"true" long:"pdm-author-name" description:"Pdm author name." json:"PdmAuthorName"`
		PdmAuthorEmail string `required:"true" long:"pdm-author-email" description:"Pdm author email." json:"PdmAuthorEmail"`
		PdmLicense     string `required:"true" long:"pdm-license" description:"Pdm license." json:"PdmLicense"`
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
	_, err := flags.Parse(&opts)
	checkErr(err)
	err = os.MkdirAll(opts.ProjectPath, os.ModePerm)
	checkErr(err)
	opts.ProjectName = strings.ReplaceAll(path.Base(opts.ProjectPath), "-", "_")
	err = os.MkdirAll(path.Join(opts.ProjectPath, opts.ProjectName), os.ModePerm)
	checkErr(err)
	initPythonFile := path.Join(opts.ProjectPath, opts.ProjectName, "__init__.py")
	if _, err = os.Stat(initPythonFile); os.IsNotExist(err) {
		_, err = os.Create(initPythonFile)
		checkErr(err)
	}
	templates, err := template.ParseFS(fs, "templates/*.tmpl")
	checkErr(err)
	for _, file := range templates.Templates() {
		var (
			output = path.Join(opts.ProjectPath, fmt.Sprintf("%s", file.Name()[:len(file.Name())-5]))
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

			err = file.Execute(f, &opts)
			checkErr(err)
		}()

	}
}

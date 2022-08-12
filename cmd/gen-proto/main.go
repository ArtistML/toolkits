package main

import (
	"embed"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/jessevdk/go-flags"
)

const (
	header = "// Code generated by github.com/artistml/toolkits/cmd/gen-proto. DO NOT EDIT.\n"
)

type Opt struct {
	Output                string `json:"Output"`
	DefaultHost           string `json:"DefaultHost"`
	EntityHeaders         string `json:"EntityHeaders"`
	PackageName           string `json:"PackageName"`
	PackageVersionNo      string `required:"true" description:"Proto package version number." json:"PackageVersionNo"`
	EntityName            string `required:"true" description:"Entity name for CRUD." json:"EntityName"`
	EntityIdExpr          string `required:"true" json:"EntityIdExpr"`
	EntityResourceType    string `required:"true" json:"EntityResourceType"`
	EntityResourcePattern string `required:"true" json:"EntityResourcePattern"`
	CapitalEntityName     string `required:"true" json:"CapitalEntityName"`
	ParentUri             string `required:"true" json:"ParentUri"`
	ParentUriPattern      string `required:"true" json:"ParentUriPattern"`
	ParentUriType         string `required:"true" json:"ParentUriType"`
	HasImportRequest      bool   `required:"true" json:"HasImportRequest"`
	HasExportRequest      bool   `required:"true" json:"HasExportRequest"`
	AuthImport            string `required:"true" json:"AuthImport"`
	AuthOption            string `required:"true" json:"AuthOption"`
	SwaggerHost           string `required:"true" json:"SwaggerHost"`
	SwaggerUrl            string `required:"true" json:"SwaggerUrl"`
}

type ApiResource struct {
	Type       string `json:"Type"`
	Pattern    string `json:"Pattern"`
	Uri        string `json:"Uri"`
	AuthImport string `json:"AuthImport"`
	AuthOption string `json:"AuthOption"`
}

var (
	// Example:
	// go run "github.com/artistml/toolkits/cmd/gen-proto" --host=github.com/artistml/toolkits -f echo/v1 -f echo/v2
	opts struct {
		DefaultHost string   `required:"true" long:"host" description:"Default host for whole proto." json:"DefaultHost"`
		SwaggerHost string   `required:"true" long:"swagger-host" description:"Swagger host for whole proto." json:"SwaggerHost"`
		SwaggerUrl  string   `required:"true" long:"swagger-url" description:"Swagger host for whole proto." json:"SwaggerUrl"`
		Filters     []string `required:"false" short:"f" long:"filter" description:"Filter paths for generation." json:"Filters"`
	}

	optList = []Opt{}
	//go:embed templates/*.tmpl
	fs             embed.FS
	uriRegexp      = regexp.MustCompile(`\{[a-zA-Z_\d]+\}`)
	uriReplacement = "*"
	ignoreFiles    = map[string]bool{
		"common_resources.proto": true,
		"types.proto":            true,
		"enums.proto":            true,
	}
	wordRegexp = regexp.MustCompile(`[a-zA-Z][a-zA-Z\d]*`)
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getCapitalEntityName(entityName string) string {
	capitalEntityName := ""
	for _, word := range wordRegexp.FindAll([]byte(entityName), -1) {
		w := strings.ToUpper(string(word[:1]))
		w = w + string(word[1:])
		capitalEntityName = capitalEntityName + w
	}
	fmt.Println("capitalEntityName", capitalEntityName)
	return capitalEntityName
}

func readLines(messageFile string, filters map[string]string) ([]string, map[string][]int) {
	fileExtension := filepath.Ext(messageFile)
	authFile := strings.ReplaceAll(messageFile, fileExtension, ".auth")
	lines := []string{}
	for _, filePath := range []string{messageFile, authFile} {
		if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
			continue
		}
		body, err := ioutil.ReadFile(filePath)
		checkErr(err)
		bodyStr := string(body)
		lines = append(lines, strings.Split(bodyStr, "\n")...)
	}

	if len(filters) == 0 {
		return lines, make(map[string][]int)
	}
	output := make(map[string][]int)
	for i, line := range lines {
		for filter, value := range filters {
			if !strings.Contains(line, filter) {
				continue
			}
			_, ok := output[filter]
			if !ok {
				output[filter] = []int{}
			}
			output[value] = append(output[value], i)
		}
	}
	return lines, output
}

func readResource(messageFile string) *ApiResource {
	lines, matchFilters := readLines(messageFile, map[string]string{"google.api.resource": "google.api.resource", "type:": "type:", "type :": "type:", "pattern:": "pattern:", "pattern :": "pattern:", "authImport": "authImport", "authOption": "authOption"})
	if _, ok := matchFilters["google.api.resource"]; !ok {
		return nil
	}
	resourceType := strings.Trim(lines[matchFilters["type:"][0]], " \n\t")
	start, end := strings.IndexAny(resourceType, "\""), strings.LastIndex(resourceType, "\"")
	rType := resourceType[start+1 : end]
	resourcePattern := strings.Trim(lines[matchFilters["pattern:"][0]], " \n\t")
	start, end = strings.IndexAny(resourcePattern, "\""), strings.LastIndex(resourcePattern, "\"")
	rPattern := resourcePattern[start+1 : end]
	resourcePattern = resourcePattern[strings.IndexAny(resourcePattern, "\""):strings.LastIndex(resourcePattern, "\"")]
	apiResource := &ApiResource{
		Type:    rType,
		Pattern: rPattern,
		Uri:     uriRegexp.ReplaceAllString(rPattern, uriReplacement),
	}
	if strings.Contains(apiResource.Uri, "{") {
		fmt.Printf("Uri %s can't contains '{' or '}'. \n", apiResource.Uri)
		os.Exit(1)
	}
	if value, ok := matchFilters["authImport"]; ok {
		apiResource.AuthImport = strings.Trim(strings.Split(lines[value[0]], "authImport")[1], " :\n\t")
	}
	if value, ok := matchFilters["authOption"]; ok {
		apiResource.AuthOption = strings.Trim(strings.Split(lines[value[0]], "authOption")[1], " :\n\t")
	}
	return apiResource
}

func main() {
	rootPath, err := os.Getwd()
	checkErr(err)

	_, err = flags.Parse(&opts)
	checkErr(err)

	var pkgPaths []string
	err = filepath.Walk(rootPath, func(walkPath string, info os.FileInfo, err error) error {
		fmt.Printf("rootPath: %s, walkPath: %s.\n", rootPath, walkPath)
		if !info.IsDir() {
			return nil
		}
		if path.Dir(path.Dir(walkPath)) != rootPath {
			return nil
		}
		match := true
		if len(opts.Filters) > 0 {
			match = false
			// 定义了 filter 的情况下只对包含某个 filter 的路径进行 service proto 生成
			for _, subPath := range opts.Filters {
				if !strings.Contains(walkPath, subPath) {
					continue
				}
				match = true
				break
			}
		}
		if !match {
			return nil
		}
		pkgPaths = append(pkgPaths, walkPath)
		return nil
	})
	checkErr(err)
	fmt.Printf("pkgPaths: %s.\n", pkgPaths)
	if len(pkgPaths) == 0 {
		return
	}
	for _, pkgPath := range pkgPaths {
		parentResource := readResource(path.Join(pkgPath, "common_resources.proto"))
		if parentResource == nil {
			fmt.Println("common_resources.proto must contains google.api.resource_definition")
			os.Exit(1)
		}
		pkgName := path.Base(path.Dir(pkgPath))
		err = filepath.Walk(pkgPath, func(walkPath string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if path.Dir(walkPath) != pkgPath {
				return nil
			}
			if strings.Contains(walkPath, "_service.proto") {
				return nil
			}
			if _, ok := ignoreFiles[path.Base(walkPath)]; ok {
				return nil
			}
			items := strings.Split(path.Base(walkPath), ".")
			entityName := items[0]
			suffix := items[1]
			if pkgName == entityName {
				// 对于与 model 同名的 entity，认为是自定义的服务，不需要再生成 service
				return nil
			}
			if suffix == "auth" {
				// 对于 {entity}.auth 这样用于 {entity} 登录验证所使用的配置，不需要生成 service
				return nil
			}

			entityResource := readResource(walkPath)
			if entityResource == nil {
				entityResource = &ApiResource{}
			} else {
				if !strings.HasPrefix(entityResource.Pattern, parentResource.Pattern) {
					fmt.Printf("Resource %s of entity should start with parent's resource %s!\n", entityResource.Pattern, parentResource.Pattern)
					os.Exit(1)
				}
			}
			capitalEntityName := getCapitalEntityName(entityName)
			importConfigName, exportConfigName := fmt.Sprintf("Import%sConfig", capitalEntityName), fmt.Sprintf("Export%sConfig", capitalEntityName)
			lines, matchFilters := readLines(walkPath, map[string]string{"import": "import", importConfigName: importConfigName, exportConfigName: exportConfigName})
			i := matchFilters["import"][0]
			headers := strings.Join(lines[:i], "\n")
			_, hasImportRequest := matchFilters[importConfigName]
			_, hasExportRequest := matchFilters[exportConfigName]
			optList = append(optList, Opt{
				Output: pkgPath, DefaultHost: opts.DefaultHost, EntityHeaders: headers,
				PackageName:           pkgName,
				PackageVersionNo:      path.Base(pkgPath)[1:],
				EntityName:            entityName,
				EntityIdExpr:          fmt.Sprintf("{%s.id=*}", entityName),
				EntityResourceType:    entityResource.Type,
				EntityResourcePattern: entityResource.Pattern,
				CapitalEntityName:     capitalEntityName,
				ParentUri:             parentResource.Uri,
				ParentUriPattern:      parentResource.Pattern,
				ParentUriType:         parentResource.Type,
				HasImportRequest:      hasImportRequest,
				HasExportRequest:      hasExportRequest,
				AuthImport:            entityResource.AuthImport,
				AuthOption:            entityResource.AuthOption,
				SwaggerHost:           opts.SwaggerHost,
				SwaggerUrl:            opts.SwaggerUrl,
			})
			return nil
		})
		checkErr(err)
	}

	templates, err := template.ParseFS(fs, "templates/*.tmpl")
	checkErr(err)
	for _, file := range templates.Templates() {
		for _, opt := range optList {
			var (
				output = path.Join(opt.Output, fmt.Sprintf("%s_%s", opt.EntityName, file.Name()[:len(file.Name())-5]))
			)
			fmt.Printf("Generate server.proto: %s\n", output)
			func() {
				f, err := os.Create(output)
				checkErr(err)
				defer func() {
					_ = f.Sync()
					_ = f.Close()
				}()

				_, err = f.WriteString(header)
				checkErr(err)
				err = file.Execute(f, &opt)
				checkErr(err)
			}()
		}

	}
}

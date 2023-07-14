/*
Copyright © 2023 Xu Wu <ixw1991@126.com>
Use of this source code is governed by a MIT style
license that can be found in the LICENSE file.
*/
package scaffold

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/imxw/gitlab-scaffold/internal/gitlabx"
	"github.com/imxw/gitlab-scaffold/internal/stringx"
	"github.com/imxw/gitlab-scaffold/internal/util"
)

const (
	DefaultTemplateGroup = "template"
)

var defaultExtensions = []string{
	".go", ".java", ".py", ".vue", // Source code files
	".md",                           // Documentation files
	".html", ".css", ".js", ".scss", // Web files
	".json", ".xml", ".yml", ".yaml", // Data files
	".kt", ".gradle", // Android specific
}

var defaultBase64Extensions = []string{
	".png",
	".jar",
	".jpg",
	".jks",
}

var defaultFiles = []string{
	"Dockerfile",
	"Makefile",
}

type TemplateData struct {
	Name string
	Port int
}

type Config struct {
	Namespace        string   `mapstructure:"namespace"`
	Extensions       []string `mapstructure:"extensions"`
	Base64Extensions []string `mapstructure:"base64_extensions"`
	Files            []string `mapstructure:"files"`
}

func replacePathName(serviceName, pathName string) string {

	replacements := map[string]string{
		"{{Name}}":                      serviceName,
		"{{Name_SkipFirstPart}}":        stringx.SkipFirstPart(serviceName),
		"{{Name_SkipFirstAndLastPart}}": stringx.SkipFirstAndLastParts(serviceName),
		"{{Name_ToPascalCase}}":         stringx.ToPascalCase(serviceName),
		"{{Name_ToCamelCase}}":          stringx.ToCamelCase(serviceName),
	}

	for k, v := range replacements {
		pathName = strings.Replace(pathName, k, v, -1)
	}

	return pathName
}

func ListTemplates(client *gitlabx.Client) (string, error) {

	projects, err := client.ListProjectsInGroup(DefaultTemplateGroup)
	if err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(projects, "", "  ")

	return string(jsonData), err

}

// DownloadAndUnpackTemplateToTempDir 是一个函数，它下载指定模板并将其解压到临时目录，
// 然后返回解压后数据的路径。它接收两个参数：一个 gitlabx.Client 实例和一个字符串模板名。
// 它返回解压后的数据路径及错误信息。
func DownloadAndUnpackTemplateToTempDir(client *gitlabx.Client, templateName string) (string, error) {

	data, err := client.GetProjectArchive(DefaultTemplateGroup + "/" + templateName)
	if err != nil {
		return "", err
	}

	// 创建一个临时目录来保存解压后的文件
	tempDir, err := os.MkdirTemp("", "template")
	if err != nil {
		return "", err
	}

	rootPath, err := util.UnpackTarGz(data, tempDir)
	if err != nil {
		return "", err
	}

	return rootPath, nil

}

func VisitAndModifyFiles(rootPath, path string, data TemplateData, f os.FileInfo, err error) (map[string]*gitlabx.FileData, error) {

	fileMap := make(map[string]*gitlabx.FileData)

	if err != nil {
		fmt.Printf("error visiting %v: %v\n", path, err)
		return nil, err
	}

	newpath := replacePathName(data.Name, path)
	base := filepath.Base(newpath)

	newpath = strings.TrimPrefix(newpath, rootPath)
	encode := "text"
	var fileContent string
	if !f.IsDir() {
		ext := filepath.Ext(newpath)
		ext = strings.ToLower(ext)
		var content []byte
		content, err = os.ReadFile(path)

		if err != nil {
			return nil, err
		}

		if stringx.StringInSlice(ext, defaultExtensions) || stringx.StringInSlice(base, defaultFiles) {

			// Template processing
			tmpl, err := template.New("content").Funcs(template.FuncMap{
				"SkipFirstPart":        stringx.SkipFirstPart,
				"SkipLastPart":         stringx.SkipLastPart,
				"SkipFirstAndLastPart": stringx.SkipFirstAndLastParts,
				"ToCamelCase":          stringx.ToCamelCase,
				"ToPascalCase":         stringx.ToPascalCase,
			}).Parse(string(content))
			if err != nil {
				fmt.Printf("error parsing the template %v: %v\n", path, err)
				return nil, err
			}

			var tpl bytes.Buffer
			err = tmpl.Execute(&tpl, data)
			if err != nil {
				fmt.Printf("error executing the template %v: %v\n", path, err)
				return nil, err
			}
			fileContent = tpl.String()

		} else if stringx.StringInSlice(ext, defaultBase64Extensions) {

			fileContent = base64.StdEncoding.EncodeToString(content)
			encode = "base64"

		} else {
			fileContent = string(content)
		}

		fileMap[newpath] = &gitlabx.FileData{
			Content:  fileContent,
			Encoding: encode,
		}

	}

	return fileMap, nil
}

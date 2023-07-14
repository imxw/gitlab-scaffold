/*
Copyright © 2023 Xu Wu <ixw1991@126.com>
Use of this source code is governed by a MIT style
license that can be found in the LICENSE file.
*/
package cmd

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/imxw/gitlab-scaffold/internal/config"
	"github.com/imxw/gitlab-scaffold/internal/gitlabx"
	"github.com/imxw/gitlab-scaffold/internal/scaffold"
)

var projectName string
var port int
var groupName string
var description string

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Use a scaffold template to create a new project.",
	Long: `Use the 'scaffold use' command followed by a template name, project name, port and group name to create a new project using the specified scaffold template.
	
The command uses the format: 'scaffold use TEMPLATE_NAME -n PROJECT_NAME -p PORT -g GROUP_NAME'. 
This command creates a new project using the specified scaffold template and names the project as PROJECT_NAME. For backend applications, the command assigns the specified PORT.
Additionally, the new project is associated with the given GROUP_NAME. If the project is frontend-based, specifying a port is not necessary.
	`,
	Example: "  `scaffold use backend-java-service -n tope-test -p 8955 -g team1/backend`",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		// 获取模板名称
		templateName := args[0]

		url := config.C().GetGitlab().BaseURL
		token := config.C().GetGitlab().Token
		client, err := gitlabx.NewClient(url, token)
		if err != nil {
			panic(err)
		}

		// 判断gitlab项目是否存在
		nameWithNamespace := groupName + "/" + projectName

		exist, err := client.IsProjectExist(nameWithNamespace)
		if err != nil {
			panic(err)
		}

		if exist {
			log.Fatalf("%s already exists, please use a different project name.", projectName)
		}
		// 创建项目
		if description == "" {
			description = projectName
		}
		err = client.CreateProjectInGroup(projectName, groupName, description)
		if err != nil {
			log.Fatal(err)
		}
		if err := client.CopyProjectVariables(scaffold.DefaultTemplateGroup+"/"+templateName, nameWithNamespace); err != nil {
			log.Fatal(err)
		}

		if err := client.EnableRunner(scaffold.DefaultTemplateGroup+"/"+templateName, nameWithNamespace); err != nil {
			log.Fatal(err)
		}

		// 下载模板压缩包到本地
		rootPath, err := scaffold.DownloadAndUnpackTemplateToTempDir(client, templateName)
		if err != nil {
			log.Fatal(err)
		}

		// 检查本地文件夹是否存在
		if _, err := os.Stat(rootPath); os.IsNotExist(err) {
			log.Fatalf("Local folder '%s' does not exist", rootPath)
		}

		// 渲染模板并修改文件及文件夹名
		fileMap := make(map[string]*gitlabx.FileData)

		data := scaffold.TemplateData{
			Name: projectName,
			Port: port,
		}

		err = filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
			visitFileMap, err := scaffold.VisitAndModifyFiles(rootPath, path, data, info, err)
			if err != nil {
				fmt.Printf("error visiting and modifying files in %v: %v\n", path, err)
				return err
			}

			for k, v := range visitFileMap {
				fileMap[k] = v
			}

			return nil
		})

		if err != nil {
			fmt.Printf("error walking the path %v: %v\n", rootPath, err)
			// 在此删除项目
			delErr := client.DeleteProject(nameWithNamespace)
			if delErr != nil {
				fmt.Println("Error in DeleteProject: ", delErr)
			} else {
				fmt.Println("Project deleted due to error in filepath.Walk")
			}
			return
		}

		// 提交commit
		if err := client.CreateCommitFromFiles(nameWithNamespace, fileMap); err != nil {
			log.Fatal(err)
		}

		// 创建dev分支
		devBranch := "dev"
		mainbranch := "master"

		if err := client.CreateBranch(nameWithNamespace, devBranch, mainbranch); err != nil {
			log.Fatal(err)
		}

		if err := client.SetDefaultBranch(nameWithNamespace, devBranch); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Success!")

	},
}

func init() {
	rootCmd.AddCommand(useCmd)

	useCmd.Flags().StringVarP(&projectName, "name", "n", "", "name of the new project")
	useCmd.Flags().IntVarP(&port, "port", "p", -1, "port for the application (optional)")
	useCmd.Flags().StringVarP(&groupName, "group", "g", "", "group of the new project")
	useCmd.Flags().StringVarP(&description, "desc", "d", "", "description of the new project")

	useCmd.MarkFlagRequired("name")
	useCmd.MarkFlagRequired("group")

}

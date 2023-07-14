/*
Copyright © 2023 Xu Wu <ixw1991@126.com>
Use of this source code is governed by a MIT style
license that can be found in the LICENSE file.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/imxw/gitlab-scaffold/internal/config"
	"github.com/imxw/gitlab-scaffold/internal/gitlabx"
	"github.com/imxw/gitlab-scaffold/internal/scaffold"
)

type Project struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Projects []Project

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available scaffold templates",
	Long:  `List available scaffold templates`,
	Run: func(cmd *cobra.Command, args []string) {
		url := config.C().GetGitlab().BaseURL
		token := config.C().GetGitlab().Token
		client, err := gitlabx.NewClient(url, token)
		if err != nil {
			panic(err)
		}

		templates, err := scaffold.ListTemplates(client)
		if err != nil {
			panic(err)
		}
		fmt.Println(templates)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

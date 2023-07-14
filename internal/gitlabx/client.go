/*
Copyright © 2023 Xu Wu <ixw1991@126.com>
Use of this source code is governed by a MIT style
license that can be found in the LICENSE file.
*/
package gitlabx

import (
	"errors"
	"fmt"

	"github.com/xanzy/go-gitlab"
)

type Config struct {
	BaseURL string `mapstructure:"baseurl" validate:"url"`
	Token   string `mapstructure:"token" validate:"required"`
}

// Client 结构体包含一个go-gitlab客户端实例
type Client struct {
	git *gitlab.Client
}

type FileData struct {
	Content  string
	Encoding string
}

// 默认的GitLab URL
const defaultGitLabUrl = "https://gitlab.com"

// NewClient 函数创建一个新的GitLab客户端，接收GitLab的URL和token作为参数
// 如果没有提供URL或者URL是默认的GitLab URL，那么会使用默认的GitLab URL创建客户端
// 如果没有提供token，那么会返回一个错误
func NewClient(url, token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("empty gitlab token provided")
	}

	var git *gitlab.Client
	var err error

	if url == "" || url == defaultGitLabUrl {
		git, err = gitlab.NewClient(token)
	} else {
		git, err = gitlab.NewClient(token, gitlab.WithBaseURL(url))
	}

	if err != nil {
		return nil, err
	}

	return &Client{git: git}, nil
}

// getGroupID 方法获取给定组名的组ID，如果组不存在则返回错误
func (c *Client) getGroupID(groupName string) (int, error) {
	groups, _, err := c.git.Groups.ListGroups(&gitlab.ListGroupsOptions{
		Search: gitlab.String(groupName),
	})
	if err != nil {
		return 0, err
	}

	for _, group := range groups {
		if group.FullPath == groupName {
			return group.ID, nil
		}
	}

	return 0, fmt.Errorf("group %s not found", groupName)
}

// ListProjectsInGroup 方法根据提供的组名获取组内的所有项目，返回一个包含项目名和项目描述的map，如果组不存在则返回错误
func (c *Client) ListProjectsInGroup(groupName string) (map[string]string, error) {

	groupId, err := c.getGroupID(groupName)
	if err != nil {
		return nil, err
	}

	projects, _, err := c.git.Groups.ListGroupProjects(groupId, &gitlab.ListGroupProjectsOptions{
		IncludeSubGroups: gitlab.Bool(false),
	})

	res := make(map[string]string)

	for _, p := range projects {
		res[p.Name] = p.Description
	}

	return res, err

}

// IsProjectExist 函数用于检查在 GitLab 中是否存在特定的项目。
// 该函数接收一个包含项目名称和命名空间的字符串作为输入（例如，"namespace/project"）。
//
// 函数返回两个值：
//   - 一个布尔值，表示项目是否存在（true）或不存在（false）。
//   - 一个 error，如果在过程中发生错误，它将是非空的。如果项目不存在，
//     但是向 GitLab 的查询是成功的（即，项目只是不存在），
//     则函数将返回 false 和一个 nil error。
//
// 该函数通过尝试从 GitLab 获取项目信息来工作。如果请求成功
// （即使项目不存在），函数将返回一个 nil error。
// 如果请求失败（例如，由于网络问题、认证问题等），
// 函数将返回相应的错误。
func (c *Client) IsProjectExist(projectWithNamespace string) (bool, error) {

	_, resp, err := c.git.Projects.GetProject(projectWithNamespace, &gitlab.GetProjectOptions{})
	if err != nil {
		// 如果在尝试获取项目时发生错误，我们检查状态码。
		// 如果状态码是 404，这意味着项目不存在，但请求是成功的。
		// 在这种情况下，我们返回 false 和一个 nil error。
		// 对于任何其他错误，我们返回 false 和错误。
		if resp != nil && resp.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}

	// 如果我们到了这一点，这意味着项目存在，所以我们返回 true 和一个 nil error
	return true, nil
}

// GetProjectArchive 通过项目名（包括命名空间）获取指定项目的源码压缩包
// 输入参数 projectWithNamespace 是包括命名空间的 GitLab 项目名。
// 返回值是一个字节切片，其中包含了项目的 tar.gz 归档文件。如果在获取归档文件过程中发生错误，会返回一个非 nil 的 error。
func (c *Client) GetProjectArchive(projectWithNamespace string) ([]byte, error) {

	// 首先，需要根据项目名获取项目的详细信息
	project, _, err := c.git.Projects.GetProject(projectWithNamespace, &gitlab.GetProjectOptions{})
	if err != nil {
		// 如果在获取项目信息过程中发生错误，返回 nil 和错误
		return nil, err
	}

	// 然后，使用 GitLab 客户端的 Repositories.Archive 方法获取项目归档
	data, _, err := c.git.Repositories.Archive(project.ID, &gitlab.ArchiveOptions{
		Format: gitlab.String("tar.gz"),
	})

	if err != nil {
		// 如果在获取归档文件过程中发生错误，返回 nil 和错误
		return nil, err
	}

	// 如果一切正常，返回归档文件的字节切片和 nil error
	return data, nil
}

// CreateProjectInGroup 在指定的 GitLab 组内创建一个新项目。
// name 参数是新项目的名称。
// group 参数是项目所属的 GitLab 组的名称。
// desc 参数是新项目的描述。
// 如果项目创建成功，返回 nil error。
// 如果在查找组 ID 或创建项目过程中出现错误，返回对应的 error。
func (c *Client) CreateProjectInGroup(name, group, desc string) error {
	namespaceID, err := c.getGroupID(group)
	if err != nil {
		return err
	}

	opt := &gitlab.CreateProjectOptions{
		Name:        gitlab.String(name),
		Description: gitlab.String(desc),
		NamespaceID: gitlab.Int(namespaceID),
		Visibility:  gitlab.Visibility(gitlab.PrivateVisibility),
	}
	_, _, err = c.git.Projects.CreateProject(opt)
	return err
}

func (c *Client) EnableRunner(sourceProjectID, targetProjectID string) error {
	runners, _, err := c.git.Runners.ListProjectRunners(sourceProjectID, &gitlab.ListProjectRunnersOptions{})
	if err != nil {
		return err
	}

	for _, runner := range runners {
		if !runner.IsShared {
			_, _, err := c.git.Runners.EnableProjectRunner(targetProjectID, &gitlab.EnableProjectRunnerOptions{
				RunnerID: runner.ID,
			})
			if err != nil {
				fmt.Printf("Failed to enable runner %d for project %s: %v\n", runner.ID, targetProjectID, err)
			} else {
				fmt.Printf("Runner %d enabled for project %s\n", runner.ID, targetProjectID)
			}
		}
	}

	return nil
}

func (c *Client) CopyProjectVariables(sourceProjectID, targetProjectID string) error {
	vars, _, err := c.git.ProjectVariables.ListVariables(sourceProjectID, &gitlab.ListProjectVariablesOptions{})
	if err != nil {
		return err
	}

	if len(vars) == 0 {
		fmt.Printf("No variables found in source project %s\n", sourceProjectID)
		return nil
	}

	for _, v := range vars {
		_, _, err := c.git.ProjectVariables.CreateVariable(targetProjectID, &gitlab.CreateProjectVariableOptions{
			Key:              gitlab.String(v.Key),
			Value:            gitlab.String(v.Value),
			Protected:        gitlab.Bool(v.Protected),
			Masked:           gitlab.Bool(v.Masked),
			EnvironmentScope: gitlab.String(v.EnvironmentScope),
		})
		if err != nil {
			fmt.Printf("Failed to copy variable %s to project %s: %v\n", v.Key, targetProjectID, err)
		} else {
			fmt.Printf("Variable %s copied to project %s\n", v.Key, targetProjectID)
		}
	}

	return nil
}

func (c *Client) CreateCommitFromFiles(projectID string, files map[string]*FileData) error {
	// 这个切片用于保存 CommitActionOptions
	var actions []*gitlab.CommitActionOptions

	for path, fileData := range files {
		options := &gitlab.CommitActionOptions{
			Action:   gitlab.FileAction(gitlab.FileCreate),
			FilePath: gitlab.String(path),
			Content:  gitlab.String(fileData.Content),
			Encoding: gitlab.String(fileData.Encoding),
		}
		actions = append(actions, options)
	}

	_, _, err := c.git.Commits.CreateCommit(projectID, &gitlab.CreateCommitOptions{
		Actions:       actions,
		Branch:        gitlab.String("master"),
		CommitMessage: gitlab.String("init project [skip ci]"),
	})
	return err
}

func (c *Client) DeleteProject(projectID string) error {
	_, err := c.git.Projects.DeleteProject(projectID)
	return err
}

func (c *Client) CreateBranch(projectID, branch, ref string) error {
	_, _, err := c.git.Branches.CreateBranch(projectID, &gitlab.CreateBranchOptions{
		Branch: gitlab.String(branch),
		Ref:    gitlab.String(ref),
	})
	return err
}

func (c *Client) SetDefaultBranch(projectID, branch string) error {
	_, _, err := c.git.Projects.EditProject(projectID, &gitlab.EditProjectOptions{
		DefaultBranch: gitlab.String(branch),
	})
	return err
}

/*
Copyright © 2023 Xu Wu <ixw1991@126.com>
Use of this source code is governed by a MIT style
license that can be found in the LICENSE file.
*/
package main

import (
	"log"

	"github.com/imxw/gitlab-scaffold/cmd"
	"github.com/imxw/gitlab-scaffold/internal/config"
)

func main() {
	if err := config.InitializeConfig(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	cmd.Execute()
}

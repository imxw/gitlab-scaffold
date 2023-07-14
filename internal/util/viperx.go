/*
Copyright © 2023 Xu Wu <ixw1991@126.com>
Use of this source code is governed by a MIT style
license that can be found in the LICENSE file.
*/
package util

import (
	"log"

	"github.com/spf13/viper"
)

func CheckRequiredConfigs(configs ...string) {
	for _, config := range configs {
		if !viper.IsSet(config) {
			log.Fatalf("ERROR: %s configuration is required", config)
		}
	}
}

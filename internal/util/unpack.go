/*
Copyright © 2023 Xu Wu <ixw1991@126.com>
Use of this source code is governed by a MIT style
license that can be found in the LICENSE file.
*/
package util

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UnpackTarGz 解压一个 tar.gz 文件到指定的目的地，返回解压后的根目录全路径。
// 输入参数 data 是包含压缩文件内容的字节切片，dst 是解压文件的目标目录。
// 如果在解压过程中发生错误，会返回一个非 nil 的 error。
func UnpackTarGz(data []byte, dst string) (string, error) {
	// 创建 gzip 解压器
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer gzipReader.Close()

	// 创建 tar 解压器
	tr := tar.NewReader(gzipReader)

	var rootDir string

	// 遍历 tar 压缩包中的文件
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// 如果是 pax_global_header 文件，跳过处理
		if hdr.Name == "pax_global_header" {
			continue
		}

		// 从 tar 压缩包中第一个文件的名称获取根目录
		if rootDir == "" {
			// 根目录是第一个文件的路径的第一部分
			rootDir = strings.Split(hdr.Name, "/")[0]
		}

		// 创建完整的目的地路径
		destPath := filepath.Join(dst, hdr.Name)

		// 检查文件是否是一个目录
		if hdr.FileInfo().IsDir() {
			// 创建目录
			err := os.MkdirAll(destPath, hdr.FileInfo().Mode())
			if err != nil {
				return "", err
			}
		} else {
			// 创建文件并将内容写入文件
			file, err := os.OpenFile(destPath, os.O_CREATE|os.O_RDWR, hdr.FileInfo().Mode())
			if err != nil {
				return "", err
			}
			defer file.Close()

			if _, err := io.Copy(file, tr); err != nil {
				return "", err
			}
		}
	}

	return filepath.Join(dst, rootDir), nil
}

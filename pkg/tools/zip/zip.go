package zip

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ZipFiles 压缩文件
func ZipFiles(filename string, files []string) error {
	zipfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	for _, file := range files {
		err := addFileToZip(archive, file)
		if err != nil {
			return err
		}
	}

	return nil
}

// UnzipFile 解压文件，含 Zip Slip 防护
func UnzipFile(zipFile string, destDir string) error {
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 确保目标目录以分隔符结尾，用于路径穿越检测
	destDir = filepath.Clean(destDir) + string(os.PathSeparator)

	for _, file := range reader.File {
		err := extractFile(file, destDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func addFileToZip(archive *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer, err := archive.Create(filepath.Base(filename))
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

func extractFile(file *zip.File, destDir string) error {
	// Zip Slip 防护：校验目标路径是否在 destDir 范围内
	targetPath := filepath.Join(destDir, file.Name)

	// 清理并检查路径穿越
	cleanTarget := filepath.Clean(targetPath)
	if !strings.HasPrefix(cleanTarget, destDir) {
		return nil // 跳过恶意文件，不返回错误以继续解压其他文件
	}

	if file.FileInfo().IsDir() {
		return os.MkdirAll(cleanTarget, os.ModePerm)
	}

	if err := os.MkdirAll(filepath.Dir(cleanTarget), os.ModePerm); err != nil {
		return err
	}

	dest, err := os.OpenFile(cleanTarget, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer dest.Close()

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	_, err = io.Copy(dest, src)
	return err
}

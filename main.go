package main

import (
	"archive/zip"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"strings"

	"os"
	"path/filepath"
	"golang.org/x/image/webp"

)


func gatherZips(dirPath string)(zs []string, err error){
	// 遍历目录下的所有zip文件
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 如果是 WebP 文件，则进行转换
		if filepath.Ext(path) == ".zip" {

			z := filepath.Join(filepath.Dir(path), filepath.Base(path))

			zipFile, err := zip.OpenReader(path)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			defer zipFile.Close()


			// 遍历 zip 文件中的所有文件
			iswebp := false
			for _, zipEntry := range zipFile.File {
				// fmt.Println("File:", zipEntry.Name)

				// 如果是要读取的文件，则读取其内容
				// fmt.Println("filepath.ext:", filepath.Ext(zipEntry.Name))
				if filepath.Ext(zipEntry.Name) == ".webp" {
					iswebp = true
					break
					fileReader, err := zipEntry.Open()
					if err != nil {
						fmt.Println(err)
						return nil
					}
					defer fileReader.Close()

					fileContents, err := io.ReadAll(fileReader)
					if err != nil {
						fmt.Println(err)
						return nil
					}

					fmt.Println("Contents:", string(fileContents))
				}
			}

			if iswebp{
				zs = append(zs,z)
			}

			// jpgPath := filepath.Join(filepath.Dir(path), filepath.Base(path[0:len(path)-5])+".jpg")
			// err := convertToJpg(path, jpgPath)
			// if err != nil {
			// 	return err
			// }

		}

		return nil
	})

	if err != nil {

		return nil,err
	}

	return zs,nil
}


func main() {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get current directory:", err)
	}
	fmt.Println("Current directory is:", currentDir)

	zs,err := gatherZips(currentDir)
	if err != nil{
		log.Fatal(err)
	}

	if len(zs) == 0{
		fmt.Println("No zips needed help")
		os.Exit(0)
	}
	for _,z := range zs {
		helpWithJpg(z)
		rezip(z)
	}
}

func rezip(path string)  {
	baseName := path[:len(path)-4]


	err := zipDir(baseName, path)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Rezip successful:", path)
	}

	// delete source files
	os.RemoveAll(baseName)
}

func zipDir(source,destFile string)error{
	// 预防：旧文件无法覆盖
	os.RemoveAll(destFile)

	// 创建：zip文件
	zipfile, _ := os.Create(destFile)
	defer zipfile.Close()

	// 打开：zip文件
	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	// 遍历路径信息
	err := filepath.Walk(source, func(path string, info os.FileInfo, _ error) error {

		// 如果是源路径，提前进行下一个遍历
		if path == source {
			return nil
		}

		// 获取：文件头信息
		header, _ := zip.FileInfoHeader(info)
		header.Name = strings.TrimPrefix(path, source+`\`)

		// 判断：文件是不是文件夹
		if info.IsDir() {
			header.Name += `/`
		} else {
			// 设置：zip的文件压缩算法
			header.Method = zip.Deflate
		}

		// 创建：压缩包头部信息
		writer, _ := archive.CreateHeader(header)
		if !info.IsDir() {
			file, _ := os.Open(path)
			defer file.Close()
			io.Copy(writer, file)
		}
		return nil
	})
	if err != nil{
		return nil
	}
	return nil
}

func helpWithJpg(path string){
	// 打开 zip 文件
	fmt.Println("Help:",path)
	zipFile, err := zip.OpenReader(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer zipFile.Close()

	baseName := path[:len(path)-4]
	// fmt.Println("BasePath:",baseName)

	os.Mkdir(baseName, 660)

	// 遍历 zip 文件中的所有文件
	for _, zipEntry := range zipFile.File {
		// 如果是要读取的文件，则读取其内容
		if filepath.Ext(zipEntry.Name) == ".webp" {
			fileReader, err := zipEntry.Open()
			if err != nil {
				fmt.Println(err)
				return
			}

			// 创建 JPG 文件
			outFile, err := os.Create(filepath.Join(baseName, zipEntry.Name[:len(zipEntry.Name)-5])+".jpeg")
			if err != nil {
				fmt.Println("Failed to create file:", err)
				return
			}

			// 解码 WebP 图片
			img, err := webp.Decode(fileReader)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// 编码为 JPG 格式
			err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 100})
			if err != nil {
				fmt.Println("jpeg encode error",err)
			}

			fileReader.Close()
			outFile.Close()
		}
	}

}

func convertToJpg(webpPath, jpgPath string) error {
	// 读取 WebP 文件
	file, err := os.Open(webpPath)
	if err != nil {
		return err
	}
	defer file.Close()


	// 解码 WebP 图片
	img, err := webp.Decode(file)
	if err != nil {
		return err
	}

	// 创建 JPG 文件
	outFile, err := os.Create(jpgPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// 编码为 JPG 格式
	err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 80})
	if err != nil {
		return err
	}

	return nil
}

func batchConvertToJpg(dirPath string) error {
	// 遍历目录下的所有文件
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 如果是 WebP 文件，则进行转换
		if filepath.Ext(path) == ".webp" {
			jpgPath := filepath.Join(filepath.Dir(path), filepath.Base(path[0:len(path)-5])+".jpg")
			err := convertToJpg(path, jpgPath)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}


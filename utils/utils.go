package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func GetMimetypeFromFilePath(filepath string) (fileType string, err error) {
	file, err := os.Open(filepath)
	if err != nil {
		return fileType, err
	}
	defer file.Close()

	// 读取前 512 个字节的内容
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return fileType, err
	}

	// 判断文件类型
	fileType = http.DetectContentType(buffer)
	return fileType, nil
}

func DownloadFileFromUrl(filepath string, url string) (err error) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func RemoveFileIfExist(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		err = os.Remove(filePath) //remove the file using built-in functions
		if err == nil {
			return true
		} else {
			fmt.Println(err)
		}
	}
	return false
}

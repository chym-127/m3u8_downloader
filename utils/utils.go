package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"io/ioutil"
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

func DecryptFile(key []byte, filename string, outFilename string) (string, error) {
	if len(outFilename) == 0 {
		outFilename = filename + ".dec"
	}

	ciphertext, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	of, err := os.Create(outFilename)
	if err != nil {
		return "", err
	}
	defer of.Close()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize]) //初始向量的长度必须等于块block的长度16字节
	origData := make([]byte, len(ciphertext))
	blockMode.CryptBlocks(origData, ciphertext)
	origData = PKCS5UnPadding(origData)

	if _, err := of.Write(origData); err != nil {
		return "", err
	}
	return outFilename, nil
}

//@brief:填充明文
func PKCS5Padding(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padtext...)
}

//@brief:去除填充数据
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func ReadFile2Str(filePath string) string {
	b, _ := ioutil.ReadFile(filePath) // b has type []byte
	str := string(b)
	return str
}

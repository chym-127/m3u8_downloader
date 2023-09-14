package utils

import (
	"fmt"
	"testing"
)

// func TestGetMimetypeFromFilePath(t *testing.T) {
// 	fileType, err := GetMimetypeFromFilePath("D:\\Download\\downloads\\segments\\192.ts")
// 	if err != nil {
// 		t.Error(`GetMimetypeFromFilePath("D:\\Download\\downloads\\segments\\110.ts") = false`)
// 	}
// 	fmt.Println(fileType)
// }

// func TestDownloadFileFromUrl(t *testing.T) {
// 	err := DownloadFileFromUrl("1.ts", "https://vip.lz-cdn12.com/20230818/9346_61aa9917/2000k/hls/b269a1b2081000001.ts")
// 	if err != nil {
// 		t.Error(`TestDownloadFileFromUrl("1.ts", "https://vip.lz-cdn12.com/20230818/9346_61aa9917/2000k/hls/b269a1b2081000001.ts") = false`)
// 	}
// }

func TestDecryptFile(t *testing.T) {
	key := []byte("68a2cf3899dfbbcf")
	_, err := DecryptFile(key, "F:\\movies\\黄海\\xfnbHlFLSaQU\\segments\\0.ts", "F:\\movies\\黄海\\xfnbHlFLSaQU\\segments\\0-dec.ts")
	if err != nil {
		fmt.Println(err)
		t.Error(`faile") = false`)
	}
}

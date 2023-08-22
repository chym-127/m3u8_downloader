package utils

import (
	"fmt"
	"testing"
)

func TestGetMimetypeFromFilePath(t *testing.T) {
	fileType, err := GetMimetypeFromFilePath("D:\\Download\\downloads\\segments\\192.ts")
	if err != nil {
		t.Error(`GetMimetypeFromFilePath("D:\\Download\\downloads\\segments\\110.ts") = false`)
	}
	fmt.Println(fileType)
}

func TestDownloadFileFromUrl(t *testing.T) {
	err := DownloadFileFromUrl("1.ts", "https://vip.lz-cdn12.com/20230818/9346_61aa9917/2000k/hls/b269a1b2081000001.ts")
	if err != nil {
		t.Error(`TestDownloadFileFromUrl("1.ts", "https://vip.lz-cdn12.com/20230818/9346_61aa9917/2000k/hls/b269a1b2081000001.ts") = false`)
	}
}

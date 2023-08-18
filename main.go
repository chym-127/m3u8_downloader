package main

import (
	"bufio"
	"errors"
	"fmt"
	"m3u8_downloader/utils"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"

	"github.com/jessevdk/go-flags"
	"github.com/panjf2000/ants/v2"
	"github.com/schollz/progressbar/v3"
)

var (
	DOWNLOAD_PATH                  = ""
	SEGMENTS_PATH                  = ""
	TS_FILE_NAME                   = "all.txt"
	OUTPUT_FILE_PATH               = ""
	DEFAULT_INPUT_FILE_PATH string = ""
)

type Option struct {
	M3u8FilePath   string `short:"i" long:"input" description:"m3u8文件路径"`
	OutPutFileName string `short:"o" long:"output" description:"输出文件名" default:"output"`
}

func main() {
	var opt Option
	flags.Parse(&opt)
	filePath := opt.M3u8FilePath
	outputFileName := opt.OutPutFileName
	initEnv(outputFileName)

	if filePath == "" {
		return
	} else {
		url_regexp, _ := regexp.Compile("http.*")
		if url_regexp.MatchString(filePath) {
			_ = utils.DownloadFileFromUrl(DEFAULT_INPUT_FILE_PATH, filePath)
			filePath = DEFAULT_INPUT_FILE_PATH
		}
	}

	info := parseM3u8(filePath, outputFileName)
	bar := progressbar.NewOptions(int(info.total_segment),
		progressbar.OptionEnableColorCodes(true),
		// progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[reset]Download file..."),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	execStep1(info, bar)
	bar.Exit()
	defer bar.Close()

	execStep2(info)

}

func execStep2(info M3u8Info) {
	outputMp4(info)
}

func execStep1(info M3u8Info, bar *progressbar.ProgressBar) {
	var wg sync.WaitGroup
	p, _ := ants.NewPool(15)
	defer p.Release()

	for _, item := range info.segments {
		wg.Add(1)
		filePath := filepath.Join(SEGMENTS_PATH, strconv.Itoa(int(item.index-1))+".ts")
		p.Submit(taskFuncWrapper(filePath, item.url, &wg, bar))
	}

	wg.Wait()
}

func initEnv(outputFileName string) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	DOWNLOAD_PATH = filepath.Join(exPath, "downloads")
	SEGMENTS_PATH = filepath.Join(DOWNLOAD_PATH, "segments")
	_ = os.Mkdir(DOWNLOAD_PATH, os.ModeDir)
	_ = os.Mkdir(SEGMENTS_PATH, os.ModeDir)

	DEFAULT_INPUT_FILE_PATH = filepath.Join(DOWNLOAD_PATH, "video.m3u8")
	OUTPUT_FILE_PATH = filepath.Join(DOWNLOAD_PATH, outputFileName+".mp4")
}

func removeFileIfExist(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		err = os.Remove(filePath) //remove the file using built-in functions
		if err == nil {
			return true
		}
	}
	return false
}

func outputMp4(info M3u8Info) {
	fmt.Println()
	fmt.Println("Generating video files in progress")

	removeFileIfExist(OUTPUT_FILE_PATH)
	fileName := createTsFile(info)

	args := []string{"-f", "concat", "-i", fileName, "-c", "copy", "-bsf:a", "aac_adtstoasc", OUTPUT_FILE_PATH}

	cmd := exec.Command("ffmpeg", args...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully！Have fun.")

}

//暂时不用
func createTsFile(info M3u8Info) string {
	fileName := filepath.Join(SEGMENTS_PATH, TS_FILE_NAME)
	removeFileIfExist(fileName)
	file, _ := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	datawriter := bufio.NewWriter(file)

	for i, item := range info.segments {
		str := "file " + strconv.Itoa(i) + ".ts\n"
		str += "duration " + item.duration_str

		_, _ = datawriter.WriteString(str + "\n")
	}

	datawriter.Flush()
	file.Close()
	return fileName
}

type taskFunc func()

func taskFuncWrapper(filepath string, url string, wg *sync.WaitGroup, bar *progressbar.ProgressBar) taskFunc {
	return func() {
		if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
			err := utils.DownloadFileFromUrl(filepath, url)
			if err != nil {
				fmt.Println(err)
			}
		}
		bar.Add(1)
		wg.Done()
	}
}

type M3u8Info struct {
	name           string    //名称
	total_segment  int32     //子元素的个数
	total_duration float64   //总时长
	segments       []Segment //子元素
}

type Segment struct {
	index        int32   //位置索引
	duration     float64 //时长
	duration_str string
	url          string //地址
}

func parseM3u8(filePath string, name string) M3u8Info {
	info := M3u8Info{
		name:           name,
		total_segment:  0,
		total_duration: 0.0,
		segments:       []Segment{},
	}
	url_regexp, _ := regexp.Compile("http.*.ts")
	duration_regexp, _ := regexp.Compile("#EXTINF:.*")
	number_re := regexp.MustCompile("[0-9]+.[0-9]+")
	number_re1 := regexp.MustCompile("[0-9]+")

	readFile, err := os.Open(filePath)

	if err != nil {
		fmt.Println(err)
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	var count_index int32 = 1
	for fileScanner.Scan() {
		line := fileScanner.Text()
		if duration_regexp.MatchString(line) {
			str := number_re.FindString(line)
			if len(str) == 0 {
				str = number_re1.FindString(line)
			}
			var d float64 = 0.0
			if len(str) > 0 {
				v, _ := strconv.ParseFloat(str, 64)
				d, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", v), 64)
			}
			s := Segment{
				index:        count_index,
				duration:     d,
				duration_str: str,
			}
			if fileScanner.Scan() {
				line = fileScanner.Text()
				if url_regexp.MatchString(line) {
					s.url = line
				}
			}
			info.total_segment += 1
			info.total_duration += s.duration
			count_index += 1
			info.segments = append(info.segments, s)
		}
	}
	return info
}

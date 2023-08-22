package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"m3u8_downloader/utils"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/jessevdk/go-flags"
	"github.com/panjf2000/ants/v2"
	"github.com/schollz/progressbar/v3"
)

var (
	DOWNLOAD_PATH                  = ""
	SEGMENTS_PATH                  = ""
	LOG_PATH                       = ""
	TS_FILE_NAME                   = "all.txt"
	OUTPUT_FILE_PATH               = ""
	DEFAULT_INPUT_FILE_PATH string = ""
)

var (
	RETRY_DOWNLOAD_COUNT = 3  //文件下载重试次数
	MAXIMUM_CONCURRENCY  = 15 // 最大并发数
	IGNORED_DOWNFAIL     = false
)

type Option struct {
	M3u8FilePath       string `short:"i" long:"input" description:"m3u8文件路径"`
	OutPutFileName     string `short:"o" long:"output" description:"输出文件名" default:"output"`
	RetryDownloadCount int    `short:"r" long:"retry" description:"文件下载重试次数" default:"3"`
	MaximumCoucurrency int    `short:"d" long:"download" description:"下载最大并发数" default:"15"`
	IgnoredDownFail    bool   `short:"g" long:"ignore" description:"忽略下载失败的文件"`
}

func main() {
	var opt Option
	var baseUrl = ""
	flags.Parse(&opt)

	filePath := opt.M3u8FilePath
	outputFileName := opt.OutPutFileName
	RETRY_DOWNLOAD_COUNT = opt.RetryDownloadCount
	MAXIMUM_CONCURRENCY = opt.MaximumCoucurrency
	IGNORED_DOWNFAIL = opt.IgnoredDownFail

	initEnv(outputFileName)
	logFile := utils.InitLogger(LOG_PATH)
	defer logFile.Close()

	// 先下载m3u8文件
	if filePath == "" {
		return
	} else {
		url_regexp, _ := regexp.Compile("http.*")
		if url_regexp.MatchString(filePath) {
			utils.Println("Start downloading m3u8 file:", filePath)
			baseUrl = getBaseUrl(filePath)
			err := utils.DownloadFileFromUrl(DEFAULT_INPUT_FILE_PATH, filePath)
			if err != nil {
				utils.Println("m3u8 file download failure")
				return
			}
			utils.Println("m3u8 file download successfully")
			filePath = DEFAULT_INPUT_FILE_PATH
		}
	}

	utils.Println("Start parsing m3u8 files")
	info := parseM3u8(filePath, outputFileName, baseUrl)
	utils.Println("Name: " + info.name)
	utils.Println("BaseUrl: " + info.base_url)
	utils.Println("Duration: " + fmt.Sprintf("%.2f", info.total_duration) + "s")
	utils.Println("SegmentLength: " + strconv.Itoa(int(info.total_segment)))

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

	execStep1(&info, bar)

	fmt.Println()
	utils.Println("Ts file downloaded successfully")

	bar.Exit()
	defer bar.Close()

	execStep2(info)

}

func getBaseUrl(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		utils.Println(err)
	}
	baseUrl := u.Scheme + "://" + u.Host
	fileNameReg, _ := regexp.Compile(".?.m3u8")
	name := path.Base(uri)
	if fileNameReg.MatchString(name) {
		baseUrl += path.Dir(u.Path)
	} else {
		baseUrl += u.Path
	}
	return baseUrl
}

func execStep1(info *M3u8Info, bar *progressbar.ProgressBar) {
	var wg sync.WaitGroup
	p, _ := ants.NewPool(MAXIMUM_CONCURRENCY)
	defer p.Release()

	for _, item := range info.segments {
		wg.Add(1)
		filePath := filepath.Join(SEGMENTS_PATH, strconv.Itoa(int(item.index-1))+".ts")
		p.Submit(taskFuncWrapper(info, int(item.index-1), filePath, &wg, bar))
	}

	wg.Wait()
}

func initEnv(outputFileName string) {
	exPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// exPath := filepath.Dir(ex)
	DOWNLOAD_PATH = filepath.Join(exPath, "downloads")
	SEGMENTS_PATH = filepath.Join(DOWNLOAD_PATH, "segments")
	LOG_PATH = filepath.Join(DOWNLOAD_PATH, "log.txt")

	_ = os.Mkdir(DOWNLOAD_PATH, os.ModeDir)
	_ = os.Mkdir(SEGMENTS_PATH, os.ModeDir)

	DEFAULT_INPUT_FILE_PATH = filepath.Join(DOWNLOAD_PATH, "video.m3u8")
	OUTPUT_FILE_PATH = filepath.Join(DOWNLOAD_PATH, outputFileName+".mp4")
}

func execStep2(info M3u8Info) {
	utils.Println("Generating video files in progress")

	utils.RemoveFileIfExist(OUTPUT_FILE_PATH)
	fileName, failCount := createTsFile(info)
	if !IGNORED_DOWNFAIL && failCount > 0 {
		utils.Error("Partial file download failure, export failure")
		return
	}

	args := []string{"-f", "concat", "-i", fileName, "-c", "copy", "-bsf:a", "aac_adtstoasc", OUTPUT_FILE_PATH}
	log.Println("ffmpeg command :ffmpeg " + strings.Join(args, " "))
	cmd := exec.Command("ffmpeg", args...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		utils.Println(err)
	}

	utils.Println("Successfully！Have fun.")
}

func createTsFile(info M3u8Info) (string, int) {
	fileName := filepath.Join(SEGMENTS_PATH, TS_FILE_NAME)
	utils.RemoveFileIfExist(fileName)
	file, _ := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	datawriter := bufio.NewWriter(file)

	failCount := 0
	for i, item := range info.segments {
		if item.is_download {
			str := "file " + strconv.Itoa(i) + ".ts\n"
			str += "duration " + item.duration_str
			_, _ = datawriter.WriteString(str + "\n")
		} else {
			failCount += 1
		}
	}

	datawriter.Flush()
	file.Close()
	return fileName, failCount
}

type taskFunc func()

func taskFuncWrapper(info *M3u8Info, index int, filepath string, wg *sync.WaitGroup, bar *progressbar.ProgressBar) taskFunc {
	return func() {
		url := info.segments[index].url
		for i := 1; i <= RETRY_DOWNLOAD_COUNT; i++ {
			if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
				utils.DownloadFileFromUrl(filepath, url)
			}
			fileType, err := utils.GetMimetypeFromFilePath(filepath)

			if err == nil && fileType == "application/octet-stream" {
				break
			} else {
				if i == RETRY_DOWNLOAD_COUNT {
					info.segments[index].is_download = false
					log.Println(strconv.Itoa(int(index)) + ".ts: " + url + " 下载失败")
				}
				utils.RemoveFileIfExist(filepath)
			}
		}
		bar.Add(1)
		wg.Done()
	}
}

type M3u8Info struct {
	name           string //名称
	base_url       string
	total_segment  int32     //子元素的个数
	total_duration float64   //总时长
	segments       []Segment //子元素
}

type Segment struct {
	index        int32   //位置索引
	duration     float64 //时长
	duration_str string
	url          string //地址
	is_download  bool
}

func parseM3u8(filePath string, name string, baseUrl string) M3u8Info {
	info := M3u8Info{
		name:           name,
		base_url:       baseUrl,
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
		utils.Println(err)
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
				d, _ = strconv.ParseFloat(fmt.Sprintf("%.10f", v), 64)
			}
			s := Segment{
				index:        count_index,
				duration:     d,
				duration_str: str,
				is_download:  true,
			}
			if fileScanner.Scan() {
				line = fileScanner.Text()
				if url_regexp.MatchString(line) {
					s.url = line
				} else {
					s.url = info.base_url + "/" + line
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

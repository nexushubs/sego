package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/nexushubs/sego"
)

var (
	host         = flag.String("host", "", "HTTP服务器主机名")
	port         = flag.Int("port", 5678, "HTTP服务器端口")
	dict         = flag.String("dict", getDict(), "词典文件")
	staticFolder = flag.String("static_folder", "static", "静态页面存放的目录")
	segmenter    = sego.Segmenter{}
)

type JsonResponse struct {
	Segments []*Segment `json:"segments"`
}

type Segment struct {
	Text string `json:"text"`
	Pos  string `json:"pos"`
}

func JsonRpcServer(w http.ResponseWriter, req *http.Request) {
	// 得到要分词的文本
	text := req.URL.Query().Get("text")
	if text == "" {
		text = req.PostFormValue("text")
	}

	// 分词
	segments := segmenter.Segment([]byte(text))

	// 整理为输出格式
	ss := []*Segment{}
	for _, segment := range segments {
		ss = append(ss, &Segment{Text: segment.Token().Text(), Pos: segment.Token().Pos()})
	}
	response, _ := json.Marshal(&JsonResponse{Segments: ss})

	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(response))
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getDict() string {
	dictionary := os.Getenv("DICT")
	if !fileExists(dictionary) {
		dictionary = "./dict.txt"
		if !fileExists(dictionary) {
			dictionary = "../data/dict.txt"
			if !fileExists(dictionary) {
				panic("Unable to locate dict file.")
			}
		}
	}
	return dictionary
}

func main() {
	flag.Parse()

	// 将线程数设置为CPU数
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 初始化分词器
	segmenter.LoadDictionary(*dict)

	http.HandleFunc("/json", JsonRpcServer)
	http.Handle("/", http.FileServer(http.Dir(*staticFolder)))
	log.Print("服务器启动")
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", *host, *port), nil)
	if err != nil {
		log.Println(err)
	}
}

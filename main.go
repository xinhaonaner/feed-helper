package main

import (
	"embed"
	"flag"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"feed-helper/internal/xml2csv"
)

// utf8BOM 让 Excel 等 Windows 程序正确识别 UTF-8，避免乱码和错位。
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// crlfWriter 将 \n 转为 \r\n，便于 Windows 下 Excel 正确识别换行。
type crlfWriter struct{ w io.Writer }

func (c *crlfWriter) Write(p []byte) (n int, err error) {
	start := 0
	for i := 0; i < len(p); i++ {
		if p[i] == '\n' {
			if _, err = c.w.Write(p[start:i]); err != nil {
				return len(p), err
			}
			if _, err = c.w.Write([]byte{'\r', '\n'}); err != nil {
				return len(p), err
			}
			start = i + 1
		}
	}
	if start < len(p) {
		_, err = c.w.Write(p[start:])
	}
	return len(p), err
}

//go:embed static
var staticFS embed.FS

func main() {
	port := flag.String("port", "8080", "HTTP 监听端口")
	openBrowser := flag.Bool("open-browser", false, "启动时在默认浏览器中打开页面")
	flag.Parse()

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/convert", handleConvert)

	addr := ":" + *port
	if *openBrowser {
		go func() {
			url := "http://localhost" + addr
			openURL(url)
		}()
	}
	log.Printf("监听 %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "需要 POST", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("xml")
	if err != nil {
		http.Error(w, "请选择 XML 文件: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	rowTag := strings.TrimSpace(r.FormValue("rowTag"))
	if rowTag == "" {
		rowTag = "item"
	}

	w.Header().Set("Content-Disposition", `attachment; filename="result.csv"`)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")

	// Windows 兼容：先写 UTF-8 BOM，Excel 才能正确识别编码；再用 CRLF 换行避免错位。
	if _, err := w.Write(utf8BOM); err != nil {
		log.Printf("write BOM: %v", err)
		return
	}
	dst := &crlfWriter{w: w}
	opts := []xml2csv.Option{xml2csv.RowTag(rowTag)}
	if err := xml2csv.Convert(file, dst, opts...); err != nil {
		log.Printf("convert error: %v", err)
		// 不调用 http.Error：可能已向 w 写过数据（如 broken pipe 时客户端已断开），
		// 再写状态码会触发 superfluous response.WriteHeader。
		return
	}
}

func openURL(url string) {
	switch runtime.GOOS {
	case "windows":
		exec.Command("cmd", "/c", "start", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	default:
		// linux etc.
		exec.Command("xdg-open", url).Start()
	}
}

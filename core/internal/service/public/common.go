package public

import (
	"bufio"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"math/big"
	mRand "math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/util/gconv"

	"github.com/gogf/gf/v2/debug/gdebug"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/g0rbe/go-chattr"

	"billionmail-core/internal/consts"
)

// Check if it is a development environment
var isDev = false

// Project root directory
var ROOT_PATH = GetCwdPath()

type CodeStatus struct {
	Status   bool
	Msg      string
	ErrorMsg string
}

type StandardRes struct {
	g.Meta   `mime:"application/json"`
	Status   bool        `json:"status" description:"Status"`
	Code     int         `json:"code" description:"Status code"`
	Msg      string      `json:"msg" description:"Message"`
	ErrorMsg string      `json:"error_msg" description:"Error message"`
	Message  interface{} `json:"message"  description:"Data"`
}

type AccountSession struct {
	Username  string `json:"username" description:"username"`
	AccountId int    `json:"account_id" description:"Account ID"`
	Email     string `json:"email" description:"Email"`
}

type ReturnData struct {
	Status   int         `json:"status" description:"状态 0:成功 -1:失败"`
	Code     int         `json:"code" description:"状态码,请参考CodeMap"`
	Msg      string      `json:"msg" description:"提示信息"`
	ErrorMsg string      `json:"error_msg" description:"错误信息"`
	Message  interface{} `json:"message" description:"数据"`
}

type DefaultYamlConfig struct {
	Server struct {
		Address   string `yaml:"address"`
		HttpsAddr string `yaml:"httpsAddr"`
	} `yaml:"server"`
}

// Response status code
var CodeMap = map[int]CodeStatus{
	101: {
		Status:   true,
		Msg:      "Server is waiting for the next request",
		ErrorMsg: "",
	},
	200: {
		Status:   true,
		Msg:      "Operation successful",
		ErrorMsg: "",
	},
	401: {
		Status:   false,
		Msg:      "Not logged in",
		ErrorMsg: "Not logged in",
	},
	402: {
		Status:   false,
		Msg:      "Incorrect account or password",
		ErrorMsg: "Incorrect account or password",
	},
	403: {
		Status:   false,
		Msg:      "No permission",
		ErrorMsg: "No permission",
	},
	404: {
		Status:   false,
		Msg:      "Specified object does not exist",
		ErrorMsg: "Specified object does not exist",
	},
	405: {
		Status:   false,
		Msg:      "Incorrect verification code",
		ErrorMsg: "Incorrect verification code",
	},
	406: {
		Status:   false,
		Msg:      "Please enter the verification code",
		ErrorMsg: "Please enter the verification code",
	},
	412: {
		Status:   false,
		Msg:      "Parameter error",
		ErrorMsg: "Parameter error",
	},
	500: {
		Status:   false,
		Msg:      "Unknown error",
		ErrorMsg: "Unknown error",
	},
	501: {
		Status:   false,
		Msg:      "Unsupported request",
		ErrorMsg: "Unsupported request",
	},
	502: {
		Status:   false,
		Msg:      "Failed to connect to upstream service",
		ErrorMsg: "Failed to connect to upstream service",
	},
	503: {
		Status:   false,
		Msg:      "Service is under maintenance, please try again later",
		ErrorMsg: "Service is under maintenance, please try again later",
	},
	504: {
		Status:   false,
		Msg:      "Timeout waiting for upstream service response",
		ErrorMsg: "Timeout waiting for upstream service response",
	},
	509: {
		Status:   false,
		Msg:      "Access restricted, rate limit exceeded",
		ErrorMsg: "Access restricted, rate limit exceeded",
	},
}

// Convert string to integer
func IpToLong(ip string) int64 {
	if IsIpv4(ip) {
		return int64(Ip2Long(ip))
	} else {
		return IP2LongV6(ip)
	}

}

func Ip2Long(ip string) uint32 {
	quads := strings.Split(ip, ".")
	var result uint32 = 0
	a, _ := strconv.Atoi(quads[3])
	result += uint32(a)
	b, _ := strconv.Atoi(quads[2])
	result += uint32(b) << 8
	c, _ := strconv.Atoi(quads[1])
	result += uint32(c) << 16
	d, _ := strconv.Atoi(quads[0])
	result += uint32(d) << 24
	return result
}

func Long2Ip(ip uint32) string {
	var result string
	b := strconv.Itoa(int(ip & 0xFF))
	c := strconv.Itoa(int((ip >> 8) & 0xFF))
	d := strconv.Itoa(int((ip >> 16) & 0xFF))
	a := strconv.Itoa(int((ip >> 24) & 0xFF))
	result = a + "." + d + "." + c + "." + b
	return result
}

func IP2LongV6(ip string) int64 {
	ipv6 := net.ParseIP(ip)
	if ipv6 == nil {
		return 0
	}
	ipv6Int := big.NewInt(0)
	ipv6Int.SetBytes(ipv6.To16())
	return ipv6Int.Int64()
}

func Long2IPV6(ip int64) string {
	ipv6Int := big.NewInt(ip)
	return net.IP(ipv6Int.Bytes()).String()
}

// Convert string to integer
func bytesToLong(a, b, c, d byte) uint32 {
	a1 := uint32(a)
	b1 := uint32(b)
	c1 := uint32(c)
	d1 := uint32(d)
	return (a1 & 0xFF) | ((b1 << 8) & 0xFF00) | ((c1 << 16) & 0xFF0000) | ((d1 << 24) & 0xFF000000)
}

func bytesToLong3(a, b, c byte) uint32 {
	a1 := uint32(a)
	b1 := uint32(b)
	c1 := uint32(c)
	return (a1 & 0xFF) | ((b1 << 8) & 0xFF00) | ((c1 << 16) & 0xFF0000)

}

// Check if it is a LAN IP (for ipLong)
func IsLan(IpLong int64) bool {
	// Loopback address
	if IpLong >= 2130706432 && IpLong <= 2147483647 {
		return true
	}

	// LAN address
	if IpLong >= 167772161 && IpLong <= 184549375 || IpLong >= 2886729728 && IpLong <= 2887778303 || IpLong >= 3232235520 && IpLong <= 3232301055 {
		return true
	}

	return false
}

// Check if it is a LAN IP (for ipString)
func IsLanString(Ip string) bool {
	if Ip == "::1" {
		return true
	}
	ip_long := IpToLong(Ip)
	return IsLan(ip_long)
}

// Get database connection
func M(table string) *gdb.Model {
	return g.Model(table)
}

// Get database connection (specified database)
func MR(db string, table string) *gdb.Model {
	return g.DB(db).Model(table)
}

// Respond with JSON
func ReturnJson(ctx context.Context, data ReturnData) {
	jsonStr, _ := json.Marshal(data)

	// Respond with JSON and immediately exit the goroutine, terminating subsequent code execution
	g.RequestFromCtx(ctx).Response.WriteJsonExit(jsonStr)
}

// Respond with JSON
func Json(ctx context.Context, data ReturnData) {
	ReturnJson(ctx, data)
}

// Default response
func ReturnDefault(data interface{}) ReturnData {
	return ReturnData{Status: 0, Code: 200, Msg: "ok", Message: data}
}

// Response with text
func ReturnText(ctx context.Context, data string) {
	g.RequestFromCtx(ctx).Response.Write(data)
}

// Standard response
func Return(ctx context.Context, status bool, code int, msg string, data interface{}, error_msg string) {
	int_status := 0
	if !status {
		int_status = -1
	}

	rdata := ReturnData{}
	rdata.Status = int_status
	rdata.Code = code
	rdata.Msg = msg
	rdata.ErrorMsg = error_msg
	rdata.Message = data
	if rdata.Msg == "" {
		rdata.Msg = error_msg
	}

	ReturnJson(ctx, rdata)
}

// Standard response -- success
func Success(ctx context.Context, msg string, data ...interface{}) {
	length := len(data)
	error_msg := ""
	switch length {
	case 0:
		Return(ctx, true, 200, msg, nil, error_msg)
	case 1:
		Return(ctx, true, 200, msg, data[0], error_msg)
	default:
		Return(ctx, true, 200, msg, data, error_msg)
	}
}

// Standard response -- failure
func Error(ctx context.Context, msg string, error_msg string) {
	Return(ctx, false, 500, msg, nil, error_msg)
}

// Response based on status code
func ReturnCode(ctx context.Context, code int, data interface{}) {
	Return(ctx, CodeMap[code].Status, code, CodeMap[code].Msg, data, CodeMap[code].ErrorMsg)
}

// Get HTTP client object
func GetHttpClient(timeout int) *http.Client {
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // Skip certificate verification
		DisableKeepAlives: true,                                  // Disable connection pooling
	}
	HttpClient := &http.Client{
		Timeout:   time.Duration(timeout) * time.Second, // Set timeout
		Transport: tr,
	}
	return HttpClient
}

// Native HTTP POST request
func HttpPostSrc(url string, data url.Values, timeout int) (string, error) {
	client := GetHttpClient(timeout)
	resp, err := client.PostForm(url, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	context, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(context), nil
}

// Native HTTP GET request
func HttpGetSrc(url string, timeout int) (respBody string, err error) {
	client := GetHttpClient(timeout)
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < http.StatusOK && resp.StatusCode >= http.StatusMultipleChoices {
		err = fmt.Errorf("http get error: %s", resp.Status)
	}

	var bs []byte
	bs, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	respBody = string(bs)

	return
}

// Send GET request and return JSON data
func HttpGetJson(url string, timeout int, pointer interface{}) (err error) {
	var respBody string

	if respBody, err = HttpGetSrc(url, timeout); err != nil {
		return
	}

	return json.NewDecoder(strings.NewReader(respBody)).Decode(pointer)
}

// Calculate MD5 hash of a string
func Md5(str string) string {
	obj := md5.New()
	obj.Write([]byte(str))
	return hex.EncodeToString(obj.Sum(nil))
}

// Calculate MD5 hash of a file
func FileMd5(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	const bufferSize = 65536
	obj := md5.New()
	for buf, reader := make([]byte, bufferSize), bufio.NewReader(f); ; {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		obj.Write(buf[:n])
	}
	return hex.EncodeToString(obj.Sum(nil)), nil
}

// Decode JSON string
func JsonDecode(json_str string) (map[string]interface{}, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(json_str), &data)
	return data, err
}

// Convert ISO8601 time to timestamp
func Iso8601_To_Time(timestr string) int64 {
	// 2006-01-02 15:04:05 is similar to YYYY-MM-dd HH:mm:ss
	result, err := time.ParseInLocation("2006-01-02T15:04:05+08:00", timestr, time.Local)
	// Exit if error occurs
	if err != nil {
		return -1
	}
	return result.Unix()
}

// Get current date
func GetNowDate() string {
	return time.Now().Format("2006-01-02")
}

// Get current date as integer (e.g., 20210102)
func GetNowDateByInt() int {
	date_str := time.Now().Format("20060102")
	date_int, _ := strconv.Atoi(date_str)
	return date_int
}

// Get current date, hour, minute, and timestamp
func GetNowDateSplit() (int, int, int, int, int64) {
	unit_time := time.Now().Unix()
	date_full := time.Unix(unit_time, 0).Format("20060102 15:04:05")
	date_arr := strings.Split(date_full, " ")
	date, _ := strconv.Atoi(date_arr[0])

	time_arr := strings.Split(date_arr[1], ":")
	hour, _ := strconv.Atoi(time_arr[0])
	minute, _ := strconv.Atoi(time_arr[1])
	sencond, _ := strconv.Atoi(time_arr[2])

	return date, hour, minute, sencond, unit_time
}

// Get current timestamp
func GetNowTime() int64 {
	return time.Now().Unix()
}

// Get current timestamp in milliseconds
func GetNowTimeMillisecond() int64 {
	return time.Now().UnixMilli()
}

// Get current date and time
func GetNowTimeStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// Get timestamp for the next day at 00:00
func GetNextDayTime() int64 {
	t := time.Now()
	t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())
	return t.Unix()
}

// Get timestamp for the previous day at 00:00
func GetLastDayTime() int64 {
	t := time.Now()
	t = time.Date(t.Year(), t.Month(), t.Day()-1, 0, 0, 0, 0, t.Location())
	return t.Unix()
}

// Get timestamp for the previous hour
func GetLastHourTime() int64 {
	t := time.Now()
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour()-1, 0, 0, 0, t.Location())
	return t.Unix()
}

// Get timestamp for the previous minute
func GetLastMinuteTime() int64 {
	t := time.Now()
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute()-1, 0, 0, t.Location())
	return t.Unix()
}

// Get timestamp for the previous second
func GetLastSecondTime() int64 {
	t := time.Now()
	return t.Unix() - 1
}

// Get timestamp for the next hour
func GetNextHourTime() int64 {
	t := time.Now()
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, t.Location())
	return t.Unix()
}

// Get timestamp for the next minute
func GetNextMinuteTime() int64 {
	t := time.Now()
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute()+1, 0, 0, t.Location())
	return t.Unix()
}

// Get timestamp for the next second
func GetNextSecondTime() int64 {
	t := time.Now()
	return t.Unix() + 1
}

// Convert timestamp to date
func TimeToDate(timestamp int64, format string) string {
	tm := time.Unix(timestamp, 0)
	return tm.Format(format)
}

// URL decode
func UrlDecode(str string) string {
	if len(str) < 2 {
		return str
	}
	body, err := url.QueryUnescape(str)
	if err != nil {
		return UrlDecode(str[:len(str)-2])
	}
	return body
}

// URL encode
func UrlEncode(str string) string {
	return url.QueryEscape(str)
}

// Check if a key exists in an array
func InArray(key any, arr []any) bool {
	for _, v := range arr {
		if v == key {
			return true
		}
	}
	return false
}

// Get file list in a specified directory (only one level deep)
func GetDirFiles(path string) []string {
	var files []string
	_list, err := os.ReadDir(path)
	if err != nil {
		return files
	}
	for _, file := range _list {
		if file.IsDir() {
			continue
		}

		files = append(files, file.Name())
	}
	return files
}

// Check if a directory is empty
func IsEmptyDir(path string) bool {
	_list, err := os.ReadDir(path)
	if err != nil {
		return true
	}
	return len(_list) == 0
}

// Reverse DNS lookup by IP
func DnsLookup(ip string) (string, error) {
	arr, err := net.LookupAddr(ip)
	if err != nil {
		return "", err
	}

	result := ""
	for _, v := range arr {
		result += v + "\n"
	}

	return result, nil
}

// Get file modification time
func FileMtime(filename string) int64 {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return 0
	}
	return fileInfo.ModTime().Unix()
}

// Execute shell command
func ExecShell(command string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Get file size
func FileSize(filename string) int64 {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}

// Get current working directory
func GetCwdPath() string {
	if len(os.Args) == 0 {
		p, _ := os.Getwd()
		return p
	}

	// Get executable path
	exePath := filepath.Dir(os.Args[0])

	// Check if it is a development environment
	if isDev || strings.Contains(filepath.ToSlash(exePath), "/go-build") {
		isDev = true
		p, _ := os.Getwd()
		return p
	}

	// Get absolute path
	p, _ := filepath.Abs(exePath)
	return p
}

// Get absolute path
func AbsPath(p string) string {
	// In Linux, if the path starts with "/", it is an absolute path
	if strings.HasPrefix(p, "/") {
		return p
	}

	// In Windows, if the path starts with "C:", it is an absolute path
	if len(p) > 1 && p[1] == ':' {
		return p
	}

	p, _ = filepath.Abs(filepath.Join(ROOT_PATH, p))

	return p
}

// Calculate directory size
func CalcDirSize(path string) (size int64) {
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		size += info.Size()
		return nil
	})

	return
}

// Check if it is a directory
func IsDir(filename string) bool {
	fs, err := os.Stat(filename)

	if err != nil {
		return false
	}

	return fs.IsDir()
}

// Check if a file exists
func FileExists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

// Get relative path (specified root path)
func PathRelative(path, rootPath string) (pathRel string) {
	// Compatible with Linux path rules
	rootPath = strings.TrimPrefix(strings.TrimPrefix(filepath.ToSlash(rootPath), "./"), "/")
	pathRel = strings.TrimPrefix(strings.TrimPrefix(filepath.ToSlash(path), "./"), "/")
	pathRel = strings.TrimPrefix(strings.TrimPrefix(pathRel, rootPath), "/")
	return
}

// Read file bytes
func ReadFileBytes(filename string) ([]byte, error) {

	if !FileExists(filename) {
		return nil, os.ErrNotExist
	}

	fd, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	defer fd.Close()
	context, err := io.ReadAll(fd)
	if err != nil {
		return nil, err
	}

	return context, nil
}

// Read file content (as string)
func ReadFile(filename string) (string, error) {
	conntext, err := ReadFileBytes(filename)
	return string(conntext), err
}

// Write file
func WriteFile(filename string, context string) (bool, error) {
	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return false, err
	}
	defer fd.Close()

	fd.WriteString(context)
	return true, nil
}

// Write file bytes
func WriteFileBytes(filename string, context []byte) (bool, error) {
	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return false, err
	}
	defer fd.Close()
	fd.Write(context)
	return true, nil
}

// Append content to a file
func FileAppend(filename string, content string) (err error) {
	var fp *os.File

	fp, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	if err != nil {
		return
	}

	defer fp.Close()

	_, err = fp.WriteString(content)

	return
}

// Get language
func GetLanguage() string {
	defaultLanguage := "en"
	r := GetReq()
	if r == nil {
		return defaultLanguage
	}

	return GetLanguageFromCtx(r.Context())
}

// Get language list
func GetLanguageList() []map[string]interface{} {
	filename := ROOT_PATH + "/languages/settings.json"

	langCache := GetCache(filename)
	if langCache != nil {
		return langCache.([]map[string]interface{})
	}

	default_languages := []g.Map{
		{
			"name":   "zh",
			"google": "zh-cn",
			"title":  "Simplified Chinese",
			"cn":     "Simplified Chinese",
		},
		{
			"name":   "en",
			"google": "en",
			"title":  "English",
			"cn":     "English",
		},
	}

	if !FileExists(filename) {
		return default_languages
	}

	// Read language pack
	lang_data, err := ReadFile(filename)
	if err != nil {
		return default_languages
	}

	// Convert language pack to map
	lang_map := make([]map[string]interface{}, 0)
	err = json.Unmarshal([]byte(lang_data), &lang_map)
	if err != nil {
		return default_languages
	}

	SetCache(filename, lang_map, 60)
	return lang_map
}

// Get language from context
func GetLanguageFromCtx(ctx context.Context) string {
	defaultLanguage := "en"

	r := g.RequestFromCtx(ctx)
	if r == nil {
		return defaultLanguage
	}

	// Get language from lang parameter
	lang := r.Get("lang").String()
	if lang != "" {
		lang = strings.ToLower(lang)
		defaultLanguage = strings.Split(lang, "-")[0]
		r.Session.Set("language", defaultLanguage)
	} else {
		language, err := r.Session.Get("language")
		if err == nil && !language.IsEmpty() {
			return language.String()
		}
		// Get from cache
		langCache := GetCache("language")
		if langCache != nil {
			defaultLanguage = langCache.(string)
		} else {
			// Get from request header
			lang := r.Header.Get("Accept-Language")
			if lang != "" {
				lang = strings.ToLower(lang)
				defaultLanguage = strings.Split(lang, ",")[0]
				defaultLanguage = strings.Split(defaultLanguage, "-")[0]

				if defaultLanguage == "ja" {
					defaultLanguage = "jp"
				}
			}
		}
	}

	return defaultLanguage
}

// Replace language pack format variables
func ReplaceLangVar(msg string, args []interface{}) string {

	if len(args) == 0 {
		return msg
	}

	for _, v := range args {
		msg = strings.Replace(msg, "{}", gconv.String(v), 1)
	}

	return msg
}

// Get language variable
func LangVar(now_lang string, msg string, args []interface{}) string {
	hash := Md5(msg)

	filename := ROOT_PATH + "/languages/" + now_lang + "/server.json"
	// First read from cache
	cacheLang := GetCache(filename)
	lang_map := make(map[string]string)

	if cacheLang == nil {
		// Read from file
		if now_lang == "" || !FileExists(filename) {
			fmt.Println("File not exists: ", filename)
			return ReplaceLangVar(msg, args)
		}

		// Read language pack
		lang_data, err := ReadFile(filename)
		if err != nil {
			fmt.Println("Read File Error: ", filename)
			return ReplaceLangVar(msg, args)
		}

		// Convert language pack to map
		err = json.Unmarshal([]byte(lang_data), &lang_map)
		if err != nil {
			fmt.Println("Parse Error: ", err)
			return ReplaceLangVar(msg, args)
		}
	} else {
		lang_map = cacheLang.(map[string]string)
	}

	// Check if it exists in the language pack
	if _, ok := lang_map[hash]; ok {
		// Get content from language pack
		msg = lang_map[hash]
		// fmt.Println(now_lang, hash, msg)
	}

	return ReplaceLangVar(msg, args)
}

// Get language
func Lang(msg string, args ...interface{}) string {
	now_lang := GetLanguage()
	return LangVar(now_lang, msg, args)
}

// Get language -- with context
func LangCtx(ctx context.Context, msg string, args ...interface{}) string {
	now_lang := GetLanguageFromCtx(ctx)
	return LangVar(now_lang, msg, args)
}

// Initialize cache
var Cache = gcache.New()

// Initialize cache context
var Ctx = gctx.New()

// Read/write cache
func S(key interface{}, value ...interface{}) interface{} {
	vlen := len(value)
	expire := time.Duration(0)
	is_exists, _ := Cache.Contains(Ctx, key)
	if vlen == 0 {
		// Return nil if cache does not exist
		if !is_exists {
			return nil
		}
		// Read cache
		value, err := Cache.Get(Ctx, key)
		if err != nil {
			return nil
		}
		return value.Interface()
	} else if vlen == 2 {
		// Convert expiration time
		expire = time.Duration(value[1].(int)) * time.Second
	}

	// Delete cache
	if value[0] == nil && is_exists {
		Cache.Remove(Ctx, key)
	}

	if is_exists {
		// Update cache
		Cache.Update(Ctx, key, value[0])
	}

	// Set cache
	err := Cache.Set(Ctx, key, value[0], expire)

	return err == nil
}

// Get cache
func GetCache(key interface{}) interface{} {
	return S(key)
}

// Set cache
func SetCache(key interface{}, value interface{}, expire int) bool {
	return S(key, value, expire).(bool)
}

// Update cache
func UpdateCache(key interface{}, value interface{}) bool {
	is_exists, _ := Cache.Contains(Ctx, key)
	if !is_exists {
		return false
	}

	Cache.Update(Ctx, key, value)
	return true
}

// Update cache expiration time
func UpdateCacheExpire(key interface{}, expire int) bool {
	is_exists, _ := Cache.Contains(Ctx, key)
	if !is_exists {
		return false
	}

	Cache.UpdateExpire(Ctx, key, time.Duration(expire))
	return true
}

// Delete cache
func RemoveCache(key interface{}) bool {
	is_exists, _ := Cache.Contains(Ctx, key)
	if !is_exists {
		return false
	}
	Cache.Remove(Ctx, key)
	return true
}

// Increment cache value by 1
func IncrCache(key interface{}) int {
	v := GetCache(key)
	if v == nil {
		SetCache(key, 1, 0)
		return 1
	}
	num := v.(int) + 1
	Cache.Update(Ctx, key, num)
	return v.(int) + 1
}

// Decrement cache value by 1
func DecrCache(key interface{}) int {
	v := GetCache(key)
	if v == nil {
		SetCache(key, 1, 0)
		return 1
	}
	num := v.(int) - 1
	Cache.Update(Ctx, key, num)
	return num
}

// Convert struct to map
func StructToMap(obj any) (data map[string]any) {
	// Convert struct to map using reflection
	data = make(map[string]any)
	objV := reflect.Indirect(reflect.ValueOf(obj))
	objT := objV.Type()

	if objT.Kind() == reflect.Map {
		return objV.Interface().(map[string]any)
	}

	numField := objT.NumField()
	for i := 0; i < numField; i++ {
		field, ok := objT.Field(i).Tag.Lookup("json")
		if ok {
			data[field] = objV.Field(i).Interface()
		} else {
			data[objT.Field(i).Name] = objV.Field(i).Interface()
		}
	}

	return data
}

// Generate random string
func RandomStr(n int) string {

	// Set random seed
	r := mRand.New(mRand.NewSource(time.Now().UnixNano()))
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	num := len(letterBytes)
	for i := range b {
		b[i] = letterBytes[r.Intn(num)]
	}
	return string(b)
}

// IsComplexPassword Check password complexity
func IsComplexPassword(password string, score int) bool {
	password = strings.TrimSpace(password)

	// Password length must be at least 8 characters
	if len(password) < 7 {
		return false
	}

	//// Contains uppercase letters
	//if regexp.MustCompile(`[A-Z]+`).MatchString(password) {
	//	score++
	//}
	//
	//// Contains lowercase letters
	//if regexp.MustCompile(`[a-z]+`).MatchString(password) {
	//	score++
	//}
	//
	//// Contains numbers
	//if regexp.MustCompile(`[0-9]+`).MatchString(password) {
	//	score++
	//}
	//
	//// Contains special characters
	//if regexp.MustCompile(`[^A-Za-z0-9]+`).MatchString(password) {
	//	score++
	//}

	return score > 2
}

// Paginate data
func PageData(data interface{}, pageNumber, pageSize int) map[string]interface{} {
	lst := make([]interface{}, 0, 256)

	dataV := reflect.Indirect(reflect.ValueOf(data))
	switch dataV.Type().Kind() {
	case reflect.Slice, reflect.Array:
		size := dataV.Len()
		for i := 0; i < size; i++ {
			lst = append(lst, dataV.Index(i).Interface())
		}
	default:
		panic("unsupported pagination type " + dataV.Type().String())
	}

	startIndex := (pageNumber - 1) * pageSize
	if startIndex >= len(lst) {
		return map[string]interface{}{
			"list": []interface{}{},
			"page": map[string]interface{}{
				"count":      len(lst),              // Total rows
				"rows":       pageSize,              // Rows per page
				"p":          pageNumber,            // Current page
				"start_line": startIndex + 1,        // Start line
				"end_line":   startIndex + pageSize, // End line
			},
		}
	}

	endIndex := startIndex + pageSize
	if endIndex > len(lst) {
		endIndex = len(lst)
	}

	return map[string]interface{}{
		"list": lst[startIndex:endIndex],
		"page": map[string]interface{}{
			"count":      len(lst),       // Total rows
			"rows":       pageSize,       // Rows per page
			"p":          pageNumber,     // Current page
			"start_line": startIndex + 1, // Start line
			"end_line":   endIndex,       // End line
		},
	}
}

// Paginate
func Page(count int, pageNumber int, pageSize int) map[string]interface{} {

	startIndex := (pageNumber - 1) * pageSize

	// Calculate total pages
	page_num := count / pageSize
	if count%pageSize > 0 {
		page_num++
	}

	if startIndex >= count {
		return map[string]interface{}{
			"count":      count,                 // Total rows
			"rows":       pageSize,              // Rows per page
			"p":          pageNumber,            // Current page
			"start_line": startIndex + 1,        // Start line
			"end_line":   startIndex + pageSize, // End line
			"page_num":   page_num,              // Total pages
		}
	}

	endIndex := startIndex + pageSize
	if endIndex > count {
		endIndex = count
	}

	return map[string]interface{}{
		"count":      count,          // Total rows
		"rows":       pageSize,       // Rows per page
		"p":          pageNumber,     // Current page
		"start_line": startIndex + 1, // Start line
		"end_line":   endIndex,       // End line
		"page_num":   page_num,       // Total pages
	}
}

// Write operation log
func WriteOperationLog(ctx context.Context, logType, logLevel, logBody, remoteAddress string, logStatus, accountId int) bool {
	// Log information
	logsInfo := g.Map{
		"log_type":    logType,       // Log type
		"log_level":   logLevel,      // Log level
		"log_body":    logBody,       // Log content
		"remote_addr": remoteAddress, // Operation address
		"log_status":  logStatus,     // Operation status
		"account_id":  accountId,     // User ID
	}
	// Insert data
	_, err := MR("log", "logs").Insert(logsInfo)
	if err != nil {
		g.Log().Error(ctx, "Failed to insert data into database:", err)
		return false
	}
	return true
}

// Get client IP address
func GetClientIp(ctx context.Context) string {
	r := ghttp.RequestFromCtx(ctx)
	// Get client IP address
	return GetClientIpFromRequest(r)
}

// Get client IP address from request
func GetClientIpFromRequest(r *ghttp.Request) string {
	// Get client IP address
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}

// Get SESSION value
func GetSessionInterface(ctx context.Context, key string, default_value ...interface{}) interface{} {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		if len(default_value) > 0 {
			return default_value[0]
		}
		return nil
	}
	if r.Session == nil {
		if len(default_value) > 0 {
			return default_value[0]
		}
		return nil
	}
	sessionValue, err := r.Session.Get(key)
	if err != nil {
		// Return default value if it exists
		if len(default_value) > 0 {
			return default_value[0]
		}
		return nil
	}
	value := sessionValue.Val()
	if sessionValue.IsEmpty() {
		if len(default_value) > 0 {
			return default_value[0]
		}
	}
	return value
}

// Get SESSION value
func GetSession(ctx context.Context, key string, default_value ...interface{}) interface{} {
	return GetSessionInterface(ctx, key, default_value...)
}

// Get SESSION value (string)
func GetSessionString(ctx context.Context, key string, default_value ...string) string {
	sessValue := GetSession(ctx, key, nil)
	if sessValue == nil {
		if len(default_value) == 0 {
			return ""
		}
		return default_value[0]
	}
	return sessValue.(string)
}

// Get SESSION value (int)
func GetSessionInt(ctx context.Context, key string, default_value ...int) int {
	sessValue := GetSession(ctx, key, nil)
	if sessValue == nil {
		if len(default_value) == 0 {
			return 0
		}
		return default_value[0]
	}
	return sessValue.(int)
}

// Get SESSION value (int64)
func GetSessionInt64(ctx context.Context, key string, default_value ...int64) int64 {
	sessValue := GetSession(ctx, key, nil)
	if sessValue == nil {
		if len(default_value) == 0 {
			return 0
		}
		return default_value[0]
	}
	return sessValue.(int64)
}

// Get SESSION value (bool)
func GetSessionBool(ctx context.Context, key string, default_value ...bool) bool {
	sessValue := GetSession(ctx, key, nil)
	if sessValue == nil {
		if len(default_value) == 0 {
			return false
		}
		return default_value[0]
	}
	return sessValue.(bool)
}

// Set SESSION value
func SetSession(ctx context.Context, key string, value interface{}) bool {
	err := ghttp.RequestFromCtx(ctx).Session.Set(key, value)
	return err == nil
}

// Get request URL
func GetRequestUrl(ctx context.Context, urlPath string) string {
	r := ghttp.RequestFromCtx(ctx)
	protocol := GetRequestProtocol(ctx)
	host := r.Request.Host
	// Construct full URL
	if urlPath != "" {
		host = host + "/" + urlPath
	} else {
		host = host + "/"
	}
	host = strings.Replace(host, "//", "/", -1)
	fullURL := protocol + "://" + host
	return fullURL
}

// Get IP address from request
func GetRequestIp(ctx context.Context) string {
	r := ghttp.RequestFromCtx(ctx)
	clientIP := r.Header.Get("X-Real-IP")
	return clientIP
}

// Get request protocol
func GetRequestProtocol(ctx context.Context) string {
	r := ghttp.RequestFromCtx(ctx)
	// Get protocol
	protocol := r.Request.URL.Scheme
	if protocol == "" {
		protocol = "http"
	}
	return protocol
}

// Restart web server
func RestartWebServer() {
	s := g.Server(consts.DEFAULT_SERVER_NAME)
	if s != nil {
		s.Shutdown()
		s.Start()
	}
}

// Retrieve mountpoint by user
func RetrieveMountpointByUser(username string) (mountpoint string, err error) {
	// Get user home directory
	userInfo, err := user.Lookup(username)
	if err != nil {
		return
	}

	// Get home directory
	homepath := userInfo.HomeDir
	if !strings.Contains(homepath, "/home/") {
		err = errors.New("specified user not valid")
		return
	}

	// Use /home/ to split home directory, take the first part as mountpoint
	mountpoint = strings.Split(homepath, "/home/")[0]

	if mountpoint == "" {
		mountpoint = "/"
	}

	if !FileExists(mountpoint) {
		err = errors.New("mountpoint not exists")
		return
	}

	return
}

// Get current username
func GetUserName(ctx context.Context) string {
	val := GetSession(ctx, "username", "www")
	if val == nil {
		return "www"
	}
	return val.(string)
}

// Get UID and GID of a specified user
func GetUidAndGid(username string) (int, int) {
	userInfo, err := user.Lookup(username)
	if err != nil {
		fmt.Println(err.Error())
		return 0, 0
	}

	return gconv.Int(userInfo.Uid), gconv.Int(userInfo.Gid)
}

// Change group of a file or directory (recursive)
func Chgrp(path, groupName string) error {
	grp, err := user.LookupGroup(groupName)

	if err != nil {
		return err
	}

	gid := gconv.Int(grp.Gid)

	if gid < 1 {
		return errors.New("group not exists: " + groupName)
	}

	return ChgrpWithGidRecursive(path, gid)
}

// Change group of a file or directory with GID (recursive)
func ChgrpWithGidRecursive(path string, gid int) error {
	// File does not exist
	if !FileExists(path) {
		return errors.New("file not exists: " + path)
	}

	// Directory
	if IsDir(path) {
		ds, err := os.ReadDir(path)

		if err != nil {
			return err
		}

		for _, d := range ds {
			err = ChgrpWithGidRecursive(path+"/"+d.Name(), gid)

			if err != nil {
				return err
			}
		}
	}

	return os.Chown(path, -1, gid)
}

// Change group of a file or directory (non-recursive)
func ChgrpOnly(path, groupName string) error {
	grp, err := user.LookupGroup(groupName)

	if err != nil {
		return err
	}

	gid := gconv.Int(grp.Gid)

	if gid < 1 {
		return errors.New("group not exists: " + groupName)
	}

	return os.Chown(path, -1, gid)
}

// Change owner of a file or directory (recursive)
func Chown(path, username string) error {
	uid, gid := GetUidAndGid(username)

	if uid == 0 || gid == 0 {
		return errors.New("user not exists: " + username)
	}

	return ChownWithUidAndGidRecursive(path, uid, gid)
}

// Change owner of a file or directory with UID and GID (recursive)
func ChownWithUidAndGidRecursive(path string, uid, gid int) error {
	// File does not exist
	if !FileExists(path) {
		return errors.New("file not exists: " + path)
	}

	// Directory
	if IsDir(path) {
		ds, err := os.ReadDir(path)

		if err != nil {
			return err
		}

		for _, d := range ds {
			err = ChownWithUidAndGidRecursive(path+"/"+d.Name(), uid, gid)

			if err != nil {
				return err
			}
		}
	}

	return os.Chown(path, uid, gid)
}

// Change owner of a file or directory (non-recursive)
func ChownOnly(path, username string) error {
	uid, gid := GetUidAndGid(username)

	if uid == 0 || gid == 0 {
		return errors.New("user not exists: " + username)
	}

	return os.Chown(path, uid, gid)
}

// Change file or directory permissions (recursive)
func Chmod(path string, mode os.FileMode) error {
	return ChmodRecursive(path, mode)
}

// Change file or directory permissions (recursive)
func ChmodRecursive(path string, mode os.FileMode) error {
	// File does not exist
	if !FileExists(path) {
		return errors.New("file not exists: " + path)
	}

	// Directory
	if IsDir(path) {
		ds, err := os.ReadDir(path)

		if err != nil {
			return err
		}

		for _, d := range ds {
			err = ChmodRecursive(path+"/"+d.Name(), mode)

			if err != nil {
				return err
			}
		}
	}

	return os.Chmod(path, mode)
}

// Change file or directory attributes
func Chattr(path string, attr string) error {
	return ChattrRecursive(path, attr)
}

// Change file or directory attributes (recursive)
func ChattrRecursive(path string, attr string) error {
	// Check if attr is valid
	if len(attr) != 2 || (attr[0] != '+' && attr[0] != '-') {
		return errors.New("invalid attr: " + attr)
	}

	if attr[1] != 'i' && attr[1] != 'a' && attr[1] != 'd' && attr[1] != 's' && attr[1] != 'c' {
		return errors.New("invalid attr: " + attr)
	}

	// File does not exist
	if !FileExists(path) {
		return errors.New("file not exists: " + path)
	}

	// Directory
	if IsDir(path) {
		ds, err := os.ReadDir(path)

		if err != nil {
			return err
		}

		for _, d := range ds {
			err = ChattrRecursive(path+"/"+d.Name(), attr)

			if err != nil {
				return err
			}
		}
	}

	// Change file attributes
	fp, err := os.Open(path)

	if err != nil {
		return err
	}

	defer fp.Close()

	var flag int32

	switch attr[1] {
	case 'i':
		flag = chattr.FS_IMMUTABLE_FL
	case 'a':
		flag = chattr.FS_APPEND_FL
	case 'd':
		flag = chattr.FS_NODUMP_FL
	case 's':
		flag = chattr.FS_SYNC_FL
	case 'c':
		flag = chattr.FS_COMPR_FL
	}

	if attr[0] == '+' {
		return chattr.SetAttr(fp, flag)
	}

	return chattr.UnsetAttr(fp, flag)
}

// Create directory (for user)
func MkDirAll(dir, username string) (err error) {
	uid, gid := GetUidAndGid(username)

	if uid == 0 || gid == 0 {
		return errors.New("user not exists: " + username)
	}

	if !FileExists(dir) {
		if err = os.MkdirAll(dir, 0600); err != nil {
			return
		}
		err = os.Chown(dir, uid, gid)
	}
	return
}

// Get current user ID
func GetAccountId(ctx context.Context) int {
	return gconv.Int(GetSession(ctx, "account_id", 0))
}

// Get user ID by username
func GetAccountIdByUsername(username string) int {
	dbInfo, err := M("account").Where("username=?", username).Fields("account_id").One()
	if err != nil {
		return 0
	}
	return dbInfo["account_id"].Int()
}

// Get current request object -- Note: This method has low performance, use sparingly
func GetReq() *ghttp.Request {
	routineId := gdebug.GoroutineId()
	ctxVar, err := Cache.Get(Ctx, routineId)
	if err != nil {
		return nil
	}
	if ctxVar.IsEmpty() {
		return nil
	}

	return ctxVar.Interface().(*ghttp.Request)
}

// Get user ID by context
func GetAccountIdByCtx(ctx context.Context) int {
	username := GetUserName(ctx)
	// Get user ID
	accountInfo, err := M("account").Where("username=?", username).Fields("account_id").One()
	if err != nil && accountInfo == nil {
		return 0
	}
	accountId := accountInfo["account_id"].Int()
	return accountId
}

// Compare version numbers
func VersionCompare(version1, version2, opt string) bool {
	// Check operator
	if opt != ">" && opt != ">=" && opt != "<" && opt != "<=" && opt != "==" && opt != "=" {
		panic("Invalid operator: " + opt)
	}

	v1 := strings.Split(version1, ".")
	v2 := strings.Split(version2, ".")
	if len(v1) != len(v2) {
		// Pad
		if len(v1) > len(v2) {
			for i := 0; i < len(v1)-len(v2); i++ {
				v2 = append(v2, "0")
			}
		} else {
			for i := 0; i < len(v2)-len(v1); i++ {
				v1 = append(v1, "0")
			}
		}
	}

	length := len(v1)

	for i := 0; i < length; i++ {
		// Convert to numbers for comparison
		v1i, err := strconv.Atoi(v1[i])
		if err != nil {
			v1i = 0
		}

		v2i, err := strconv.Atoi(v2[i])
		if err != nil {
			v2i = 0
		}

		if v1i == v2i {
			continue
		}
		if v1i > v2i && (opt == ">" || opt == ">=") {
			return true
		}
		if v1i < v2i && (opt == "<" || opt == "<=") {
			return true
		}
	}

	if opt == "=" || opt == "==" || opt == ">=" || opt == "<=" {
		return true
	}

	return false
}

// Get configuration item
func GetConfig(key string) interface{} {
	configFile := "config/settings.json"
	defaultConfig := `{
        "defaultMountpoint":"",
        "defaultLanguage":"en"
    }`
	if !FileExists(configFile) {
		// Write default configuration
		_, err := WriteFile(configFile, defaultConfig)
		if err != nil {
			return nil
		}
	}

	// Read configuration file
	configData, err := ReadFile(configFile)
	if err != nil {
		return nil
	}

	// Convert to map
	configMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(configData), &configMap)
	if err != nil {
		// If parsing fails, rewrite configuration file
		_, err := WriteFile(configFile, defaultConfig)
		if err != nil {
			return nil
		}
		// Parse default configuration
		json.Unmarshal([]byte(configData), &configMap)
	}

	// Get configuration item
	if value, ok := configMap[key]; ok {
		return value
	}

	return nil
}

// Set configuration item
func SetConfig(key string, value interface{}) bool {
	configFile := "config/settings.json"
	if !FileExists(configFile) {
		return false
	}

	// Read configuration file
	configData, err := ReadFile(configFile)
	if err != nil {
		return false
	}

	// Convert to map
	configMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(configData), &configMap)
	if err != nil {
		return false
	}

	// Set configuration item
	configMap[key] = value

	// Convert to JSON
	configJson, err := json.Marshal(configMap)
	if err != nil {
		return false
	}

	// Write configuration file
	_, err = WriteFile(configFile, string(configJson))

	return err == nil
}

// Substring
func Substring(s string, start, length int) string {
	rs := []rune(s)
	rl := len(rs)

	if start < 0 {
		start = 0
	}

	if start >= rl {
		return ""
	}

	end := start + length

	if end > rl {
		end = rl
	}

	return string(rs[start:end])
}

// Read last n lines of a file
func Tail(filename string, n int) (string, error) {
	result := make([]string, 0)

	err := ReadEachReverse(filename, func(row string, cnt int) bool {
		result = append([]string{row}, result...)
		return cnt < n-1
	})

	return strings.Join(result, "\n"), err
}

// Read file in reverse line by line
func ReadEachReverse(filename string, handler func(row string, cnt int) bool) error {
	fp, err := os.Open(filename)

	if err != nil {
		return err
	}

	defer fp.Close()

	endpos, err := fp.Seek(0, 2)

	if err != nil {
		return err
	}

	var offset int64
	lastOffset := endpos
	chunkSize := int64(4096)
	buf := make([]byte, chunkSize)
	last := ""

	loops := endpos / chunkSize
	i := int64(0)
	cnt := 0

	for ; i < loops; i++ {
		offset, err = fp.Seek(endpos+(chunkSize+chunkSize*i)*-1, 0)

		if err != nil {
			break
		}

		e := lastOffset - offset

		if _, err = fp.Read(buf); err != nil {
			break
		}

		lines := strings.Split(string(buf[:e]), "\n")
		j := len(lines) - 1

		if last != "" || buf[e-1] != '\n' {
			lines[j] = lines[j] + last
		}

		last = lines[0]

		for ; j > 0; j-- {
			cnt++
			if !handler(lines[j], cnt) {
				return nil
			}
		}

		lastOffset = offset
	}

	if i < loops {
		return nil
	}

	remainder := endpos % chunkSize

	if remainder == 0 {
		return nil
	}

	if _, err = fp.Seek(0, 0); err != nil {
		return err
	}

	if _, err = fp.Read(buf); err != nil {
		return err
	}

	lines := strings.Split(string(buf[:remainder]), "\n")
	j := len(lines) - 1

	if last != "" {
		lines[j] = lines[j] + last
		last = ""
	}

	for ; j > -1; j-- {
		cnt++
		if !handler(lines[j], cnt) {
			return nil
		}
	}

	return nil
}

// Read first n lines of a file
func Head(filename string, n int) (string, error) {
	result := make([]string, 0)

	err := ReadEach(filename, func(row string, cnt int) bool {
		result = append(result, row)
		return cnt < n-1
	})

	return strings.Join(result, "\n"), err
}

// Read file line by line
func ReadEach(filename string, handler func(row string, cnt int) bool) error {
	fp, err := os.Open(filename)

	if err != nil {
		return err
	}

	defer fp.Close()

	// Create an IO stream reader
	reader := bufio.NewReader(fp)
	cnt := 0

	// Read file line by line
	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				handler(line, cnt)
			}
			break
		}

		if !handler(line, cnt) {
			break
		}

		// Increment line number
		cnt++
	}

	return nil
}

// Download file write counter
type downloadFileWriteCounter struct {
	Total          uint64
	ContentLength  uint64
	startTimeMilli int64
	handler        func(contentLength, total, speed uint64)
}

// Implement io.Writer interface
func (w *downloadFileWriteCounter) Write(p []byte) (n int, err error) {
	n = len(p)
	w.Total += uint64(n)

	if w.handler != nil {
		w.handler(w.ContentLength, w.Total, uint64(float64(w.Total)/float64(time.Now().UnixMilli()-w.startTimeMilli)*1000))
	}

	return
}

// Download file
func DownloadFile(ctx context.Context, url, dst string, timeout time.Duration, headers map[string]string, watcher func(contentLength, downloaded, speed uint64)) (err error) {
	// Create an HTTP client
	client := GetHttpClient(int(timeout.Seconds()))

	req, err := http.NewRequest("GET", url, nil)

	// Set request headers
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	// Send request
	var resp *http.Response

	if resp, err = client.Do(req.WithContext(ctx)); err != nil {
		return
	}

	defer resp.Body.Close()

	// Download failed
	if resp.StatusCode < http.StatusOK && resp.StatusCode >= http.StatusMultipleChoices {
		return errors.New("download file failed: " + resp.Status)
	}

	// Create directory
	dir := filepath.Dir(dst)
	if !FileExists(dir) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return
		}
	}

	// Create temporary file
	tmpFile := dst + ".tmp"

	// Open file
	var fp *os.File
	if fp, err = os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
		return
	}

	defer fp.Close()

	// Use buffered writer
	bufWriter := bufio.NewWriter(fp)

	// Write downloaded data and monitor progress
	if _, err = io.Copy(bufWriter, io.TeeReader(resp.Body, &downloadFileWriteCounter{
		ContentLength:  uint64(resp.ContentLength),
		startTimeMilli: time.Now().UnixMilli(),
		handler:        watcher,
	})); err != nil {
		return
	}

	// Flush buffer
	if err = bufWriter.Flush(); err != nil {
		return
	}

	// Rename file
	if err = os.Rename(tmpFile, dst); err != nil {
		return
	}

	return
}

// Calculate program execution time
func TimeCost() func() {
	start := time.Now()

	return func() {
		tc := time.Since(start)
		fmt.Printf("Time cost = %v\n", tc)
	}
}

// Get server port
func GetServerPort(ctx context.Context) string {
	r := ghttp.RequestFromCtx(ctx)
	hostArr := strings.Split(r.Host, ":")

	port := "80"
	if r.Proto == "https" {
		port = "443"
	}
	if len(hostArr) == 2 {
		port = hostArr[1]
	}
	defaultYaml := filepath.Join(GetCwdPath(), "manifest", "config", "default.yaml")
	// Get port from default.yaml server.httpsAddr
	if FileExists(defaultYaml) {
		content, err := ReadFile(defaultYaml)
		if err == nil {
			// Parse content into Config
			var config DefaultYamlConfig
			err = yaml.Unmarshal([]byte(content), &config)
			if err == nil {
				if config.Server.HttpsAddr != "" {
					port = strings.Split(config.Server.HttpsAddr, ":")[1]
				} else {
					if config.Server.Address != "" {
						port = strings.Split(config.Server.Address, ":")[1]
					}
				}
			}
		}
	}
	return port
}

// Check if a port is in use
func PortInUse(port int) bool {
	// Try to listen on the port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", strconv.Itoa(port)))

	if err != nil {
		// Port is in use
		return true

	}
	// Close listener
	defer listener.Close()

	return false
}

// Allow port
func AllowPort(port int) (err error) {
	// Check if port number is valid
	if !IsPort(strconv.Itoa(port)) {
		return errors.New("invalid port: " + strconv.Itoa(port))
	}

	defer func() {
		if err == nil {
			// Reload firewall
			_ = ReloadFirewall()
		}
	}()

	// Command line output
	var s string

	// Check if it is Ubuntu
	if FileExists("/usr/sbin/ufw") || FileExists("/usr/bin/ufw") {
		s, err = ExecShell(fmt.Sprintf("ufw allow %d/tcp", port))
		if err != nil {
			glog.Error(context.Background(), "AllowPort error: ", err, " ", s)
		}

		return
	}

	// Check if it is CentOS
	if FileExists("/usr/sbin/firewalld") || FileExists("/usr/bin/firewalld") {
		s, err = ExecShell(fmt.Sprintf("firewall-cmd --permanent --zone=public --add-port=%d/tcp", port))
		if err != nil {
			glog.Error(context.Background(), "AllowPort error: ", err, " ", s)
			return
		}
	}

	// Use iptables
	s, err = ExecShell(fmt.Sprintf("iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport %d -j ACCEPT", port))
	if err != nil {
		glog.Error(context.Background(), "AllowPort error: ", err, " ", s)
		return
	}

	return
}

// Deny port
func DeniedPort(port int) (err error) {
	// Check if port number is valid
	if !IsPort(strconv.Itoa(port)) {
		return errors.New("invalid port: " + strconv.Itoa(port))
	}

	defer func() {
		if err == nil {
			// Reload firewall
			_ = ReloadFirewall()
		}
	}()

	// Command line output
	var s string

	// Check if it is Ubuntu
	if FileExists("/usr/sbin/ufw") || FileExists("/usr/bin/ufw") {
		s, err = ExecShell(fmt.Sprintf("ufw delete allow %d/tcp", port))
		if err != nil {
			glog.Error(context.Background(), "DeniedPort error: ", err, " ", s)
		}

		return
	}

	// Check if it is CentOS
	if FileExists("/usr/sbin/firewalld") || FileExists("/usr/bin/firewalld") {
		s, err = ExecShell(fmt.Sprintf("firewall-cmd --permanent --zone=public --remove-port=%d/tcp", port))
		if err != nil {
			glog.Error(context.Background(), "DeniedPort error: ", err, " ", s)
			return
		}
	}

	// Use iptables
	s, err = ExecShell(fmt.Sprintf("iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport %d -j ACCEPT", port))
	if err != nil {
		glog.Error(context.Background(), "DeniedPort error: ", err, " ", s)
		return
	}

	return
}

// Reload firewall
func ReloadFirewall() (err error) {
	// Command line output
	var s string

	// Check if it is Ubuntu
	if FileExists("/usr/sbin/ufw") || FileExists("/usr/bin/ufw") {
		s, err = ExecShell("ufw reload")
		if err != nil {
			glog.Error(context.Background(), "ReloadFirewall error: ", err, " ", s)
		}

		return
	}

	// Check if it is CentOS
	if FileExists("/usr/sbin/firewalld") || FileExists("/usr/bin/firewalld") {
		s, err = ExecShell("firewall-cmd --reload")
		if err != nil {
			glog.Error(context.Background(), "ReloadFirewall error: ", err, " ", s)
			return
		}
	}

	// Use iptables
	s, err = ExecShell("/etc/init.d/iptables save && /etc/init.d/iptables restart")
	if err != nil {
		glog.Error(context.Background(), "ReloadFirewall error: ", err, " ", s)
		return
	}

	return
}

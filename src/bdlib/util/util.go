package util

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"math/rand"
	mrand "math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

var (
	// gRandTime _
	gRandTime *rand.Rand
)

func init() {
	gRandTime = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// F2S convert float64 to string
func F2S(f float64, prec int) string {
	return strconv.FormatFloat(f, 'f', prec, 64)
}

// RegexText 正则匹配, 返回string数组
func RegexText(partern string, response []byte) ([]string, bool) {
	re := regexp.MustCompile(partern)
	con := re.FindAllStringSubmatch(string(response), -1)
	if len(con) == 0 || len(con[0]) < 2 {
		return nil, false
	}
	return con[0][1:], true
}

// GenHmac 生成请求校验参数
func GenHmac(input url.Values, fixKey string) string {
	keys := make([]string, len(input))
	idx := 0
	for k := range input {
		keys[idx] = k
		idx++
	}
	sort.Sort(sort.StringSlice(keys))
	h := hmac.New(sha256.New, []byte(fixKey))
	rndnum := String(gRandTime.Int())

	var space string
	for idx, k := range keys {
		if idx%2 == 0 {
			space = "_"
		} else {
			space = "^"
		}
		if len(input[k]) > 0 {
			fmt.Fprintf(h, "%s%s", input[k][0], space)
		} else {
			fmt.Fprintf(h, "%s", space)
		}
	}
	fmt.Fprintf(h, "%s", rndnum)
	clientMac := fmt.Sprintf("%x", h.Sum(nil))[:16]
	return fmt.Sprintf("%s;%s", clientMac, rndnum)
}

// RedirectToPanicFile 将panic信息打到.panic
func RedirectToPanicFile() {
	var discard *os.File
	var err error
	fileName := fmt.Sprintf(".panic_%s", ToString(os.Getpid()))
	discard, err = os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		discard, err = os.OpenFile("/dev/null", os.O_RDWR, 0)
	}
	if err == nil {
		fd := discard.Fd()
		syscall.Dup2(int(fd), int(os.Stderr.Fd()))
	}
}

// ToString _
func ToString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

// MD5 hash
func MD5(any interface{}) string {
	switch val := any.(type) {
	case []byte:
		return fmt.Sprintf("%x", md5.Sum(val))
	case string:
		return fmt.Sprintf("%x", md5.Sum([]byte(val)))
	default:
		h := md5.New()
		fmt.Fprintf(h, "%v", val)
		return fmt.Sprintf("%x", h.Sum(nil))
	}
}

// Base64Encode encode
func Base64Encode(key interface{}) string {
	switch val := key.(type) {
	case string:
		return base64.StdEncoding.EncodeToString([]byte(val))
	case []byte:
		return base64.StdEncoding.EncodeToString(val)
	default:
		str := fmt.Sprintf("%v", key)
		return base64.StdEncoding.EncodeToString([]byte(str))
	}
}

// Base64Decode decode
func Base64Decode(key interface{}) (plain string, err error) {

	var plainByte []byte
	switch val := key.(type) {
	case string:
		plainByte, err = base64.StdEncoding.DecodeString(val)
		plain = string(plainByte)
	case []byte:
		plainByte, err = base64.StdEncoding.DecodeString(string(val))
		plain = string(plainByte)
	default:
		err = fmt.Errorf("unsupport type")
	}
	return
}

// String convert to string
func String(any interface{}) string {
	switch val := any.(type) {
	case int, uint, int64, uint64, uint32, int32, uint8, int8, int16, uint16:
		return fmt.Sprintf("%d", val)
	case string, []byte:
		return fmt.Sprintf("%s", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Int64 convert string to int64
func Int64(str string) int64 {
	num, _ := strconv.ParseInt(str, 10, 64)
	return num
}

// Float64 convert string to float64
func Float64(str string) float64 {
	num, _ := strconv.ParseFloat(str, 64)
	return num
}

// DeepCopyObj ...
func DeepCopyObj(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

// GetDistance _
func GetDistance(lat1, lat2, lng1, lng2 float64) (dist float64, err error) {
	rad := math.Pi / 180
	lat1 = lat1 * rad
	lat2 = lat2 * rad
	lng1 = lng1 * rad
	lng2 = lng2 * rad
	c := math.Sin(lat1)*math.Sin(lat2) + math.Cos(lat1)*math.Cos(lat2)*math.Cos(lng2-lng1)
	dist = 6367 * math.Acos(c)
	return
}

// GetImageConfig 获取图片信息
func GetImageConfig(url string) (config image.Config, err error) {
	var (
		resp *http.Response
	)
	if resp, err = http.Get(url); err != nil {
		return
	}
	defer resp.Body.Close()

	if config, _, err = image.DecodeConfig(resp.Body); err != nil {
		return
	}
	return
}

// ConvertGCJ02ToBD09 腾讯地图坐标转百度地图坐标
func ConvertGCJ02ToBD09(lng, lat float64) (newLng, newLat float64) {
	var (
		xPi, z, t float64
	)
	xPi, newLng, newLat = math.Pi*3000/180, lng, lat
	z = math.Sqrt(newLng*newLng+newLat*newLat) + 0.00002*math.Sin(newLat*xPi)
	t = math.Atan2(newLat, newLng) + 0.000003*math.Cos(newLng*xPi)
	newLng, newLat = z*math.Cos(t)+0.0065, z*math.Sin(t)+0.006

	return
}

//CoverUp 屏蔽信息用
func CoverUp(str string) string {
	if str == "" {
		return str
	}
	start := len(str) / 3
	if start <= 0 {
		return str
	}
	ucode := []byte(str)
	for i := range ucode[start:] {
		ucode[start+i] = '*'
		if i == 3 {
			break
		}
	}
	return string(ucode)
}

// ValidPNG _
func ValidPNG(r io.Reader) bool {
	if _, err := png.DecodeConfig(r); err != nil {
		return false
	}
	return true
}

// ValidJPEG _
func ValidJPEG(r io.Reader) bool {
	if _, err := jpeg.DecodeConfig(r); err != nil {
		return false
	}
	return true
}

// GenerateQrCode 生成二维码
func GenerateQrCode(srcQrCode string) (buffer bytes.Buffer, err error) {
	var (
		qrCode barcode.Barcode
	)
	srcQrCode = " " + srcQrCode
	if qrCode, err = qr.Encode(srcQrCode, qr.L, qr.Auto); err != nil {
		return
	}
	qrCode, err = barcode.Scale(qrCode, 500, 500)
	if err != nil {
		return
	}
	png.Encode(&buffer, qrCode)

	return
}

// WeekDayCN 返回星期
func WeekDayCN(t time.Time) string {
	weekStrDict := map[string]string{
		"Monday":    "星期一",
		"Tuesday":   "星期二",
		"Wednesday": "星期三",
		"Thursday":  "星期四",
		"Friday":    "星期五",
		"Saturday":  "星期六",
		"Sunday":    "星期日",
	}
	if rs, ok := weekStrDict[t.Weekday().String()]; ok {
		return rs
	}
	return ""
}

// Anything2Json json转换函数
func Anything2Json(in interface{}) (out []byte) {
	out, _ = json.Marshal(in)

	return
}

// String2Int string to int
func String2Int(in string) (out int) {
	out, _ = strconv.Atoi(in)

	return
}

// Int2String int to string
func Int2String(in int) (out string) {
	out = strconv.Itoa(in)
	return
}

// Base64 encoding
func Base64(str string) (result string) {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// DeBase64 decode base64
func DeBase64(str string) (result string, err error) {
	resultByte, err := base64.StdEncoding.DecodeString(str)
	result = string(resultByte)
	return
}

// Base64URLDecode decode base64 url decode
func Base64URLDecode(str string) (result string, err error) {
	resultByte, err := base64.StdEncoding.DecodeString(str)
	if result, err = url.QueryUnescape(string(resultByte)); err != nil {
		return
	}
	return
}

// RandomString 生成随机字符串
func RandomString(l int) string {
	var result bytes.Buffer
	var temp string
	for i := 0; i < l; {
		if string(RandInt(65, 90)) != temp {
			temp = string(RandInt(65, 90))
			result.WriteString(temp)
			i++
		}
	}
	return result.String()
}

// RandInt 生成随机整形
func RandInt(min int, max int) int {
	mrand.Seed(time.Now().UTC().UnixNano())
	return min + mrand.Intn(max-min)
}

// IP2Int 将IP string 转换成 int 型
func IP2Int(ip string) int64 {
	if len(ip) == 0 {
		return 0
	}
	bits := strings.Split(ip, ".")
	if len(bits) < 4 {
		return 0
	}

	b0 := String2Int(bits[0])
	b1 := String2Int(bits[1])
	b2 := String2Int(bits[2])
	b3 := String2Int(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}

// ToInt convert string to int64
func ToInt(str interface{}) int64 {
	return ToInt(ToString(str))
}

// ToFloat convert string to float64
func ToFloat(str string) float64 {
	num, _ := strconv.ParseFloat(str, 64)
	return num
}

// Int2IP 将 int 型 转换成 net.IP
func Int2IP(ip int64) net.IP {
	var bytes [4]byte
	bytes[0] = byte(ip & 0xFF)
	bytes[1] = byte((ip >> 8) & 0xFF)
	bytes[2] = byte((ip >> 16) & 0xFF)
	bytes[3] = byte((ip >> 24) & 0xFF)

	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}

// GetPartsOfIP 针对ip ABC类网关字符串截取
// ip D类ip
// parts 取A/B/C类ip 4-D 3-C 2-B 1-A
// mark 间隔符
func GetPartsOfIP(ip string, parts int, mark string) (partIP string) {
	if parts > 4 || parts < 1 {
		return
	}
	if parts == 4 {
		return ip
	}
	// C类
	for i := 3; i >= parts; i-- {
		ip = ip[:strings.LastIndex(ip, mark)]
	}
	return ip
}

// TypeOf _
func TypeOf(v interface{}) reflect.Type {
	return reflect.TypeOf(v)
}

// Substr _
func Substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0
	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length
	if start > end {
		start, end = end, start
	}
	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[start:end])
}

// ReplaceRune _
func ReplaceRune(str string, oldRune []rune, newRune []rune) (newStr string) {
	if len(oldRune) == 0 || str == "" {
		newStr = str
		return
	}
	reader := strings.NewReader(str)
	var err error
	var ch rune
	var index int
	var list, queue []rune
	for {
		ch, _, err = reader.ReadRune()
		if err == io.EOF {
			if index > 0 {
				for _, r := range queue {
					list = append(list, r)
				}
			}
			break
		} else if err != nil {
			return
		}
		if ch == oldRune[index] {
			index++
			if index < len(oldRune) {
				queue = append(queue, ch)
			} else {
				// match
				index = 0
				queue = []rune{}
				for _, r := range newRune {
					list = append(list, r)
				}
			}
		} else {
			if index > 0 {
				index = 0
				for _, r := range queue {
					list = append(list, r)
				}
				queue = []rune{}
			}
			list = append(list, ch)
		}
	}
	return string(list)
}

// Addslashes mysql 防注入 转义.
func Addslashes(in string) (out string) {
	r := strings.NewReplacer("'", "\\'", "\"", "\\\"", "\\", "\\\\", "NULL", "\\NULL")
	out = r.Replace(in)
	return
}

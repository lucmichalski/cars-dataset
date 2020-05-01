package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/base64"
	"io"
	"os"
	"reflect"
	"time"
	"math/rand"
	"path/filepath"
	"strings"
	"io/ioutil"
)

func init() {
    rand.Seed(time.Now().UnixNano()) // do it once during app initialization
}

func DecodeToFile(url, data string) (*os.File, string, error) {

	var segments []string
	if strings.Contains(strings.ToLower(url), "jpg") {
		segments = append(segments, ".jpg")
	} else if strings.Contains(strings.ToLower(url), "png") {
		segments = append(segments, ".png")
	} else {
		segments = strings.Split(url, "/")
	}

	fileName := GetMD5Hash(url) + "-" + segments[len(segments)-1]
	if strings.Contains(fileName, "?") {
        // clean up query string
        fileParts := strings.Split(fileName, "?")
        if len(fileParts) > 0 {
                fileName = fileParts[0]
        }
	}
	filePath := filepath.Join(os.TempDir(), fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return nil, "", err
	}

    reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))

    // defer file.Close()
    buf, err := ioutil.ReadAll(reader)
    if err != nil {
		return nil, "", err
    }
    _, err = file.Write(buf)

	checksum, err := GetMD5File(filePath)
    if err != nil {
		return nil, "", err
    }

    return file, checksum, nil

}

func Shuffle(slice interface{}) {
    rv := reflect.ValueOf(slice)
    swap := reflect.Swapper(slice)
    length := rv.Len()
    for i := length - 1; i > 0; i-- {
        j := rand.Intn(i + 1)
        swap(i, j)
    }
}

func EnsureDir(path string) error {
	d, err := os.Open(path)
	if err != nil {
		os.MkdirAll(path, os.FileMode(0755))
	} else {
		return err
	}
	d.Close()
	return nil
}

func GetMD5File(filePath string) (result string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return
	}

	result = hex.EncodeToString(hash.Sum(nil))
	return
}

func RemoveDuplicates(elements []string) []string {
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func Escape(sql string) string {
	dest := make([]byte, 0, 2*len(sql))
	var escape byte
	for i := 0; i < len(sql); i++ {
		c := sql[i]

		escape = 0

		switch c {
		case 0: /* Must be escaped for 'mysql' */
			escape = '0'
			break
		case '\n': /* Must be escaped for logs */
			escape = 'n'
			break
		case '\r':
			escape = 'r'
			break
		case '\\':
			escape = '\\'
			break
		case '\'':
			escape = '\''
			break
		case '"': /* Better safe than sorry */
			escape = '"'
			break
		case '\032': /* This gives problems on Win32 */
			escape = 'Z'
		}

		if escape != 0 {
			dest = append(dest, '\\', escape)
		} else {
			dest = append(dest, c)
		}
	}

	return string(dest)
}

func Stripslashes(str string) string {
	var dstRune []rune
	strRune := []rune(str)
	strLenth := len(strRune)
	for i := 0; i < strLenth; i++ {
		if strRune[i] == []rune{'\\'}[0] {
			i++
		}
		dstRune = append(dstRune, strRune[i])
	}
	return string(dstRune)
}

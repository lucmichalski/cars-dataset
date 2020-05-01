package main

import (
    "bufio"
    "encoding/base64"
    "fmt"
    "io/ioutil"
    "os"
    "log"
    "strings"
)


func main() {
    base64Str := encodeBase64("./sample.jpg")
    decodeBase64(base64Str)
}

func encodeBase64(fp string) string {
    // Open file on disk.
    f, err := os.Open(fp)
    checkErr(err)

    // Read entire JPG into byte slice.
    reader := bufio.NewReader(f)
    content, err := ioutil.ReadAll(reader)
    checkErr(err)

    // Encode as base64.
    encoded := base64.StdEncoding.EncodeToString(content)

    // Print encoded data to console.
    // ... The base64 image can be used as a data URI in a browser.
    // fmt.Println("ENCODED: " + encoded)
    return encoded
}

func decodeBase64(data string) {
    reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
    f, err := os.Create("./sample_decoded.jpg")
    checkErr(err)
    defer f.Close()
    buf, err := ioutil.ReadAll(reader)
    if err != nil {
        log.Fatal(err)
    }
    _, err = f.Write(buf)
    checkErr(err)
}

func checkErr(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

/*
    file, _, err := req.FormFile("image")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer file.Close()
    img, _, err := image.Decode(file)
    m := resize.Resize(135, 115, img, resize.Lanczos3)
    buf := new(bytes.Buffer)
    err = jpeg.Encode(buf, m, &jpeg.Options{35})
    imageBit := buf.Bytes()
    // Defining the new image size

    photoBase64 := b64.StdEncoding.EncodeToString([]byte(imageBit))
    fmt.Println("Photo Base64.............................:" + fotoBase64)
*/

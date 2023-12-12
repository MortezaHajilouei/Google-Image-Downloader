package utils

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func ToBase64ImageFromBytes(bytes []byte) string {
	mimeType := http.DetectContentType(bytes)

	var fullBase64Encoding string
	switch mimeType {
	case "image/jpeg":
		fullBase64Encoding += "data:image/jpeg;base64,"
	case "image/png":
		fullBase64Encoding += "data:image/png;base64,"
	}

	base64Encoding := base64.StdEncoding.EncodeToString(bytes)

	fullBase64Encoding += base64Encoding
	return fullBase64Encoding
}

func ToBytesFromBase64Image(in string) []byte {
	if strings.ContainsAny(in, ",") {
		base64Data := strings.Split(in, ",")[1]
		bytes, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			return nil
		}
		return bytes
	}
	return nil
}

func BytesToImage(b []byte) image.Image {
	reader := bytes.NewReader(b)
	img, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	return img
}

func ImageToBytes(img image.Image) []byte {
	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, nil)
	if err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}

func IsValidURL(in string) bool {
	u, err := url.Parse(in)
	return err == nil && u.Scheme != "" && u.Host != ""
}

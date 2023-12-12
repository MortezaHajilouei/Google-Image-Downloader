package downloader

import (
	"fmt"

	"database/sql"
	"ggd/utils"
	"image"
	"io"
	"net/http"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/disintegration/imaging"
	_ "github.com/lib/pq"
)

func GetImageURLs(query string) ([]string, []string, error) {
	url := fmt.Sprintf("https://www.google.com/search?q=%s&tbm=isch", query)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	var urls []string
	var images []string
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, ok := s.Attr("src")
		if ok {
			if !utils.IsValidURL(src) {
				images = append(images, src)
			} else {
				urls = append(urls, src)
			}
		}
	})

	return urls, images, nil
}

func DownloadImage(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func ResizeImage(img image.Image) (image.Image, error) {
	resized := imaging.Resize(img, 300, 300, imaging.Lanczos)
	return resized, nil
}

func StoreImage(img image.Image, db *sql.DB) error {
	data := utils.ImageToBytes(img)
	secureData := utils.ToBase64ImageFromBytes(data)
	_, err := db.Exec("INSERT INTO images (data) VALUES ($1)", secureData)
	if err != nil {
		return err
	}

	return nil
}

func ProcessImage(data []byte, db *sql.DB) error {
	img := utils.BytesToImage(data)
	resized, err := ResizeImage(img)
	if err != nil {
		return err
	}

	err = StoreImage(resized, db)
	if err != nil {
		return err
	}

	return nil
}

func ProcessImages(query string, max int, db *sql.DB) error {
	urls, b64Images, err := GetImageURLs(query)
	if err != nil {
		return err
	}

	downloaded := 0
	var wg sync.WaitGroup
	ch := make(chan error)

	// save base64-data src
	for _, b64 := range b64Images {
		if downloaded >= max {
			break
		}
		downloaded++
		wg.Add(1)

		data := utils.ToBytesFromBase64Image(b64)
		go func(data []byte) {
			defer wg.Done()
			ch <- ProcessImage(data, db)
		}(data)
	}

	// download image url (src)
	for _, url := range urls {
		if downloaded >= max {
			break
		}
		downloaded++
		wg.Add(1)

		go func(url string) {
			defer wg.Done()
			img, err := DownloadImage(url)
			if err != nil {
				ch <- err
				return
			}

			ch <- ProcessImage(img, db)
		}(url)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var errors []error
	for err := range ch {
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%d out of %d images failed to process", len(errors), max)
	}

	return nil
}

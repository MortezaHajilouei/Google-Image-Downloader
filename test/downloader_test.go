package test

import (
	"database/sql"
	"fmt"
	"ggd/downloader"
	"ggd/utils"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestGetImageURLs(t *testing.T) {
	query := "cat"
	urls, b64Images, err := downloader.GetImageURLs(query)
	if err != nil {
		t.Errorf("GetImageURLs(%q) returned an error: %v", query, err)
	}
	if len(urls) == 0 && len(b64Images) == 0 {
		t.Errorf("GetImageURLs(%q) returned no urls, want some urls", query)
	}
}

func TestDownloadImage(t *testing.T) {
	url := "https://www.google.com/images/branding/googlelogo/1x/googlelogo_color_272x92dp.png"
	img, err := downloader.DownloadImage(url)
	if err != nil {
		t.Errorf("DownloadImage(%q) returned an error: %v", url, err)
	}
	if img == nil {
		t.Errorf("DownloadImage(%q) returned a nil image, want an image", url)
	}
}

func TestResizeImage(t *testing.T) {
	url := "https://www.google.com/images/branding/googlelogo/1x/googlelogo_color_272x92dp.png"
	body, err := downloader.DownloadImage(url)
	if err != nil {
		t.Fatal(err)
	}
	img := utils.BytesToImage(body)
	resized, err := downloader.ResizeImage(img)
	if err != nil {
		t.Errorf("ResizeImage returned an error: %v", err)
	}
	if resized.Bounds().Dx() != 300 || resized.Bounds().Dy() != 300 {
		t.Errorf("ResizeImage returned an image with size %dx%d, want 300x300", resized.Bounds().Dx(), resized.Bounds().Dy())
	}
}

func TestStoreImage(t *testing.T) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dbInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM images")
	if err != nil {
		t.Fatal(err)
	}

	url := "https://www.google.com/images/branding/googlelogo/1x/googlelogo_color_272x92dp.png"
	body, err := downloader.DownloadImage(url)
	if err != nil {
		t.Fatal(err)
	}
	img := utils.BytesToImage(body)
	resized, err := downloader.ResizeImage(img)
	if err != nil {
		t.Fatal(err)
	}

	err = downloader.StoreImage(resized, db)
	if err != nil {
		t.Errorf("StoreImage returned an error: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM images").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("StoreImage inserted %d images, want 1", count)
	}
}

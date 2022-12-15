package main

import (
	"bytes"
	"fmt"
	"github.com/Kagami/go-avif"
	"github.com/valyala/bytebufferpool"
	tele "gopkg.in/telebot.v3"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"time"
)

func main() {
	pref := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle(tele.OnPhoto, func(c tele.Context) error {
		photo := c.Message().Photo
		bb := bytebufferpool.Get()
		log.Printf("Received image %v, on disk %v", photo.FileID, photo.OnDisk())
		tempFile, err := os.CreateTemp("", "photo_*.avif")
		if err != nil {
			return c.Send(fmt.Sprintf("Can't create temp file %v", err))
		}
		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				log.Fatalf("Error removing temp file %v", err)
			}
		}(tempFile.Name())

		err = b.Download(&photo.File, tempFile.Name())
		if err != nil {
			return c.Send(fmt.Sprintf("Can't download file %v", err))
		}
		log.Printf("Downloaded image %s", tempFile.Name())
		src, err := os.Open(tempFile.Name())
		if err != nil {
			return c.Send(fmt.Sprintf("Can't open source file: %v", err))
		}

		img, ext, err := image.Decode(src)
		if err != nil {
			return c.Send(fmt.Sprintf("Can't decode source file: %v", err))
		}
		log.Printf("Decoded image %s - %s", ext, tempFile.Name())

		err = avif.Encode(bb, img, nil)
		if err != nil {
			return c.Send(fmt.Sprintf("Can't encode avif file: %v", err))
		}

		f := &tele.Document{File: tele.FromReader(bytes.NewReader(bb.B)), FileName: fmt.Sprintf("%s.avif", photo.FileID)}
		defer bytebufferpool.Put(bb)
		return c.Send(f)
	})

	b.Start()
}

package util

import (
	"github.com/pkg/errors"
	"github.com/schollz/progressbar/v3"
	"io"
	"net/http"
	"os"
)

func DownloadFileWithProgress(url, dest string) error {
	req, _ := http.NewRequest("GET", url, nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	f, _ := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	defer f.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"downloading",
	)
	_, err := io.Copy(io.MultiWriter(f, bar), resp.Body)
	if err != nil {
		return errors.Wrapf(err, "could not download file from %v", url)
	}
	return nil
}

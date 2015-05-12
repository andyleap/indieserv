package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"net/url"
	"time"
)

func TimeToID(id time.Time) []byte {
	return []byte(id.In(time.UTC).Format(time.RFC3339Nano))
}

func TimeToSlug(id time.Time) string {
	buf := make([]byte, 16)
	n := binary.PutVarint(buf, id.Unix())
	n2 := binary.PutVarint(buf[n:], int64(id.Nanosecond()))
	return base64.URLEncoding.EncodeToString(buf[:n+n2])
}

func SlugToTime(slug string) time.Time {
	slugbytes, _ := base64.URLEncoding.DecodeString(slug)
	buf := bytes.NewBuffer(slugbytes)
	sec, _ := binary.ReadVarint(buf)
	nsec, _ := binary.ReadVarint(buf)
	return time.Unix(sec, nsec)
}

func UrlToPath(url *url.URL, err error) string {
	if err != nil {
		panic(err)
	}
	return url.Path
}

func UrlToAbsPath(url *url.URL, err error) string {
	if err != nil {
		panic(err)
	}
	return url.String()
}

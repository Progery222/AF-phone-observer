package domain

import "time"

type ScreenState struct {
	Serial      string
	Activity    string
	PackageName string
	XMLDump     string
}

type Screenshot struct {
	Serial     string
	ObjectKey  string
	StorageURL string
	LocalPath  string
	Bytes      []byte
	SizeBytes  int64
	Width      int32
	Height     int32
	TakenAt    time.Time
}

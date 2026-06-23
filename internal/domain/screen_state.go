package domain

type ScreenState struct {
	Serial      string
	Activity    string
	PackageName string
	XMLDump     string
}

type Screenshot struct {
	Serial   string
	ObjectKey string
	LocalPath string
	Bytes    []byte
	Width    int32
	Height   int32
}

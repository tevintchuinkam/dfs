package main

type ChunckHandle struct {
	Handle string
}
type ChunckServer struct {
	Port string
	ID   string
}
type Master struct {
	FileToHandlesMap  map[string][]ChunckHandle
	HandleToServerMap map[string]ChunckServer
}

type Orchastrator interface {
	CreateFile(name string, dir string, data []byte) error
	ReadFile(name string, dir string) ([]byte, error)
	DeleteFile(name string, dir string) error
	LS(dir string) error
	CreateDir(name string) error
	DeleteDir(name string) error
}

type Server interface {
}

func main() {

}

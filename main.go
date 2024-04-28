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
	CreateDir(dir string) error
	DeleteDir(dir string) error
}

type Server interface {
}

func NewMaster() Orchastrator {
	return &Master{
		FileToHandlesMap:  make(map[string][]ChunckHandle),
		HandleToServerMap: make(map[string]ChunckServer),
	}
}

func (m *Master) CreateFile(name, dir string, data []byte) error {
	return nil
}
func (m *Master) ReadFile(name, dir string) ([]byte, error) {
	return []byte{}, nil
}
func (m *Master) DeleteFile(name, dir string) error {
	return nil
}
func (m *Master) CreateDir(dir string) error {
	return nil
}
func (m *Master) DeleteDir(dir string) error {
	return nil
}
func (m *Master) LS(dir string) error {
	return nil
}

func main() {

}

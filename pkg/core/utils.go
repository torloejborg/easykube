package core

type IUtils interface {
	HasBinary(bin string) bool
	FileOrDirExists(path string) bool
	CopyResourceToConfigDir(src, dest string) error
	SaveFile(data string, dest string)
	SaveFileByte(data []byte, dest string)
	ReadPropertyFile(path string) (map[string]string, error)
	IntSliceToStrings(input []int) []string
	ReadFileToBytes(filename string) ([]byte, error)
}

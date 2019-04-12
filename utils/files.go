package utils

import "os"

func Create_folder(folder string) (err error) {
	_, err = os.Stat(folder)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(folder, 0755)
		if errDir != nil {
			panic(err)
		}
	}
	return
}
func Save_file(file_path string, content []byte) (int, error) {
	file, err := os.Create(file_path)
	if err == nil {
		n, err := file.Write(content)
		file.Close()
		if err == nil {
			return n, err
		}
	}
	return 0, err
}

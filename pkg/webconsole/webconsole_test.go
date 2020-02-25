package webconsole

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestConsoleYamlSamples(t *testing.T) {
	loadTestFiles("name","./examples", "consoleyamlsamples")
}
func TestWebConsole(t *testing.T) {
	loadTestFiles("name","./examples", "webconsole")
}


func loadTestFiles(boxname string, path string, folder string) {
	fullpath := strings.Join([]string{path, folder}, "/")

	fileList, err := ioutil.ReadDir(fullpath)
	if err != nil {
		fmt.Println( fmt.Errorf("%s not found with io ", fullpath))
	}

	var files []string
	for _, filename := range fileList {
		yamlStr, err := ioutil.ReadFile(fullpath + "/" + filename.Name())
		if err != nil {
			fmt.Println(err)
			continue
		}
		files = append(files, string(yamlStr))
	}
	//for _, f := range files {
	//	a := []rune(f)
	//	fmt.Println("filename: ", string(a[0: 10]))
	//
	//	err := ApplyWebConsoleYaml(f)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//}

	resMap, err := ApplyMultipleWebConsoleYamls(files)
	if err != nil {
		fmt.Println(err)
	}
	for k, v := range resMap {
		//fmt.Println(files[k])
		fmt.Println(k, " - ", v)
	}

}

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/micro/go-platform/config"
	"github.com/micro/go-platform/config/source/file"
)

var (
	configFile = filepath.Join(os.TempDir(), "config.example")
)

func writeFile(i int) error {
	return ioutil.WriteFile(configFile, []byte(fmt.Sprintf(`{"key": "value-%d"}`, i)), 0600)
}

func removeFile() error {
	return os.Remove(configFile)
}

func main() {
	flag.Parse()

	// Write our first entry
	if err := writeFile(0); err != nil {
		fmt.Println(err)
		return
	}
	defer os.Remove(configFile)

	// Create a config instance
	config := config.NewConfig(
		// aggressive config polling
		config.PollInterval(time.Millisecond*500),
		// use file as a config source
		config.WithSource(file.NewSource(config.SourceName(configFile))),
	)

	fmt.Println("Starting config runner")

	// Start the config runner which polls config
	config.Start()

	// lets read the value while editing it a number of times
	for i := 0; i < 10; i++ {
		val := config.Get("key").String("default")
		fmt.Println("Got ", val)
		writeFile(i + 1)
		time.Sleep(time.Second)
	}

	fmt.Println("Stopping config runner")

	// Stop the runner. The cache is still populated
	config.Stop()
}

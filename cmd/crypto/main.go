package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/spf13/viper"
)

var cryptogenPath = "../../cryptogen"

func getAbsPath(path string) string {
	abs, err := filepath.Abs(filepath.Join(cryptogenPath, path))
	if err != nil {
		log.Panicf("Failed to make abs path: %s", path)
	}
	return abs
}

func main() {
	viper.SetConfigName("cryptogen")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../cryptogen")
	viper.AddConfigPath(".")
	viper.AddConfigPath("cryptogen")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	scriptPath := viper.GetString("scriptPath")
	scriptPath = getAbsPath(scriptPath)
	configs := viper.GetStringMap("configs")

	for org := range configs {
		configPath := fmt.Sprintf("configs.%s.path", org)
		outputPath := fmt.Sprintf("configs.%s.output", org)

		configPath = getAbsPath(viper.GetString(configPath))
		outputPath = getAbsPath(viper.GetString(outputPath))

		command := exec.Command("/bin/bash",
			scriptPath,
			"-c", configPath,
			"-o", outputPath,
			"-i", org,
		)
		out, err := command.Output()
		log.Fatalf("Executed command [%s] %s\nErrorCode = %s\nOutput = %s\n", command.Dir, command.Args, err, out)
	}
}

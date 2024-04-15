/*
env is a general-purpose package to load environment variables from files.
*/
package env

import (
	"bufio"
	"os"
	"strings"
)

/*
LoadEnv assumes default values for LoadEnvFromFile, where the filePath is equal to ".env"
and the returnNilErrOnNoFile is true. See LoadEnvFromFile for specific documentation on
parameters.
*/
func LoadEnv() (map[string]string, error) {
	return LoadEnvFromFile(".env", true)
}

/*
LoadEnvFromFile reads a given "env" filePath and populates the current environment with
the values inside. Also returns the map object that contains all the environment
variable keys along with their values that were present in the file.
returnNilErrOnNoFile tells the function to ignore an error when the filePath given does
not exist on the filesystem. This is useful in the case where loading the environment
file contents isn't critical to the execution of the application.
*/
func LoadEnvFromFile(filePath string, returnNilErrOnNoFile bool) (map[string]string, error) {
	envFile, err := os.Open(filePath)
	if err != nil {
		if returnNilErrOnNoFile {
			return make(map[string]string), nil
		}

		return nil, err
	}
	defer envFile.Close()

	envMap := make(map[string]string)

	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		firstEqualsIndex := strings.Index(line, "=")
		if firstEqualsIndex == -1 {
			continue
		}

		key := line[:firstEqualsIndex]
		value := line[firstEqualsIndex+1:]

		envMap[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	for key, value := range envMap {
		if err := os.Setenv(key, value); err != nil {
			return nil, err
		}
	}

	return envMap, nil
}

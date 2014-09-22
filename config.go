package main

import (
	"bufio"
	"os"
  "strings"
)

func ReadConfig(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func ProcessConfig(path string) (map[string]string, error) {
  lines, err := ReadConfig(path)
 
  if err != nil {
    return nil, err
  }

  config := make(map[string]string)

  for _, line := range lines {
    if Contains(line, "ircServer") {
      ConfigParse(config, line, "ircServer")
    } else if Contains(line, "ircPassword") {
      ConfigParse(config, line, "ircPassword")
    } else if Contains(line, "ircChannel") {
      ConfigParse(config, line, "ircChannel")
    } else if Contains(line, "ircNickname") {
      ConfigParse(config, line, "ircNickname")
    } else if Contains(line, "ircUsername") {
      ConfigParse(config, line, "ircUsername") 
    } else if Contains(line, "csServer") {
      ConfigParse(config, line, "csServer") 
    } else if Contains(line, "csMaps") {
      ConfigParse(config, line, "csMaps")
    } else if Contains(line, "csRconPassword") {
      ConfigParse(config, line, "csRconPassword")
    } else if Contains(line, "csPugAdminPassword") {
      ConfigParse(config, line, "csPugAdminPassword")
    }
  }
  return config, nil
}

func Contains(line, value string) (bool) {
  if !strings.Contains(line, value) {
    return false
  }
  return true
}

func ConfigParse(config map[string]string, line string, value string) {
  s := value + "="
  config[value] = strings.TrimPrefix(line, s)
}
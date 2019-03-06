package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"gopkg.in/yaml.v2"
)

func exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

type configItem struct {
	Condition map[string][]string `yaml:"when"`
	Commands []string `yaml:"command"`
}

func main() {
	var config string
	var logfile string
	flag.StringVar(&config, "config", "/etc/default/condrun.yaml", "configuration file")
	flag.StringVar(&logfile, "log", "-", "log file")
	flag.Parse()
	if logfile == "-" {
		log.SetOutput(os.Stdout)
	} else {
		f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("error opening log file %v: %v", logfile, err)
		}
		defer f.Close()
		log.SetOutput(f)
	}
	log.Printf("Invoked with arguments: %v", flag.Args())
	if len(flag.Args()) == 0 {
		log.Fatal("No conditions to run")
	}
	if exists, _ := exists(config); exists != true {
		log.Fatalf("Config does not exists: %v", config)
	}
	data, err := ioutil.ReadFile(config)
	if err != nil {
		log.Fatalf("Cannot read file %v: %v", config, err)
	}
	var ci map[string]configItem
	err = yaml.Unmarshal([]byte(data), &ci)
	if err != nil {
		log.Fatalf("Cannot parse %v: %v", config, err)
	}
	findNumber := regexp.MustCompile("[0-9]+")
	for item, definition := range ci {
		log.Printf("Processing '%v' entry. Checking conditions...", item)
		matched := true
		for argument, variants := range definition.Condition {
			index, err := strconv.Atoi(findNumber.FindString(argument))
			if err != nil {
				matched = false
				log.Printf("Wrong argument for condition: %v, should be like 'arg1' or just '1'", argument)
				break
			}
			if index > len(flag.Args()) || index < 1 {
				matched = false
				log.Printf("Argument index %v out of bounds [1:%v]", index, len(flag.Args()) + 1)
				break
			}
			arg := flag.Args()[index - 1]
			equals := false
			for _, variant := range variants {
				if arg == variant {
					log.Printf("%v = %v", argument, variant)
					equals = true
					break
				}
			}
			if equals == false {
				matched = false
				log.Printf("Argument '%v' (%v) didn't matched any of values: %v", arg, argument, variants)
				break
			}
		}
		if matched == true {
			log.Println("All conditions passed! Executing commands...")
			for _, command := range definition.Commands {
				log.Printf("Executing with uid %v: %v", syscall.Getuid(), command)
				args := strings.Split(command, " ")
				var cmd *exec.Cmd
				if len(args) > 1 {
					cmd = exec.Command(args[0], args[1:]...)
				} else {
					cmd = exec.Command(args[0])
				}
				err := cmd.Run()
				if err != nil {
					log.Fatal(err)
				}
			}
			log.Println("done")
		} else {
			log.Println("Skipping, because arguments didn't matched to conditions")
		}
	}
}

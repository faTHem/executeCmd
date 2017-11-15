package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/howeyc/gopass"
	"golang.org/x/crypto/ssh"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Printf("You must use flags, use -h for examples\n")
		os.Exit(1)
	}

	commandPtr := flag.String("c", "", "This is for the list of commands to be run.")
	devicePtr := flag.String("d", "", "This is for the list of hosts")
	flag.Parse()

	cmd, err := readLines(*commandPtr)
	if err != nil {
		fmt.Printf("%v \n", err)
	}

	hosts, err := readLines(*devicePtr)
	if err != nil {
		fmt.Printf("%v \n", err)
	}

	config := getClientConfig()

	//Create a results channel to pass output from executeCmd
	results := make(chan string)

	// Here is where we loop over the list of commands and hosts using a goroutine
	// then pass the output to the results channel for display
	wg := &sync.WaitGroup{}
	wg.Add(len(cmd) * len(hosts))

	for _, hostname := range hosts {
		fmt.Printf("RUNNING ON DEVICE: %v\n", hostname)
		for _, command := range cmd {
			fmt.Printf("RUNNING COMMAND: %v\n", command)
			go func(command, hostname string) {
				results <- executeCmd(command, hostname, config)
			}(command, hostname)
		}
	}
	for i := 0; i < len(cmd)*len(hosts); i++ {
		res := <-results
		fmt.Println(res)
		wg.Done()
	}
	wg.Wait()
	close(results)
}

//executeCmd allows the program to ssh into a device, run a command and return the output
//func executCmdWithAuth(host, username string)
func executeCmd(command, hostname string, config *ssh.ClientConfig) string {
	conn, err := ssh.Dial("tcp", hostname+":22", config)
	if err != nil {
		return fmt.Sprintf("Error: could not open  - %v", err)
	}
	session, err := conn.NewSession()
	if err != nil {
		return fmt.Sprintf("Error: could not establish session to host - %v", err)
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Run(command)

	return hostname + ": " + stdoutBuf.String()
}

//readLines opens a file and reads in the contents into a slice of strings
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Error: could not open  - %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func getClientConfig() *ssh.ClientConfig {
	var pass []byte
	for len(pass) == 0 {
		fmt.Printf("Password: ")
		pass, _ = gopass.GetPasswd()
	}
	config := &ssh.ClientConfig{
		User: os.Getenv("USER"),
		Auth: []ssh.AuthMethod{
			ssh.Password(string(pass)),
		},
		Timeout: time.Second * 5,
	}
	return config
}

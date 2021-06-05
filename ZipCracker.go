package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	zip "github.com/yeka/zip"
)

func main() {
	//Creating a channel for the goroutine
	c := make(chan string)
	//vars for the input
	var list string
	var zipFile string

	InputScanner := bufio.NewScanner(os.Stdin)
	//user input
	fmt.Println("Input the zip file location:")
	InputScanner.Scan()
	zipFile = InputScanner.Text()

	fmt.Println("Input the wordlist file location:")
	InputScanner.Scan()
	list = InputScanner.Text()

	//starts the timer to see who long the script took to finish
	start := time.Now()

	//error handling
	file, err := os.Open(list)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	//this opens the wordlist file and turns it into an extension
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var password []string
	for scanner.Scan() {
		password = append(password, scanner.Text())
	}
	file.Close()

	//this opens the zip file
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	//passsword := []string{"lol", "testa", "adad", "lmao", "tesad", "asdasd", "test", "hehe"} <= this was a small test
	//counter vars
	i := 0
	p := 0

	//Creating a new wait group
	var wg sync.WaitGroup
	wg.Add(len(password))

	//The wordlist brute force loop that inputs different passwords
	for range password {
		//This goroutine inputs the password and returns the result
		//Expected the goroutine to speed up the process. It shifted the time needed from 3s to 2.6s (14-20% boost)
		go func(i int) {
			defer wg.Done()
			//Sends array item to the function
			fmt.Println("password attempt:", password[i])
			x := fmt.Sprint(attempt(r, password[i]))
			//forwards the function result into a channel
			c <- x
		}(i)

		//Receving info from the channel and printing it
		x := <-c
		fmt.Print(x)
		//if condition to see when the posswords has been found
		if strings.HasPrefix(x, "Password correct") {
			fmt.Printf("***PASSWORD FOUND: %v ***\n", password[i])
			//Timer stop
			duration := time.Since(start)
			//Time needed and the amount of passwords tried
			fmt.Printf("time needed: %.2f s. %v passwords tried.", duration.Seconds(), p)
			break
		}
		//counter incrementation
		i++
		p++
	}

	wg.Wait()

}

//attempt is a function that sends a password to the zip file and returns the result
func attempt(r *zip.ReadCloser, password string) string {
	var result string
	//loop that goes trough every file in the zip file
	for _, f := range r.File {
		//if the zip is encrypted then it will attempt the password
		if f.IsEncrypted() {
			f.SetPassword(password)
		}
		//opens the file and sends an error if the passwords is incorrect
		r, err := f.Open()
		if err != nil {
			log.Println("File: "+f.Name, err)
			continue
		}
		//Simple function the just reads how big the file is
		buf, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		defer r.Close()
		//In the case that the password is correct, it will display this string with its name and file size
		result = result + fmt.Sprintf("Password correct for: %v, file size: %v\n", f.Name, len(buf))
	}
	return result
}

package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main(){
	//fmt.Println(string.Reverse("hello, world\n"))
	cmd := exec.Command("gcloud", 
		"auth",
		"application-default",
		"print-access-token",)
	//var a []string
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", out)
}

package main

import (
	"log"
	"time"

	"github.com/TheRedScreen64/go-exiftool"
)

func main() {
	// Measure time that it takes to execute
	start := time.Now()

	et, err := exiftool.NewExifTool("./exiftool.exe")
	if err != nil {
		log.Fatal(err)
	}

	// _, err = et.Exec([]string{"-json", "./image.jpg"})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	for i := 0; i < 1000; i++ {
		start := time.Now()
		_, err := et.GetMetadata("./image.jpg", "./image-2.jpg")
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Printf("%p\n", m)
		elapsed := time.Since(start)
		log.Println("Execution time:", elapsed)
	}

	time.Sleep(60 * time.Second)

	et.Close()

	elapsed := time.Since(start)
	log.Println("Execution time:", elapsed)
}

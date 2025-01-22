package main

import (
	"fmt"
	"log"
	"net/http"

	"cpa/handlers"
)

func main() {
	if err := handlers.LoadPalettes("palettes/palettes.json"); err != nil {
		log.Fatalf("Palettes could not be loaded: %v", err)
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handlers.IndexHandler)
	http.HandleFunc("/palettes", handlers.PalettesHandler)
	http.HandleFunc("/extractPalette", handlers.ExtractPaletteHandler)
	http.HandleFunc("/process", handlers.ProcessHandler)

	fmt.Println("The server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

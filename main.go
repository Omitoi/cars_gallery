package main

import (
	"errors"
	"log"
	"net/http"
	"os/exec"
	"sync"
)

type Manufacturer struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Country      string `json:"country"`
	FoundingYear int    `json:"foundingYear"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type carModel struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	ManufacturerID int    `json:"manufacturerId"`
	CategoryID     int    `json:"categoryId"`
	Year           int    `json:"year"`
	Specifications struct {
		Engine       string `json:"engine"`
		Horsepower   int    `json:"horsepower"`
		Transmission string `json:"transmission"`
		Drivetrain   string `json:"drivetrain"`
	} `json:"specifications"`
	Image string `json:"image"`
}

type searchData struct {
	Search        string
	Category      int
	Manufacturer  int
	Year          int
	Engine        string
	Horsepower    int
	Transmission  string
	Drivetrain    string
	EasterEggCars []carModel
}

var years []int
var categories []Category
var manufacturers []Manufacturer
var engines []string
var horsepowers []int
var hpmax int
var hpmin int
var transmissions []string
var drivetrains []string
var models []carModel
var currentPage int
var currentPageMutex sync.RWMutex
var loaded bool
var errorMessage string
var lastSearch searchData

func getCurrentPage() int {
	currentPageMutex.RLock()
	defer currentPageMutex.RUnlock()
	return currentPage
}

func setCurrentPage(page int) {
	currentPageMutex.Lock()
	defer currentPageMutex.Unlock()
	if page < 1 {
		page = 1
	}
	currentPage = page
}

func main() {
	go func() {
		cmd := exec.Command("node", "main.js")
		cmd.Dir = "api"
		if err := cmd.Start(); err != nil {
			log.Printf("Failed to start API: %v", err)
			return
		}
		log.Println("API server started on port 3000")
	}()

	loaded = true
	setCurrentPage(1)

	http.HandleFunc("/", handler)
	http.HandleFunc("/car", showCar)
	http.HandleFunc("/search", search)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Go server starting on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func removeDuplicateInts(elements []int) []int {
	encountered := map[int]bool{}
	result := []int{}

	for v := range elements {
		if !encountered[elements[v]] {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}
	return result
}

func removeDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if !encountered[elements[v]] {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}
	return result
}

func MinMax(nums []int) (int, int, error) {
	if len(nums) == 0 {
		return 0, 0, errors.New("slice is empty")
	}

	min, max := nums[0], nums[0]

	for _, num := range nums[1:] {
		if num < min {
			min = num
		}
		if num > max {
			max = num
		}
	}

	return min, max, nil
}

func getCategoryName(id int) string {
	for _, c := range categories {
		if c.ID == id {
			return c.Name
		}
	}
	return ""
}

func getCategoryId(name string) int {
	for _, c := range categories {
		if c.Name == name {
			return c.ID
		}
	}
	return 0
}

func getManuName(id int) string {
	for _, c := range manufacturers {
		if c.ID == id {
			return c.Name
		}
	}
	return ""
}

func getManuId(name string) int {
	for _, c := range manufacturers {
		if c.Name == name {
			return c.ID
		}
	}
	return 0
}

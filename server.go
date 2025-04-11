package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method == http.MethodGet {
		if !loaded {
			w.Write([]byte(mainPage(r)))
			return
		}
		updateData()
		nextPage, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if nextPage > 0 {
			setCurrentPage(nextPage)
		} else {
			setCurrentPage(1)
		}
		w.Write([]byte(mainPage(r)))
	}
}

func search(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		loaded = updateData()
		if !loaded {
			w.Write([]byte(mainPage(r)))
			return
		}
		var s searchData
		// Get form data
		nextPage, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if nextPage > 0 {
			setCurrentPage(nextPage)
		} else {
			setCurrentPage(1)
		}
		s.Search = r.URL.Query().Get("search")
		s.Category, _ = strconv.Atoi(r.URL.Query().Get("cat"))
		s.Manufacturer, _ = strconv.Atoi(r.URL.Query().Get("manu"))
		s.Year, _ = strconv.Atoi(r.URL.Query().Get("year"))
		s.Engine = r.URL.Query().Get("engine")
		hp, _ := strconv.Atoi(r.URL.Query().Get("horsepower"))
		if hp < hpmin {
			s.Horsepower = hpmin // Use minimum horsepower as default
		} else {
			s.Horsepower = hp
		}
		s.Transmission = r.URL.Query().Get("transmission")
		s.Drivetrain = r.URL.Query().Get("drivetrain")
		lastSearch = s

		// Filter cars
		var filteredCars []carModel

		//Easter Eggs
		if strings.ToLower(s.Search) == "felicia" {
			var felicia carModel
			for i := 0; i < 16; i++ {
				felicia.Name = "Škoda Felicia"
				felicia.ID = i + 982
				felicia.CategoryID = 999
				felicia.ManufacturerID = 999
				felicia.Year = 1998
				felicia.Specifications.Engine = "All of them!"
				felicia.Specifications.Horsepower = 999
				felicia.Specifications.Transmission = "Manual of course"
				felicia.Specifications.Drivetrain = "Basically floats."
				felicia.Image = "/static/felicia" + strconv.Itoa(i+1) + ".jpg"
				filteredCars = append(filteredCars, felicia)
			}
			setCurrentPage(1)
			// Store easter egg cars in lastSearch for searchPage to use
			lastSearch.EasterEggCars = filteredCars
			w.Write([]byte(searchPage(r)))
			return
		}
		if strings.ToLower(s.Search) == "peugeot" {
			elmo := carModel{
				Name:  "Le Baguette Forever",
				Image: "/static/elmo.gif",
			}
			var elmos []carModel
			elmos = append(elmos, elmo)
			setCurrentPage(1)
			// Store easter egg cars in lastSearch for searchPage to use
			lastSearch.EasterEggCars = elmos
			w.Write([]byte(searchPage(r)))
			return
		}

		for _, c := range models {
			hpCheck := c.Specifications.Horsepower >= s.Horsepower
			if s.Horsepower == hpmin {
				hpCheck = true
			}
			if (s.Search == "" || strings.Contains(strings.ToLower(c.Name), strings.ToLower(s.Search))) &&
				(s.Category == 0 || c.CategoryID == s.Category) &&
				(s.Manufacturer == 0 || c.ManufacturerID == s.Manufacturer) &&
				(s.Year == 0 || c.Year == s.Year) &&
				(s.Engine == "0" || c.Specifications.Engine == s.Engine) &&
				hpCheck &&
				(s.Transmission == "0" || c.Specifications.Transmission == s.Transmission) &&
				(s.Drivetrain == "0" || c.Specifications.Drivetrain == s.Drivetrain) {
				filteredCars = append(filteredCars, c)
			}
		}
		// Store filtered cars in lastSearch for searchPage to use
		lastSearch.EasterEggCars = filteredCars
		// Reset page to 1 if no cars found
		if len(filteredCars) == 0 {
			setCurrentPage(1)
		}
		w.Write([]byte(searchPage(r)))
	}
}

func showCar(w http.ResponseWriter, r *http.Request) {
	updateData()
	single := false
	if r.Method == http.MethodGet {
		ids := r.URL.Query()["id"]
		show := r.URL.Query()["show"]
		if len(show) > 0 {
			if show[0] == "true" {
				single = true
			}
		}

		var carModels []carModel
		for _, idStr := range ids {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}
			for _, c := range models {
				if c.ID == id {
					carModels = append(carModels, c)
				}
			}
		}
		w.Write([]byte(comparePage(carModels, single, r)))
	}
}

func updateData() bool {
	categoriesURL := "http://localhost:3000/api/categories/"
	modelsURL := "http://localhost:3000/api/models/"
	manufacturersURL := "http://localhost:3000/api/manufacturers/"
	imagesURL := "http://localhost:3000/api/images/"

	// Create channels to receive results
	categoriesChan := make(chan []Category, 1)
	manufacturersChan := make(chan []Manufacturer, 1)
	modelsChan := make(chan []carModel, 1)
	errChan := make(chan error, 3)

	// Fetch categories in a goroutine
	go func() {
		resp, err := http.Get(categoriesURL)
		if err != nil {
			errChan <- fmt.Errorf("failed to fetch categories: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errChan <- fmt.Errorf("categories API returned non-OK status")
			return
		}

		var cats []Category
		if err := json.NewDecoder(resp.Body).Decode(&cats); err != nil {
			errChan <- fmt.Errorf("failed to parse categories JSON: %v", err)
			return
		}
		categoriesChan <- cats
	}()

	// Fetch manufacturers in a goroutine
	go func() {
		resp, err := http.Get(manufacturersURL)
		if err != nil {
			errChan <- fmt.Errorf("failed to fetch manufacturers: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errChan <- fmt.Errorf("manufacturers API returned non-OK status")
			return
		}

		var mans []Manufacturer
		if err := json.NewDecoder(resp.Body).Decode(&mans); err != nil {
			errChan <- fmt.Errorf("failed to parse manufacturers JSON: %v", err)
			return
		}
		manufacturersChan <- mans
	}()

	// Fetch models in a goroutine
	go func() {
		resp, err := http.Get(modelsURL)
		if err != nil {
			errChan <- fmt.Errorf("failed to fetch models: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errChan <- fmt.Errorf("models API returned non-OK status")
			return
		}

		var mods []carModel
		if err := json.NewDecoder(resp.Body).Decode(&mods); err != nil {
			errChan <- fmt.Errorf("failed to parse models JSON: %v", err)
			return
		}
		modelsChan <- mods
	}()

	// Collect results and check for errors
	for i := 0; i < 3; i++ {
		select {
		case err := <-errChan:
			errorMessage = err.Error()
			return false
		case categories = <-categoriesChan:
		case manufacturers = <-manufacturersChan:
		case models = <-modelsChan:
		}
	}

	// Process the data
	for _, c := range models {
		years = append(years, c.Year)
		engines = append(engines, c.Specifications.Engine)
		horsepowers = append(horsepowers, c.Specifications.Horsepower)
		transmissions = append(transmissions, c.Specifications.Transmission)
		drivetrains = append(drivetrains, c.Specifications.Drivetrain)
	}

	years = removeDuplicateInts(years)
	sort.Sort(sort.Reverse(sort.IntSlice(years)))
	engines = removeDuplicates(engines)
	sort.Strings(engines)
	hpmin, hpmax, _ = MinMax(horsepowers)
	transmissions = removeDuplicates(transmissions)
	sort.Strings(transmissions)
	drivetrains = removeDuplicates(drivetrains)
	sort.Strings(transmissions)

	for i := range models {
		models[i].Image = imagesURL + models[i].Image
	}
	return true
}

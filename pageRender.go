package main

import (
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Helper function to check if a car is favorited
func isCarFavorited(carID int, favorites []string) bool {
	return getFavoritePosition(carID, favorites) >= 0
}

// Helper function to check if a car is favorited and get its position
func getFavoritePosition(carID int, favorites []string) int {
	for i, id := range favorites {
		if id == strconv.Itoa(carID) {
			return i
		}
	}
	return -1
}

// Helper function to sort cars by favorite status
func sortCarsByFavorite(cars []carModel, favorites []string) []carModel {
	// Create a copy of the slice to avoid modifying the original
	sortedCars := make([]carModel, len(cars))
	copy(sortedCars, cars)

	// Sort the cars
	sort.SliceStable(sortedCars, func(i, j int) bool {
		iPos := getFavoritePosition(sortedCars[i].ID, favorites)
		jPos := getFavoritePosition(sortedCars[j].ID, favorites)

		// If both are favorited, maintain the order from the favorites list
		if iPos >= 0 && jPos >= 0 {
			return iPos < jPos
		}

		// If only one is favorited, it should come first
		if iPos >= 0 {
			return true
		}
		if jPos >= 0 {
			return false
		}

		// If neither is favorited, maintain original order
		return i < j
	})

	return sortedCars
}

func mainPage(r *http.Request) string {
	lastSearch = searchData{}
	// Create a channel to receive the file content
	contentChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Read file in a goroutine
	go func() {
		content, err := os.ReadFile("./index.html")
		if err != nil {
			errChan <- err
			return
		}
		contentChan <- string(content)
	}()

	// Wait for file content
	var content string
	select {
	case err := <-errChan:
		panic(err)
	case content = <-contentChan:
	}

	page := strings.Split(content, "<!--Separator-->")
	var result string

	// Process the page sections concurrently
	type sectionResult struct {
		index   int
		content string
	}
	resultChan := make(chan sectionResult, 9)
	errorChan := make(chan error, 9)

	// Process each section in a goroutine
	go func() {
		resultChan <- sectionResult{0, page[0]} // Search value
	}()

	go func() {
		var section string
		section = page[1] // Categories form dropdown
		var toDisplayCategory []string
		for _, cat := range categories {
			toDisplayCategory = append(toDisplayCategory, cat.Name)
		}
		sort.Strings(toDisplayCategory)
		for _, cat := range toDisplayCategory {
			section += "<option value=\"" + strconv.Itoa(getCategoryId(cat)) + "\">" + cat + "</option>"
		}
		resultChan <- sectionResult{1, section}
	}()

	go func() {
		var section string
		section = page[2] // Manufacturers form dropdown
		var toDisplayManu []string
		for _, manu := range manufacturers {
			toDisplayManu = append(toDisplayManu, manu.Name)
		}
		sort.Strings(toDisplayManu)
		for _, manu := range toDisplayManu {
			section += "<option value=\"" + strconv.Itoa(getManuId(manu)) + "\">" + manu + "</option>"
		}
		resultChan <- sectionResult{2, section}
	}()

	go func() {
		var section string
		section = page[3] // Years form dropdown
		for _, year := range years {
			section += "<option value=\"" + strconv.Itoa(year) + "\">" + strconv.Itoa(year) + "</option>"
		}
		resultChan <- sectionResult{3, section}
	}()

	go func() {
		var section string
		section = page[4] // Engine form dropdown
		for _, engine := range engines {
			section += "<option value=\"" + engine + "\">" + engine + "</option>"
		}
		resultChan <- sectionResult{4, section}
	}()

	go func() {
		var section string
		section = page[5] // Horsepower form dropdown
		section += "<label for=\"horsepower\">Minimal horsepower: <output id=\"hpValue\">" + strconv.Itoa(hpmin) + "</output></label>"
		section += "<input type=\"range\" id=\"horsepower\" name=\"horsepower\" form=\"filter\" min=\"" + strconv.Itoa(hpmin) + "\" max=\"" + strconv.Itoa(hpmax) + "\" oninput=\"hpValue.value = this.value\" value=\"" + strconv.Itoa(hpmin) + "\">"
		resultChan <- sectionResult{5, section}
	}()

	go func() {
		var section string
		section = page[6] // Transmission form dropdown
		for _, ts := range transmissions {
			section += "<option value=\"" + ts + "\">" + ts + "</option>"
		}
		resultChan <- sectionResult{6, section}
	}()

	go func() {
		var section string
		section = page[7] // Drivetrain form dropdown
		for _, dt := range drivetrains {
			section += "<option value=\"" + dt + "\">" + dt + "</option>"
		}
		resultChan <- sectionResult{7, section}
	}()

	go func() {
		resultChan <- sectionResult{8, page[8]}
	}()

	// Collect results in order
	sections := make([]string, 9)
	for i := 0; i < 9; i++ {
		select {
		case err := <-errorChan:
			panic(err)
		case section := <-resultChan:
			sections[section.index] = section.content
		}
	}

	// Combine sections in order
	for _, section := range sections {
		result += section
	}

	result += "<div class=\"pagebuttons\">"
	result += "<a class=\"button"
	if getCurrentPage() <= 1 {
		result += " disabled"
	}
	result += "\" href=\"/?page=" + strconv.Itoa(getCurrentPage()-1) + "\">Previous Page</a>"

	availablePages := len(models) / 16
	if len(models)%16 > 0 {
		availablePages++
	}
	result += "<span>Current Page:" + strconv.Itoa(getCurrentPage()) + "/" + strconv.Itoa(availablePages) + "</span>"

	result += "<a class=\"button"
	if getCurrentPage() >= availablePages {
		result += " disabled"
	}
	result += "\" href=\"/?page=" + strconv.Itoa(getCurrentPage()+1) + "\">Next Page</a></div>"

	result += page[9]

	if !loaded {
		result += "<main class=\"error\">" +
			"<h1>" + errorMessage + "</h1>" +
			"<img src=\"./static/error.png\">" +
			"</main>"
		return result
	}

	result += "<main class=\"content\">"

	// Get favorites from cookie
	favorites := []string{}
	if cookie, err := r.Cookie("favorites"); err == nil {
		favorites = strings.Split(cookie.Value, ",")
	}

	// Sort models by favorite status
	models = sortCarsByFavorite(models, favorites)

	// Calculate the range of cars to display
	startIdx := (getCurrentPage() - 1) * 16
	endIdx := getCurrentPage() * 16
	if endIdx > len(models) {
		endIdx = len(models)
	}

	// Create a slice to store car results in order
	carResults := make([]string, endIdx-startIdx)
	carChan := make(chan struct {
		index   int
		content string
	}, endIdx-startIdx)

	// Launch goroutines for each car in the current page
	for i := startIdx; i < endIdx; i++ {
		go func(idx int, car carModel) {
			carChan <- struct {
				index   int
				content string
			}{
				index: idx - startIdx,
				content: "<div class=\"car-card\">" +
					"<a href=\"/car?id=" + strconv.Itoa(car.ID) + "&show=true\"> <div class=\"image-box\" style=\"background-image: url(" + car.Image + ");\">" +
					"<input type=\"checkbox\" name=\"id\" value=\"" + strconv.Itoa(car.ID) + "\" form=\"compare\">" +
					"<label for=\"car" + strconv.Itoa(car.ID) + "\">" + car.Name + "</label>" +
					"</div></a>" +
					"<button class=\"favorite-btn\" onclick=\"toggleFavorite(" + strconv.Itoa(car.ID) + ")\" data-car-id=\"" + strconv.Itoa(car.ID) + "\" aria-label=\"Add " + car.Name + " to favorites\">" +
					"<svg class=\"star-icon\" width=\"24\" height=\"24\" viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\">" +
					"<path d=\"M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z\"></path>" +
					"</svg>" +
					"</button>" +
					"</div>",
			}
		}(i, models[i])
	}

	// Collect results in order
	for i := 0; i < endIdx-startIdx; i++ {
		carResult := <-carChan
		carResults[carResult.index] = carResult.content
	}

	// Combine results in order
	for _, carResult := range carResults {
		result += carResult
	}

	result += "</main>"
	return result
}

func searchPage(r *http.Request) string {
	if lastSearch.Search == "" && lastSearch.Category == 0 && lastSearch.Manufacturer == 0 && lastSearch.Year == 0 && lastSearch.Engine == "0" && lastSearch.Horsepower == 139 && lastSearch.Transmission == "0" && lastSearch.Drivetrain == "0" {
		return mainPage(r)
	}

	// Use easter egg cars if available, otherwise filter cars based on lastSearch
	var carsToShow []carModel
	if len(lastSearch.EasterEggCars) > 0 {
		carsToShow = lastSearch.EasterEggCars
	} else {
		for _, c := range models {
			hpCheck := c.Specifications.Horsepower >= lastSearch.Horsepower
			if lastSearch.Horsepower == hpmin {
				hpCheck = true
			}
			if (lastSearch.Search == "" || strings.Contains(strings.ToLower(c.Name), strings.ToLower(lastSearch.Search))) &&
				(lastSearch.Category == 0 || c.CategoryID == lastSearch.Category) &&
				(lastSearch.Manufacturer == 0 || c.ManufacturerID == lastSearch.Manufacturer) &&
				(lastSearch.Year == 0 || c.Year == lastSearch.Year) &&
				(lastSearch.Engine == "0" || c.Specifications.Engine == lastSearch.Engine) &&
				hpCheck &&
				(lastSearch.Transmission == "0" || c.Specifications.Transmission == lastSearch.Transmission) &&
				(lastSearch.Drivetrain == "0" || c.Specifications.Drivetrain == lastSearch.Drivetrain) {
				carsToShow = append(carsToShow, c)
			}
		}
	}

	// Create a channel to receive the file content
	contentChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Read file in a goroutine
	go func() {
		content, err := os.ReadFile("./index.html")
		if err != nil {
			errChan <- err
			return
		}
		contentChan <- string(content)
	}()

	// Wait for file content
	var content string
	select {
	case err := <-errChan:
		panic(err)
	case content = <-contentChan:
	}

	page := strings.Split(content, "<!--Separator-->")
	var result string

	// Process the page sections concurrently
	type sectionResult struct {
		index   int
		content string
	}
	resultChan := make(chan sectionResult, 9)
	errorChan := make(chan error, 9)

	// Process each section in a goroutine
	go func() {
		resultChan <- sectionResult{0, page[0] + "value=\"" + lastSearch.Search + "\""} // Search value
	}()

	go func() {
		var section string
		section = page[1] // Categories form dropdown
		seenCategories := make(map[int]bool)
		selectedCategory := getCategoryName(lastSearch.Category)
		var toDisplayCategory []string
		for _, cat := range carsToShow {
			if !seenCategories[cat.CategoryID] {
				seenCategories[cat.CategoryID] = true
				toDisplayCategory = append(toDisplayCategory, getCategoryName(cat.CategoryID))
			}
		}
		sort.Strings(toDisplayCategory)
		for _, cat := range toDisplayCategory {
			if cat == selectedCategory {
				section += "<option selected value=\"" + strconv.Itoa(getCategoryId(cat)) + "\">" + cat + "</option>"
			} else {
				section += "<option value=\"" + strconv.Itoa(getCategoryId(cat)) + "\">" + cat + "</option>"
			}
		}
		resultChan <- sectionResult{1, section}
	}()

	go func() {
		var section string
		section = page[2] // Manufacturers form dropdown
		seenManu := make(map[int]bool)
		selectedManu := getManuName(lastSearch.Manufacturer)
		var toDisplayManu []string
		for _, manu := range carsToShow {
			if !seenManu[manu.ManufacturerID] {
				seenManu[manu.ManufacturerID] = true
				manuName := getManuName(manu.ManufacturerID)
				if manuName != "" {
					toDisplayManu = append(toDisplayManu, manuName)
				}
			}
		}
		sort.Strings(toDisplayManu)
		for _, manu := range toDisplayManu {
			if manu == selectedManu && selectedManu != "0" {
				section += "<option selected value=\"" + strconv.Itoa(getManuId(manu)) + "\">" + manu + "</option>"
			} else {
				section += "<option value=\"" + strconv.Itoa(getManuId(manu)) + "\">" + manu + "</option>"
			}
		}
		resultChan <- sectionResult{2, section}
	}()

	go func() {
		var section string
		section = page[3] // Years form dropdown
		seenYear := make(map[int]bool)
		for _, year := range carsToShow {
			if !seenYear[year.Year] {
				seenYear[year.Year] = true
				if year.Year == lastSearch.Year {
					section += "<option selected value=\"" + strconv.Itoa(year.Year) + "\">" + strconv.Itoa(year.Year) + "</option>"
				} else {
					section += "<option value=\"" + strconv.Itoa(year.Year) + "\">" + strconv.Itoa(year.Year) + "</option>"
				}
			}
		}
		resultChan <- sectionResult{3, section}
	}()

	go func() {
		var section string
		section = page[4] // Engine form dropdown
		seenEngine := make(map[string]bool)
		for _, engine := range carsToShow {
			if !seenEngine[engine.Specifications.Engine] {
				seenEngine[engine.Specifications.Engine] = true
				if engine.Specifications.Engine == lastSearch.Engine {
					section += "<option selected value=\"" + engine.Specifications.Engine + "\">" + engine.Specifications.Engine + "</option>"
				} else {
					section += "<option value=\"" + engine.Specifications.Engine + "\">" + engine.Specifications.Engine + "</option>"
				}
			}
		}
		resultChan <- sectionResult{4, section}
	}()

	go func() {
		var section string
		section = page[5] // Horsepower form dropdown
		section += "<label for=\"horsepower\">Minimal horsepower: <output id=\"hpValue\">" + strconv.Itoa(lastSearch.Horsepower) + "</output></label>"
		section += "<input type=\"range\" id=\"horsepower\" name=\"horsepower\" form=\"filter\" min=\"" + strconv.Itoa(hpmin) + "\" max=\"" + strconv.Itoa(hpmax) + "\" oninput=\"hpValue.value = this.value\" value=\"" + strconv.Itoa(lastSearch.Horsepower) + "\">"
		resultChan <- sectionResult{5, section}
	}()

	go func() {
		var section string
		section = page[6] // Transmission form dropdown
		seenTs := make(map[string]bool)
		for _, ts := range carsToShow {
			if !seenTs[ts.Specifications.Transmission] {
				seenTs[ts.Specifications.Transmission] = true
				if ts.Specifications.Transmission == lastSearch.Transmission {
					section += "<option selected value=\"" + ts.Specifications.Transmission + "\">" + ts.Specifications.Transmission + "</option>"
				} else {
					section += "<option value=\"" + ts.Specifications.Transmission + "\">" + ts.Specifications.Transmission + "</option>"
				}
			}
		}
		resultChan <- sectionResult{6, section}
	}()

	go func() {
		var section string
		section = page[7] // Drivetrain form dropdown
		seenDt := make(map[string]bool)
		for _, dt := range carsToShow {
			if !seenDt[dt.Specifications.Drivetrain] {
				seenDt[dt.Specifications.Drivetrain] = true
				if dt.Specifications.Drivetrain == lastSearch.Drivetrain {
					section += "<option selected value=\"" + dt.Specifications.Drivetrain + "\">" + dt.Specifications.Drivetrain + "</option>"
				} else {
					section += "<option value=\"" + dt.Specifications.Drivetrain + "\">" + dt.Specifications.Drivetrain + "</option>"
				}
			}
		}
		resultChan <- sectionResult{7, section}
	}()

	go func() {
		resultChan <- sectionResult{8, page[8]}
	}()

	// Collect results in order
	sections := make([]string, 9)
	for i := 0; i < 9; i++ {
		select {
		case err := <-errorChan:
			panic(err)
		case section := <-resultChan:
			sections[section.index] = section.content
		}
	}

	// Combine sections in order
	for _, section := range sections {
		result += section
	}

	// Calculate pagination
	availablePages := len(carsToShow) / 16
	if len(carsToShow)%16 > 0 {
		availablePages++
	}
	if availablePages < 1 {
		availablePages = 1
	}
	if getCurrentPage() > availablePages {
		setCurrentPage(availablePages)
	}

	// Add pagination buttons
	result += "<div class=\"pagebuttons\">"
	result += "<a class=\"button"
	if currentPage <= 1 {
		result += " disabled"
	}
	result += "\" href=\"/search?page=" + strconv.Itoa(currentPage-1) +
		"&search=" + lastSearch.Search +
		"&cat=" + strconv.Itoa(lastSearch.Category) +
		"&manu=" + strconv.Itoa(lastSearch.Manufacturer) +
		"&year=" + strconv.Itoa(lastSearch.Year) +
		"&engine=" + lastSearch.Engine +
		"&horsepower=" + strconv.Itoa(lastSearch.Horsepower) +
		"&transmission=" + lastSearch.Transmission +
		"&drivetrain=" + lastSearch.Drivetrain + "\">Previous Page</a>"

	result += "<span>Current Page:" + strconv.Itoa(currentPage) + "/" + strconv.Itoa(availablePages) + "</span>"

	result += "<a class=\"button"
	if currentPage >= availablePages {
		result += " disabled"
	}
	result += "\" href=\"/search?page=" + strconv.Itoa(currentPage+1) +
		"&search=" + lastSearch.Search +
		"&cat=" + strconv.Itoa(lastSearch.Category) +
		"&manu=" + strconv.Itoa(lastSearch.Manufacturer) +
		"&year=" + strconv.Itoa(lastSearch.Year) +
		"&engine=" + lastSearch.Engine +
		"&horsepower=" + strconv.Itoa(lastSearch.Horsepower) +
		"&transmission=" + lastSearch.Transmission +
		"&drivetrain=" + lastSearch.Drivetrain + "\">Next Page</a></div>"

	result += page[9]

	if len(carsToShow) == 0 {
		errorMessage = "No cars found"
		result += "<main class=\"error\">" +
			"<h1>" + errorMessage + "</h1>" +
			"<img src=\"./static/error.png\">" +
			"</main>"
		return result
	}

	result += "<main class=\"content\">"

	// Get favorites from cookie
	favorites := []string{}
	if cookie, err := r.Cookie("favorites"); err == nil {
		favorites = strings.Split(cookie.Value, ",")
	}

	// Sort carsToShow by favorite status
	carsToShow = sortCarsByFavorite(carsToShow, favorites)

	// Calculate the range of cars to display
	startIdx := (currentPage - 1) * 16
	endIdx := currentPage * 16
	if endIdx > len(carsToShow) {
		endIdx = len(carsToShow)
	}

	// Create a slice to store car results in order
	carResults := make([]string, endIdx-startIdx)
	carChan := make(chan struct {
		index   int
		content string
	}, endIdx-startIdx)

	// Launch goroutines for each car in the current page
	for i := startIdx; i < endIdx; i++ {
		go func(idx int, car carModel) {
			carChan <- struct {
				index   int
				content string
			}{
				index: idx - startIdx,
				content: "<div class=\"car-card\">" +
					"<a href=\"/car?id=" + strconv.Itoa(car.ID) + "&show=true\"> <div class=\"image-box\" style=\"background-image: url(" + car.Image + ");\">" +
					"<input type=\"checkbox\" name=\"id\" value=\"" + strconv.Itoa(car.ID) + "\" form=\"compare\">" +
					"<label for=\"car" + strconv.Itoa(car.ID) + "\">" + car.Name + "</label>" +
					"</div></a>" +
					"<button class=\"favorite-btn\" onclick=\"toggleFavorite(" + strconv.Itoa(car.ID) + ")\" data-car-id=\"" + strconv.Itoa(car.ID) + "\" aria-label=\"Add " + car.Name + " to favorites\">" +
					"<svg class=\"star-icon\" width=\"24\" height=\"24\" viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\">" +
					"<path d=\"M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z\"></path>" +
					"</svg>" +
					"</button>" +
					"</div>",
			}
		}(i, carsToShow[i])
	}

	// Collect results in order
	for i := 0; i < endIdx-startIdx; i++ {
		carResult := <-carChan
		carResults[carResult.index] = carResult.content
	}

	// Combine results in order
	for _, carResult := range carResults {
		result += carResult
	}

	result += "</main>"
	return result
}

func comparePage(carsToCompare []carModel, single bool, r *http.Request) string {
	var page []string
	var result string
	content, err := os.ReadFile("./comparison.html")
	if err != nil {
		panic(err)
	}

	page = strings.Split(string(content), "<!--Separator-->")

	result += page[0]
	numCars := len(carsToCompare)
	if single {
		result += "<h2>Specifications</h2>"
	} else {
		if numCars > 4 {
			result += "<h2>Comparison</h2><p style=\"color: red;\">Max 4 cars can be compared at once</p>"
		} else {
			result += "<h2>Comparison</h2>"
		}
	}

	// Add related cars section for single car view
	if single && len(carsToCompare) > 0 {
		// Get the category of the current car
		currentCategory := carsToCompare[0].CategoryID

		// Find other cars from the same category
		var relatedCars []carModel
		for _, car := range models {
			if car.CategoryID == currentCategory && car.ID != carsToCompare[0].ID {
				relatedCars = append(relatedCars, car)
			}
		}

		// Sort related cars by favorite status
		favorites := []string{}
		if cookie, err := r.Cookie("favorites"); err == nil {
			favorites = strings.Split(cookie.Value, ",")
		}
		relatedCars = sortCarsByFavorite(relatedCars, favorites)

		// Take up to 4 cars
		if len(relatedCars) > 4 {
			relatedCars = relatedCars[:4]
		}

		// Add the related cars section
		result += "<div class=\"related-cars\">"
		result += "<h3>Related Cars</h3>"
		result += "<div class=\"mini-car-grid\">"

		for _, car := range relatedCars {
			result += "<div class=\"mini-car-card\">" +
				"<a href=\"/car?id=" + strconv.Itoa(car.ID) + "&show=true\">" +
				"<div class=\"mini-image-box\" style=\"background-image: url(" + car.Image + ");\">" +
				"<input type=\"checkbox\" name=\"id\" value=\"" + strconv.Itoa(car.ID) + "\" form=\"compare\">" +
				"<label>" + car.Name + "</label>" +
				"</div></a>" +
				"<button class=\"favorite-btn\" onclick=\"toggleFavorite(" + strconv.Itoa(car.ID) + ")\" data-car-id=\"" + strconv.Itoa(car.ID) + "\" aria-label=\"Add " + car.Name + " to favorites\">" +
				"<svg class=\"star-icon\" width=\"24\" height=\"24\" viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\">" +
				"<path d=\"M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z\"></path>" +
				"</svg>" +
				"</button>" +
				"</div>"
		}
		result += "</div></div>"
	}

	result += "<div class=\"button-group tmarg40\">"
	if lastSearch.Search == "" && lastSearch.Category == 0 && lastSearch.Manufacturer == 0 && lastSearch.Year == 0 && lastSearch.Engine == "" && lastSearch.Horsepower == 0 && lastSearch.Transmission == "" && lastSearch.Drivetrain == "" {
		result += "<a href=\"/\" class=\"button\">Back to the Gallery</a>"
	} else {
		result += "<a href=\"/search?search=" + lastSearch.Search + "&cat=" + strconv.Itoa(lastSearch.Category) + "&manu=" + strconv.Itoa(lastSearch.Manufacturer) + "&year=" + strconv.Itoa(lastSearch.Year) + "&engine=" + lastSearch.Engine + "&horsepower=" + strconv.Itoa(lastSearch.Horsepower) + "&transmission=" + lastSearch.Transmission + "&drivetrain=" + lastSearch.Drivetrain + "\" class=\"button\">Back to the Gallery</a>"
	}
	if single && len(carsToCompare) > 0 {
		result += "<form action=\"/car\" method=\"GET\" id=\"compare\">"
		// Add hidden input for current car
		result += "<input type=\"hidden\" name=\"id\" value=\"" + strconv.Itoa(carsToCompare[0].ID) + "\">"
		result += "<input type=\"submit\" value=\"Compare with Selected\" class=\"button\">"
		result += "</form>"
	}
	result += "</div>"
	result += page[1]
	var carCat string
	if len(carsToCompare) != 0 {
		if single {
			for _, cate := range categories {
				if carsToCompare[0].CategoryID == cate.ID {
					carCat = cate.Name
				}
			}
			// Get favorites from cookie
			favorites := []string{}
			if cookie, err := r.Cookie("favorites"); err == nil {
				favorites = strings.Split(cookie.Value, ",")
			}
			result += "<main class=\"content-single\">" +
				"<div class=\"image-box\" style=\"background-image: url(" + carsToCompare[0].Image + ");\"><img src=\"" + carsToCompare[0].Image + "\"></div>" +
				"<button class=\"favorite-btn" + func() string {
				if isCarFavorited(carsToCompare[0].ID, favorites) {
					return " active"
				}
				return ""
			}() + "\" onclick=\"toggleFavorite(" + strconv.Itoa(carsToCompare[0].ID) + ")\" data-car-id=\"" + strconv.Itoa(carsToCompare[0].ID) + "\" aria-label=\"" + func() string {
				if isCarFavorited(carsToCompare[0].ID, favorites) {
					return "Remove " + carsToCompare[0].Name + " from favorites"
				}
				return "Add " + carsToCompare[0].Name + " to favorites"
			}() + "\">" +
				"<svg class=\"star-icon\" width=\"24\" height=\"24\" viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\">" +
				"<path d=\"M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z\"></path>" +
				"</svg>" +
				"</button>" +
				"<div class=\"columns\">" +
				"<h3>Car Name:</h3><h3>" + carsToCompare[0].Name + "</h3>" +
				"<p>Category:</p><p>" + carCat + "</p>" +
				"<p>Year:</p><p>" + strconv.Itoa(carsToCompare[0].Year) + "</p>" +
				"<p>Engine:</p><p>" + carsToCompare[0].Specifications.Engine + "</p>" +
				"<p>Horsepower:</p><p>" + strconv.Itoa(carsToCompare[0].Specifications.Horsepower) + " hp</p>" +
				"<p>Transmission:</p><p>" + carsToCompare[0].Specifications.Transmission + "</p>" +
				"<p>Drivetrain:</p><p>" + carsToCompare[0].Specifications.Drivetrain + "</p>"
			for _, manu := range manufacturers {
				if carsToCompare[0].ManufacturerID == manu.ID {
					result += "<p>Manufacturer:</p><p>" + manu.Name + "</p>" +
						"<p>Country:</p><p>" + manu.Country + "</p>" +
						"<p>Founding Year:</p><p>" + strconv.Itoa(manu.FoundingYear) + "</p>"
				}
			}
			result += "</div></main>"
		} else {
			result += "<main class=\"content-compare\">" +
				"<div class=\"image-box\" id=\"no-shadow\"></div>" +
				"<h3>Car Name</h3>" +
				"<p>Category</p>" +
				"<p>Year</p>" +
				"<p>Engine</p>" +
				"<p>Horsepower</p>" +
				"<p>Transmission</p>" +
				"<p>Drivetrain</p>" +
				"<p>Manufacturer</p>" +
				"<p>Country</p>" +
				"<p>Founding Year</p>"

			// Calculate how many cars to show (max 4)

			if numCars > 4 {
				numCars = 4
			}

			// Create a slice to store car results in order
			carResults := make([]string, numCars)
			carChan := make(chan struct {
				index   int
				content string
			}, numCars)

			// Launch goroutines for each car
			for i := 0; i < numCars; i++ {
				go func(idx int, car carModel) {
					var carContent string
					var carCat string
					for _, cate := range categories {
						if car.CategoryID == cate.ID {
							carCat = cate.Name
						}
					}
					carContent = "<a href=\"/car?id=" + strconv.Itoa(car.ID) + "&show=true\">" +
						"<div class=\"image-box\" style=\"background-image: url(" + car.Image + ");\">" +
						"<button class=\"favorite-btn\" onclick=\"toggleFavorite(" + strconv.Itoa(car.ID) + ")\" data-car-id=\"" + strconv.Itoa(car.ID) + "\" aria-label=\"Add " + car.Name + " to favorites\">" +
						"<svg class=\"star-icon\" width=\"24\" height=\"24\" viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\">" +
						"<path d=\"M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z\"></path>" +
						"</svg>" +
						"</button>" +
						"</div>" +
						"</a>" +
						"<h3>" + car.Name + "</h3>" +
						"<p>" + carCat + "</p>" +
						"<p>" + strconv.Itoa(car.Year) + "</p>" +
						"<p>" + car.Specifications.Engine + "</p>" +
						"<p>" + strconv.Itoa(car.Specifications.Horsepower) + " hp</p>" +
						"<p>" + car.Specifications.Transmission + "</p>" +
						"<p>" + car.Specifications.Drivetrain + "</p>"
					for _, manu := range manufacturers {
						if car.ManufacturerID == manu.ID {
							carContent += "<p>" + manu.Name + "</p>" +
								"<p>" + manu.Country + "</p>" +
								"<p>" + strconv.Itoa(manu.FoundingYear) + "</p>"
						}
					}
					carChan <- struct {
						index   int
						content string
					}{
						index:   idx,
						content: carContent,
					}
				}(i, carsToCompare[i])
			}

			// Collect results in order
			for i := 0; i < numCars; i++ {
				carResult := <-carChan
				carResults[carResult.index] = carResult.content
			}

			// Combine results in order
			for _, carResult := range carResults {
				result += carResult
			}

			result += "</main>"
		}
	} else {
		errorMessage = "No cars selected to compare"
		result += "<main class=\"error\">" +
			"<h1>" + errorMessage + "</h1>" +
			"<img src=\"./static/error.png\">" +
			"</main>"
	}
	result += page[2]
	return result
}

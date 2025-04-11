# Cars Information Website

## Overview

This project is a user-friendly website that fetches and displays car model information from a provided Cars API. Users can search, filter, compare models, and receive personalized recommendations based on their interactions.

## Features

- **Fetch & Display Data:** Retrieves car model, manufacturer, and specifications from the API and presents them in tables or cards.
- **Interactive UI:** Users can click on car models to fetch additional details dynamically.
- **Advanced Filtering:** Users can filter cars by Name, Category, Manufacturer, Year, Engine Type, Horsepower, Transmission, and Drivetrain.
- **Favorite cars** Users can click on stars that appear in the model windows and favorite them so they appear first in the sorting
- **Comparisons:** Allows side-by-side comparison of different car models.
- **Easter Eggs:** Searching for "Felicia" or "Peugeot" triggers hidden surprises.

## Installation & Setup

### Updating Go with Homebrew (Linux)

If you have Go installed via Homebrew, update it using the following commands:

1. Update Homebrew:
   ```sh
   brew update
   ```
2. Upgrade Go:
   ```sh
   brew upgrade go
   ```
3. Verify the installation:
   ```sh
   go version
   ```

### Cloning the Repository

To get started, clone the repository using the following command:

```sh
git clone https://gitea.koodsisu.fi/petrkubec/cars.git
```

Then navigate into the project directory:

```sh
cd cars
```

### Running the API

1. Navigate to the `./api/` folder.
2. Follow the instructions in the `README` inside `./api/` to start the API.

### Running the Project

1. Open a terminal and navigate to the project root.
2. Run the command:
   ```sh
   go run .
   ```
3. Open a web browser and navigate to the provided local server address.

## API Endpoints

The project communicates with a Cars API that provides car data. The API is required to be running in a separate terminal for the project to function correctly.

## Usage

- Use the search page to filter cars based on various attributes.
- Click on a car model to fetch and display additional details.
- Compare multiple cars side-by-side.
- Click on stars in model windows to favorite that certain model and make it appear first.
- Try searching for "Felicia" or "Peugeot" to discover hidden Easter eggs!

This project is open-source and licensed under MIT.

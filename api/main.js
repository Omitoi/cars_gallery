const express = require("express");
const expressStatic = require("express-static");
const data = require("./data.json");


const app = express();
const PORT = process.env.PORT || 3000;


// Middleware to parse JSON requests
app.use(express.json());


app.get("/api", (req, res) => {
  res.json({
    models: "/api/models",
    categories: "/api/categories",
    manufacturers: "/api/manufacturers",
  });
});


// Static files (Used to serve images)
app.use("/api/images", expressStatic("img"));


const carModels = data.carModels
, categories = data.categories
, manufacturers = data.manufacturers;


// Car Models Handler
app.get("/api/models", (req, res) => {
  res.json(carModels);
});

app.get("/api/models/:id", (req, res) => {
  const id = parseInt(req.params.id);
  const model = carModels.find((model) => model.id === id);

  if (!model) {
    return res.status(404).json({ message: "Car model not found" });
  }

  res.json(model);
});


// Categories Handler
app.get("/api/categories", (req, res) => {
  res.json(categories);
});

app.get("/api/categories/:id", (req, res) => {
  const id = parseInt(req.params.id);
  const category = categories.find((category) => category.id === id);

  if (!category) {
    return res.status(404).json({ message: "Category not found" });
  }

  res.json(category);
});


// Manufacturers Handler
app.get("/api/manufacturers", (req, res) => {
  res.json(manufacturers);
});

app.get("/api/manufacturers/:id", (req, res) => {
  const id = parseInt(req.params.id);
  const manufacturer = manufacturers.find(
    (manufacturer) => manufacturer.id === id
  );

  if (!manufacturer) {
    return res.status(404).json({ message: "Manufacturer not found" });
  }

  res.json(manufacturer);
});


// Serve
app.listen(PORT, () => {
  console.log(`Server is running on http://localhost:${PORT}`);
});

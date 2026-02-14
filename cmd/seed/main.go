// Package main provides a CLI tool to seed the database with sample products.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/one-backend-go/internal/config"
	"github.com/one-backend-go/internal/db"
	"github.com/one-backend-go/internal/domain/product"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	mongoDB, err := db.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		slog.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer func() { _ = db.Disconnect(ctx, mongoDB) }()

	if err := db.EnsureIndexes(ctx, mongoDB); err != nil {
		slog.Error("failed to create indexes", "error", err)
		os.Exit(1)
	}

	repo := product.NewRepository(mongoDB)

	products := []product.Product{
		{Name: "Classic Cheeseburger", Description: "Juicy beef patty with cheddar cheese, lettuce, tomato, and pickles", PriceCents: 999, Category: "burgers", ImageURL: "https://example.com/img/cheeseburger.jpg", IsAvailable: true},
		{Name: "Margherita Pizza", Description: "Traditional pizza with fresh mozzarella, tomato sauce, and basil", PriceCents: 1299, Category: "pizza", ImageURL: "https://example.com/img/margherita.jpg", IsAvailable: true},
		{Name: "Caesar Salad", Description: "Romaine lettuce with parmesan, croutons, and Caesar dressing", PriceCents: 799, Category: "salads", ImageURL: "https://example.com/img/caesar.jpg", IsAvailable: true},
		{Name: "Chicken Tacos", Description: "Three soft corn tortillas with grilled chicken, salsa, and guacamole", PriceCents: 1099, Category: "mexican", ImageURL: "https://example.com/img/tacos.jpg", IsAvailable: true},
		{Name: "Spaghetti Carbonara", Description: "Classic Italian pasta with pancetta, egg, and pecorino cheese", PriceCents: 1399, Category: "pasta", ImageURL: "https://example.com/img/carbonara.jpg", IsAvailable: true},
		{Name: "Fish and Chips", Description: "Beer-battered cod with crispy fries and tartar sauce", PriceCents: 1199, Category: "seafood", ImageURL: "https://example.com/img/fishnchips.jpg", IsAvailable: true},
		{Name: "Veggie Wrap", Description: "Grilled vegetables with hummus in a whole wheat tortilla", PriceCents: 849, Category: "healthy", ImageURL: "https://example.com/img/veggiewrap.jpg", IsAvailable: true},
		{Name: "Chocolate Brownie", Description: "Rich dark chocolate brownie served with vanilla ice cream", PriceCents: 599, Category: "desserts", ImageURL: "https://example.com/img/brownie.jpg", IsAvailable: true},
		{Name: "Mango Smoothie", Description: "Fresh mango blended with yogurt and honey", PriceCents: 499, Category: "drinks", ImageURL: "https://example.com/img/mango-smoothie.jpg", IsAvailable: true},
		{Name: "BBQ Chicken Wings", Description: "Crispy chicken wings tossed in smoky BBQ sauce", PriceCents: 999, Category: "appetizers", ImageURL: "https://example.com/img/wings.jpg", IsAvailable: true},
	}

	if err := repo.InsertMany(ctx, products); err != nil {
		slog.Error("failed to seed products", "error", err)
		os.Exit(1)
	}

	slog.Info("successfully seeded 10 products")
}

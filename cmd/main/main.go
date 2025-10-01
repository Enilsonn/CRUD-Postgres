package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Enilsonn/CRUD-Postgres/cmd/configs"
	"github.com/Enilsonn/CRUD-Postgres/database"
	"github.com/Enilsonn/CRUD-Postgres/database/migrations"
	"github.com/Enilsonn/CRUD-Postgres/internal/controller"
	"github.com/Enilsonn/CRUD-Postgres/internal/repository"
	"github.com/Enilsonn/CRUD-Postgres/internal/service"
	"github.com/Enilsonn/CRUD-Postgres/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	err := configs.Load("./cmd/main")
	if err != nil {
		panic(err)
	}

	conn, err := database.OpenConnection()
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close()

	err = migrations.Up(conn)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	clientRepository := repository.NewClientRepository(conn)
	productRepository := repository.NewProductRepository(conn)
	walletRepository := repository.NewWalletRepository(conn)
	sellerRepository := repository.NewSellerRepository(conn)
	orderRepository := repository.NewOrderRepository(conn)
	reportRepository := repository.NewReportRepository(conn)

	planService := service.NewPlanService(productRepository)
	orderService := service.NewOrderService(orderRepository, clientRepository, sellerRepository, productRepository, walletRepository)
	reportService := service.NewReportService(reportRepository)

	clientHandler := controller.NewClientHandler(clientRepository)
	productHandler := controller.NewProductHandler(productRepository, planService)
	walletHandler := controller.NewWalletHandler(walletRepository, productRepository)
	orderHandler := controller.NewOrderHandler(orderService)
	reportHandler := controller.NewReportHandler(reportService)
	chatHandler := controller.NewChatHandler(walletRepository)

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.Logger)
	r.Use(utils.JsonMiddleware)

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/clients", func(r chi.Router) {
		r.Post("/", clientHandler.CreateClient)
		r.Get("/", clientHandler.GetAllClients)
		r.Get("/{id}", clientHandler.GetClientByID)
		r.Get("/name/{name}", clientHandler.GetClientByName)
		r.Put("/{id}", clientHandler.UpdateClients)
		r.Delete("/{id}", clientHandler.DeleteClient)
	})

	r.Route("/api/plans", func(r chi.Router) {
		r.Post("/", productHandler.CreateClientProduct)
		r.Get("/", productHandler.GetAllClientProduct)
		r.Get("/search", productHandler.SearchPlans)
		r.With(utils.RequireEmployee).Get("/low-stock", productHandler.LowStock)
		r.Get("/name/{name}", productHandler.GetClientProductByName)
		r.Get("/{id}", productHandler.GetProductByID)
		r.Put("/{id}", productHandler.UpdateClientProduct)
		r.Delete("/{id}", productHandler.DeleteClientProduct)
	})

	r.Route("/api/orders", func(r chi.Router) {
		r.Post("/", orderHandler.CreateOrder)
		r.Post("/{id}/finalize", orderHandler.FinalizeOrder)
	})

	r.Get("/api/clients/{id}/orders", orderHandler.ListClientOrders)

	r.Route("/api/wallets", func(r chi.Router) {
		r.Get("/{client_id}", walletHandler.GetWalletBalance)
		r.Get("/{client_id}/ledger", walletHandler.GetLedgerEntries)
		r.Post("/{client_id}/topups", walletHandler.TopUpCredits)
	})

	r.Post("/api/usage", walletHandler.ProcessUsage)
	r.Post("/api/chat/ollama", chatHandler.ChatOllama)

	r.With(utils.RequireEmployee).Get("/api/reports/sales/monthly", reportHandler.SellerMonthlySales)

	// Serve the admin dashboard
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/", http.StatusFound)
	})
	r.Get("/ui", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "./ui/index.html")
	})
	r.Get("/ui/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "./ui/index.html")
	})

	serverPort := configs.GetServerPort()
	log.Printf("Server starting on port %s...", serverPort)
	log.Printf("Health check endpoint: http://localhost:%s/api/health", serverPort)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", serverPort), r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

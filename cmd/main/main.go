package main

import (
	"log"

	"github.com/Enilsonn/CRUD-Postgres/database"
	"github.com/Enilsonn/CRUD-Postgres/internal/controller"
	"github.com/Enilsonn/CRUD-Postgres/internal/repository"
	"github.com/Enilsonn/CRUD-Postgres/internal/utils"
	"github.com/go-chi/chi/v5"
)

func main() {
	conn, err := database.OpenConecction()
	if err != nil {
		log.Fatalf("No possible to connect with database: %v", err)
	}
	defer conn.Close()

	clientRepository := repository.NewClientRepository(conn)
	productRepository := repository.NewProductRepository(conn)

	clientHandler := controller.NewClientHandler(clientRepository)
	productHandler := controller.NewProductHandler(productRepository)

	r := chi.NewRouter()
	r.Use(utils.JsonMiddleare)

	r.Route("/client", func(r chi.Router) {
		r.Post("/", clientHandler.CreateClient)
		r.Get("/", clientHandler.GetAllClients)
		r.Get("/{id}", clientHandler.GetClientByID)
		r.Get("/name", clientHandler.GetClientByName)
		r.Put("/{id}", clientHandler.UpdateClients)
		r.Put("/{id}", clientHandler.DeleteClient)
	})

	r.Route("/product", func(r chi.Router) {
		r.Post("/", productHandler.CreateClientProduct)
		r.Get("/", productHandler.GetAllClientProduct)
		r.Get("/{id}", productHandler.GetProductByID)
		r.Get("/name", productHandler.GetClientProductByName)
		r.Put("/{id}", productHandler.UpdateClientProduct)
		r.Put("/{id}", productHandler.DeleteClientProduct)
	})
}

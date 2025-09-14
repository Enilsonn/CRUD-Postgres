package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Enilsonn/CRUD-Postgres/cmd/configs"
	"github.com/Enilsonn/CRUD-Postgres/database"
	"github.com/Enilsonn/CRUD-Postgres/internal/controller"
	"github.com/Enilsonn/CRUD-Postgres/internal/repository"
	"github.com/Enilsonn/CRUD-Postgres/internal/utils"
	"github.com/go-chi/chi/v5"
)

func main() {
	err := configs.Load(".")
	if err != nil {
		panic(err)
	}

	conn, err := database.OpenConecction(configs.GetDB())
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

	if err := http.ListenAndServe(fmt.Sprintf(":%s", configs.GetDB().Port), r); err != nil {
		log.Fatalf("Não foi possível iniciar o servidor: %v", err)
	}
}

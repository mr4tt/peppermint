package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	// add postgres info here so the functions can access it
}

func (b Handler) SaveUserInfo(w http.ResponseWriter, r *http.Request)    {}
func (b Handler) SaveNewCategory(w http.ResponseWriter, r *http.Request) {}

func (b Handler) AddTransaction(w http.ResponseWriter, r *http.Request)  {}
func (b Handler) EditTransaction(w http.ResponseWriter, r *http.Request) {}

func (b Handler) GetRemainingMoney(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Write([]byte(id))
}
func (b Handler) GetNewTransactionsFromTeller(w http.ResponseWriter, r *http.Request) {}
func (b Handler) GetTransactionsFromDB(w http.ResponseWriter, r *http.Request)        {}

func (b Handler) GetCategories(w http.ResponseWriter, r *http.Request) {}

func (b Handler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {}
func (b Handler) DeleteCategory(w http.ResponseWriter, r *http.Request)    {}

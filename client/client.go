package main

import (
	"context"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 300 * time.Millisecond)
	defer cancel()

	url := "http://localhost:8080/cotacao"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Panicf("Erro ao fazer a requisição HTTP: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Panicf("Erro ao fazer a requisição HTTP: %v", err)
	}
	defer resp.Body.Close()

	type Cotacao struct {
		Valor float64 `json:"bid"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panicf("Erro ao ler o corpo da resposta: %v", err)
	}
	var valor Cotacao

	err = json.Unmarshal(body, &valor)
	if err != nil {
		log.Panicf("Erro ao decodificar o JSON: %v", err)
	}

	arquivo, err := os.OpenFile("cotacao.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Panicf("Erro ao criar o arquivo: %v", err)
	}
	defer arquivo.Close()

	tmp := template.New("Cliente")
	tmp, _ = tmp.Parse("Dólar: {{.Valor}} \n")
	tmp.Execute(arquivo, valor)

}

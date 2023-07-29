package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Usdbrl struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type Cotacao struct {
	gorm.Model
	Valor float64    `json:"bid"`
}

func main() {
	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		cotacao, err := CotacaoDiaria()
		if err != nil {
			log.Println(err)
			if strings.Contains(err.Error(), "context deadline exceeded") {
				w.WriteHeader(http.StatusRequestTimeout)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
			return
 		} 
		err = PersistirLocal(cotacao)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		type Retorno struct {
			Bid float64 `json:"bid"`
		}
		bid, _ := strconv.ParseFloat(cotacao.Usdbrl.Bid, 64)
		retorno := Retorno{ Bid: bid }
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jsonData,_ := json.Marshal(retorno)
		w.Write(jsonData)	
	})
	http.ListenAndServe(":8080",nil)
}

func PersistirLocal(cotacao Usdbrl) error {
	ctx := context.Background()
	ctx,cancel := context.WithTimeout(ctx, 10 * time.Millisecond)
	defer cancel()

	db, err := gorm.Open(sqlite.Open("server.db"), &gorm.Config{})
	if err != nil {		
		return errors.New(fmt.Sprintf("Erro ao abrir o banco de dados: %v", err))
	}

	// Executar migrações para criar a tabela (opcional)
	err = db.AutoMigrate(&Cotacao{})
	if err != nil {
		return errors.New(fmt.Sprintf("Erro ao executar as migrações: %v", err))
	}

	bid, _ := strconv.ParseFloat(cotacao.Usdbrl.Bid, 64)

	var c Cotacao = Cotacao{ Valor: bid}
	
	tx := db.WithContext(ctx).Create(&c)
	if tx.Error != nil {
		return errors.New(fmt.Sprintf("Erro ao persistir: %v", err))
	}

	return nil

}

func CotacaoDiaria() (Usdbrl, error) {
	ctx := context.Background() 
	ctx,cancel := context.WithTimeout(ctx, 200 * time.Millisecond)
	defer cancel()
	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {		
		return Usdbrl{}, err
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return Usdbrl{}, ctx.Err()			
		}
		return Usdbrl{}, err		
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Usdbrl{}, err
	}

	var resposta Usdbrl
	err = json.Unmarshal(body, &resposta)
	if err != nil {		
		return Usdbrl{}, err
	}

	return resposta, nil
}
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type CepBrasilApiResponse struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

type CepViaCepResponse struct {
	Cep          string `json:"cep"`
	Street       string `json:"logradouro"`
	Complement   string `json:"complemento"`
	Unidade      string `json:"unidade"`
	Neighborhood string `json:"bairro"`
	City         string `json:"localidade"`
	State        string `json:"uf"`
	Region       string `json:"regiao"`
	Ibge         string `json:"ibge"`
	Gia          string `json:"gia"`
	Ddd          string `json:"ddd"`
	Siafi        string `json:"siafi"`
}

func main() {
	r := chi.NewRouter()

	r.Get("/cep/{cep}", HandlerCep)

	fmt.Println("Server is running on port 8080")

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}

func HandlerCep(w http.ResponseWriter, r *http.Request) {
	cep := chi.URLParam(r, "cep")

	if cep == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c1 := make(chan *CepBrasilApiResponse)
	c2 := make(chan *CepViaCepResponse)

	go func() {
		cepBrasilApiResponse, error := GetCepBrasilApi(cep)
		if error != nil {
			fmt.Println("Cep não encontrado no Brasil API")
		}
		c1 <- cepBrasilApiResponse
	}()

	go func() {
		cepViaCepResponse, error := GetCepViaCep(cep)
		if error != nil {
			fmt.Println("Cep não encontrado no Via Cep API")
		}
		c2 <- cepViaCepResponse
	}()

	select {
	case cepBrasilApiResponse := <-c1:
		fmt.Printf("Cep encontrado no Brasil API: %s\n", cepBrasilApiResponse.Cep)
	case cepViaCepResponse := <-c2:
		fmt.Printf("Cep encontrado no Via Cep API: %s\n", cepViaCepResponse.Cep)
	case <-time.After(time.Second):
		fmt.Println("Tempo de espera excedido")
	}
}

func GetCepBrasilApi(cep string) (*CepBrasilApiResponse, error) {
	req, err := http.NewRequest("GET", "https://brasilapi.com.br/api/cep/v1/"+cep, nil)
	if err != nil {
		return nil, err
	}

	resp, error := http.DefaultClient.Do(req)
	if error != nil {
		return nil, error
	}
	defer resp.Body.Close()

	body, error := io.ReadAll(resp.Body)
	if error != nil {
		return nil, error
	}

	var cepBrasilApiResponse CepBrasilApiResponse
	error = json.Unmarshal(body, &cepBrasilApiResponse)
	if error != nil {
		return nil, error
	}
	return &cepBrasilApiResponse, nil
}

func GetCepViaCep(cep string) (*CepViaCepResponse, error) {
	req, err := http.NewRequest("GET", "http://viacep.com.br/ws/"+cep+"/json/", nil)
	if err != nil {
		return nil, err
	}

	resp, error := http.DefaultClient.Do(req)
	if error != nil {
		return nil, error
	}
	defer resp.Body.Close()

	body, error := io.ReadAll(resp.Body)
	if error != nil {
		return nil, error
	}

	var cepViaCepResponse CepViaCepResponse
	error = json.Unmarshal(body, &cepViaCepResponse)
	if error != nil {
		return nil, error
	}
	return &cepViaCepResponse, nil
}

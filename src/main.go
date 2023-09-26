package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type ViaCEP struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
}

type ApiCEP struct {
	Code     string `json:"code"`
	State    string `json:"state"`
	City     string `json:"city"`
	District string `json:"district"`
	Address  string `json:"address"`
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		panic("CEP deve ser informado")
	}
	cep := getCepFormatado(args[0])

	c1 := make(chan ViaCEP)
	c2 := make(chan ApiCEP)

	go func() {
		for {
			cepResult := buscaCEPFromViaCEP(cep)
			c1 <- cepResult
		}
	}()

	go func() {
		for {
			cepResult := buscaCEPFromApiCEP(cep)
			c2 <- cepResult
		}
	}()

	select {
	case msg := <-c1:
		fmt.Printf("CEP Recebido com sucesso! Fonte: Via CEP. Dados: %v", msg)
		return
	case msg := <-c2:
		fmt.Printf("CEP Recebido com sucesso! Fonte: API CEP. Dados: %v", msg)
		return
	case <-time.After(time.Second * 1):
		println("Erro: timeout!")
		return
	}

}

func getCepFormatado(cep string) string {
	if !strings.Contains(cep, "-") {
		cep = cep[:5] + "-" + cep[5:]
	}
	return cep
}

func buscaCEPFromViaCEP(cep string) ViaCEP {
	url := "http://viacep.com.br/ws/" + cep + "/json/"
	res := buscaCEP(url, "Via CEP")
	var data ViaCEP
	err := json.Unmarshal(res, &data)
	if err != nil {
		panic(err)
	}
	return data
}

func buscaCEPFromApiCEP(cep string) ApiCEP {
	url := "https://cdn.apicep.com/file/apicep/" + cep + ".json"
	res := buscaCEP(url, "API CEP")
	var data ApiCEP
	err := json.Unmarshal(res, &data)
	if err != nil {
		panic(err)
	}
	return data
}

func buscaCEP(url, source string) []byte {
	req, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	if req.StatusCode != http.StatusOK {
		panic("Erro ao fazer requisição para " + source + ": status code diferente de 200: " + strconv.Itoa(req.StatusCode))
	}
	defer req.Body.Close()
	res, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	return res
}

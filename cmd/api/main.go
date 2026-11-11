// Command api é o ponto de entrada da API HTTP do e-commerce.
//
// Nesta fase (configuração base) ele apenas valida que o projeto compila,
// roda e encerra. O servidor HTTP com Chi, a configuração por variáveis de
// ambiente, o PostgreSQL e o graceful shutdown entram na etapa 3 (fundação do
// backend).
package main

import (
	"fmt"
	"io"
	"os"
)

const serviceName = "ecommerce-backend"

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", serviceName, err)
		os.Exit(1)
	}
}

// run concentra a lógica do main e devolve um erro em vez de encerrar o
// processo. Isso permite testá-la sem chamar os.Exit.
func run(out io.Writer) error {
	_, err := fmt.Fprintf(out, "%s: api inicializada\n", serviceName)
	return err
}

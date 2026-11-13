// Command api é o ponto de entrada da API HTTP do e-commerce.
//
// Nesta fase ele conecta ao PostgreSQL (DATABASE_URL), confirma que o banco
// responde e encerra. O servidor HTTP com Chi, o pacote de configuração e o
// graceful shutdown entram nas próximas tarefas da fundação do backend.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/pedrohrodrig/ecommerce-backend/internal/platform/database"
)

const serviceName = "ecommerce-backend"

func main() {
	if err := run(context.Background(), os.Stdout, os.Getenv("DATABASE_URL")); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", serviceName, err)
		os.Exit(1)
	}
}

// run concentra a lógica do main e devolve um erro em vez de encerrar o
// processo. Isso permite testá-la sem chamar os.Exit.
func run(ctx context.Context, out io.Writer, databaseURL string) error {
	if databaseURL == "" {
		return errors.New("DATABASE_URL não definida")
	}

	pool, err := database.Open(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	_, err = fmt.Fprintf(out, "%s: api inicializada\n", serviceName)
	return err
}

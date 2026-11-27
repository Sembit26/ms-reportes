package e2e_test

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func puertoLibre(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("no se pudo reservar puerto libre: %v", err)
	}
	defer func() { _ = listener.Close() }()
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("no se pudo obtener el puerto: %v", err)
	}
	return port
}

func obtenerBinario(t *testing.T) string {
	t.Helper()
	if binario := os.Getenv("API_BINARY"); binario != "" {
		return binario
	}
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	binario := filepath.Join(t.TempDir(), "ms-reportes-test"+ext)
	compilar := exec.Command("go", "build", "-o", binario, "../cmd/api")
	if salida, err := compilar.CombinedOutput(); err != nil {
		t.Fatalf("error al compilar binario E2E: %v\n%s", err, salida)
	}
	return binario
}

func TestBinarioSaludYDocs(t *testing.T) {
	binario := obtenerBinario(t)
	puerto := puertoLibre(t)
	baseURL := fmt.Sprintf("http://127.0.0.1:%s", puerto)

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://reportes:clave-prueba@127.0.0.1:1/reportes?sslmode=disable"
	}

	cmd := exec.Command(binario)
	cmd.Env = append(os.Environ(),
		"API_PORT="+puerto,
		"API_ENV=test",
		"DATABASE_URL="+dbURL,
		"FRONTEND_URL=http://localhost:5110",
		"TZ=America/Santiago",
	)
	if err := cmd.Start(); err != nil {
		t.Fatalf("no se pudo iniciar el binario de la API: %v", err)
	}

	defer func() {
		if cmd.Process == nil {
			return
		}
		if runtime.GOOS == "windows" {
			_ = cmd.Process.Kill()
		} else {
			_ = cmd.Process.Signal(syscall.SIGTERM)
		}
		_ = cmd.Wait()
	}()

	cliente := &http.Client{Timeout: 2 * time.Second}
	var listo bool
	for range 30 {
		resp, err := cliente.Get(baseURL + "/healthz")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				listo = true
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !listo {
		t.Fatalf("el binario no respondió 200 en %s/healthz", baseURL)
	}

	respDocs, err := cliente.Get(baseURL + "/api/docs")
	if err != nil {
		t.Fatalf("error consultando /api/docs: %v", err)
	}
	_ = respDocs.Body.Close()
	if respDocs.StatusCode != http.StatusOK {
		t.Fatalf("/api/docs devolvió %d, se esperaba 200", respDocs.StatusCode)
	}

	respReady, err := cliente.Get(baseURL + "/readyz")
	if err != nil {
		t.Fatalf("error consultando /readyz: %v", err)
	}
	_ = respReady.Body.Close()
	if os.Getenv("TEST_DATABASE_URL") != "" && respReady.StatusCode != http.StatusOK {
		t.Fatalf("/readyz con TEST_DATABASE_URL devolvió %d, se esperaba 200", respReady.StatusCode)
	}
	if os.Getenv("TEST_DATABASE_URL") == "" && respReady.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("/readyz sin PostgreSQL devolvió %d, se esperaba 503", respReady.StatusCode)
	}
}

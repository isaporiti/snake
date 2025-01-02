build:
	env GOOS=js GOARCH=wasm go build -o snake.wasm github.com/isaporiti/snake

serve:
	go run github.com/hajimehoshi/wasmserve@latest .
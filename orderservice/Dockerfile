Dockerfile\nFROM golang:1.20-alpine\nWORKDIR /app\nCOPY . .\nRUN go mod tidy\nRUN go build -o main\nCMD ["./main"]

FROM golang:1.24.6

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

# 安裝 air（hot reload 工具）
RUN go install github.com/air-verse/air@v1.61.7

# 預設啟動 air
CMD ["air"]
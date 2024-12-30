# Sử dụng image chính thức của Golang
FROM golang:1.21.1

# Thiết lập thư mục làm việc
WORKDIR /app

# Sao chép tệp `go.mod` và `go.sum` trước
COPY go.mod go.sum ./

# Tải các thư viện phụ thuộc
RUN go mod download

# Sao chép mã nguồn dự án
COPY . .

# Biên dịch ứng dụng
RUN go build -o main .

# Chạy ứng dụng
CMD ["./main"]

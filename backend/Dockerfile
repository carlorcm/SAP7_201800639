# Usa una imagen oficial de Go como base
FROM golang:1.16

# Crea un directorio de trabajo dentro del contenedor
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# Copia el código de tu proyecto dentro del directorio de trabajo
COPY main.go /app

# Compila el código de tu proyecto
RUN go build -o app

# Expone el puerto 8080
EXPOSE 8080

# Especifica el comando para iniciar el servicio
CMD ["./app"]

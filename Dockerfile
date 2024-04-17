# syntax=docker/dockerfile:1
# STAGE 1 - BUILDING
FROM golang:1.22-alpine AS build

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o fleetman/ ./...

# STEP 2 - DEPLOY
FROM scratch AS run

WORKDIR /

COPY --from=build /app/fleetman/ fleetman/

EXPOSE 8000

ENTRYPOINT [ "/fleetman/fleetman" ]
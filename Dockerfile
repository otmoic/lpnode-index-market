FROM docker.io/library/golang:1.19.3-buster
ADD ./ /data/lp_market/
EXPOSE 18080
WORKDIR /data/lp_market/
RUN go build -o lp_market cmd/main.go 
ENV GO_ENV=production
ENV SERVICE_PORT=18080

CMD ["./lp_market" ]


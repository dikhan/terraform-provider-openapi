FROM golang:latest

ENV WORK_DIR "$GOPATH/src/github.com/dikhan/terraform-provider-openapi/examples/swaggercodegen/api"

RUN mkdir -p $WORK_DIR
ADD . $WORK_DIR
WORKDIR $WORK_DIR

RUN go build -o cdn-service-provider-api .

EXPOSE 80
EXPOSE 443

CMD ["./cdn-service-provider-api"]
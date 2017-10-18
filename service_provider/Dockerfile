FROM golang:latest

ENV WORK_DIR "$GOPATH/src/github.com/dikhan/terraform-provider-api/service_provider"

RUN mkdir -p $WORK_DIR
ADD . $WORK_DIR
WORKDIR $WORK_DIR

RUN go get
RUN go build -o service-provider-api .

EXPOSE 8080

CMD ["./service-provider-api"]
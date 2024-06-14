FROM heroiclabs/nakama-pluginbuilder:3.22.0 AS go-builder

ENV GO111MODULE on
ENV CGO_ENABLED 1

WORKDIR /backend

COPY go.mod .
COPY main.go .

# Download all dependencies. Dependencies will be cached if the go.mod file is not changed
RUN go get nakama-rpc
RUN go mod download

# Create vendor directory
RUN go mod vendor

RUN go build --trimpath --mod=vendor --buildmode=plugin -o ./backend.so

FROM registry.heroiclabs.com/heroiclabs/nakama:3.22.0

COPY --from=go-builder /backend/backend.so /nakama/data/modules/

# Copy the sample files into the appropriate directory
COPY sample_files /nakama/data/sample_files

FROM golang:1.22

# Set destination for COPY
WORKDIR /pb

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY *.go ./


# Build
RUN CGO_ENABLED=0 go build -o /pb/pocketbase

# Figure out how to get data to persist with volumes

# uncomment to copy the local pb_migrations dir into the image
# COPY ./pb_data /pb/pb_data

# auth pages
COPY ./pb_public /pb/pb_public

EXPOSE 8090

# start PocketBase
CMD ["/pb/pocketbase", "serve", "--http=0.0.0.0:8090" ]
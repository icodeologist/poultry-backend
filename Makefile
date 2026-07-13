BINARY_NAME=avin-shop
PORT=3000
PHONE_NUMBER=6666666666

.PHONY: build run test clean create-customer

build:
	go build -o $(BINARY_NAME) .

run: build
	./$(BINARY_NAME)

dev:
	go run .

clean:
	rm -f $(BINARY_NAME)

create-customer:
	curl -X POST http://localhost:$(PORT)/create \
		-H "Content-Type: application/json" \
		-d '{"name": "Denzil", "phone_number": "6666666666"}'

add-product:
	curl -X POST http://localhost:$(PORT)/add/product \
		-H "Content-Type: application/json" \
		-d '{"title": "Chicken tighs", "price": 176.54, "instock":40}'

search-user:
	curl -X GET http://localhost:$(PORT)/search/6666666666 \
		-H "Content-Type: application/json"

tidy:
	go mod tidy



# Ethereum Exchange Server

This is an Ethereum exchange server built using Go. It allows users to place limit and market orders for trading Ether (ETH). The server integrates with an Ethereum client and handles order placement, matching, and cancellation.


## Features

- **Market Orders:** Users can place market orders to buy or sell Ether at the current market price.

- **Limit Orders:** Users can place limit orders specifying a desired price for buying or selling Ether.

- **Order Matching:** The server matches orders between buyers and sellers to facilitate trades.

- **Orderbook Management:** It maintains orderbooks for different markets to track orders.

- **User Registration:** Users can register with their Ethereum private keys.

- **ETH Transfers:** The server supports ETH transfers between users when orders are matched.

## Installation

To run this project, you need to have Go and an Ethereum client (e.g., geth) installed on your machine.

1. Clone the repository:

   ```bash
   git clone https://github.com/inagib21/crypto-exchange.git
   ```

2. Install dependencies using Go Modules:

   ```bash
   go mod tidy
   ```

3. Configure Ethereum client:
   - Ensure that you have an Ethereum client (e.g., geth) running on `http://localhost:8545`. You can adjust this in the code if necessary.

4. Use the Makefile commands to build, run, and test the server:

   - Build the server:

     ```bash
     make build
     ```

   - Run the server:

     ```bash
     make run
     ```

   - Run tests:

     ```bash
     make test
     ```

The server should be up and running on `http://localhost:3000`.


## Usage

### Registering Users

Users need to register with their Ethereum private keys. This can be done programmatically by calling the `registerUser` function or through a user registration API.

### Placing Orders

Users can place orders using the `/order` API endpoint. They can specify the order type (market or limit), bid or ask, order size, price, and market.

Example of placing a limit order:
```bash
curl -X POST http://localhost:3000/order -d '{
  "UserID": 1,
  "Type": "LIMIT",
  "Bid": true,
  "Size": 1.0,
  "Price": 200.0,
  "Market": "ETH"
}'
```

### Viewing Orders

Users can view their orders and order history using the `/order/:userID` API endpoint.

Example of viewing orders for a user with ID 1:
```bash
curl http://localhost:3000/order/1
```

### Order Matching

The server automatically matches buy and sell orders when conditions are met. The matched orders are then executed.

### Cancelling Orders

Users can cancel their orders using the `/order/:id` API endpoint. Specify the order ID to cancel.

Example of cancelling an order with ID 123:
```bash
curl -X DELETE http://localhost:3000/order/123
```

### Getting Market Data

Users can retrieve market data, including the order book, best bid, best ask, and recent trades using various API endpoints.

## Acknowledgments

Special thanks to [AnthonyGG](https://www.youtube.com/@anthonygg_) for his excellent tutorial on building an Ethereum exchange server in Go. His tutorial was a valuable resource for creating this project.

## Next Steps

Create the front end for this project

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.


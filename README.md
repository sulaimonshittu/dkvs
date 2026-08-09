# Key Value Store
A Raft-enabled Distributed Key Value Store that covers CAP theorem, consistency, and replication models

## Requirements
- Go (v1.20+)

## ⚙️ How to run

1. *Clone the repo:*

```bash
git clone https://github.com/sulaimonshittu/dkvs.git
```

2. *Install dependencies:*

```bash
go mod tidy
```

3. *Run the application*

```bash
$ go run main.go
```

4. *Call the key store service endpoints*

- GET /v1/{key} Gets the value that matches the key
- PUT /v1/{key} Update the value that matches the key with a new value
- DELETE /v1/{key} Delete the value that matches the key

## 🔥 Future Improvements
- Ability for service to join an existing cluster 
- Ability for sevice to persist and restore its state.
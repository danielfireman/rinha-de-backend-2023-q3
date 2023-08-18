# rinha-de-backend-2023-q3
Repo para participar do evento Rinha de Backend 2023 Q3

### Subir ambiente
```sh
docker-compose up --build --force-recreate
```

#### Subir o servidor
```sh
go run .
```

### Operações

#### Adicionar 
```sh
curl -v -X POST -H "Content-Type: application/json" -d "@exemplo_pessoa.json" http://localhost:9999/pessoas
```

#### Consultar
```sh
curl "http://localhost:1323/pessoas/xxxxxxx"
curl "http://localhost:1323/pessoas?t=go"  # Encontra
curl "http://localhost:1323/pessoas?t=node"  # Vazio
curl "http://localhost:1323/pessoas?t="  # Bad request
```

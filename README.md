# 🚀 EventCounter

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)

## 📋 Índice

- [Funcionalidades](#-funcionalidades)
- [Arquitetura](#-arquitetura)
- [Estrutura do Projeto](#-estrutura-do-projeto)
- [Início Rápido](#-início-rápido)
- [Configuração](#-configuração)
- [Testes](#-testes)
- [Padrões de Design](#-padrões-de-design)
- [Contribuindo](#-contribuindo)

## ✨ Funcionalidades

- **Alta Performance**: Processamento concorrente com goroutines e canais
- **Deduplicação de Mensagens**: Processamento idempotente
- **Agregação de Eventos**: Agrupa eventos por ID do usuário e tipo de evento
- **Integração RabbitMQ**: Consumo mensagens com reconexão automática
- **Saída JSON**: Resultados estruturados em arquivos separados por tipo de evento
- **Encerramento Elegante**: Término com contexto e limpeza adequada
- **Testes Abrangentes**: Testes unitários e de integração incluídos


## 📁 Estrutura do Projeto

```
├── cmd/
│   ├── consumer/           # Aplicação principal do consumidor
│   │   ├── main.go         # Ponto de entrada da aplicação
│   │   ├── config/         # Gerenciamento de configuração
│   │   ├── connection/     # Lógica de conexão RabbitMQ
│   │   └── domain/         # Lógica de negócio
│   │       ├── event.go
│   │       ├── event_counter.go    # Lógica principal de contagem
│   │       ├── dispatcher.go       # Roteamento de eventos e workers
│   │       └── event_counter_test.go
│   └── generator/          # Gerador de mensagens de teste
├── pkg/                    # Pacotes compartilhados e interfaces
├── logger/                 # Utilitário simples de logging
├── bin/                    # Binários compilados
└── results/                # Arquivos JSON de saída
    ├── created.json
    ├── updated.json
    └── deleted.json
```


## 🚀 Início Rápido

### Pré-requisitos
- Go 1.23+
- Docker (para RabbitMQ)
- PowerShell (Windows) ou shell compatível

### Desenvolvimento Local

1. **Iniciar ambiente RabbitMQ**
   ```powershell
   make env-up
   ```

2. **Gerar dados de teste** (cria exchange/fila e publica 100 mensagens de teste)
   ```powershell
   make generator-publish
   ```

3. **Executar o consumidor**
   ```powershell
   go run ./cmd/consumer/main.go
   ```

4. **Verificar resultados** (gerados após execução bem-sucedida)
   - `created.json`
   - `updated.json` 
   - `deleted.json`

   Cada arquivo contém contagens agregadas:
   ```json
   [
     { "user_id": "user123", "count": 42 },
     { "user_id": "user456", "count": 10 }
   ]
   ```

5. **Limpeza**
   ```powershell
   make env-down
   ```

### Opções de Configuração
Para modificar as configurações de porta/exchange do RabbitMQ, edite as variáveis no topo do `Makefile`.

## 🧪 Testes

### Testes Unitários
Cobertura abrangente de testes disponível em `cmd/consumer/domain/event_counter_test.go`, cobrindo:
- Lógica de contagem e deduplicação de eventos
- Mecanismos de roteamento
- Tratamento de concorrência
- Validação de persistência
- Cenários de integração

**Executar testes:**
```powershell
go test ./cmd/consumer/domain/
```

**Executar testes com cobertura:**
```powershell
go test -cover ./cmd/consumer/domain/
```

### Testes de Integração
Testes adicionais podem ser adicionados para cobrir:
- Tratamento de timeout e contexto
- Casos extremos de concorrência
- Verificação de persistência de resultados

## 🎯 Configuração

A aplicação pode ser configurada via variáveis de ambiente ou flags de linha de comando. 
As principais opções de configuração são gerenciadas em `cmd/consumer/config/config.go`:

- URL de conexão RabbitMQ
- Nome do exchange
- Configuração da fila
- Configurações de timeout

## 🏛 Padrões de Design

### Princípios Fundamentais
- **Inversão de Dependência**: Favorece inversão de controle para baixo acoplamento
- **Responsabilidade Única**: Cada componente tem um propósito bem definido

### Modelo de Concorrência
- **Canais + Goroutines**: Processamento paralelo por tipo de evento
- **Sincronização**: `sync.WaitGroup` para coordenação de workers
- **Segurança de Thread**: `sync.Mutex` protege estruturas de dados compartilhadas
- **Propagação de Contexto**: `context.Context` para timeouts elegantes (encerramento após 5 segundos de inatividade)

### Integridade de Dados
- **Idempotência**: Mensagens são deduplicadas por ID para prevenir contagem dupla
- **Operações Atômicas**: Acesso seguro e concorrente a contadores e conjuntos de mensagens processadas
- **Saída Estruturada**: Arquivos JSON separados por tipo de evento para fácil consumo

## 📄 Arquivos Principais

| Arquivo | Propósito |
|---------|-----------|
| `cmd/consumer/main.go` | Inicialização da aplicação e configuração de dependências |
| `cmd/consumer/connection/rabbitmq.go` | Conexão RabbitMQ e consumo de mensagens |
| `cmd/consumer/domain/event_counter.go` | Lógica principal de contagem e deduplicação |
| `cmd/consumer/domain/dispatcher.go` | Roteamento de eventos e gerenciamento de workers |
| `pkg/consumer.go` | Contrato da interface Consumer |
| `logger/logger.go` | Utilitário simples de logging |

## 🔧 Dicas de Depuração

1. **Variáveis de Ambiente**: Verifique a configuração em `cmd/consumer/config/config.go`
2. **Logging**: Logs do console estão disponíveis via o logger simples em `logger/logger.go`
3. **Gerenciamento RabbitMQ**: Acesse a interface web em `http://localhost:15672` (guest/guest)
4. **Inspeção de Mensagens**: Use a interface de gerenciamento do RabbitMQ para 
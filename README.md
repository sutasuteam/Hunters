<p align="center">
  <img src="Logo/Hunters.png" width="180" alt="Hunters Logo">
</p>

<h1 align="center">Hunters</h1>

<p align="center">
A decentralized social platform built on the Canopy Network.
</p>

---

# One-line Pitch

**Hunters is a decentralized Social-Fi platform on Canopy where users own their identity, reputation, social assets, and NFTs through on-chain transactions.**

---

# Why Hunters?

Hunters explores how social interactions can become native blockchain transactions.

Instead of storing identity and social activity entirely off-chain, Hunters integrates directly with the Canopy Network to demonstrate a decentralized Social-Fi ecosystem.

Every user receives a Canopy wallet and interacts with the blockchain through custom plugin transactions.

---

# Social-Fi Components

Hunters includes several Social-Fi concepts:

- ✅ Decentralized Identity
- ✅ On-chain User Profiles
- ✅ NFT-powered Social Assets
- ✅ Community Content
- ✅ Social Reputation Foundation
- ✅ User-owned Wallet Identity
- ✅ Token Economy
- ✅ Blockchain-native Transactions

---

# Canopy Integration

Hunters is **not** a standalone web application.

It communicates directly with a local Canopy blockchain.

The application architecture consists of:

```
Frontend
      │
      ▼
Express Backend
      │
      ▼
Custom Go Plugin
      │
      ▼
Canopy RPC
(50002 / 50003)
      │
      ▼
Local Canopy Chain
```

The backend communicates with the Canopy node using RPC while custom transaction types are implemented inside the Go plugin.

---

# Custom Transactions

Hunters extends Canopy through custom transaction types.

Current plugin interactions include:

- Wallet generation
- Token creation
- NFT minting
- Asset transfers
- On-chain profile operations

TransactionTypeUrls: []string{
    "type.googleapis.com/types.MessageSend",
    "type.googleapis.com/types.MessageMintNFT",
    "type.googleapis.com/types.MessageSetStake",
    "type.googleapis.com/types.MessageSendToken",
    "type.googleapis.com/types.MessageLikePost",
},


These transactions are executed through the Go plugin and submitted to the local Canopy chain.

---

# RPC Usage

Hunters interacts directly with the local Canopy RPC.

```
RPC Port : 50002
REST Port: 50003
```

The frontend never uses mocked blockchain data.

Every blockchain action is forwarded through the backend and executed on the local Canopy chain.

---

# Built with Canopy Template

Hunters is built using the official **Canopy Go Template**.

The project extends the template by adding:

- Custom signer service
- Social application backend
- NFT module
- Token module
- Wallet onboarding
- Frontend integration

---

# Running Hunters

## 1. Start Canopy

```bash
./canopy start
```

---

## 2. Enable Go Plugin

Open

```
config/config.json
```

Set

```json
{
  "plugin": "go"
}
```

Restart the node.

---

## 3. Start Hunters Backend

Open another terminal.

```bash
npm install

npm run dev
```

---

## 4. Edit ENV

```
Adjust data validation in ENV
```

---


## 5. Open Hunters

```
http://127.0.0.1:8080/
```

---


# Architecture

```
                 MetaMask
                     │
                     ▼
             Hunters Frontend
                     │
                     ▼
             Express Backend
             (Node.js Server)
                     │
         ┌───────────┴───────────┐
         ▼                       ▼
  Canopy Signer             Local Database
         │                  (users/posts)
         ▼
     Canopy RPC
         │
         ▼
    Canopy Network
```

# Development Flow

```
Start Canopy
        │
        ▼
Enable Go Plugin
        │
        ▼
Run npm run dev
        │
        ▼
Open Frontend
        │
        ▼
MetaMask Login
        │
        ▼
Create Canopy Wallet
        │
        ▼
Interact with Canopy RPC
        │
        ▼
Execute Custom Transactions
```
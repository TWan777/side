# Cross Chain Liquidity

The Cross Chain Liquidity is a CCA app that establishes a cross-chain liquidity pool, enabling users to swap tokens within the pool without the need for transferring tokens between different chains.

## Technical Specification

### Transaction Flow

![flow](./cross_chain_liquidity_workflow.png)

### Data Structure

```ts
interface PoolAsset {
  channelId: string; // native or channel_id
  balance: Coin;
  // percentage: 50 for 50%
  weight: int32;
}
```

```ts
interface WeightedPool {
  creator: string,
  id: string;
  assets: []PoolAsset;
  swapFee: int32;
  // the issued amount of pool token in the pool. the denom is pool id
  supply: Coin;
  status: PoolStatus;
}

interface Deposit {
    sender: string, // local sender
    poolId: string,
    token: Coin[],
    requestId: string,
    status: string,
    createdAt: u64,
    completedAt: u64,
}

interface Withdraw {
    sender: string, // local sender
    poolId: string,
    token: Coin,
    requestId: string,
    status: string,
    createdAt: u64,
    completedAt: u64,
}

interface Trade {
    sender: string, // local sender
    poolId: string,
    requestId: string,
    tokenIn: Coin,
    tokenOut: Coin,
    slippage: u64,
    status: string,
    createdAt: u64,
    completedAt: u64,
}
```

- Deposit: 0x01 | sender | desired_sender | channelId -> ProtocolBuffer(Deposit)

```ts
class DepositEthereumResposne extends DefaultEthereumResponseAdapter<Deposit> {
  verify() {
    if (super.verify()) return false;
    // check ERC20 tokens later
    if (this.txResponse.value !== this.state.token.amount) return false;
    return true;
  }
}
```

### Messages

```proto
message PoolToken {
    string channelId = 1;
    Coin token = 2;
    string desiredSender = 3;
    u32 weight = 4;
}

message MsgCreatePool {
    u32 swapFee = 1;
    repeated PoolToken tokens = 2;
}
```

```proto
message MsgMultiDeposit {
    string desiredSender = 1;
    string poolId = 2;
    repeated Coin tokens = 3;
    string channelId = 4;
}
message MsgSingleDeposit {
    string desired_sender = 1;
    string poolId = 2;
    Coin token = 3;
    string channelId = 4;
}
```

```proto
message MsgWithdraw {
    string recipient = 1;
    string poolId = 2;
    Coin token = 3;
    string channelId = 4;
}
```

```proto
message MsgSwap {
    string sender = 1;
    string recipient = 2;
    string poolId = 3;
    Coin tokenIn = 4;
    Coin tokenOut = 5;
    u64 slippage = 6;
}
```

### MessageHandler

```ts
function handleMsgCreatePool(msg: MsgCreatePool) {

    let poolId = store.getIncrementalPoolId() //
    let poolAssets = []
    let totalWeight = 0;
    let remoteAsset = 0;
    for(t in msg.tokens) {
        if(t.channelId === 'native') {
            let escrowedAddress = getEscrowedAccount(`${AppName}/${poolId}`)
            bank.sendTokenToAccount(msg.sender, escrowedAddress, t.token)

            poolAssets.push({
                channelId: t.channelId;  // native or channel_id
                balance: t.token;
                weight: t.weight;
            })
        } else {
            // request remote deposit
            const request: IntentRequest  = {
                sender: msg.sender,
                channelId: t.channelId,
                action: "CreatePool",
                expectedSender: t.desired_sender, // the expected sender of inboundTx on
                expectedReceivedToken: t.token,
                hash: "",
                referenceId: poolId,
                status: "INITIATED",
                inboundTx: [],
                createAt: block.timestamp,
            }

            store.registerInboundSigningRequest(request)
            remoteAsset++;

            poolAssets.push({
                channelId: t.channelId;  // native or channel_id
                balance: new Coin(0, t.token.denom);
                weight: t.weight;
            })
        }

        totalWeight += t.weight;
    }
    abortTransactionUnless(totalWeight === 100);

    let supplyAmount = 0
    if(remoteAssets === 0 ) {
        supplyAmount = calculateInitialSupply(msg.tokens);
    }
    let supplyToken = new Coin(supplyAmount, supplyToken);
    if(supplyAmount > 0) {
        // if tokens are native, pool will created in the handler.
        // otherwise, pool will created when the remoted deposited is finalised.
        bank.mint(ModuleName, supplyToken);
        bank.sendTokenFromModuleToAccount(ModuleName, msg.sender, supplyToken);
    }

    let newPool = {
        id: poolId;
        assets: poolAssets;
        swapFee: msg.swapFee;
        supply: supplyToken;
        status: supplyAmount==0? "Initial": "Ready";
    }

    store.save(newPool)

}
function handleMsgSingleDeposit(msg: MsgSingleDeposit) {

    // process request
    if(isNativeToken(msg.token)) {
        // lock assets on escrowed account
        let pool = getPool(msg.poolId)
        let escrowedAddress = getEscrowedAccount(`${AppName}/${msg.poolId}`)
        bank.sendTokenToAccount(msg.sender, escrowedAddress, t.token)
        let supplyAmount = calculateInitialSupply(msg.tokens);
        let supplyToken = new Coin(supplyAmount, pool.supply.denom);
        if(supplyAmount > 0) {
            // if tokens are native, pool will created in the handler.
            // otherwise, pool will created when the remoted deposited is finalised.
            bank.mint(ModuleName, supplyToken);
            bank.sendTokenFromModuleToAccount(ModuleName, msg.sender, supplyToken);
        }
    } else {
        // request remote deposit
        const request: IntentRequest  = {
            sender: msg.sender,
            channelId: msg.channelId,
            action: "SingleDeposit",
            expectedSender: msg.desired_sender, // the expected sender of inboundTx on
            expectedReceivedToken: msg.token,
            hash: "",
            status: "INITIATED",
            inboundTx: [],
            createAt: block.timestamp,
        }

        store.registerInboundRequest(request)
    }
}

function handleMsgMultiDeposit(msg: MsgMultiDeposit) {

    // process request
    const request: IntentRequest  = {
        channelId: msg.channelId,
        action: "MultiDeposit",
        expectedSender: msg.desired_sender, // the expected sender of inboundTx on counterparty chain
        hash: "",
        status: "INITIATED",
        inboundTx: [],
        createAt: block.timestamp,
    }

    store.registerInboundRequest(request)

    cosnt deposit = {
        sender: msg.sender,
        desired_sender: msg.desired_sender, // remote sender
        token: msg.token,
        channelId: msg.channeId,
        status: "INITIATED",
        createdAt: block.timestamp,
        completedAt: 0,
    }
    store.save(deposit)
}
```

```ts
function handleMsgWithdraw(msg: MsgWithdraw) {

    abortTransactionUnless(msg.token.amount > 0)

    let channel = store.getChannel(msg.channelId);
    let adapter = TX_REGISTRY.getAdapter(msg.channelId);
    let pool = store.getPool(msg.poolId)
    abortTransactionUnless(pool.supply.denom === msg.token.denom)

    let escrowedAddress = getEscrowedAccount(`${AppName}/${poolId}`)
    bank.sendTokenToAccount(msg.sender, escrowedAddress, t.token)

    let requestId = store.registerOutboundSigningRequest(
        adapter.buildSigningRequest("WITHDRAW", channel, msg.recipient, msg.token)
    );

    let newWithdraw = {
        sender: msg.sender, // local sender
        poolId: msg.poolId,
        token: msg.token,
        requestId,
        status: "Initial",
        createdAt: block.timestamp,
        completedAt: 0,
    }
    store.save(newWithdraw)
}
```

```ts
function handleMsgSwap(msg: MsgSwap) {
    // process request
    let trade = {
        sender: msg.sender, // local sender
        poolId: msg.poolId,
        tokenIn: msg.TokenIn,
        tokenOut: msg.TokenOut,
        slippage: msg.slippage,
        createdAt: block.timestamp,
        completedAt: 0,
    }
    if(isNativeToken(msg.tokenIn)) {
        // lock assets on escrowed account
        let pool = getPool(msg.poolId)
        let escrowedAddress = getEscrowedAccount(`${AppName}/${msg.poolId}`)
        bank.sendTokenToAccount(msg.sender, escrowedAddress, t.tokenIn)
        let outAmount = calculateSwapOut(msg.tokenIn);
        if(outAmount > 0) {
            let outToken = new Coin(outAmount, msg.tokenOut.denom)
            if(isNativeToken(msg.tokenOut)) {
                bank.sendTokenToAccount(escrowedAddress, msg.sender, outToken)
                // update pool states
                for (t in pool.assets) {
                    if (t.denom == msg.tokenIn.denom) {
                        t.amount += msg.tokenIn.amount
                    }
                    if (t.denom == outToken.denom) {
                        t.amount -= outToken.amount
                    }
                }
                store.save(pool)
                trade.status = 'Completed'
            } else {
                let requestId = store.registerOutboundSigningRequest(
                    adapter.buildSigningRequest("Swap", channel, msg.recipient, outToken)
                );
                trade.status = 'Initial'
            }
        }
    } else {
        // request remote deposit
        const request: IntentRequest  = {
            sender: msg.sender,
            channelId: msg.channelId,
            action: "Swap",
            referenceId: trade.id
            expectedSender: msg.desired_sender, // the expected sender of inboundTx on
            expectedReceivedToken: msg.token,
            hash: "",
            status: "INITIATED",
            inboundTx: [],
            createAt: block.timestamp,
        }

        let requestId = store.registerInboundRequest(request)
        trade.status = 'Initial'
        trade.requestId = requestId
    }

    store.save(trade)
}
```

### Transaction Handler

```ts
function onInboundExecuted(request: IntentRequest) {}
```

```ts
function onInboundConfirmed(request: IntentRequest) {}
```

```ts
function onInboundFinalized(request: IntentRequest) {
  const channel = store.getChannel(request.channelId);
  const adapter = new DepositEthereumResposne(request, channel, deposit);
  const ok = adapter.verify();
  if (!ok) return;

  // Mint voucher tokens
  if (request.action === "CreatePool") {
    let pool = store.findPoolbyId(request.referenceId);
    for (t in pool.assets) {
      if (t.balance.denom === request.expectedReceivedToken.denom) {
        t.amount = request.expectedReceivedToken.amount;
      }
    }
    // check if deposit completed
    let doneDeposited = true;
    for (t in pool.assets) {
      if (t.amount == 0) {
        doneDeposited = false;
      }
    }
    if (doneDeposited) {
      supplyAmount = calculateInitialSupply(pool.assets);
      let supplyToken = new Coin(supplyAmount, supplyToken);
      // send pool token to the sender
      bank.mint(ModuleName, supplyToken);
      bank.sendTokenFromModuleToAccount(
        ModuleName,
        request.sender,
        supplyToken
      );
      pool.status = "Ready";
    }

    store.save(pool);
  } else if (request.action === "SingleDeposit") {
    let pool = store.getPool(request.referenceId);
    let newSupplyAmount = calculateSupply(pool, request.ExpectedReceivedToken);
    let newSupplyToken = new Coin(newSupplyAmount, pool.supply.denom);
    if (newSupplyAmount > 0) {
      // if tokens are native, pool will created in the handler.
      // otherwise, pool will created when the remoted deposited is finalised.
      bank.mint(ModuleName, newSupplyToken);
      bank.sendTokenFromModuleToAccount(ModuleName, msg.sender, newSupplyToken);
    }
    // update pool states
    for (t in pool.assets) {
      if (t.balance.denom === request.expectedReceivedToken.denom) {
        t.amount += request.expectedReceivedToken.amount;
      }
    }
    pool.supply.amount += newSupplyAmount;
    store.save(pool);
  } else if (request.action === 'Swap') {
    let trade = store.getWithdraw(request.referenceId);
    let pool = store.getPool(withdraw.poolId)
    let outAmount = calculateOutToken( pool, trade.tokenIn );
    let outToken = new Coin(outAmount, msg.tokenOut.denom)
    if(isNativeToken(msg.tokenOut)) {
        bank.sendTokenToAccount(escrowedAddress, msg.sender, outToken)
        // update pool states
        for (t in pool.assets) {
            if (t.denom == msg.tokenIn.denom) {
                t.amount += msg.tokenIn.amount
            }
            if (t.denom == outToken.denom) {
                t.amount -= outToken.amount
            }
        }
        store.save(pool)
        trade.status = 'Completed'
    } else {
        let requestId = store.registerOutboundSigningRequest(
            adapter.buildSigningRequest("Swap", channel, msg.recipient, outToken)
        );
        trade.status = 'Executed'
    }
    store.save(trade)
  }
}
```

```ts
function onInboundExpired(request: IntentRequest) {}
```

```ts
function onOutboundSigned(request: SigningRequest) {
  if (request.action === "Withdraw") {
    let withdraw = store.getWithdraw(request.referenceId);
    let pool = store.getPool(withdraw.poolId)
    let out = calculateOutToken( pool, withdraw.token );
    // update pool states
    for (t in pool.assets) {
        for(o in out) {
            if (t.balance.denom === o.denom) {
                t.amount -= o.amount;
            }
        }
    }
    pool.supply.amount -= withdraw.token.amount;
    store.save(pool);
    withdraw.status = 'Executed'
    store.save(withdraw)
  } else if(request.action === "Swap") {
    let trade = store.getTrade(request.referenceId)
    let pool = store.getPool(trade.poolId)
    let outAmount = calculateOutToken( pool, trade.tokenIn );
    let outToken = new Coin(outAmount, msg.tokenOut.denom)
    // update pool states
    for (t in pool.assets) {
        if (t.denom == trade.tokenIn.denom) {
            t.amount += trade.tokenIn.amount
        }
        if (t.denom == outToken.denom) {
            t.amount -= outToken.amount
        }
    }
    store.save(pool)
    trade.status = 'Executed'
    store.save(trade)
  }
}
```

```ts
function onOutboundBroadcasted(request: SigningRequest) {}
```

```ts
function onOutboundExecuted(request: SigningRequest) {}
```

```ts
function onOutboundConfirmed(request: SigningRequest) {}
```

```ts
function onOutboundFinalized(request: SigningRequest) {
    if (request.action === "Withdraw") {
        let withdraw = store.getWithdraw(request.referenceId);
        withraw.status = 'Finalised'
        store.save(withdraw)
    } else if(request.action == 'Swap') {
        // update pool states
        let trade = store.getTrade(request.referenceId)
        trade.status = 'Completed'
        store.save(trade)
    }
}
```
# Cross Chain Liquidity

The Cross Chain Liquidity is a CCA app that establishes a cross-chain liquidity pool, enabling users to swap tokens within the pool without the need for transferring tokens between different chains.

## Technical Specification
### Transaction Flow
![flow](./cross_chain_liquidity_workflow.png)

### Data Structure

```ts
interface Deposit {
    sender: string, //local sender
    desired_sender: string, // remote sender
    token: Coin,
    channelId: string,
    status: string,
    createdAt: u64,
    completedAt: u64,
}
```
- Deposit: 0x01 | sender | desired_sender | channelId  -> ProtocolBuffer(Deposit)

```ts
interface DenomTrace {
    channelId: string,
    vaultAddress: string,
    denom: string,
}
```
- Deposit: 0x02 | channelId | vaultAddress | denom  -> ProtocolBuffer(DenomTrace)



```ts
class DepositEthereumResposne extends DefaultEthereumResponseAdapter<Deposit> {
    verify() {
        if (super.verify()) return false;
        // check ERC20 tokens later
        if(this.txResponse.value !== this.state.token.amount) return false
        return true;
    }
}
```


### Messages
```proto
message MsgMultiDeposit {
    string desired_sender = 1;
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
    string channelId = 1;
    string recipient = 2;
    string poolId = 3;
    Coin tokenIn = 4;
    Coin tokenOut = 5;
    u64 slippage = 6;
}
```
### MessageHandler

```ts
function handleMsgDeposit(msg: MsgDeposit) {

    // process request
    const request: IntentRequest  = {
        channelId: msg.channelId,
        action: "Deposit",
        expectedSender: msg.desired_sender, // the expected sender of inboundTx on counterparty chain
        hash: "",
        status: "INITIATED",
        inboundTx: [],
        createAt: block.timestamp,
    }

    store.registerInboundSigningRequest(request)

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
    const channel = store.getChannel(msg.channelId)
    const adapter = TX_REGISTRY.getAdapter(msg.channelId)

    // naming check
    const tokenMeta = store.getTokenMeta(msg.token.denom)
    if (msg.token.denom !== hash(`${channel.id}/${channel.vaultAddress}/${tokenMeta.denom}`)) {
        throw new Error("Can not withdraw the tokens")
    }
    // convert voucher coin to remote tokens
    // TODO: process ERC20 later
    const value = parseInt(msg.token.amount)
    const data = ""
    store.registerOutboundSigningRequest(adapter.buildSigningRequest(
        "WITHDRAW", channel, msg.recipient, value, data
    ))
}
```

### Transaction Handler

```ts
function onInboundExecuted(request: IntentRequest) {
}
```
```ts
function onInboundConfirmed(request: IntentRequest) {

}
```
```ts
function onInboundFinalized(request: IntentRequest) {

    const key = `0x01|${request.sender}|${request.desired_sender}|${request.channelId}`
    const deposit = store.getDeposit(key)
    const channel = store.getChannel(request.channelId)

    const adapter = new DepositEthereumResposne(request, channel,  deposit)
    const ok = adapter.verify()
    if(!ok) return

    // Mint voucher tokens
    const denom = hash(`${channel.id}/${channel.vaultAddress}/${deposit.token.denom}`);
    const voucherToken = new Coin(deposit.token.amount, denom);
    bank.mintToken(voucherToken)
    bank.sendToken(moduleAddress, deposit.sender, voucherToken)

    // save denom trace
    store.save(denom, {
       channelId: channel.id,
       vaultAddress: channel.vaultAddress,
       denom: deposit.token.denom,
    })

}
```
```ts
function onInboundExpired(request: IntentRequest) {

}
```

```ts
function onOutboundSigned(request: SigningRequest) {
    // Burn voucher tokens
    bank.burnToken(request.token)

}
```
```ts
function onOutboundBroadcasted(request: SigningRequest) {

}
```
```ts
function onOutboundExecuted(request: SigningRequest) {

}
```
```ts
function onOutboundConfirmed(request: SigningRequest) {

}
```
```ts
function onOutboundFinalized(request: SigningRequest) {

}
```
